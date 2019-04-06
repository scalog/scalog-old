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

		// Check for missing log requests
		if req.MinLogNum != 0 && req.MaxLogNum != 0 {
			server.provideMissingCuts(int(req.MinLogNum), int(req.MaxLogNum), int(req.ShardID), stream)
			continue
		}

		server.rc.leaderMu.RLock()
		if !server.rc.isLeader {
			//forward requests to the leader
			server.rc.toLeaderStream.Send(req)
			server.rc.leaderMu.RUnlock()
		} else {
			server.rc.leaderMu.RUnlock()

			if req.Finalized {
				// finalize cut
				server.notifyFinalizedShard(int(req.ShardID))
				server.deleteShard(int(req.ShardID))
			} else {
				// save the new cut info
				server.saveTentativeCut(req)
			}
		}
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
		server.rc.leaderMu.RLock()
		if !server.rc.isLeader {
			server.rc.leaderMu.RUnlock()
			logger.Printf("Received forwarded message, but is not the leader")
			continue
		}
		server.rc.leaderMu.RUnlock()

		server.saveTentativeCut(req)
	}
}

func (server *orderServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	return nil, nil
}

func (server *orderServer) Finalize(ctx context.Context, req *pb.FinalizeRequest) (*pb.FinalizeResponse, error) {
	server.rc.leaderMu.RLock()

	if server.rc.isLeader {
		server.rc.leaderMu.RUnlock()
		for _, shardID := range req.ShardIDs {
			server.finalizationResponseChannels[shardID] = make(chan bool)
		}
	} else {
		//drop the message if no leader exists
		if server.rc.toLeaderStream == nil {
			server.rc.leaderMu.RUnlock()
			logger.Printf("Received finalize request with no Raft leader")
			return nil, nil
		}

		for _, shardID := range req.ShardIDs {
			server.rc.toLeaderStream.Send(&pb.ReportRequest{ShardID: shardID, Finalized: true})
		}
		server.rc.leaderMu.RUnlock()
	}

	// Wait for raft to commit the finalization requests before responding
	for _, shardID := range req.ShardIDs {
		<-server.finalizationResponseChannels[shardID]
		delete(server.finalizationResponseChannels, shardID)
	}
	return &pb.FinalizeResponse{}, nil
}
