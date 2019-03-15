package data

import (
	"context"
	"fmt"
	"io"
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
	om "github.com/scalog/scalog/order/messaging"
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

/* Note: This is a blocking operation -- waits until all servers in a shard are up before trying to open
communication channels with them. Only then will this server become ready to take requests.
*/
func newDataServer() *dataServer {
	var shardPods []chan messaging.ReplicateRequest

	// Run this only if running within a kubernetes cluster
	if !viper.GetBool("localRun") {
		logger.Printf("Server %d searching for other %d servers in %s\n", viper.Get("id"), viper.Get("serverCount"), viper.GetString("shardGroup"))
		clientset := initKubernetesClient()
		listOptions := metav1.ListOptions{
			LabelSelector: "tier=" + viper.GetString("shardGroup"),
		}
		pods := getShardPods(clientset, listOptions)
		for _, pod := range pods.Items {
			if pod.Status.PodIP == viper.Get("pod_ip") {
				continue
			}
			shardPods = append(shardPods, make(chan messaging.ReplicateRequest))
			go createIntershardPodConnection(pod.Status.PodIP, shardPods[len(shardPods)-1])
		}
	}

	fs := filesystem.New("scalog-db")
	s := &dataServer{
		stableStorage: &fs,
		serverBuffers: make([][]Record, viper.GetInt("serverCount")),
		mu:            sync.Mutex{},
		shardServers:  shardPods,
	}
	s.sendTentativeCuts()
	return s
}

// Should form a channel which writes to a specific pod on a specific pod ip
func createIntershardPodConnection(podIP string, ch chan messaging.ReplicateRequest) {
	var opts []grpc.DialOption
	// TODO: Use a secured connection
	opts = append(opts, grpc.WithInsecure())

	logger.Printf(fmt.Sprintf("Dialing %s:%s", podIP, viper.GetString("port")))
	ticker := time.NewTicker(100 * time.Millisecond)
	connChannel := make(chan messaging.Data_ReplicateClient)

	go func() {
		for range ticker.C {
			conn, err := grpc.Dial(podIP+":"+viper.GetString("port"), opts...)
			if err != nil {
				logger.Printf(err.Error())
				continue
			}
			client := messaging.NewDataClient(conn)
			stream, err := client.Replicate(context.Background())
			if err != nil {
				logger.Printf("Failed to dial... Trying again")
				continue
			}
			connChannel <- stream
			return
		}
	}()

	stream := <-connChannel
	ticker.Stop()
	logger.Printf("Successfully set up channels with " + podIP)

	for req := range ch {
		if err := stream.Send(&req); err != nil {
			logger.Panicf(err.Error())
		}
	}
}

func (server *dataServer) sendTentativeCuts() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	logger.Printf("Utilizing DNS to dial Ordering Layer at scalog-order-service.scalog:" + viper.GetString("orderPort"))
	conn, err := grpc.Dial("dns:///scalog-order-service.scalog:"+viper.GetString("orderPort"), opts...)
	if err != nil {
		panic(err)
	}
	client := om.NewOrderClient(conn)
	stream, err := client.Report(context.Background())

	// Anonymous function for sending tentative cuts to the ordering layer
	go func() {
		ticker := time.NewTicker(8 * time.Millisecond) // todo remove hard-coded interval
		for range ticker.C {
			cut := make([]int32, viper.GetInt("serverCount"))
			for idx, buf := range server.serverBuffers {
				cut[idx] = int32(len(buf))
			}
			reportReq := &om.ReportRequest{
				ShardID:      viper.GetInt32("shardID"),
				ReplicaID:    viper.GetInt32("id"),
				TentativeCut: cut,
			}
			stream.Send(reportReq)
		}
	}()

	// Anonymous function for receiving stream inputs
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				logger.Panicf(err.Error())
			}

			if len(in.CommittedCuts) != len(server.serverBuffers) {
				logger.Panicf(
					fmt.Sprintf(
						"[Data] Received cut is of irregular length (%d vs %d)",
						len(in.CommittedCuts),
						len(server.serverBuffers),
					),
				)
			}

			gsn := int(in.StartGlobalSequenceNum)
			for idx, cut := range in.CommittedCuts {
				offset := int(in.Offsets[idx])
				numLogs := int(cut) - offset

				for i := 0; i < numLogs; i++ {
					server.serverBuffers[idx][offset+i].commitResp <- gsn
					gsn++
				}
			}
		}
	}()
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

func allPodsAreRunning(pods []v1.Pod) bool {
	for _, pod := range pods {
		if string(pod.Status.Phase) != "Running" {
			return false
		}
	}
	return true
}

func getShardPods(clientset *kubernetes.Clientset, listOptions metav1.ListOptions) *v1.PodList {
	pods, err := clientset.CoreV1().Pods(viper.GetString("namespace")).List(listOptions)
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
	return pods
}
