package discovery

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/scalog/scalog/discovery/rpc"
	"github.com/scalog/scalog/logger"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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
		logger.Panicf("Error when querying kubernetes api for list of data service IPs")
	}

	serviceIPs := make([]string, len(services.Items))
	for i, service := range services.Items {
		// We assume that each one of these services only has one external IP
		var externalIPs []string
		if len(service.Spec.ExternalIPs) == 0 {
			return nil, errors.New("External data service does not have an external IP")
		}
		serviceIPs[i] = externalIPs[0]
	}

	resp := rpc.DiscoverResponse{
		ServerAddresses: serviceIPs,
	}
	return &resp, nil
}

func newDiscoveryServer() *discoveryServer {
	clientset := initKubernetesClient()
	listOptions := metav1.ListOptions{
		LabelSelector: "type=exposedDataService",
	}
	ds := &discoveryServer{
		client:       clientset,
		serverLabels: listOptions,
	}
	return ds
}

//////////////////	Utility Functions

func initKubernetesClient() *kubernetes.Clientset {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset
}
