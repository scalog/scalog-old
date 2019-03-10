package order

import (
	"sync"

	"github.com/scalog/scalog/logger"

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

// Map<Shard ID, Channels to write responses to for each server in shard>
type ResponseChannels map[int][]chan pb.ReportResponse

type orderServer struct {
	committedGlobalCut CommittedGlobalCut
	contestedGlobalCut ContestedGlobalCut
	globalSequenceNum  int
	shardIds           []int
	numServersPerShard int
	mu                 sync.Mutex
	responseChannels   ResponseChannels
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

func initResponseChannels(shardIds []int, numServersPerShard int) ResponseChannels {
	channels := make(ResponseChannels, len(shardIds))
	for _, shardId := range shardIds {
		channelsForShard := make([]chan pb.ReportResponse, numServersPerShard, numServersPerShard)
		for i := 0; i < numServersPerShard; i++ {
			channelsForShard[i] = make(chan pb.ReportResponse)
		}
		channels[shardId] = channelsForShard
	}
	return channels
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
		response := pb.ReportResponse{
			StartGlobalSequenceNum: int32(server.globalSequenceNum),
			Offsets:                intSliceToInt32Slice(server.committedGlobalCut[shardId]),
		}

		for i := 0; i < server.numServersPerShard; i++ {
			delta := deltas[shardId][i]

			// calculate new committed cuts
			server.committedGlobalCut[shardId][i] += delta
			server.globalSequenceNum += delta
		}

		response.CommittedCuts = intSliceToInt32Slice(server.committedGlobalCut[shardId])

		for i := 0; i < server.numServersPerShard; i++ {
			server.responseChannels[shardId][i] <- response
		}
	}
}

// Reports final cuts to the data layer periodically
func (server *orderServer) reportResponseRoutine(stream pb.Order_ReportServer, req *pb.ReportRequest) {
	shardId := int(req.ShardID)
	num := req.ReplicaID
	for response := range server.responseChannels[shardId][num] {
		logger.Printf(response.String())
		stream.Send(&response)
	}
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
