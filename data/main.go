package data

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/scalog/data/filesystem"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/scalog/data/messaging"
	"github.com/scalog/logger"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

// Should form a channel which writes to a specific pod on a specific pod ip
func podConn(podIP string, ch chan messaging.ReplicateRequest) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial(podIP+":"+viper.GetString("port"), opts...)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := messaging.NewDataClient(conn)
	stream, err := client.Replicate(context.Background())
	if err != nil {
		panic(err)
	}

	for req := range ch {
		if err := stream.Send(&req); err != nil {
			panic(err)
		}
	}
}

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

func newDataServer() *dataServer {
	clientset := initKubernetesClient()

	listOptions := metav1.ListOptions{
		LabelSelector: viper.GetString("shard"),
	}
	// Recover all servers within the shard
	pods, err := clientset.CoreV1().Pods("").List(listOptions)
	if err != nil {
		panic(err)
	}

	// Setup pod communication channels
	var shardPods []chan messaging.ReplicateRequest
	for idx, pod := range pods.Items {
		if idx == viper.GetInt("id") {
			shardPods = append(shardPods, nil)
			continue
		}
		shardPods = append(shardPods, make(chan messaging.ReplicateRequest))
		go podConn(pod.Status.PodIP, shardPods[len(shardPods)-1])
	}

	fs := filesystem.New("tmp")
	s := &dataServer{
		stableStorage: &fs,
		serverBuffers: make([][]Record, viper.GetInt("serverCount")),
		mu:            sync.Mutex{},
		shardServers:  shardPods,
	}
	return s
}

// Start RPC server
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
