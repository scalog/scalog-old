package order

import (
	"context"
	"io"

	"github.com/golang/protobuf/proto"
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
		propData, err := proto.Marshal(req)
		if err != nil {
			logger.Panicf(err.Error())
		}
		prop := raftProposal{
			proposalType: REPORT,
			proposalData: propData,
		}
		server.rc.proposeC <- prop
	}
}

func (server *orderServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	return nil, nil
}

func (server *orderServer) Finalize(ctx context.Context, req *pb.FinalizeRequest) (*pb.FinalizeResponse, error) {
	propData, err := proto.Marshal(req)
	if err != nil {
		logger.Printf("Could not marshal finalization request message")
		return nil, err
	}
	// Create finalizationResponseChannels for each shard. We should wait until we hear back from
	// all of these shards
	for _, shardID := range req.Shards {
		server.finalizationResponseChannels[shardID] = make(chan struct{})
	}
	prop := raftProposal{
		proposalType: FINALIZE,
		proposalData: propData,
	}
	server.rc.proposeC <- prop
	// Wait for raft to commit the finalization requests before responding
	for _, shardID := range req.Shards {
		<-server.finalizationResponseChannels[shardID]
		delete(server.finalizationResponseChannels, shardID)
	}
	return &pb.FinalizeResponse{}, nil
}
