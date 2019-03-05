package order

import (
	"context"
	pb "github.com/scalog/scalog/order/messaging"
)

type orderServer struct {
}

func (s *orderServer) Report(ctx context.Context, req *pb.ReportRequest) (*pb.ReportResponse, error) {
	return nil, nil
}

func (s *orderServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	return nil, nil
}

func (s *orderServer) Finalize(ctx context.Context, req *pb.FinalizeRequest) (*pb.FinalizeResponse, error) {
	return nil, nil
}
