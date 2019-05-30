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
		if !server.rc.isLeader() {
			server.forwardC <- req
		} else {
			serializedReq, err := proto.Marshal(req)
			if err != nil {
				logger.Panicf(err.Error())
			}
			server.rc.proposeC <- serializedReq
		}
	}
}

func (server *orderServer) Forward(stream pb.Order_ForwardServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if !server.rc.isLeader() {
			server.forwardC <- req
		} else {
			serializedReq, err := proto.Marshal(req)
			if err != nil {
				logger.Panicf(err.Error())
			}
			server.rc.proposeC <- serializedReq
		}
	}
}

func (server *orderServer) Finalize(ctx context.Context, req *pb.FinalizeRequest) (*pb.FinalizeResponse, error) {
	proposedShardFinalization := pb.ReportRequest{
		Shards:   nil,
		Finalize: req.Shards,
		Batch:    false,
	}
	raftProposal, err := proto.Marshal(&proposedShardFinalization)
	if err != nil {
		logger.Printf("Could not marshal finalization request message")
		return nil, err
	}
	// Create finalizationResponseChannels for each shard. We should wait until we hear back from
	// all of these shards
	for _, shardID := range req.Shards {
		server.finalizationResponseChannels[shardID] = make(chan struct{})
	}
	server.rc.proposeC <- raftProposal
	// Wait for raft to commit the finalization requests before responding
	for _, shardID := range req.Shards {
		<-server.finalizationResponseChannels[shardID]
		delete(server.finalizationResponseChannels, shardID)
	}
	return &pb.FinalizeResponse{}, nil
}
