package order

import (
	"context"
	"fmt"
	"io"

	"github.com/golang/protobuf/proto"
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
			server.forwardC <- req
			return fmt.Errorf("Forwarded to non-leader: update leader connection")
		}
		server.updateContestedCut(req)
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
