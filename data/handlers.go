package data

import (
	"context"
	"errors"
	"io"
	"sync"

	"k8s.io/client-go/kubernetes"

	"github.com/scalog/scalog/data/filesystem"
	"github.com/scalog/scalog/data/messaging"
	pb "github.com/scalog/scalog/data/messaging"
)

// Record is an internal representation of a log record
type Record struct {
	cid        int32
	csn        int32
	gsn        int32
	record     string
	commitResp chan int // Should send back a GSN to be forwarded to client
}

type dataServer struct {
	// ID of this data server within a shard
	replicaID int
	// Number of servers within a data shard
	replicaCount int
	// Shard ID
	shardID int32
	// True if this replica has been finalized
	isFinalized bool
	// Stable storage for entries into this shard
	stableStorage *filesystem.RecordStorage
	// Main storage stacks for incoming records
	// TODO: evantzhao use goroutines and channels to update this instead of mutex
	serverBuffers [][]Record
	// Protects all internal structures
	mu sync.Mutex
	// Live connections to other servers inside the shard -- used for replication
	shardServers []chan messaging.ReplicateRequest
	// Client for API calls with kubernetes
	kubeClient *kubernetes.Clientset
}

func (server *dataServer) Append(c context.Context, req *pb.AppendRequest) (*pb.AppendResponse, error) {
	if server.isFinalized {
		return nil, errors.New("this shard has been finalized. No further appends will be permitted")
	}

	server.mu.Lock()
	r := &Record{
		cid:        req.Cid,
		csn:        req.Csn,
		record:     req.Record,
		commitResp: make(chan int),
	}
	server.serverBuffers[server.replicaID] = append(server.serverBuffers[server.replicaID], *r)
	server.mu.Unlock()

	replicateRequest := &messaging.ReplicateRequest{
		ServerID: int32(server.replicaID),
		Record:   req.Record,
	}
	for _, ch := range server.shardServers {
		ch <- *replicateRequest
	}
	gsn := int32(<-r.commitResp)
	return &pb.AppendResponse{Csn: r.csn, Gsn: gsn}, nil
}

func (server *dataServer) Replicate(stream pb.Data_ReplicateServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		server.mu.Lock()
		server.serverBuffers[in.ServerID] = append(server.serverBuffers[in.ServerID], Record{record: in.Record})
		server.mu.Unlock()
	}
}

func (server *dataServer) Subscribe(*pb.SubscribeRequest, pb.Data_SubscribeServer) error {
	// TODO: Implement
	return nil
}

func (server *dataServer) Trim(context.Context, *pb.TrimRequest) (*pb.TrimResponse, error) {
	// TODO: Implement
	return nil, nil
}
