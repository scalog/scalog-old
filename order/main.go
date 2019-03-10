package order

import (
	"fmt"
	"go.etcd.io/etcd/raft/raftpb"
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
	logger.Printf("Ordering layer server %d started on %d\n", viper.Get("asdf"), viper.Get("port"))
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", viper.Get("port")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	id := 1                                              //TODO change to be unique for all ordering servers
	peers := strings.Split("http://127.0.0.1:9021", ",") //TODO change string to list of URLs
	raftProposeChannel := make(chan string)
	raftCommitChannel := make(chan *string)
	raftConfigChangeChannel := make(chan raftpb.ConfChange)

	grpcServer := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
		}),
	)
	orderServer := newOrderServer(make([]int, 1, 1), 2, raftProposeChannel, raftCommitChannel) // todo fix hard-coded parameters
	raftErrorChannel, raftSnapshotReadyChannel := newRaftNode(id, peers, false, orderServer.getSnapshot, raftProposeChannel, raftCommitChannel, raftConfigChangeChannel)

	messaging.RegisterOrderServer(grpcServer, orderServer)
	logger.Printf("Order layer server %d available on %d\n", viper.Get("asdf"), viper.Get("port"))
	grpcServer.Serve(lis)

	go orderServer.respondToDataLayer()
	go orderServer.listenForRaftCommits()
	go listenForErrors(raftErrorChannel)
}

func listenForErrors(errorChannel <-chan error) {
	for err := range errorChannel {
		logger.Printf("Raft error: " + err.Error())
	}
}
