package order

import (
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
