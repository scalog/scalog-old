package data

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/scalog/scalog/data/filesystem"
	"github.com/scalog/scalog/data/messaging"
	"github.com/scalog/scalog/logger"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

// Should form a channel which writes to a specific pod on a specific pod ip
func podConn(podIP string, ch chan messaging.ReplicateRequest) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	logger.Printf(fmt.Sprintf("Dialing %s:%s", podIP, viper.GetString("port")))
	conn, err := grpc.Dial(podIP+":"+viper.GetString("port"), opts...)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := messaging.NewDataClient(conn)
	stream, err := client.Replicate(context.Background())

	for err != nil {
		time.Sleep(100 * time.Millisecond)
		conn, err := grpc.Dial(podIP+":"+viper.GetString("port"), opts...)
		if err != nil {
			panic(err)
		}
		defer conn.Close()
		client := messaging.NewDataClient(conn)
		stream, err = client.Replicate(context.Background())
	}
	logger.Printf("Successfully set up channels with " + podIP)
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

func allPodsAreRunning(pods []v1.Pod) bool {
	for _, pod := range pods {
		if string(pod.Status.Phase) != "Running" {
			return false
		}
	}
	return true
}

/* Note: This is a blocking operation -- waits until all servers in a shard are up before trying to open
communication channels with them. Only then will this server become ready to take requests.
*/
func newDataServer() *dataServer {
	clientset := initKubernetesClient()

	logger.Printf("Waiting for " + viper.GetString("shardGroup"))
	listOptions := metav1.ListOptions{
		LabelSelector: "tier=" + viper.GetString("shardGroup"),
	}

	logger.Printf("Server %d searching for other %d servers\n", viper.Get("id"), viper.Get("serverCount"))

	var pods *v1.PodList
	var err error

	// Recover all servers within the shard
	pods, err = clientset.CoreV1().Pods(viper.GetString("namespace")).List(listOptions)
	if err != nil {
		panic(err)
	}
	for len(pods.Items) != viper.Get("serverCount") || !allPodsAreRunning(pods.Items) {
		time.Sleep(1000 * time.Millisecond)
		pods, err = clientset.CoreV1().Pods(viper.GetString("namespace")).List(listOptions)
		if err != nil {
			panic(err)
		}
	}

	// Setup pod communication channels
	var shardPods []chan messaging.ReplicateRequest

	for _, pod := range pods.Items {
		// Ignore yourself. This pod should also be present in pods.Items
		if pod.Status.PodIP == viper.Get("pod_ip") || pod.Status.PodIP == "" {
			continue
		}
		shardPods = append(shardPods, make(chan messaging.ReplicateRequest))
		go podConn(pod.Status.PodIP, shardPods[len(shardPods)-1])
	}

	fs := filesystem.New("scalog-db")
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
	logger.Printf("Data layer server %d booting on %d\n", viper.Get("id"), viper.Get("port"))
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", viper.Get("port")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	messaging.RegisterDataServer(grpcServer, newDataServer())
	logger.Printf("Data layer server %d available on %d\n", viper.Get("id"), viper.Get("port"))
	grpcServer.Serve(lis)
}
