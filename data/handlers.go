package data

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/spf13/viper"

	"github.com/scalog/data/filesystem"
	"github.com/scalog/data/messaging"
	pb "github.com/scalog/data/messaging"
)

// Record is a pending message to the ordering layer
type Record struct {
	cid        int32
	csn        int32
	record     string
	commitResp chan int // Should send back a GSN
}

// We should mount some persistent volume. the filesystem package will be used to write
// to this mounted volume.
type dataServer struct {
	// Storage for this server in particular
	stableStorage *filesystem.RecordStorage

	// Main storage stacks for incoming records
	// TODO: evantzhao make this faster by using a separate go routine for each
	// server stack.
	serverBuffers [][]Record

	// lastCommitedRecords stores the integer value of the last commited
	// value in each server buffer
	lastCommitedRecords []int

	// TODO: evantzhao make this more efficient by either utilizing mutexes for
	// each server log. Also consider channel
	// Protects all internal structures
	mu sync.Mutex

	// Items written to these channels will be sent to these corresponding pods
	shardServers []chan messaging.ReplicateRequest
}

// Append adds the specified record to the local data server, and this server additionally
// forwards the append request to other servers within the shard with Replicate. Responds
// to the client with a global sequence number when the Ordering layer commits
// this record.
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
	for idx, ch := range s.shardServers {
		if idx == viper.GetInt("id") {
			continue
		}
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

		if len(in.Cut) != len(s.serverBuffers) {
			panic(
				fmt.Sprintf("[Data] Received cut is of irregular length (%d vs %d)",
					len(in.Cut),
					len(s.serverBuffers),
				),
			)
		}

		// TODO evanzhao: implement cut update logic. Should push GSN's to record channel
	}
}

func (s *dataServer) Subscribe(*pb.SubscribeRequest, pb.Data_SubscribeServer) error {
	return nil
}

func (s *dataServer) Trim(context.Context, *pb.TrimRequest) (*pb.TrimResponse, error) {
	return nil, nil
}
