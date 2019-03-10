package order

import (
	"context"
	pb "github.com/scalog/scalog/order/messaging"
	"io"
)

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

		cut := make(ShardCut, server.numServersPerShard, server.numServersPerShard)
		for i := 0; i < len(cut); i++ {
			cut[i] = int(req.TentativeCut[i])
		}
		server.mu.Lock()
		for i := 0; i < len(cut); i++ {
			curr := server.contestedGlobalCut[int(req.ShardID)][int(req.ReplicaID)][i]
			server.contestedGlobalCut[int(req.ShardID)][int(req.ReplicaID)][i] = max(curr, cut[i])
		}
		server.mu.Unlock()
	}
}

func (server *orderServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	return nil, nil
}

func (server *orderServer) Finalize(ctx context.Context, req *pb.FinalizeRequest) (*pb.FinalizeResponse, error) {
	return nil, nil
}
