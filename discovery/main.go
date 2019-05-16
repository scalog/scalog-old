package discovery

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/scalog/scalog/discovery/rpc"
	"github.com/scalog/scalog/internal/pkg/kube"
	log "github.com/scalog/scalog/logger"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Start serving requests for the data layer
func Start() {
	log.Printf("Discovery server starting on %d\n", viper.Get("port"))
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", viper.Get("port")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	rpc.RegisterDiscoveryServer(grpcServer, newDiscoveryServer())
	log.Printf("Discovery server available on %d\n", viper.Get("port"))
	grpcServer.Serve(lis)
}

type discoveryServer struct {
	client       *kubernetes.Clientset
	serverLabels metav1.ListOptions
}

func (ds *discoveryServer) DiscoverServers(ctx context.Context, req *rpc.DiscoverRequest) (*rpc.DiscoverResponse, error) {
	services, err := ds.client.CoreV1().Services(viper.GetString("namespace")).List(ds.serverLabels)
	if err != nil {
		log.Panicf(err.Error())
	}
	serviceIPs := make([]*rpc.DataServer, len(services.Items))
	for i, service := range services.Items {
		parts := strings.Split(service.Name, "-")
		serverID, err := strconv.ParseInt(parts[len(parts)-1], 10, 32)
		if err != nil {
			log.Printf(err.Error())
			return nil, err
		}
		shardID, err := strconv.ParseInt(parts[len(parts)-2], 10, 32)
		if err != nil {
			log.Printf(err.Error())
			return nil, err
		}
		if len(service.Spec.Ports) != 1 {
			return nil, errors.New("expected only a single port service for each data server")
		}
		serviceIPs[i] = &rpc.DataServer{
			ServerID: int32(serverID),
			ShardID:  int32(shardID),
			Port:     service.Spec.Ports[0].NodePort,
			Ip:       service.Spec.ClusterIP,
		}
	}
	resp := &rpc.DiscoverResponse{
		Servers: serviceIPs,
	}
	return resp, nil
}

func newDiscoveryServer() *discoveryServer {
	clientset := kube.InitKubernetesClient()
	listOptions := metav1.ListOptions{LabelSelector: "role=scalog-exposed-data-service"}
	ds := &discoveryServer{
		client:       clientset,
		serverLabels: listOptions,
	}
	return ds
}
