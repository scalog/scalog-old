package order

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/scalog/scalog/internal/pkg/golib"
	log "github.com/scalog/scalog/logger"
	"github.com/scalog/scalog/order/orderpb"

	"github.com/gogo/protobuf/proto"
	"github.com/spf13/viper"
	"go.etcd.io/etcd/etcdserver/api/snap"
	"go.etcd.io/etcd/raft"
)

// Cut is a replica's latest view of cuts for all replicas in its shard
type Cut []int

// All cuts from all replicas within a shard
type ShardCuts []Cut

// Map<Shard ID, Shard cuts reported by all servers in shard>
type ContestedCut map[int]ShardCuts

// Map<Shard ID, Cuts in the shard>
type CommittedCut map[int32]*orderpb.Cut

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
	finalizeShardRequests  []*orderpb.FinalizeRequest
	finalizedShards        *golib.Set32
	globalSequenceNum      int32
	shardIds               *golib.Set
	numServersPerShard     int
	mu                     sync.RWMutex
	reportResponseChannels []chan *orderpb.ReportResponse
	rc                     *raftNode
	viewID                 int32
	viewMu                 sync.RWMutex
	forwardC               chan *orderpb.ReportRequest
}

// Fields in orderServer necessary for state replication. Used in Raft.
type orderServerState struct {
	committedCut      CommittedCut
	globalSequenceNum int32
	shardIds          *golib.Set
}

func newOrderServer(shardIds *golib.Set, numServersPerShard int) *orderServer {
	return &orderServer{
		committedCut:          initCommittedCut(shardIds, numServersPerShard),
		contestedCut:          initContestedCut(shardIds, numServersPerShard),
		finalizeShardRequests: make([]*orderpb.FinalizeRequest, 0),
		finalizedShards:       golib.NewSet32(),
		globalSequenceNum:     0,
		shardIds:              shardIds,
		numServersPerShard:    numServersPerShard,
		mu:                    sync.RWMutex{},
		reportResponseChannels: make([]chan *orderpb.ReportResponse, 0),
		viewID:                 0,
		viewMu:                 sync.RWMutex{},
		forwardC:               make(chan *orderpb.ReportRequest),
	}
}

func (server *orderServer) connectToLeader() {
	if server.rc.isLeader() {
		server.listenForForwards()
		return
	}
	server.rc.leaderMu.RLock()
	leaderID := server.rc.leaderID
	server.rc.leaderMu.RUnlock()
	if leaderID == raft.None {
		log.Printf("Failed to connect to leader: no leader")
		return
	}
	address := strings.TrimPrefix(server.rc.peers[leaderID-1], "http://")
	conn := golib.ConnectTo(address)
	client := orderpb.NewOrderClient(conn)
	stream, err := client.Forward(context.Background())
	if err != nil {
		log.Printf(err.Error())
	}
	for req := range server.forwardC {
		if err := stream.Send(req); err != nil {
			log.Printf(err.Error())
			go server.connectToLeader()
			return
		}
	}
}

func (server *orderServer) listenForForwards() {
	for req := range server.forwardC {
		if !server.rc.isLeader() {
			go server.connectToLeader()
			return
		}
		server.updateContestedCut(req)
	}
}

////////////// INITIALIZERS

func initCommittedCut(shardIds *golib.Set, numServersPerShard int) CommittedCut {
	cut := make(CommittedCut)
	for shardID := range shardIds.Iterable() {
		cut[int32(shardID)] = &orderpb.Cut{
			Cut: make([]int32, numServersPerShard),
		}
	}
	return cut
}

func initContestedCut(shardIds *golib.Set, numServersPerShard int) ContestedCut {
	cut := make(ContestedCut)
	for shardID := range shardIds.Iterable() {
		cut[shardID] = newShardCuts(numServersPerShard)
	}
	return cut
}

func newShardCuts(numReplicasPerShard int) ShardCuts {
	s := make(ShardCuts, numReplicasPerShard)
	for i := 0; i < numReplicasPerShard; i++ {
		s[i] = make(Cut, numReplicasPerShard)
	}
	return s
}

////////////// ORDER SERVER STATE MUTATORS

func (server *orderServer) updateContestedCut(req *orderpb.ReportRequest) {
	server.mu.Lock()
	defer server.mu.Unlock()
	for shardID, shardView := range req.Shards {
		server.addShard(int(shardID))
		for replicaID, cut := range shardView.Replicas {
			for i := 0; i < server.numServersPerShard; i++ {
				prior := server.contestedCut[int(shardID)][int(replicaID)][i]
				server.contestedCut[int(shardID)][int(replicaID)][i] = golib.Max(prior, int(cut.Cut[i]))
			}
		}
	}
}

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

// Assumes server lock is acquired
func (server *orderServer) addShard(shardID int) {
	in := server.shardIds.Contains(int(shardID))
	if in {
		return
	}
	server.shardIds.Add(shardID)
	server.committedCut[int32(shardID)] = &orderpb.Cut{Cut: make([]int32, server.numServersPerShard)}
	server.contestedCut[shardID] = newShardCuts(server.numServersPerShard)
}

func (server *orderServer) deleteShard(shardID int32) {
	server.mu.RLock()
	in := server.shardIds.Contains(int(shardID))
	server.mu.RUnlock()
	if !in {
		return
	}
	server.mu.Lock()
	defer server.mu.Unlock()
	server.shardIds.Remove(int(shardID))
	delete(server.committedCut, shardID)
	delete(server.contestedCut, int(shardID))
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
		resp := &orderpb.ReportResponse{
			CommitedCuts: server.committedCut,
			StartGSN:     server.globalSequenceNum,
		}
		server.mu.Unlock()
		propData, err := json.Marshal(resp)
		if err != nil {
			log.Printf(err.Error())
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
	shardsToFinalize := make([]*orderpb.FinalizeRequest, 0)
	shardsPendingFinalize := make([]*orderpb.FinalizeRequest, 0)
	server.mu.Lock()
	for _, req := range server.finalizeShardRequests {
		if req.Limit <= 0 {
			shardsToFinalize = append(shardsToFinalize, req)
		} else {
			req.Limit--
			shardsPendingFinalize = append(shardsPendingFinalize, req)
		}
	}
	server.finalizeShardRequests = shardsPendingFinalize
	server.mu.Unlock()
	for _, req := range shardsToFinalize {
		propData, err := proto.Marshal(req)
		if err != nil {
			log.Printf("Could not marshal finalization request message")
			continue
		}
		prop := raftProposal{
			proposalType: FINALIZE,
			proposalData: propData,
		}
		server.rc.proposeC <- prop
	}
}

func (server *orderServer) listenForRaftCommits() {
	for entry := range server.rc.commitC {
		if entry == nil {
			server.attemptRecoverFromSnapshot()
			continue
		}

		prop := &raftProposal{}
		if err := json.Unmarshal(entry.Data, prop); err != nil {
			log.Printf(err.Error())
			continue
		}

		switch prop.proposalType {
		case REPORT:
			resp := &orderpb.ReportResponse{}
			if err := proto.Unmarshal(entry.Data, resp); err != nil {
				log.Printf(err.Error())
				continue
			}
			server.viewMu.RLock()
			resp.ViewID = server.viewID
			server.viewMu.RUnlock()
			server.mu.Lock()
			server.committedCut = resp.CommitedCuts
			server.globalSequenceNum = resp.StartGSN
			for shardID, replicaCuts := range resp.CommitedCuts {
				server.addShard(int(shardID))
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
			resp := &orderpb.ReportResponse{ViewID: server.viewID}
			server.viewMu.Unlock()
			for _, respC := range server.reportResponseChannels {
				respC <- resp
			}
		case FINALIZE:
			req := &orderpb.FinalizeRequest{}
			if err := proto.Unmarshal(entry.Data, req); err != nil {
				log.Printf(err.Error())
				continue
			}
			server.viewMu.Lock()
			server.viewID++
			resp := &orderpb.ReportResponse{
				ViewID:           server.viewID,
				FinalizeShardIDs: req.ShardIDs,
			}
			server.viewMu.Unlock()
			for _, respC := range server.reportResponseChannels {
				respC <- resp
			}
			for _, shardID := range req.ShardIDs {
				server.finalizedShards.Add(shardID)
				server.deleteShard(shardID)
			}
		default:
			log.Printf("Invalid raft proposal committed")
		}
	}
}

// Extracts all variables necessary in state replication in orderServer.
func (server *orderServer) getState() orderServerState {
	server.mu.RLock()
	defer server.mu.RUnlock()
	return orderServerState{
		server.committedCut,
		server.globalSequenceNum,
		server.shardIds,
	}
}

// Overwrites current server data with state data.
// NOTE: assumes that shardIds and numServersPerShard do not change. If they change, then we must recreate channels.
func (server *orderServer) loadState(state *orderServerState) {
	server.mu.Lock()
	defer server.mu.Unlock()
	server.committedCut = state.committedCut
	server.globalSequenceNum = state.globalSequenceNum
	server.shardIds = state.shardIds
}

// Returns all stored data.
func (server *orderServer) getSnapshot() ([]byte, error) {
	return json.Marshal(server.getState())
}

// Use snapshot to reload server state.
func (server *orderServer) attemptRecoverFromSnapshot() {
	snapshot, err := server.rc.snapshotter.Load()
	if err == snap.ErrNoSnapshot {
		return
	}
	if err != nil {
		log.Panicf(err.Error())
	}

	log.Printf("Ordering layer attempting to recover from Raft snapshot")
	state := &orderServerState{}
	err = json.Unmarshal(snapshot.Data, state)

	if err != nil {
		log.Panicf(err.Error())
	}

	server.loadState(state)
}
