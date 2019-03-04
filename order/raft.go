package order

import (
	"github.com/coreos/etcd/raft/raftpb"
	"go.etcd.io/etcd/raft"
)

type raftNode struct {
	incoming <-chan string
	outgoing <-chan string

	configuration raftpb.ConfState

	node    raft.Node
	storage *raft.MemoryStorage
}
