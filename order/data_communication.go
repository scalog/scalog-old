package order

import (
	"context"
	"io"

	"github.com/scalog/scalog/logger"
	pb "github.com/scalog/scalog/order/messaging"
)

/**
Receives messages from the data layer. Spawns response function with a pointer to the stream to respond to data layer with newly committed cuts.
*/
func (server *orderServer) Report(stream pb.Order_ReportServer) error {
	spawned := false
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if !spawned {
			// Boot routine ONLY ONCE for periodically responding to data layer
			go server.reportResponseRoutine(stream, req)
			spawned = true
		}

		server.leaderMu.RLock()
		if *server.isLeader {
			server.saveTentativeCut(req)
		} else {
			toLeaderStream := *server.toLeaderStream
			toLeaderStream.Send(req)
		}
		server.leaderMu.RUnlock()
	}
}

/**
Receives messages from other ordering servers. Store if leader, otherwise ignore.
*/
func (server *orderServer) Forward(stream pb.Order_ForwardServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		// do nothing if you're not the leader
		server.leaderMu.RLock()
		if !*server.isLeader {
			server.leaderMu.RUnlock()
			logger.Printf("Received forwarded message, but is not the leader")
			continue
		}
		server.leaderMu.RUnlock()

		server.saveTentativeCut(req)
	}
}

func (server *orderServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	return nil, nil
}

func (server *orderServer) Finalize(ctx context.Context, req *pb.FinalizeRequest) (*pb.FinalizeResponse, error) {
	server.leaderMu.RLock()
	defer server.leaderMu.RUnlock()

	if *server.isLeader {
		for _, shardID := range req.ShardIDs {
			server.deleteShard(int(shardID))
		}
	} else {
		//drop the message if no leader exists
		if server.toLeaderStream == nil {
			logger.Printf("Received finalize request with no Raft leader")
			return nil, nil
		}

		for _, shardID := range req.ShardIDs {
			stream := *server.toLeaderStream
			stream.Send(&pb.ReportRequest{ShardID: shardID, Finalized: true})
		}
	}

	//TODO only respond once leader sends committedCut without this shard?
	resp := &pb.FinalizeResponse{}
	return resp, nil
}
