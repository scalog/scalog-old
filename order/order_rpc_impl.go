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

func (server *orderServer) Forward(stream pb.Order_ForwardServer) error {
	// TODO
	return nil
}

func (server *orderServer) Register(req *pb.RegisterRequest, stream pb.Order_RegisterServer) error {
	return nil
}

func (server *orderServer) Finalize(ctx context.Context, req *pb.FinalizeRequest) (*pb.FinalizeResponse, error) {
	propData, err := proto.Marshal(req)
	if err != nil {
		logger.Printf("Could not marshal finalization request message")
		return nil, err
	}
	server.finalizationResponseChannels[req.ShardID] = make(chan pb.FinalizeResponse)
	prop := raftProposal{
		proposalType: FINALIZE,
		proposalData: propData,
	}
	server.rc.proposeC <- prop
	resp := <-server.finalizationResponseChannels[req.ShardID]
	delete(server.finalizationResponseChannels, req.ShardID)
	return &resp, nil
}
