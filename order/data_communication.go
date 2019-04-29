package order

import (
	"context"
	"io"

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
		server.rc.toLeaderC <- req
	}
}

func (server *orderServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	return nil, nil
}

func (server *orderServer) Finalize(ctx context.Context, req *pb.FinalizeRequest) (*pb.FinalizeResponse, error) {
	proposedShardFinalization := &pb.ReportRequest{
		Shards:   nil,
		Finalize: req.Shards,
	}
	server.rc.toLeaderC <- proposedShardFinalization
	// Create finalizationResponseChannels for each shard. We should wait until we hear back from
	// all of these shards
	for _, shardID := range req.Shards {
		server.finalizationResponseChannels[shardID] = make(chan struct{})
	}
	// Wait for raft to commit the finalization requests before responding
	for _, shardID := range req.Shards {
		<-server.finalizationResponseChannels[shardID]
		delete(server.finalizationResponseChannels, shardID)
	}
	return &pb.FinalizeResponse{}, nil
}
