package order

import (
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/scalog/scalog/internal/pkg/golib"
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
	shardIds             *golib.Set
	numServersPerShard   int
	mu                   sync.RWMutex
	dataResponseChannels ResponseChannels
	raftProposeChannel   chan<- string
	raftSnapshotter      *snap.Snapshotter
}

// Fields in orderServer necessary for state replication. Used in Raft.
type orderServerState struct {
	committedGlobalCut CommittedGlobalCut
	contestedGlobalCut ContestedGlobalCut
	globalSequenceNum  int
	shardIds           *golib.Set
}

func newOrderServer(shardIds *golib.Set, numServersPerShard int, raftProposeChannel chan<- string, raftSnapshotter *snap.Snapshotter) *orderServer {
	return &orderServer{
		committedGlobalCut:   initCommittedCut(shardIds, numServersPerShard),
		contestedGlobalCut:   initContestedCut(shardIds, numServersPerShard),
		globalSequenceNum:    0,
		shardIds:             shardIds,
		numServersPerShard:   numServersPerShard,
		mu:                   sync.RWMutex{},
		dataResponseChannels: initResponseChannels(shardIds, numServersPerShard),
		raftProposeChannel:   raftProposeChannel,
		raftSnapshotter:      raftSnapshotter,
	}
}

func initCommittedCut(shardIds *golib.Set, numServersPerShard int) CommittedGlobalCut {
	cut := make(CommittedGlobalCut)
	for shardID := range shardIds.Iterable() {
		cut[shardID] = make(ShardCut, numServersPerShard, numServersPerShard)
	}
	return cut
}

func initContestedCut(shardIds *golib.Set, numServersPerShard int) ContestedGlobalCut {
	cut := make(ContestedGlobalCut)
	for shardID := range shardIds.Iterable() {
		shardCuts := make([]ShardCut, numServersPerShard, numServersPerShard)
		for i := 0; i < numServersPerShard; i++ {
			shardCuts[i] = make(ShardCut, numServersPerShard, numServersPerShard)
		}
		cut[shardID] = shardCuts
	}
	return cut
}

func initResponseChannels(shardIds *golib.Set, numServersPerShard int) ResponseChannels {
	channels := make(ResponseChannels)
	for shardID := range shardIds.Iterable() {
		channelsForShard := make([]chan pb.ReportResponse, numServersPerShard, numServersPerShard)
		for i := 0; i < numServersPerShard; i++ {
			channelsForShard[i] = make(chan pb.ReportResponse)
		}
		channels[shardID] = channelsForShard
	}
	return channels
}

/**
Find min cut for each shard and compute changes from last committed cuts.

CAUTION: THE CALLER MUST OBTAIN THE LOCK
*/
func (server *orderServer) mergeContestedCuts() Deltas {
	deltas := initCommittedCut(server.shardIds, server.numServersPerShard)
	for shardID := range server.shardIds.Iterable() {
		// find smallest cut for shard i
		for i := 0; i < server.numServersPerShard; i++ {
			minCut := math.MaxInt64
			for j := 0; j < server.numServersPerShard; j++ {
				minCut = golib.Min(server.contestedGlobalCut[shardID][j][i], minCut)
			}

			prevCut := server.committedGlobalCut[shardID][i]
			deltas[shardID][i] = minCut - prevCut
		}
	}
	return Deltas(deltas)
}

/**
Compute global sequence number for each server and send update message.
NOTE: Must be called after updateCommittedCuts().

CAUTION: THE CALLER MUST OBTAIN THE LOCK
*/
func (server *orderServer) updateCommittedCutGlobalSeqNumAndBroadcastDeltas(deltas Deltas) {
	for shardID := range server.shardIds.Iterable() {
		startSequenceNum := server.globalSequenceNum

		response := pb.ReportResponse{
			StartGlobalSequenceNum: int32(server.globalSequenceNum),
			Offsets:                golib.IntSliceToInt32Slice(server.committedGlobalCut[shardID]),
			Finalized:              false,
		}

		for i := 0; i < server.numServersPerShard; i++ {
			delta := deltas[shardID][i]

			// calculate new committed cuts
			server.committedGlobalCut[shardID][i] += delta
			server.globalSequenceNum += delta
		}

		// Don't send responses if no new logs are committed
		if server.globalSequenceNum == startSequenceNum {
			continue
		}

		response.CommittedCuts = golib.IntSliceToInt32Slice(server.committedGlobalCut[shardID])
		for i := 0; i < server.numServersPerShard; i++ {
			server.dataResponseChannels[shardID][i] <- response
		}
	}
}

/*
Send a finalization message to each replica within [shardID].

This operation acquires a writer lock.
*/
func (server *orderServer) notifyFinalizedShard(shardID int) {
	terminationMessage := pb.ReportResponse{Finalized: true}

	server.mu.Lock()
	defer server.mu.Unlock()
	for i := 0; i < server.numServersPerShard; i++ {
		// Notify the relevant data replicas that they have been finalized
		server.dataResponseChannels[shardID][i] <- terminationMessage
	}
}

/**
Adds a shard with the shardID to the ordering layer. Does nothing if the shardID already exists.

This operation acquires a writer lock.
*/
func (server *orderServer) addShard(shardID int) {
	//Use read lock to determine if adding is necessary
	server.mu.RLock()
	exists := server.shardIds.Contains(shardID)
	server.mu.RUnlock()
	if exists {
		return //Do not attempt to add a shard that was already added
	}

	server.mu.Lock()
	defer server.mu.Unlock()

	server.shardIds.Add(shardID)
	server.committedGlobalCut[shardID] = make(ShardCut, server.numServersPerShard, server.numServersPerShard)
	server.contestedGlobalCut[shardID] = make([]ShardCut, server.numServersPerShard, server.numServersPerShard)
	server.dataResponseChannels[shardID] = make([]chan pb.ReportResponse, server.numServersPerShard, server.numServersPerShard)

	for i := 0; i < server.numServersPerShard; i++ {
		server.dataResponseChannels[shardID][i] = make(chan pb.ReportResponse)
		server.contestedGlobalCut[shardID][i] = make(ShardCut, server.numServersPerShard, server.numServersPerShard)
	}
}

/**
Remove the shardID metadata from the ordering layer. Does nothing if the shardID does not exist.

This operation acquires a writer lock.
*/
func (server *orderServer) deleteShard(shardID int) {
	//Use read lock to determine if deleting is necessary
	server.mu.RLock()
	exists := server.shardIds.Contains(shardID)
	server.mu.RUnlock()
	if !exists {
		return //Do not attempt to delete a shard that was already deleted
	}

	server.mu.Lock()
	defer server.mu.Unlock()
	server.shardIds.Remove(shardID)
	for i := 0; i < server.numServersPerShard; i++ {
		// Cleanup channels used for data layer communication
		close(server.dataResponseChannels[shardID][i])
	}
	delete(server.dataResponseChannels, shardID)
	delete(server.committedGlobalCut, shardID)
	delete(server.contestedGlobalCut, shardID)
}

////////////////////// DATA LAYER GRPC FUNCTIONS

/**
Periodically merge contested cuts and broadcast to data layer. Sends responses into ResponseChannels.
*/
func (server *orderServer) respondToDataLayer() {
	ticker := time.NewTicker(100 * time.Microsecond) // todo remove hard-coded interval
	for range ticker.C {
		server.mu.RLock()
		deltas := server.mergeContestedCuts()
		server.updateCommittedCutGlobalSeqNumAndBroadcastDeltas(deltas)
		server.mu.RUnlock()
	}
}

/**
Reports final cuts to the data layer periodically. Reads from ResponseChannels and feeds into the stream.
*/
func (server *orderServer) reportResponseRoutine(stream pb.Order_ReportServer, req *pb.ReportRequest) {
	shardID := int(req.ShardID)
	num := req.ReplicaID

	//Always attempt to add shard. Does nothing if shard already exists.
	server.addShard(shardID)
	//send response
	for response := range server.dataResponseChannels[shardID][num] {
		err := stream.Send(&response)
		if err != nil {
			logger.Panicf(err.Error())
		}
	}
	logger.Printf(fmt.Sprintf("Ordering finalizing data replica %d in shard %d", num, shardID))
}

////////////////////// RAFT FUNCTIONS

/**
Extracts all variables necessary in state replication in orderServer.
*/
func (server *orderServer) getState() orderServerState {
	server.mu.RLock()
	defer server.mu.RUnlock()
	return orderServerState{
		server.committedGlobalCut,
		server.contestedGlobalCut,
		server.globalSequenceNum,
		server.shardIds,
	}
}

/**
Overwrites current server data with state data.
NOTE: assumes that shardIds and numServersPerShard do not change. If they change, then we must recreate channels.
*/
func (server *orderServer) loadState(state *orderServerState) {
	server.mu.Lock()
	defer server.mu.Unlock()

	server.committedGlobalCut = state.committedGlobalCut
	server.contestedGlobalCut = state.contestedGlobalCut
	server.globalSequenceNum = state.globalSequenceNum

	deleted := server.shardIds.Difference(state.shardIds)
	added := state.shardIds.Difference(server.shardIds)

	//delete channels
	for shardID := range deleted.Iterable() {
		for i := 0; i < server.numServersPerShard; i++ {
			close(server.dataResponseChannels[shardID][i])
		}
		delete(server.dataResponseChannels, shardID)
	}
	//add channels
	for shardID := range added.Iterable() {
		server.dataResponseChannels[shardID] = make([]chan pb.ReportResponse, server.numServersPerShard, server.numServersPerShard)
		for i := 0; i < server.numServersPerShard; i++ {
			server.dataResponseChannels[shardID][i] = make(chan pb.ReportResponse)
		}
	}

	server.shardIds = state.shardIds
}

/**
Returns all stored data.
*/
func (server *orderServer) getSnapshot() ([]byte, error) {
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
		if err := proto.Unmarshal([]byte(*requestString), req); err != nil {
			logger.Panicf("Could not unmarshal raft commit message")
		}

		// ReportRequests can take two actions -- one is where we decide upon a shard cut.
		// The other action is if we decide on finalizing a particular shard.
		if req.Finalized {
			finalizedShardID := int(req.ShardID)
			server.notifyFinalizedShard(finalizedShardID)
			server.deleteShard(finalizedShardID)
			continue
		}

		// Save into contested cuts
		cut := make(ShardCut, server.numServersPerShard, server.numServersPerShard)
		for i := 0; i < len(cut); i++ {
			cut[i] = int(req.TentativeCut[i])
		}

		server.mu.Lock()
		for i := 0; i < len(cut); i++ {
			curr := server.contestedGlobalCut[int(req.ShardID)][int(req.ReplicaID)][i]
			server.contestedGlobalCut[int(req.ShardID)][int(req.ReplicaID)][i] = golib.Max(curr, cut[i])
		}
		server.mu.Unlock()
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

	server.loadState(state)
}
