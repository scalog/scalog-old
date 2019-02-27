package data

import (
	"fmt"
	"log"
	"net"

	"github.com/scalog/data/messaging"
	"github.com/scalog/logger"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

// We should mount some persistent volume. For now the current plan is just to start running a small sql
// database so that querying is simple.

// #TODO Figure out if creating a mysql instance everywhere is efficient

func newDataServer() *dataServer {
	s := &dataServer{}
	return s
}

// Start data server
func Start() {
	logger.Printf("Data layer server %d started on %d\n", viper.Get("id"), viper.Get("port"))
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", viper.Get("port")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	messaging.RegisterDataServer(grpcServer, newDataServer())
	grpcServer.Serve(lis)
}
