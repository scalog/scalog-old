package order

import (
	"context"
	"fmt"
	"io"

	"github.com/golang/protobuf/proto"
	"github.com/scalog/scalog/internal/pkg/golib"
	pb "github.com/scalog/scalog/order/messaging"
)

func (server *orderServer) Report(stream pb.Order_ReportServer) error {
	go server.respondToDataReplica(stream)
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

func (server *orderServer) respondToDataReplica(stream pb.Order_ReportServer) {
	respC := make(chan *pb.ReportResponse)
	server.mu.Lock()
	server.reportResponseChannels = append(server.reportResponseChannels, respC)
	server.mu.Unlock()
	for resp := range respC {
		if err := stream.Send(resp); err != nil {
			return
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
			// Forward request to leader so it is not lost
			for shardID := range req.Shards {
				server.addShard(int(shardID))
			}
			server.forwardC <- req
			return fmt.Errorf("Forwarded to non-leader: update leader connection")
		}
		server.updateState(req)
	}
}

func (server *orderServer) updateState(req *pb.ReportRequest) {
	server.mu.Lock()
	defer server.mu.Unlock()
	for shardID, shardView := range req.Shards {
		for replicaID, cut := range shardView.Replicas {
			for i := 0; i < server.numServersPerShard; i++ {
				prior := server.contestedCut[int(shardID)][int(replicaID)][i]
				server.contestedCut[int(shardID)][int(replicaID)][i] = golib.Max(prior, int(cut.Cut[i]))
			}
		}
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
	viewC := make(chan *pb.RegisterResponse)
	server.viewMu.Lock()
	server.viewUpdateChannels = append(server.viewUpdateChannels, viewC)
	server.viewMu.Unlock()
	for viewUpdate := range viewC {
		if err := stream.Send(viewUpdate); err != nil {
			return err
		}
	}
	return nil
}

func (server *orderServer) Finalize(ctx context.Context, req *pb.FinalizeRequest) (*pb.FinalizeResponse, error) {
	if _, in := server.finalizeShardRequests[req.ShardID]; !in {
		server.mu.Lock()
		server.finalizeShardRequests[req.ShardID] = req
		server.mu.Unlock()
	}
	return &pb.FinalizeResponse{}, nil
}
