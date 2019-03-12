package order

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/scalog/scalog/logger"
	"github.com/scalog/scalog/order/messaging"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func Start() {
	logger.Printf("Ordering layer server started on %d\n", viper.Get("port"))
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", viper.Get("port")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	id := 1                                              //TODO change to be unique for all ordering servers
	peers := strings.Split("http://127.0.0.1:9021", ",") //TODO change string to list of URLs

	grpcServer := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
		}),
	)

	var server *orderServer
	raftProposeChannel, raftCommitChannel, raftErrorChannel, raftSnapshotter :=
		newRaftNode(id, peers, false, server.getSnapshot)
	server = newOrderServer(make([]int, 1, 1), 2, raftProposeChannel, <-raftSnapshotter)

	messaging.RegisterOrderServer(grpcServer, server)
	go server.respondToDataLayer()
	go server.listenForRaftCommits(raftCommitChannel)
	go listenForErrors(raftErrorChannel)

	logger.Printf("Order layer server %d available on %d\n", viper.Get("asdf"), viper.Get("port"))
	//Blocking, must be last step
	grpcServer.Serve(lis)
}

func listenForErrors(errorChannel <-chan error) {
	for err := range errorChannel {
		logger.Printf("Raft error: " + err.Error())
	}
}
