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
type Cut []int

// All cuts from all replicas within a shard
type ShardCuts []Cut

// Map<Shard ID, Shard cuts reported by all servers in shard>
type ContestedCut map[int]ShardCuts

// Map<Shard ID, Cuts in the shard>
type CommittedCut map[int]Cut

// Number of new logs appended within the previous time period
type Deltas CommittedCut

// Map<Shard ID, Channels to write responses to for each server in shard>
type ResponseChannels map[int][]chan pb.ReportResponse

// Map<Shard ID, Channels to write responses to when a shard finalization
// request has been committed>
type FinalizationResponseChannels map[int32]chan struct{}

type orderServer struct {
	committedCut                 CommittedCut
	contestedCut                 ContestedCut
	globalSequenceNum            int
	logNum                       int // # of batch since inception
	shardIds                     *golib.Set
	numServersPerShard           int
	mu                           sync.RWMutex
	finalizationResponseChannels FinalizationResponseChannels
	dataResponseChannels         ResponseChannels
	rc                           *raftNode
}

// Fields in orderServer necessary for state replication. Used in Raft.
type orderServerState struct {
	committedCut      CommittedCut
	contestedCut      ContestedCut
	globalSequenceNum int
	logNum            int
	shardIds          *golib.Set
}

func newOrderServer(shardIds *golib.Set, numServersPerShard int) *orderServer {
	return &orderServer{
		committedCut:                 initCommittedCut(shardIds, numServersPerShard),
		contestedCut:                 initContestedCut(shardIds, numServersPerShard),
		globalSequenceNum:            0,
		logNum:                       1,
		shardIds:                     shardIds,
		numServersPerShard:           numServersPerShard,
		mu:                           sync.RWMutex{},
		dataResponseChannels:         initResponseChannels(shardIds, numServersPerShard),
		finalizationResponseChannels: make(FinalizationResponseChannels),
	}
}

func initCommittedCut(shardIds *golib.Set, numServersPerShard int) CommittedCut {
	cut := make(CommittedCut)
	for shardID := range shardIds.Iterable() {
		cut[shardID] = make(Cut, numServersPerShard)
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

func initResponseChannels(shardIds *golib.Set, numServersPerShard int) ResponseChannels {
	channels := make(ResponseChannels)
	for shardID := range shardIds.Iterable() {
		channelsForShard := make([]chan pb.ReportResponse, numServersPerShard)
		for i := 0; i < numServersPerShard; i++ {
			channelsForShard[i] = make(chan pb.ReportResponse)
		}
		channels[shardID] = channelsForShard
	}
	return channels
}

/**
Find min cut for each shard and compute changes from last committed cuts.
Compute global sequence number for each server.
Broadcasts the new CommittedCuts to data layer.

CAUTION: THE CALLER MUST OBTAIN THE LOCK
*/
func (server *orderServer) mergeContestedCuts() {
	for shardID := range server.shardIds.Iterable() {
		delta := 0

		// find smallest cut for shard i
		for i := 0; i < server.numServersPerShard; i++ {
			minCut := math.MaxInt64
			for j := 0; j < server.numServersPerShard; j++ {
				minCut = golib.Min(server.contestedCut[shardID][j][i], minCut)
			}

			prevCut := server.committedCut[shardID][i]
			server.committedCut[shardID][i] = minCut

			delta += minCut - prevCut
		}

		// broadcast cuts
		resp := pb.ReportResponse{
			StartGlobalSequenceNum: int32(server.globalSequenceNum),
			CommittedCuts:          golib.IntSliceToInt32Slice(server.committedCut[shardID]),
			Finalized:              false,
		}
		for i := 0; i < server.numServersPerShard; i++ {
			server.dataResponseChannels[shardID][i] <- resp
		}

		server.globalSequenceNum += delta
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
	server.committedCut[shardID] = make(Cut, server.numServersPerShard)
	server.contestedCut[shardID] = make(ShardCuts, server.numServersPerShard)
	server.dataResponseChannels[shardID] = make([]chan pb.ReportResponse, server.numServersPerShard)

	for i := 0; i < server.numServersPerShard; i++ {
		server.dataResponseChannels[shardID][i] = make(chan pb.ReportResponse)
		server.contestedCut[shardID][i] = make(Cut, server.numServersPerShard)
	}
}

/**
Notifies the shard that it is being finalized, then removes the shardID metadata.
Does nothing if the shardID does not exist.

@param committed True if the deletion was committed to Raft. Determines whether or not we'll notify
	the data layer or operator of this deletion.
@returns Whether or not this shard existed.

This operation acquires a writer lock.
*/
func (server *orderServer) deleteShard(shardID int) bool {
	//Use read lock to determine if deleting is necessary
	server.mu.RLock()
	exists := server.shardIds.Contains(shardID)
	server.mu.RUnlock()
	if !exists {
		return false //Do not attempt to delete a shard that was already deleted
	}

	server.mu.Lock()
	defer server.mu.Unlock()

	// Delete the shard from cuts. A gesture to signify to Raft log that the shard is deleted.
	// Meaningless to others unless a cut with this deletion is committed (consensus is reached).
	delete(server.committedCut, shardID)
	delete(server.contestedCut, shardID)

	// Cleanup channels used for data layer communication
	server.shardIds.Remove(shardID)

	// Notify the relevant data replicas that they have been finalized, then close the channels
	terminationMessage := pb.ReportResponse{Finalized: true}
	for i := 0; i < server.numServersPerShard; i++ {
		server.dataResponseChannels[shardID][i] <- terminationMessage
		close(server.dataResponseChannels[shardID][i])
	}
	delete(server.dataResponseChannels, shardID)

	// Notify blocked channel
	finalizationC, exists := server.finalizationResponseChannels[int32(shardID)]
	if exists {
		close(finalizationC)
	}
	return true
}

////////////////////// DATA LAYER GRPC FUNCTIONS

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

/**
proposalRaftBatch periodically proposes that raft batch and send a globalCut
 to the data layer
*/
func (server *orderServer) proposalRaftBatch() {
	ticker := time.NewTicker(1000 * time.Millisecond) // todo remove hard-coded interval
	for range ticker.C {
		if server.rc.node == nil {
			continue
		}
		isLeader := server.rc.leaderID == uint64(server.rc.id)

		//do nothing if you're not the leader
		if !isLeader {
			continue
		}
		// propose a batch operation to raft
		batchReq := &pb.ReportRequest{Batch: true}
		marshaledReq, err := proto.Marshal(batchReq)
		if err != nil {
			logger.Panicf("Could not marshal batch request proposal")
		}
		server.rc.proposeC <- marshaledReq
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

		req := &pb.ReportRequest{}
		if err := proto.Unmarshal(entry.Data, req); err != nil {
			logger.Panicf("Could not unmarshal raft commit message")
		}

		// ReportRequests can take two actions -- one is where we decide upon a shard cut.
		// The other action is if we decide on finalizing a particular shard.
		if req.Finalized {
			finalizedShardID := int(req.ShardID)
			server.finalizationResponseChannels[req.ShardID] <- struct{}{}
			server.deleteShard(finalizedShardID)
			continue
		} else if req.Batch {
			// req is a request to batch and send the global state to data shards
			server.mu.Lock()
			server.mergeContestedCuts()
			server.mu.Unlock()
		} else {
			// req is a set of proposed updates to the global state
			// TODO: ReportRequest should ideally be a map
			cut := make(Cut, server.numServersPerShard)
			for i := 0; i < len(cut); i++ {
				cut[i] = int(req.TentativeCut[i])
			}

			// Update the global state
			server.mu.Lock()
			for i := 0; i < server.numServersPerShard; i++ {
				prior := server.contestedCut[int(req.ShardID)][int(req.ReplicaID)][i]
				server.contestedCut[int(req.ShardID)][int(req.ReplicaID)][i] = golib.Max(prior, cut[i])
			}
			server.mu.Unlock()
		}
		// set the new state
		server.mu.Lock()
		server.logNum = int(entry.Index) + 1
		server.mu.Unlock()
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
		server.logNum,
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
	server.logNum = state.logNum

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
		server.dataResponseChannels[shardID] = make([]chan pb.ReportResponse, server.numServersPerShard)
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
