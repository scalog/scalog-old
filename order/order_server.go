package order

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	pb "github.com/scalog/scalog/order/messaging"
	"github.com/spf13/viper"
	"go.etcd.io/etcd/etcdserver/api/snap"
	"go.etcd.io/etcd/raft"

	"github.com/scalog/scalog/internal/pkg/golib"
	"github.com/scalog/scalog/logger"
)

// Cut is a replica's latest view of cuts for all replicas in its shard
type Cut []int

// All cuts from all replicas within a shard
type ShardCuts []Cut

// Map<Shard ID, Shard cuts reported by all servers in shard>
type ContestedCut map[int]ShardCuts

// Map<Shard ID, Cuts in the shard>
type CommittedCut map[int32]*pb.Cut

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
	committedCut           CommittedCut
	contestedCut           ContestedCut
	finalizeShardRequests  map[int32]*pb.FinalizeRequest
	finalizedShards        *golib.Set32
	globalSequenceNum      int32
	shardIds               *golib.Set
	numServersPerShard     int
	mu                     sync.RWMutex
	reportResponseChannels []chan *pb.ReportResponse
	rc                     *raftNode
	viewID                 int32
	viewUpdateChannels     []chan *pb.RegisterResponse
	viewMu                 sync.RWMutex
	forwardC               chan *pb.ReportRequest
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
		committedCut:           initCommittedCut(shardIds, numServersPerShard),
		contestedCut:           initContestedCut(shardIds, numServersPerShard),
		finalizeShardRequests:  make(map[int32]*pb.FinalizeRequest),
		finalizedShards:        golib.NewSet32(),
		globalSequenceNum:      0,
		shardIds:               shardIds,
		numServersPerShard:     numServersPerShard,
		mu:                     sync.RWMutex{},
		reportResponseChannels: make([]chan *pb.ReportResponse, 0),
		viewID:                 0,
		viewUpdateChannels:     make([]chan *pb.RegisterResponse, 0),
		viewMu:                 sync.RWMutex{},
		forwardC:               make(chan *pb.ReportRequest),
	}
}

func (server *orderServer) connectToLeader() {
	if server.rc.isLeader() {
		return
	}
	server.rc.leaderMu.RLock()
	leaderID := server.rc.leaderID
	server.rc.leaderMu.RUnlock()
	if leaderID == raft.None {
		logger.Printf("Failed to connect to leader: no leader")
		return
	}
	address := strings.TrimPrefix(server.rc.peers[leaderID-1], "http://")
	conn := golib.ConnectTo(address)
	client := pb.NewOrderClient(conn)
	stream, err := client.Forward(context.Background())
	if err != nil {
		logger.Printf(err.Error())
	}
	for req := range server.forwardC {
		if err := stream.Send(req); err != nil {
			logger.Printf(err.Error())
			go server.connectToLeader()
			return
		}
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
	for shardID, req := range server.finalizeShardRequests {
		if req.Limit == 0 {
			shardsToFinalize = append(shardsToFinalize, req)
			delete(server.finalizeShardRequests, shardID)
		} else {
			server.finalizeShardRequests[shardID].Limit--
		}

	}
	return shardsToFinalize
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

func (server *orderServer) proposeGlobalCutToRaft() {
	interval := time.Duration(viper.GetInt("batch_interval"))
	ticker := time.NewTicker(interval * time.Millisecond)
	for range ticker.C {
		if server.rc.node == nil {
			continue
		}
		if !server.rc.isLeader() {
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
		server.updateFinalizeShardRequests()
	}
}

func (server *orderServer) updateFinalizeShardRequests() {
	finalizeShards := make([]*pb.FinalizeRequest, 0)
	server.mu.Lock()
	for shardID, req := range server.finalizeShardRequests {
		if req.Limit == 0 {
			finalizeShards = append(finalizeShards, req)
			delete(server.finalizeShardRequests, shardID)
		} else {
			server.finalizeShardRequests[shardID].Limit--
		}
	}
	server.mu.Unlock()
	for _, req := range finalizeShards {
		propData, err := proto.Marshal(req)
		if err != nil {
			logger.Printf("Could not marshal finalization request message")
			continue
		}
		prop := raftProposal{
			proposalType: FINALIZE,
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
			resp := &pb.ReportResponse{}
			if err := proto.Unmarshal(entry.Data, resp); err != nil {
				logger.Printf(err.Error())
				continue
			}
			server.mu.Lock()
			server.committedCut = resp.CommitedCuts
			server.globalSequenceNum = resp.StartGSN
			for shardID, replicaCuts := range resp.CommitedCuts {
				for replicaID, cut := range replicaCuts.Cut {
					for i := 0; i < server.numServersPerShard; i++ {
						prior := server.contestedCut[int(shardID)][int(replicaID)][i]
						server.contestedCut[int(shardID)][int(replicaID)][i] = golib.Max(prior, int(cut))
					}
				}
			}
			for _, respC := range server.reportResponseChannels {
				respC <- resp
			}
			server.mu.Unlock()
		case REGISTER:
			server.viewMu.Lock()
			server.viewID++
			viewUpdate := &pb.RegisterResponse{
				ViewID:          server.viewID,
				FinalizeShardID: -1,
			}
			for _, viewUpdateC := range server.viewUpdateChannels {
				viewUpdateC <- viewUpdate
			}
			server.viewMu.Unlock()
		case FINALIZE:
			req := &pb.FinalizeRequest{}
			if err := proto.Unmarshal(entry.Data, req); err != nil {
				logger.Printf(err.Error())
				continue
			}
			resp := &pb.RegisterResponse{
				ViewID:          server.viewID,
				FinalizeShardID: req.ShardID,
			}
			for _, viewUpdateC := range server.viewUpdateChannels {
				viewUpdateC <- resp
			}
			server.finalizedShards.Add(req.ShardID)
			server.deleteShard(req.ShardID)
		default:
			logger.Printf("Invalid raft proposal")
		}
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
