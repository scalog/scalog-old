package order

import (
	"context"
	"fmt"
	"io"

	"github.com/golang/protobuf/proto"
	"github.com/scalog/scalog/order/orderpb"
)

func (server *orderServer) Report(stream orderpb.Order_ReportServer) error {
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

func (server *orderServer) respondToDataReplica(stream orderpb.Order_ReportServer) {
	respC := make(chan *orderpb.ReportResponse)
	server.mu.Lock()
	server.reportResponseChannels = append(server.reportResponseChannels, respC)
	server.mu.Unlock()
	for resp := range respC {
		if err := stream.Send(resp); err != nil {
			return
		}
	}
}

func (server *orderServer) Forward(stream orderpb.Order_ForwardServer) error {
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

func (server *orderServer) Register(ctx context.Context, req *orderpb.RegisterRequest) (*orderpb.RegisterResponse, error) {
	propData, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}
	prop := raftProposal{
		proposalType: REGISTER,
		proposalData: propData,
	}
	server.rc.proposeC <- prop
	server.viewMu.RLock()
	defer server.viewMu.RUnlock()
	return &orderpb.RegisterResponse{ViewID: server.viewID}, nil
}

func (server *orderServer) Finalize(ctx context.Context, req *orderpb.FinalizeRequest) (*orderpb.FinalizeResponse, error) {
	server.mu.Lock()
	server.finalizeShardRequests = append(server.finalizeShardRequests, req)
	server.mu.Unlock()
	return &orderpb.FinalizeResponse{}, nil
}
