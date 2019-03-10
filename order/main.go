package order

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/scalog/scalog/logger"
	"github.com/scalog/scalog/order/messaging"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func newOrderServer(shardIds []int, numServersPerShard int) *orderServer {
	return &orderServer{
		committedGlobalCut: initCommittedCut(shardIds, numServersPerShard),
		contestedGlobalCut: initContestedCut(shardIds, numServersPerShard),
		globalSequenceNum:  0,
		shardIds:           shardIds,
		numServersPerShard: numServersPerShard,
		mu:                 sync.Mutex{},
		responseChannels:   initResponseChannels(shardIds, numServersPerShard),
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
	logger.Printf("Ordering layer server started on %d\n", viper.Get("port"))
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", viper.Get("port")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
		}),
	)
	orderServer := newOrderServer(make([]int, 1, 1), 2) // todo fix hard-coded parameters
	messaging.RegisterOrderServer(grpcServer, orderServer)
	go startRespondingToDataLayer(orderServer)
	logger.Printf("Order layer server available on %d\n", viper.Get("port"))
	grpcServer.Serve(lis)
}
