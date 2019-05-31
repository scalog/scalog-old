package order

import (
	"context"
	"fmt"
	"io"

	"github.com/golang/protobuf/proto"
	"github.com/scalog/scalog/internal/pkg/golib"
	pb "github.com/scalog/scalog/order/messaging"
)

/**
Receives messages from the data layer. Spawns response function with a pointer to the stream to respond to data layer with newly committed cuts.
*/
func (server *orderServer) Report(stream pb.Order_ReportServer) error {
	go server.reportResponseRoutine(stream)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		for shardID := range req.Shards {
			server.addShard(int(shardID))
		}
		server.forwardC <- req
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
			return fmt.Errorf("Forward request sent to non-leader")
		}
		server.mu.Lock()
		for shardID, replicaCuts := range req.Shards {
			for replicaID, cut := range replicaCuts.Replicas {
				for i := 0; i < server.numServersPerShard; i++ {
					prior := server.contestedCut[int(shardID)][int(replicaID)][i]
					server.contestedCut[int(shardID)][int(replicaID)][i] = golib.Max(prior, int(cut.Cut[i]))
				}
			}
		}
		server.mu.Unlock()
	}
}

func (server *orderServer) Register(req *pb.RegisterRequest, stream pb.Order_RegisterServer) error {
	resp := &pb.RegisterResponse{ViewID: server.viewID}
	if err := stream.Send(resp); err != nil {
		return err
	}
	propData, err := proto.Marshal(req)
	if err != nil {
		return err
	}
	prop := raftProposal{
		proposalType: REGISTER,
		proposalData: propData,
	}
	server.rc.proposeC <- prop
	for viewUpdate := range server.viewC {
		if err := stream.Send(viewUpdate); err != nil {
			return err
		}
	}
	return nil
}

func (server *orderServer) Finalize(ctx context.Context, req *pb.FinalizeRequest) (*pb.FinalizeResponse, error) {
	if _, in := server.finalizeMap[req.ShardID]; !in {
		server.finalizeMap[req.ShardID] = req
		server.finalizationResponseChannels[req.ShardID] = make(chan pb.FinalizeResponse)

	}
	resp := <-server.finalizationResponseChannels[req.ShardID]
	delete(server.finalizationResponseChannels, req.ShardID)
	return &resp, nil
}
