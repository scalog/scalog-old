package data

import (
	"fmt"
	"log"
	"net"

	"github.com/scalog/scalog/data/messaging"
	"github.com/scalog/scalog/logger"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

// Start serving requests for the data layer
func Start() {
	logger.Printf("Data layer server %d starting on %d\n", viper.Get("id"), viper.Get("port"))
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", viper.Get("port")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	messaging.RegisterDataServer(grpcServer, newDataServer())
	logger.Printf("Data layer server %d available on %d\n", viper.Get("id"), viper.Get("port"))
	grpcServer.Serve(lis)
}
