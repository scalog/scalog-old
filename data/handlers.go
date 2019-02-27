package data

import (
	"context"

	pb "github.com/scalog/data/messaging"
)

// We should mount some persistent volume. For now the current plan is just to start running a small sql
// database so that querying is simple.

type dataServer struct {
}

// #TODO Figure out if creating a mysql instance everywhere is efficient
func (s *dataServer) Append(context.Context, *pb.AppendRequest) (*pb.AppendResponse, error) {
	return nil, nil
}

func (s *dataServer) Replicate(stream pb.Data_ReplicateServer) error {
	return nil
}

func (s *dataServer) Commit(stream pb.Data_CommitServer) error {
	return nil
}

func (s *dataServer) Subscribe(*pb.SubscribeRequest, pb.Data_SubscribeServer) error {
	return nil
}

func (s *dataServer) Trim(context.Context, *pb.TrimRequest) (*pb.TrimResponse, error) {
	return nil, nil
}
