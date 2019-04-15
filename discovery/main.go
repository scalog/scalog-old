package discovery

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/scalog/scalog/discovery/rpc"
	"github.com/scalog/scalog/internal/pkg/kube"
	"github.com/scalog/scalog/logger"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Start serving requests for the data layer
func Start() {
	logger.Printf("Discovery server starting on %d\n", viper.Get("port"))
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", viper.Get("port")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	rpc.RegisterDiscoveryServer(grpcServer, newDiscoveryServer())
	logger.Printf("Discovery server available on %d\n", viper.Get("port"))
	grpcServer.Serve(lis)
}

type discoveryServer struct {
	client       *kubernetes.Clientset
	serverLabels metav1.ListOptions
}

func (ds *discoveryServer) DiscoverServers(ctx context.Context, req *rpc.DiscoverRequest) (*rpc.DiscoverResponse, error) {
	services, err := ds.client.CoreV1().Services(viper.GetString("namespace")).List(ds.serverLabels)
	if err != nil {
		logger.Panicf(err.Error())
	}
	serviceIPs := make([]*rpc.DataServerAddress, len(services.Items))
	for i, service := range services.Items {
		if len(service.Spec.Ports) != 1 {
			return nil, errors.New("expected only a single port service for each data server")
		}
		serviceIPs[i] = &rpc.DataServerAddress{
			Port: service.Spec.Ports[0].NodePort,
			Ip:   service.Spec.ClusterIP,
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
