package data

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/spf13/viper"

	"github.com/scalog/scalog/data/filesystem"
	"github.com/scalog/scalog/data/messaging"
	pb "github.com/scalog/scalog/data/messaging"
)

// Record is an internal representation of a log record
type Record struct {
	cid        int32
	csn        int32
	record     string
	commitResp chan int // Should send back a GSN to be forwarded to client
}

type dataServer struct {
	// Stable storage for entries into this shard
	stableStorage *filesystem.RecordStorage
	// Main storage stacks for incoming records
	// TODO: evantzhao use goroutines and channels to update this instead of mutex
	serverBuffers [][]Record
	// Protects all internal structures
	mu sync.Mutex
	// Live connections to other servers inside the shard -- used for replication
	shardServers []chan messaging.ReplicateRequest
}

func (s *dataServer) Append(c context.Context, req *pb.AppendRequest) (*pb.AppendResponse, error) {
	s.mu.Lock()
	r := &Record{
		cid:        req.Cid,
		csn:        req.Csn,
		record:     req.Record,
		commitResp: make(chan int),
	}
	s.serverBuffers[viper.GetInt("id")] = append(s.serverBuffers[viper.GetInt("id")], *r)
	s.mu.Unlock()

	replicateRequest := &messaging.ReplicateRequest{
		ServerID: viper.GetInt32("id"),
		Record:   req.Record,
	}
	for _, ch := range s.shardServers {
		ch <- *replicateRequest
	}
	gsn := int32(<-r.commitResp)
	return &pb.AppendResponse{Csn: r.csn, Gsn: gsn}, nil
}

func (s *dataServer) Replicate(stream pb.Data_ReplicateServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		s.mu.Lock()
		s.serverBuffers[in.ServerID] = append(s.serverBuffers[in.ServerID], Record{record: in.Record})
		s.mu.Unlock()
	}
}

func (s *dataServer) Commit(stream pb.Data_CommitServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		if len(in.CommittedCuts) != len(s.serverBuffers) {
			return fmt.Errorf(
				fmt.Sprintf(
					"[Data] Received cut is of irregular length (%d vs %d)",
					len(in.CommittedCuts),
					len(s.serverBuffers),
				),
			)
		}

		gsn := int(in.StartGlobalSequenceNum)
		for idx, cut := range in.CommittedCuts {
			offset := int(in.Offsets[idx])
			numLogs := int(cut) - offset

			for i := 0; i < numLogs; i++ {
				s.serverBuffers[idx][offset+i].commitResp <- gsn
				gsn += 1
			}
		}
	}
}

func (s *dataServer) Subscribe(*pb.SubscribeRequest, pb.Data_SubscribeServer) error {
	// TODO: Implement
	return nil
}

func (s *dataServer) Trim(context.Context, *pb.TrimRequest) (*pb.TrimResponse, error) {
	// TODO: Implement
	return nil, nil
}
