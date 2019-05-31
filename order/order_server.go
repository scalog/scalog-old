package order

import (
	"encoding/json"
	"math"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/spf13/viper"

	"github.com/scalog/scalog/internal/pkg/golib"
	"github.com/scalog/scalog/logger"
	pb "github.com/scalog/scalog/order/messaging"
	"go.etcd.io/etcd/etcdserver/api/snap"
)

// Server's view of latest cuts in all servers of the shard
type Cut []int

// All cuts from all replicas within a shard
type ShardCuts []Cut

// Map<Shard ID, Shard cuts reported by all servers in shard>
type ContestedCut map[int]ShardCuts

// Map<Shard ID, Cuts in the shard>
type CommittedCut map[int32]*pb.Cut

// Number of new logs appended within the previous time period
type Deltas CommittedCut

// ResponseChannels is an array of channels that response to connected aggregators
type ResponseChannels []chan pb.ReportResponse

// FinalizationMap map<shardID, finalize after k cuts>
type FinalizationMap map[int32]*pb.FinalizeRequest

// Map<Shard ID, Channels to write responses to when a shard finalization
// request has been committed>
type FinalizationResponseChannels map[int32]chan pb.FinalizeResponse

type raftProposalType int

const (
	REPORT   raftProposalType = 0
	REGISTER raftProposalType = 1
	FINALIZE raftProposalType = 2
)

type raftProposal struct {
	proposalType raftProposalType
	proposalData []byte
}

type orderServer struct {
	committedCut                 CommittedCut
	contestedCut                 ContestedCut
	finalizeMap                  FinalizationMap
	finalizedShards              *golib.Set32
	globalSequenceNum            int32
	shardIds                     *golib.Set
	numServersPerShard           int
	mu                           sync.RWMutex
	finalizationResponseChannels FinalizationResponseChannels
	aggregatorResponseChannels   ResponseChannels
	rc                           *raftNode
	viewID                       int32
	viewC                        chan *pb.RegisterResponse
}

// Fields in orderServer necessary for state replication. Used in Raft.
type orderServerState struct {
	committedCut      CommittedCut
	contestedCut      ContestedCut
	globalSequenceNum int32
	shardIds          *golib.Set
}

func newOrderServer(shardIds *golib.Set, numServersPerShard int) *orderServer {
	return &orderServer{
		committedCut:                 initCommittedCut(shardIds, numServersPerShard),
		contestedCut:                 initContestedCut(shardIds, numServersPerShard),
		finalizeMap:                  make(FinalizationMap),
		finalizedShards:              golib.NewSet32(),
		globalSequenceNum:            0,
		shardIds:                     shardIds,
		numServersPerShard:           numServersPerShard,
		mu:                           sync.RWMutex{},
		aggregatorResponseChannels:   make(ResponseChannels, 0),
		finalizationResponseChannels: make(FinalizationResponseChannels),
		viewID:                       0,
		viewC:                        make(chan *pb.RegisterResponse),
	}
}

////////////// INITIALIZERS

func initCommittedCut(shardIds *golib.Set, numServersPerShard int) CommittedCut {
	cut := make(CommittedCut)
	for shardID := range shardIds.Iterable() {
		cut[int32(shardID)] = &pb.Cut{
			Cut: make([]int32, numServersPerShard),
		}
	}
	return cut
}

func initContestedCut(shardIds *golib.Set, numServersPerShard int) ContestedCut {
	cut := make(ContestedCut)
	for shardID := range shardIds.Iterable() {
		shardCuts := make(ShardCuts, numServersPerShard)
		for i := 0; i < numServersPerShard; i++ {
			shardCuts[i] = make(Cut, numServersPerShard)
		}
		cut[shardID] = shardCuts
	}
	return cut
}

////////////// ORDER SERVER STATE MUTATORS

// updateCommittedCuts updates both the GSN and committedCuts
func (server *orderServer) updateCommittedCuts() {
	for shardID := range server.shardIds.Iterable() {
		// find smallest cut for shard i
		for i := 0; i < server.numServersPerShard; i++ {
			minCut := math.MaxInt64
			for j := 0; j < server.numServersPerShard; j++ {
				minCut = golib.Min(server.contestedCut[shardID][j][i], minCut)
			}
			previousCut := server.committedCut[int32(shardID)].Cut[i]
			server.committedCut[int32(shardID)].Cut[i] = int32(minCut)
			// update gsn
			diff := int32(minCut) - previousCut
			server.globalSequenceNum += diff
		}
	}
}

// getShardsToFinalize returns all shardID's bound in this map that
// are meant to commit in 0 cuts, and decrements all other bound
// shardID's by one.
func (server *orderServer) getShardsToFinalize() []*pb.FinalizeRequest {
	shardsToFinalize := make([]*pb.FinalizeRequest, 0)
	for shardID, req := range server.finalizeMap {
		if req.Limit == 0 {
			shardsToFinalize = append(shardsToFinalize, req)
			delete(server.finalizeMap, shardID)
		} else {
			server.finalizeMap[shardID].Limit--
		}

	}
	return shardsToFinalize
}

/**
Find min cut for each shard and compute changes from last committed cuts.
Compute global sequence number for each server.
Broadcasts the new CommittedCuts to data layer.

CAUTION: THE CALLER MUST OBTAIN THE LOCK
*/
func (server *orderServer) mergeContestedCuts() {
	server.updateCommittedCuts()
	shardsToFinalize := server.getShardsToFinalize()
	// Keep track of shards that we have finalized so that we ignore pipelined requests
	for _, req := range shardsToFinalize {
		resp := &pb.FinalizeResponse{ShardID: req.ShardID}
		propData, err := proto.Marshal(resp)
		if err != nil {
			logger.Printf("Could not marshal finalization request message")
		}
		prop := raftProposal{
			proposalType: FINALIZE,
			proposalData: propData,
		}
		server.rc.proposeC <- prop
	}
	// broadcast cuts
	resp := pb.ReportResponse{
		CommitedCuts: server.committedCut,
		StartGSN:     server.globalSequenceNum,
	}
	for _, ch := range server.aggregatorResponseChannels {
		ch <- resp
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
	server.committedCut[int32(shardID)] = &pb.Cut{
		Cut: make([]int32, server.numServersPerShard),
	}
	server.contestedCut[shardID] = make(ShardCuts, server.numServersPerShard)

	for i := 0; i < server.numServersPerShard; i++ {
		server.contestedCut[shardID][i] = make(Cut, server.numServersPerShard)
	}
}

/**
Notifies the shard that it is being finalized, then removes the shardID metadata.
Does nothing if the shardID does not exist.

@param committed True if the deletion was committed to Raft. Determines whether or not we'll notify
	the data layer or operator of this deletion.
@returns Whether or not this shard existed.

NOTE: CALLER MUST ACQUIRE WRITE LOCK
*/
func (server *orderServer) deleteShard(shardID int32) bool {
	exists := server.shardIds.Contains(int(shardID))
	if !exists {
		return false //Do not attempt to delete a shard that was already deleted
	}

	// Delete the shard from cuts. A gesture to signify to Raft log that the shard is deleted.
	// Meaningless to others unless a cut with this deletion is committed (consensus is reached).
	delete(server.committedCut, shardID)
	delete(server.contestedCut, int(shardID))

	// Cleanup channels used for data layer communication
	server.shardIds.Remove(int(shardID))
	return true
}

////////////////////// DATA LAYER GRPC FUNCTIONS

/**
Reports final cuts to the data layer periodically. Reads from ResponseChannels and feeds into the stream.
*/
func (server *orderServer) reportResponseRoutine(stream pb.Order_ReportServer, req *pb.ReportRequest) {
	// For every shardID that we have not seen before, we should generate state and stuff
	for shardID := range req.Shards {
		// addShard already checks to see if shardID exists
		server.addShard(int(shardID))
	}

	// We serve responses from this channel back to the aggregator
	ch := make(chan pb.ReportResponse)
	server.aggregatorResponseChannels = append(server.aggregatorResponseChannels, ch)
	for response := range ch {
		err := stream.Send(&response)
		if err != nil {
			logger.Panicf(err.Error())
		}
	}
}

/**
proposalRaftBatch periodically proposes that raft batch and send a globalCut
 to the data layer
*/
func (server *orderServer) proposalRaftBatch() {
	interval := time.Duration(viper.GetInt("batch_interval"))
	ticker := time.NewTicker(interval * time.Millisecond)
	for range ticker.C {
		if server.rc.node == nil {
			continue
		}
		isLeader := server.rc.leaderID == uint64(server.rc.id)

		//do nothing if you're not the leader
		if !isLeader {
			continue
		}

		server.mu.Lock()
		server.updateCommittedCuts()
		resp := &pb.ReportResponse{
			CommitedCuts: server.committedCut,
			StartGSN:     server.globalSequenceNum,
		}
		server.mu.Unlock()
		propData, err := json.Marshal(resp)
		if err != nil {
			logger.Printf(err.Error())
			continue
		}
		prop := raftProposal{
			proposalType: REPORT,
			proposalData: propData,
		}
		server.rc.proposeC <- prop
	}
}

////////////////////// RAFT FUNCTIONS

/**
Triggered when Raft commits a new message.
*/
func (server *orderServer) listenForRaftCommits() {
	for entry := range server.rc.commitC {
		if entry == nil {
			server.attemptRecoverFromSnapshot()
			continue
		}

		prop := &raftProposal{}
		if err := json.Unmarshal(entry.Data, prop); err != nil {
			logger.Printf(err.Error())
		}

		switch prop.proposalType {
		case REPORT:
			// TODO
		case REGISTER:
			server.viewID++
			viewUpdate := &pb.RegisterResponse{ViewID: server.viewID}
			server.viewC <- viewUpdate
		case FINALIZE:
			resp := &pb.FinalizeResponse{}
			if err := proto.Unmarshal(entry.Data, resp); err != nil {
				logger.Printf(err.Error())
				continue
			}
			server.finalizationResponseChannels[resp.ShardID] <- *resp
			server.finalizedShards.Add(resp.ShardID)
			server.deleteShard(resp.ShardID)
		default:
			// TODO
		}
		// Logic to be moved into Forward
		// else {
		// 	// Update the global state
		// 	server.mu.Lock()
		// 	for shardID, replicaCuts := range req.Shards {
		// 		for replicaID, cut := range replicaCuts.Replicas {
		// 			for i := 0; i < server.numServersPerShard; i++ {
		// 				prior := server.contestedCut[int(shardID)][int(replicaID)][i]
		// 				server.contestedCut[int(shardID)][int(replicaID)][i] = golib.Max(prior, int(cut.Cut[i]))
		// 			}
		// 		}
		// 	}
		// 	server.mu.Unlock()
		// }
	}
}

/**
Extracts all variables necessary in state replication in orderServer.
*/
func (server *orderServer) getState() orderServerState {
	server.mu.RLock()
	defer server.mu.RUnlock()
	return orderServerState{
		server.committedCut,
		server.contestedCut,
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

	server.committedCut = state.committedCut
	server.contestedCut = state.contestedCut
	server.globalSequenceNum = state.globalSequenceNum

	// TODO: Update loadState

	server.shardIds = state.shardIds
}

/**
Returns all stored data.
*/
func (server *orderServer) getSnapshot() ([]byte, error) {
	return json.Marshal(server.getState())
}

/**
Use snapshot to reload server state.
*/
func (server *orderServer) attemptRecoverFromSnapshot() {
	snapshot, err := server.rc.snapshotter.Load()
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
