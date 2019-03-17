package order

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/scalog/scalog/logger"
	pb "github.com/scalog/scalog/order/messaging"
	"go.etcd.io/etcd/etcdserver/api/snap"
)

// Server's view of latest cuts in all servers of the shard
type ShardCut []int

// Map<Shard ID, Shard cuts reported by all servers in shard>
type ContestedGlobalCut map[int][]ShardCut

// Map<Shard ID, Cuts in the shard>
type CommittedGlobalCut map[int]ShardCut

// Number of new logs appended within the previous time period
type Deltas CommittedGlobalCut

// Map<Shard ID, Channels to write responses to for each server in shard>
type ResponseChannels map[int][]chan pb.ReportResponse

type orderServer struct {
	committedGlobalCut   CommittedGlobalCut
	contestedGlobalCut   ContestedGlobalCut
	globalSequenceNum    int
	shardIds             []int
	numServersPerShard   int
	mu                   sync.Mutex
	dataResponseChannels ResponseChannels
	raftProposeChannel   chan<- string
	raftSnapshotter      *snap.Snapshotter
}

// Fields in orderServer necessary for state replication. Used in Raft.
type orderServerState struct {
	committedGlobalCut CommittedGlobalCut
	contestedGlobalCut ContestedGlobalCut
	globalSequenceNum  int
}

func newOrderServer(shardIds []int, numServersPerShard int, raftProposeChannel chan<- string, raftSnapshotter *snap.Snapshotter) *orderServer {
	return &orderServer{
		committedGlobalCut:   initCommittedCut(shardIds, numServersPerShard),
		contestedGlobalCut:   initContestedCut(shardIds, numServersPerShard),
		globalSequenceNum:    0,
		shardIds:             shardIds,
		numServersPerShard:   numServersPerShard,
		mu:                   sync.Mutex{},
		dataResponseChannels: initResponseChannels(shardIds, numServersPerShard),
		raftProposeChannel:   raftProposeChannel,
		raftSnapshotter:      raftSnapshotter,
	}
}

func initCommittedCut(shardIds []int, numServersPerShard int) CommittedGlobalCut {
	cut := make(CommittedGlobalCut, len(shardIds))
	for _, shardId := range shardIds {
		cut[shardId] = make(ShardCut, numServersPerShard, numServersPerShard)
	}
	return cut
}

func initContestedCut(shardIds []int, numServersPerShard int) ContestedGlobalCut {
	cut := make(ContestedGlobalCut, len(shardIds))
	for _, shardId := range shardIds {
		shardCuts := make([]ShardCut, numServersPerShard, numServersPerShard)
		for i := 0; i < numServersPerShard; i++ {
			shardCuts[i] = make(ShardCut, numServersPerShard, numServersPerShard)
		}
		cut[shardId] = shardCuts
	}
	return cut
}

func initResponseChannels(shardIds []int, numServersPerShard int) ResponseChannels {
	channels := make(ResponseChannels, len(shardIds))
	for _, shardId := range shardIds {
		channelsForShard := make([]chan pb.ReportResponse, numServersPerShard, numServersPerShard)
		for i := 0; i < numServersPerShard; i++ {
			channelsForShard[i] = make(chan pb.ReportResponse)
		}
		channels[shardId] = channelsForShard
	}
	return channels
}

/**
Find min cut for each shard and compute changes from last committed cuts.
*/
func (server *orderServer) mergeContestedCuts() Deltas {
	deltas := initCommittedCut(server.shardIds, server.numServersPerShard)

	for _, shardId := range server.shardIds {
		// find smallest cut for shard i
		for i := 0; i < server.numServersPerShard; i++ {
			minCut := 0
			for j := 0; j < server.numServersPerShard; j++ {
				minCut = min(server.contestedGlobalCut[shardId][j][i], minCut)
			}

			prevCut := server.committedGlobalCut[shardId][i]
			deltas[shardId][i] = minCut - prevCut
		}
	}

	return Deltas(deltas)
}

/**
Compute global sequence number for each server and send update message.
NOTE: Must be called after updateCommittedCuts().
*/
func (server *orderServer) updateCommittedCutGlobalSeqNumAndBroadcastDeltas(deltas Deltas) {
	for _, shardId := range server.shardIds {
		startSequenceNum := server.globalSequenceNum

		response := pb.ReportResponse{
			StartGlobalSequenceNum: int32(server.globalSequenceNum),
			Offsets:                intSliceToInt32Slice(server.committedGlobalCut[shardId]),
		}

		for i := 0; i < server.numServersPerShard; i++ {
			delta := deltas[shardId][i]

			// calculate new committed cuts
			server.committedGlobalCut[shardId][i] += delta
			server.globalSequenceNum += delta
		}

		// Don't send responses if no new logs are committed
		if server.globalSequenceNum == startSequenceNum {
			continue
		}

		response.CommittedCuts = intSliceToInt32Slice(server.committedGlobalCut[shardId])

		for i := 0; i < server.numServersPerShard; i++ {
			server.dataResponseChannels[shardId][i] <- response
		}
	}
}

////////////////////// DATA LAYER GRPC FUNCTIONS

/**
Periodically merge contested cuts and broadcast to data layer. Sends responses into ResponseChannels.
*/
func (server *orderServer) respondToDataLayer() {
	ticker := time.NewTicker(100 * time.Microsecond) // todo remove hard-coded interval
	for range ticker.C {
		deltas := server.mergeContestedCuts()
		server.updateCommittedCutGlobalSeqNumAndBroadcastDeltas(deltas)
	}
}

/**
Reports final cuts to the data layer periodically. Reads from ResponseChannels and feeds into the stream.
*/
func (server *orderServer) reportResponseRoutine(stream pb.Order_ReportServer, req *pb.ReportRequest) {
	shardId := int(req.ShardID)
	num := req.ReplicaID
	for response := range server.dataResponseChannels[shardId][num] {
		logger.Printf(response.String())
		stream.Send(&response)
	}
}

////////////////////// RAFT FUNCTIONS

/**
Extracts all variables necessary in state replication in orderServer.
*/
func (server *orderServer) getState() orderServerState {
	return orderServerState{
		server.committedGlobalCut,
		server.contestedGlobalCut,
		server.globalSequenceNum,
	}
}

/**
Overwrites current server data with state data.
NOTE: assumes that shardIds and numServersPerShard do not change. If they change, then we must recreate channels.
*/
func (server *orderServer) loadState(state *orderServerState) {
	server.committedGlobalCut = state.committedGlobalCut
	server.contestedGlobalCut = state.contestedGlobalCut
	server.globalSequenceNum = state.globalSequenceNum
}

/**
Returns all stored data.
*/
func (server *orderServer) getSnapshot() ([]byte, error) {
	server.mu.Lock()
	defer server.mu.Unlock()
	return json.Marshal(server.getState())
}

/**
Add contested cuts from the data layer. Triggered when Raft commits the new message.
*/
func (server *orderServer) listenForRaftCommits(raftCommitChannel <-chan *string) {
	for requestString := range raftCommitChannel {
		if requestString == nil {
			server.attemptRecoverFromSnapshot()
			continue
		}

		req := &pb.ReportRequest{}
		err := proto.Unmarshal([]byte(*requestString), req)
		if err != nil {
			logger.Printf("Could not unmarshal raft commit message")
		}

		//save into contested cuts
		cut := make(ShardCut, server.numServersPerShard, server.numServersPerShard)
		for i := 0; i < len(cut); i++ {
			cut[i] = int(req.TentativeCut[i])
		}
		for i := 0; i < len(cut); i++ {
			curr := server.contestedGlobalCut[int(req.ShardID)][int(req.ReplicaID)][i]
			server.contestedGlobalCut[int(req.ShardID)][int(req.ReplicaID)][i] = max(curr, cut[i])
		}
	}
}

/**
Use snapshot to reload server state.
*/
func (server *orderServer) attemptRecoverFromSnapshot() {
	snapshot, err := server.raftSnapshotter.Load()
	if err == snap.ErrNoSnapshot {
		return
	}
	if err != nil {
		logger.Panicf(err.Error())
	}

	logger.Printf("Ordering layer attempting to recover from Raft snapshot")
	state := &orderServerState{}
	err = json.Unmarshal(snapshot.Data, state)

	if err != nil {
		logger.Panicf(err.Error())
	}

	server.mu.Lock()
	server.loadState(state)
	server.mu.Unlock()
}

////////////////////// UTIL FUNCTIONS

func max(x int, y int) int {
	if x < y {
		return y
	}
	return x
}
func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
func intSliceToInt32Slice(intSlice []int) []int32 {
	int32Slice := make([]int32, len(intSlice), len(intSlice))
	for idx, element := range intSlice {
		int32Slice[idx] = int32(element)
	}
	return int32Slice
}
