package order

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/scalog/scalog/internal/pkg/golib"
	log "github.com/scalog/scalog/logger"
	"github.com/scalog/scalog/order/orderpb"

	"github.com/spf13/viper"
	"go.etcd.io/etcd/etcdserver/api/snap"
	"go.etcd.io/etcd/raft"
	"go.etcd.io/etcd/raft/raftpb"
)

type raftProposalType int

const (
	REPORT raftProposalType = iota
	FINALIZE
)

type raftProposal struct {
	proposalType raftProposalType
	proposalData []byte
}

type OrderServer struct {
	rc                     *raftNode
	mu                     sync.RWMutex
	leaderMu               sync.RWMutex
	leaderStream           orderpb.Order_ForwardClient
	localCuts              map[int32]*orderpb.LocalCut
	state                  OrderServerState
	proposeC               chan string
	confChangeC            chan raftpb.ConfChange
	commitC                <-chan *string
	errorC                 <-chan error
	snapshotterC           <-chan *snap.Snapshotter
	forwardC               chan *orderpb.LocalCuts
	finalizeC              chan *orderpb.FinalizeEntry
	leaderChangeC          chan int32
	reportResponseChannels []chan *orderpb.CommittedEntry
}

// Fields in OrderServer necessary for state replication. Used in Raft.
type OrderServerState struct {
	viewID           int32
	nextGSN          int32 // next global sequence number
	lastCommittedCut map[int32]int32
	shards           map[int32]bool // true: live, false: finalized
}

func NewOrderServer(shardIds *golib.Set, numServersPerShard int) *OrderServer {
	return &OrderServer{
		proposeC:               make(chan string),
		confChangeC:            make(chan raftpb.ConfChange),
		commitC:                nil,
		errorC:                 nil,
		snapshotterC:           nil,
		forwardC:               make(chan *orderpb.LocalCuts),
		finalizeC:              make(chan *orderpb.FinalizeEntry),
		leaderChangeC:          make(chan int32),
		reportResponseChannels: make([]chan *orderpb.CommittedEntry, 0),
	}
}

func (s *OrderServer) Start() {
	s.connectToLeader()
	numReplicas := viper.GetInt("replica_count")
	interval := viper.GetInt("batching_interval")
	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
	for {
		select {
		case <-s.leaderChangeC:
			s.connectToLeader()
		case lcs := <-s.forwardC:
			if s.rc.isLeader() { // store local cuts to local storage
				for i := 0; i < len(lcs.Cuts); i++ {
					lc := lcs.Cuts[i]
					grid := lc.ShardID*int32(numReplicas) + lc.LocalReplicaID // global replica id
					s.localCuts[grid] = lc
				}
			} else { // forward local cuts to leader
				if err := s.leaderStream.Send(lcs); err != nil {
					log.Errorf("%v", err)
				}
			}
		case fe := <-s.finalizeC: // immediately propose finalization requests
			entry := &orderpb.CommittedEntry{}
			entry.Seq = 0
			entry.FinalizeShards = fe
			s.proposeC <- entry.String()
		case <-ticker.C:
			if s.rc.isLeader() { // compute minimum cut and propose
			}
		}
	}
}

func (s *OrderServer) connectToLeader() {
	s.rc.leaderMu.RLock()
	leaderID := s.rc.leaderID
	s.rc.leaderMu.RUnlock()
	if leaderID == raft.None {
		log.Printf("Failed to connect to leader: no leader")
		return
	}
	address := strings.TrimPrefix(s.rc.peers[leaderID-1], "http://")
	conn := golib.ConnectTo(address)
	client := orderpb.NewOrderClient(conn)
	stream, err := client.Forward(context.Background())
	if err != nil {
		log.Printf(err.Error())
	}
	s.leaderStream = stream
}

// Extracts all variables necessary in state replication in OrderServer.
func (s *OrderServer) getState() OrderServerState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	state := OrderServerState{}
	state.viewID = s.state.viewID
	state.nextGSN = s.state.nextGSN
	state.lastCommittedCut = make(map[int32]int32)
	for k, v := range s.state.lastCommittedCut {
		state.lastCommittedCut[k] = v
	}
	state.shards = make(map[int32]bool)
	for k, v := range s.state.shards {
		state.shards[k] = v
	}
	return state
}

// Overwrites current s data with state data.
// NOTE: assumes that shardIds and numServersPerShard do not change. If they change, then we must recreate channels.
func (s *OrderServer) loadState(state OrderServerState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = OrderServerState{}
	s.state.viewID = state.viewID
	s.state.nextGSN = state.nextGSN
	s.state.lastCommittedCut = make(map[int32]int32)
	for k, v := range state.lastCommittedCut {
		s.state.lastCommittedCut[k] = v
	}
	s.state.shards = make(map[int32]bool)
	for k, v := range state.shards {
		s.state.shards[k] = v
	}
}

// Returns all stored data.
func (s *OrderServer) getSnapshot() ([]byte, error) {
	return json.Marshal(s.getState())
}

// Use snapshot to reload s state.
func (s *OrderServer) attemptRecoverFromSnapshot() {
	snapshot, err := s.rc.snapshotter.Load()
	if err == snap.ErrNoSnapshot {
		return
	}
	if err != nil {
		log.Panicf(err.Error())
	}
	log.Printf("Ordering layer attempting to recover from Raft snapshot")
	state := OrderServerState{}
	err = json.Unmarshal(snapshot.Data, state)
	if err != nil {
		log.Panicf(err.Error())
	}
	s.loadState(state)
}
