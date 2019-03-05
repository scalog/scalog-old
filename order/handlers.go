package order

import (
	"context"
	"io"
	"sync"
	"time"

	dataPb "github.com/scalog/scalog/data/messaging"
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
	mu                 sync.Mutex
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
Compute global sequence number for each server and send update message.
NOTE: Must be called after updateCommittedCuts().
*/
func (server *orderServer) updateCommittedCutGlobalSeqNumAndBroadcastDeltas(deltas Deltas) {
	for _, shardId := range server.shardIds {
		response := dataPb.CommitRequest{
			StartGlobalSequenceNum: int32(server.globalSequenceNum),
			Offsets:                intSliceToInt32Slice(server.committedGlobalCut[shardId]),
		}

		for i := 0; i < server.numServersPerShard; i++ {
			delta := deltas[shardId][i]

			// calculate new committed cuts
			server.committedGlobalCut[shardId][i] += delta
			server.globalSequenceNum = server.globalSequenceNum + delta
		}

		response.CommittedCuts = intSliceToInt32Slice(server.committedGlobalCut[shardId])
		server.Respond(response, shardId)
	}
}

// Reports final cuts to the data layer periodically
func (server *orderServer) reportResponseRoutine(stream pb.Order_ReportServer) {
	for {
		time.Sleep(500 * time.Millisecond)
		// TODO: Send response to the data layer by calling stream.Send([obj])
	}
}

func (server *orderServer) Report(stream pb.Order_ReportServer) error {
	// Boot routine for periodically responding to data layer
	go server.reportResponseRoutine(stream)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
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
	}
}

func (server *orderServer) Respond(req dataPb.CommitRequest, shardId int) {
	for i := 0; i < server.numServersPerShard; i++ {
		//TODO send response to data layer
	}
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
func intSliceToInt32Slice(intSlice []int) []int32 {
	int32Slice := make([]int32, len(intSlice), len(intSlice))
	for idx, element := range intSlice {
		int32Slice[idx] = int32(element)
	}
	return int32Slice
}
