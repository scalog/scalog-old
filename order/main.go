package order

import (
	"fmt"
	"github.com/scalog/scalog/logger"
	"github.com/scalog/scalog/order/messaging"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"log"
	"net"
)

func newOrderServer() *orderServer {
	s := &orderServer{
	}
	return s
}

func Start() {
	logger.Printf("Ordering layer server %d started on %d\n", viper.Get("asdf"), viper.Get("port"))
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", viper.Get("port")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	messaging.RegisterOrderServer(grpcServer, newOrderServer())
	logger.Printf("Order layer server %d available on %d\n", viper.Get("asdf"), viper.Get("port"))
	grpcServer.Serve(lis)
}
