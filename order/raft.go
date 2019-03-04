package order

import (
	"go.etcd.io/etcd/raft"
	"go.etcd.io/etcd/raft/raftpb"
)

type raftNode struct {
	incoming <-chan string
	outgoing <-chan string

	configuration raftpb.ConfState

	node    raft.Node
	storage *raft.MemoryStorage
}
