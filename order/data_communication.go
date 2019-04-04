package order

import (
	"context"
	"io"

	"github.com/gogo/protobuf/proto"
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

		//propose message to raft
		marshalBytes, err := proto.Marshal(req)
		if err != nil {
			logger.Printf("Could not marshal data layer message")
			return nil
		}
		server.raftProposeChannel <- string(marshalBytes)
	}
}

func (server *orderServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	return nil, nil
}

func (server *orderServer) Finalize(ctx context.Context, req *pb.FinalizeRequest) (*pb.FinalizeResponse, error) {
	for _, shardID := range req.ShardIDs {
		proposedShardFinalization := pb.ReportRequest{ShardID: shardID, Finalized: true}
		raftReq, err := proto.Marshal(&proposedShardFinalization)
		if err != nil {
			logger.Printf("Could not marshal finalization request message")
			return nil, err
		}
		server.raftProposeChannel <- string(raftReq)
	}
	resp := &pb.FinalizeResponse{}
	return resp, nil
}
