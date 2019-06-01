package data

import (
	"fmt"
	"net"

	"github.com/scalog/scalog/data/messaging"
	log "github.com/scalog/scalog/logger"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	health "google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
)

// Start serving requests for the data layer
func Start() {
	log.Printf("Data layer server %d starting on %d\n", viper.Get("id"), viper.Get("port"))
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", viper.Get("port")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	// TODO: For more fine tuned health checking, we should pass the health server into the data serve r
	// so we can update its state with more intimiate details
	healthServer := health.NewServer()
	healthServer.Resume()
	healthgrpc.RegisterHealthServer(grpcServer, healthServer)
	messaging.RegisterDataServer(grpcServer, newDataServer(viper.GetInt32("id"), viper.GetInt32("shardID"), viper.GetInt("replica_count")))
	log.Printf("Data layer server %d available on %d\n", viper.Get("id"), viper.Get("port"))
	grpcServer.Serve(lis)
}
