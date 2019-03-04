package order

import (
	"github.com/coreos/etcd/raft/raftpb"
	"go.etcd.io/etcd/raft"
)

type raftNode struct {
	incoming <-chan string
	outgoing <-chan string
	config   raft.Config
	node     raft.Node
}

func newRaftNode(id uint64, peers []uint64, incoming <-chan string, outgoing <-chan string) raftNode {
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

	nodeWrapper := raftNode{
		incoming,
		outgoing,
		config,
		node,
	}

	return nodeWrapper
}

func (node *raftNode) readChannels() {
	select {
	case ready := <-node.node.Ready():
		node.saveToStorage(ready.RaftState, ready.Entries, ready.Snapshot)
		node.send(ready.Messages)
		if !raft.IsEmptySnap(ready.Snapshot) {
			node.processSnapshot(ready.Snapshot)
		}

		for _, entry := range ready.CommittedEntries {
			node.process(entry)

			if entry.Type == raftpb.EntryConfChange {
				var configChange raftpb.ConfChange
				configChange.Unmarshal(entry.Data)
				node.node.ApplyConfChange(configChange)
			}
		}

		node.node.Advance()
	}
}

func (node *raftNode) saveToStorage(state raft.StateType, entries []raftpb.Entry, snapshot raftpb.Snapshot) {

}

func (node *raftNode) send(messages []raftpb.Message) {

}

func (node *raftNode) processSnapshot(snapshot raftpb.Snapshot) {

}

func (node *raftNode) process(entries raftpb.Entry) {

}
