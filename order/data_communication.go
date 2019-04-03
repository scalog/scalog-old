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
	shardID := int(req.ShardID)
	terminationMessage := pb.ReportResponse{Finalized: true}

	server.shardMu.Lock()
	for i := 0; i < server.numServersPerShard; i++ {
		// Notify the relavant data replicas that they have been finalized
		server.dataResponseChannels[shardID][i] <- terminationMessage
		// Cleanup channels used for data layer communication
		close(server.dataResponseChannels[shardID][i])
	}
	// Remove shardID from dataResponseChannels
	delete(server.dataResponseChannels, shardID)
	// Remove shardID binding in order layer state
	server.shardIds = removeShardIDFromSlice(shardID, server.shardIds)
	server.shardMu.Unlock()

	resp := &pb.FinalizeResponse{}
	return resp, nil
}

func removeShardIDFromSlice(shardID int, shardIDs []int) []int {
	for idx, id := range shardIDs {
		if id == shardID {
			// Slice out this element
			return append(shardIDs[:idx], shardIDs[idx+1:]...)
		}
	}
	// Element not found
	return shardIDs
}
