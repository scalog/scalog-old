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

// Number of new logs appended within the previous time period
type Deltas CommittedGlobalCut

type orderServer struct {
	committedGlobalCut CommittedGlobalCut
	contestedGlobalCut ContestedGlobalCut
	globalSequenceNum  int
	shardIds           []string
	numServersPerShard int
}

func initCommittedCut(shardIds []string, numServersPerShard int) CommittedGlobalCut {
	cut := make(CommittedGlobalCut, len(shardIds))
	for _, shardId := range shardIds {
		cut[shardId] = make(ShardCut, numServersPerShard, numServersPerShard)
	}
	return cut
}

func initContestedCut(shardIds []string, numServersPerShard int) ContestedGlobalCut {
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
Find min cut for each shard and compute changes from last committed cuts.
*/
func (server *orderServer) mergeContestedCuts() Deltas {
	deltas := initCommittedCut(server.shardIds, server.numServersPerShard)

	for _, shardId := range server.shardIds {
		// find smallest cut for shard i
		for i := 0; i < server.numServersPerShard; i++ {
			minCut := 0
			for j := 0; j < server.numServersPerShard; j++ {
				minCut = min(server.contestedGlobalCut[shardId][j][i], minCut)
			}

			prevCut := server.committedGlobalCut[shardId][i]
			deltas[shardId][i] = minCut - prevCut
		}
	}

	return Deltas(deltas)
}

/**
Use deltas to increment committed cuts.
*/
func (server *orderServer) updateCommittedCuts(deltas Deltas) {
	for _, shardId := range server.shardIds {
		for i := 0; i < server.numServersPerShard; i++ {
			server.committedGlobalCut[shardId][i] += deltas[shardId][i]
		}
	}
}

/**
Compute global sequence number for each server and send update message.
*/
func (server *orderServer) updateGlobalSeqNumAndBroadcastDeltas(deltas Deltas) {
	for _, shardId := range server.shardIds {
		for i := 0; i < server.numServersPerShard; i++ {
			delta := deltas[shardId][i]

			prevSequenceNum := server.globalSequenceNum
			newSequenceNum := prevSequenceNum + delta

			//server.Respond() TODO
			server.globalSequenceNum = newSequenceNum
		}
	}
}

func (server *orderServer) Report(ctx context.Context, req *pb.ReportRequest) (*pb.ReportResponse, error) {
	return nil, nil
}

func (server *orderServer) Respond(ctx context.Context) error {
	return nil
}

func (server *orderServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	return nil, nil
}

func (server *orderServer) Finalize(ctx context.Context, req *pb.FinalizeRequest) (*pb.FinalizeResponse, error) {
	return nil, nil
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
