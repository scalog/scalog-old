package order

import (
	"fmt"
	"github.com/scalog/scalog/logger"
	"github.com/scalog/scalog/order/messaging"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"log"
	"net"
	"sync"
	"time"
)

func newOrderServer(shardIds []int, numServersPerShard int) *orderServer {
	return &orderServer{
		committedGlobalCut: initCommittedCut(shardIds, numServersPerShard),
		contestedGlobalCut: initContestedCut(shardIds, numServersPerShard),
		globalSequenceNum:  0,
		shardIds:           shardIds,
		numServersPerShard: numServersPerShard,
		mu:                 sync.Mutex{},
	}
}

func startRespondingToDataLayer(server *orderServer) {
	ticker := time.NewTicker(100 * time.Microsecond) // todo remove hard-coded interval
	for range ticker.C {
		deltas := server.mergeContestedCuts()
		server.updateCommittedCutGlobalSeqNumAndBroadcastDeltas(deltas)
	}
}

func Start() {
	logger.Printf("Ordering layer server %d started on %d\n", viper.Get("asdf"), viper.Get("port"))
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", viper.Get("port")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	orderServer := newOrderServer()
	messaging.RegisterOrderServer(grpcServer, orderServer)
	logger.Printf("Order layer server %d available on %d\n", viper.Get("asdf"), viper.Get("port"))
	grpcServer.Serve(lis)
	go startRespondingToDataLayer(orderServer)
}
