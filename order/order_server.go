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

// Map<Shard ID, Channels to write responses to when a shard finalization
// request has been committed>
type FinalizationResponseChannels map[int32]chan struct{}

type orderServer struct {
	committedGlobalCut           CommittedGlobalCut
	contestedGlobalCut           ContestedGlobalCut
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
	committedGlobalCut CommittedGlobalCut
	contestedGlobalCut ContestedGlobalCut
	globalSequenceNum  int
	logNum             int
	shardIds           *golib.Set
}

func newOrderServer(shardIds *golib.Set, numServersPerShard int) *orderServer {
	return &orderServer{
		committedGlobalCut:           initCommittedCut(shardIds, numServersPerShard),
		contestedGlobalCut:           initContestedCut(shardIds, numServersPerShard),
		globalSequenceNum:            0,
		logNum:                       1,
		shardIds:                     shardIds,
		numServersPerShard:           numServersPerShard,
		mu:                           sync.RWMutex{},
		dataResponseChannels:         initResponseChannels(shardIds, numServersPerShard),
		finalizationResponseChannels: make(FinalizationResponseChannels),
	}
}

func initCommittedCut(shardIds *golib.Set, numServersPerShard int) CommittedGlobalCut {
	cut := make(CommittedGlobalCut)
	for shardID := range shardIds.Iterable() {
		cut[shardID] = make(ShardCut, numServersPerShard)
	}
	return cut
}

func initContestedCut(shardIds *golib.Set, numServersPerShard int) ContestedGlobalCut {
	cut := make(ContestedGlobalCut)
	for shardID := range shardIds.Iterable() {
		shardCuts := make([]ShardCut, numServersPerShard)
		for i := 0; i < numServersPerShard; i++ {
			shardCuts[i] = make(ShardCut, numServersPerShard)
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

CAUTION: THE CALLER MUST OBTAIN THE LOCK
*/
func (server *orderServer) mergeContestedCuts() map[int]int {
	startGlobalSequenceNums := make(map[int]int)

	for shardID := range server.shardIds.Iterable() {
		delta := 0

		// find smallest cut for shard i
		for i := 0; i < server.numServersPerShard; i++ {
			minCut := math.MaxInt64
			for j := 0; j < server.numServersPerShard; j++ {
				minCut = golib.Min(server.contestedGlobalCut[shardID][j][i], minCut)
			}

			prevCut := server.committedGlobalCut[shardID][i]
			server.committedGlobalCut[shardID][i] = minCut

			delta += minCut - prevCut
		}

		startGlobalSequenceNums[shardID] = server.globalSequenceNum
		server.globalSequenceNum += delta
	}

	return startGlobalSequenceNums
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
	server.committedGlobalCut[shardID] = make(ShardCut, server.numServersPerShard)
	server.contestedGlobalCut[shardID] = make([]ShardCut, server.numServersPerShard)
	server.dataResponseChannels[shardID] = make([]chan pb.ReportResponse, server.numServersPerShard)

	for i := 0; i < server.numServersPerShard; i++ {
		server.dataResponseChannels[shardID][i] = make(chan pb.ReportResponse)
		server.contestedGlobalCut[shardID][i] = make(ShardCut, server.numServersPerShard)
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
func (server *orderServer) deleteShard(shardID int, committed bool) bool {
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
	delete(server.committedGlobalCut, shardID)
	delete(server.contestedGlobalCut, shardID)

	if !committed {
		return true
	}

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
Saves a cut received from the data layer into contested cuts.
*/
func (server *orderServer) saveTentativeCut(req *pb.ReportRequest) {
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

/**
Periodically merge contested cuts and broadcast.
*/
func (server *orderServer) batchCuts() {
	ticker := time.NewTicker(100 * time.Microsecond) // todo remove hard-coded interval
	for range ticker.C {
		//do nothing if you're not the leader
		server.rc.leaderMu.RLock()
		if !server.rc.isLeader {
			server.rc.leaderMu.RUnlock()
			continue
		}
		server.rc.leaderMu.RUnlock()

		//merge cuts
		server.mu.Lock()
		startGlobalSequenceNums := server.mergeContestedCuts()
		server.logNum += 1
		batchedCuts := createBatchCutResponse(startGlobalSequenceNums, server.committedGlobalCut,
			server.logNum, server.globalSequenceNum)
		server.mu.Unlock()

		//propose cuts to raft
		marshalBytes, err := proto.Marshal(batchedCuts)
		if err != nil {
			logger.Panicf("Could not marshal data layer message")
		}
		server.rc.proposeC <- marshalBytes
	}
}

/**
Cast types to create a ForwardResponse.
*/
func createBatchCutResponse(startGlobalSequenceNums map[int]int, committedCuts CommittedGlobalCut, logNum int, globalSequenceNum int) *pb.ForwardResponse {
	startGlobalSequenceNums32 := make(map[int32]int32)
	committedCuts32 := make(map[int32]*pb.IntList)

	for shardID, gsn := range startGlobalSequenceNums {
		startGlobalSequenceNums32[int32(shardID)] = int32(gsn)
	}
	for shardID, intList := range committedCuts {
		cutsInShard := make([]int32, len(intList))
		for replicaID, cut := range intList {
			cutsInShard[replicaID] = int32(cut)
		}
		committedCuts32[int32(shardID)] = &pb.IntList{List: cutsInShard}
	}

	return &pb.ForwardResponse{
		StartGlobalSequenceNums: startGlobalSequenceNums32,
		CommittedCuts:           committedCuts32,
		StartGlobalSequenceNum:  int32(globalSequenceNum),
	}
}

/**
Reads into orderServer values stored in ForwardResponse by casting types.
*/
func (server *orderServer) loadBatchCut(res *pb.ForwardResponse) {
	committedCuts := make(CommittedGlobalCut)

	for shardID, intList := range res.CommittedCuts {
		cutsInShard := make([]int, len(intList.List))
		for replicaID, cut := range intList.List {
			cutsInShard[replicaID] = int(cut)
		}
		committedCuts[int(shardID)] = cutsInShard
	}

	server.mu.Lock()
	server.committedGlobalCut = committedCuts
	server.globalSequenceNum = int(res.StartGlobalSequenceNum)
	server.mu.Unlock()

	//if we're not the leader, make sure Contested cuts are updated with Committed cuts
	server.rc.leaderMu.RLock()
	if server.rc.isLeader {
		server.rc.leaderMu.RUnlock()
		return
	}
	server.rc.leaderMu.RUnlock()

	for shardID, cuts := range server.committedGlobalCut {
		for i := 0; i < server.numServersPerShard; i++ {
			server.contestedGlobalCut[shardID][i] = cuts
		}
	}
}

/**
Send the requested missing cuts to the data layer, in order.
*/
func (server *orderServer) provideMissingCuts(minLogNum int, maxLogNum int, shardID int, stream pb.Order_ReportServer) {
	entries := server.rc.getEntries(minLogNum, maxLogNum)
	savedCut := &pb.ForwardResponse{}

	for _, entry := range entries {
		if err := proto.Unmarshal(entry.Data, savedCut); err != nil {
			logger.Printf("Attempted to unmarshal a non-ForwardResponse entry. This is OK :)")
			continue
		}

		response := pb.ReportResponse{
			StartGlobalSequenceNum: savedCut.StartGlobalSequenceNums[int32(shardID)],
			CommittedCuts:          savedCut.CommittedCuts[int32(shardID)].List,
			MinLogNum:              int32(minLogNum),
			MaxLogNum:              int32(entry.Index),
			Finalized:              false,
		}
		stream.Send(&response)

		minLogNum = int(entry.Index) + 1
	}
}

////////////////////// RAFT FUNCTIONS

/**
Triggered when Raft commits a new message.
*/
func (server *orderServer) listenForRaftCommits() {
	for entry := range server.rc.commitC {
		if entry.Data == nil {
			server.attemptRecoverFromSnapshot()
			continue
		}

		req := &pb.ForwardResponse{}
		if err := proto.Unmarshal(entry.Data, req); err != nil {
			logger.Panicf("Could not unmarshal raft commit message")
		}

		//respond with committed cut
		for shardID := range server.shardIds.Iterable() {
			var response pb.ReportResponse
			_, shardExists := req.CommittedCuts[int32(shardID)]

			if shardExists {
				response = pb.ReportResponse{
					StartGlobalSequenceNum: req.StartGlobalSequenceNums[int32(shardID)],
					CommittedCuts:          req.CommittedCuts[int32(shardID)].List,
					MinLogNum:              int32(server.logNum),
					MaxLogNum:              int32(entry.Index),
					Finalized:              false,
				}
			} else {
				//finalize shards that we had listed (in shardIds) but is no longer in the cut
				server.deleteShard(shardID, true)
				continue
			}

			for i := 0; i < server.numServersPerShard; i++ {
				server.dataResponseChannels[shardID][i] <- response
			}
		}

		// set the new state
		server.mu.Lock()
		server.logNum = int(entry.Index) + 1
		server.mu.Unlock()
		server.loadBatchCut(req)
	}
}

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

	server.committedGlobalCut = state.committedGlobalCut
	server.contestedGlobalCut = state.contestedGlobalCut
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
