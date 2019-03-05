package order

import (
	"context"
	"github.com/coreos/etcd/raft/raftpb"
	"go.etcd.io/etcd/raft"
	"time"
)

/**
Reference https://godoc.org/go.etcd.io/etcd/raft
TODO: Change state change type from string
*/
type RaftNodeWrapper struct {
	stateChanges  <-chan string
	configChanges <-chan raftpb.ConfChange
	config        raft.Config
	storage       *raft.MemoryStorage
	node          raft.Node
}

func NewRaftNode(id uint64, peers []uint64, stateChanges <-chan string, configChanges <-chan raftpb.ConfChange) RaftNodeWrapper {
	storage := raft.NewMemoryStorage()

	config := raft.Config{
		ID:              id,
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         storage,
		MaxSizePerMsg:   4096,
		MaxInflightMsgs: 256,
	}

	peersObject := make([]raft.Peer, len(peers))
	for i := range peers {
		peersObject[i] = raft.Peer{ID: peers[i]}
	}

	node := raft.StartNode(&config, peersObject)

	nodeWrapper := RaftNodeWrapper{
		stateChanges,
		configChanges,
		config,
		storage,
		node,
	}

	go nodeWrapper.readRaftChannels()
	go nodeWrapper.readStateChanges()

	return nodeWrapper
}

func (nodeWrapper *RaftNodeWrapper) readStateChanges() {
	for nodeWrapper.stateChanges != nil && nodeWrapper.configChanges != nil {
		select {

		case state := <-nodeWrapper.stateChanges:
			nodeWrapper.node.Propose(context.TODO(), []byte(state))
		case config := <-nodeWrapper.configChanges:
			nodeWrapper.node.ProposeConfChange(context.TODO(), config)
		}
	}
}

func (nodeWrapper *RaftNodeWrapper) readRaftChannels() {

	timer := time.NewTicker(1 * time.Millisecond) //Tick rate
	defer timer.Stop()

	for {
		select {

		//tick periodically for heartbeat
		case <-timer.C:
			nodeWrapper.node.Tick()

		case ready := <-nodeWrapper.node.Ready():
			nodeWrapper.save(ready.HardState, ready.Entries, ready.Snapshot)
			nodeWrapper.send(ready.Messages)
			if !raft.IsEmptySnap(ready.Snapshot) {
				nodeWrapper.processSnapshot(ready.Snapshot)
			}

			for _, entry := range ready.CommittedEntries {
				nodeWrapper.process(entry)

				if entry.Type == raftpb.EntryConfChange {
					var configChange raftpb.ConfChange
					configChange.Unmarshal(entry.Data)
					nodeWrapper.node.ApplyConfChange(configChange)
				}
			}

			nodeWrapper.node.Advance()
		}
	}
}

func (nodeWrapper *RaftNodeWrapper) save(state raftpb.HardState, entries []raftpb.Entry, snapshot raftpb.Snapshot) {

}

func (nodeWrapper *RaftNodeWrapper) send(messages []raftpb.Message) {

}

func (nodeWrapper *RaftNodeWrapper) processSnapshot(snapshot raftpb.Snapshot) {

}

func (nodeWrapper *RaftNodeWrapper) process(entries raftpb.Entry) {

}
