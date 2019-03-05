package order

import (
	"context"
	pb "github.com/scalog/scalog/order/messaging"
)

// Server's view of latest cuts in all servers of the shard
type ShardCut []int

// Map<Shard ID, Shard cuts reported by all servers in shard>
type ContestedGlobalCut map[string][]ShardCut

// Map<Shard ID, Cuts in the shard>
type CommittedGlobalCut map[string]ShardCut

type orderServer struct {
	committedGlobalCut CommittedGlobalCut
	contestedGlobalCut ContestedGlobalCut
	globalSequenceNum  int
}

func initGlobalCut(shardIds []string, numServersPerShard int) ContestedGlobalCut {
	cut := make(ContestedGlobalCut, len(shardIds))
	for _, shardId := range shardIds {
		shardCuts := make([]ShardCut, numServersPerShard, numServersPerShard)
		for i := 0; i < numServersPerShard; i++ {
			shardCuts[i] = make(ShardCut, numServersPerShard, numServersPerShard)
		}
		cut[shardId] = shardCuts
	}
	return cut
}

/**
Process new cut from data server.

cut: Server's view of the latest cut
shardId: Id of shard that server belongs to
num: Index of server in shard
*/
func (gCut ContestedGlobalCut) newShardCut(cut ShardCut, shardId string, num int) {

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
