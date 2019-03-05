package order

import (
	"context"
	"sync"

	pb "github.com/scalog/scalog/order/messaging"
)

// Server's view of latest cuts in all servers of the shard
type ShardCut []int

// Map<Shard ID, Shard cuts reported by all servers in shard>
type ContestedGlobalCut map[int][]ShardCut

// Map<Shard ID, Cuts in the shard>
type CommittedGlobalCut map[int]ShardCut

// Number of new logs appended within the previous time period
type Deltas CommittedGlobalCut

type orderServer struct {
	committedGlobalCut CommittedGlobalCut
	contestedGlobalCut ContestedGlobalCut
	globalSequenceNum  int
	shardIds           []int
	numServersPerShard int
	mu sync.Mutex
}

func initCommittedCut(shardIds []int, numServersPerShard int) CommittedGlobalCut {
	cut := make(CommittedGlobalCut, len(shardIds))
	for _, shardId := range shardIds {
		cut[shardId] = make(ShardCut, numServersPerShard, numServersPerShard)
	}
	return cut
}

func initContestedCut(shardIds []int, numServersPerShard int) ContestedGlobalCut {
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
	cut := make(ShardCut, server.numServersPerShard)
	for i := 0; i < len(cut); i++ {
		cut[i] = int(req.TentativeCut[i])
	}
	server.mu.Lock()
	for i := 0; i < len(cut); i++ {
		curr := server.contestedGlobalCut[int(req.ShardID)][req.ReplicaID][i]
		server.contestedGlobalCut[int(req.ShardID)][req.ReplicaID][i] = max(curr, cut[i])
	}
	server.mu.Unlock()
	// Report is not required to respond to data layer
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

func max(x int, y int) int {
	if x < y {
		return y
	}
	return x
}
func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
