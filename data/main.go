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
	"k8s.io/client-go/util/retry"
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

/*
Creates a new data server replica

Note: This is a blocking operation -- waits until all servers in a shard are up before trying to open
communication channels with them. Only then will this server become ready to take requests.
*/
func newDataServer() *dataServer {
	replicaID := viper.GetInt("id")
	replicaCount := viper.GetInt("serverCount")
	// TODO: Refactor this out
	shardName := viper.GetString("shardGroup")
	shardID := viper.GetInt32("shardID")

	var shardPods []chan messaging.ReplicateRequest
	var clientset *kubernetes.Clientset

	// Run this only if running within a kubernetes cluster
	if !viper.GetBool("localRun") {
		logger.Printf("Server %d searching for other %d servers in %s\n", replicaID, replicaCount, shardName)
		clientset = initKubernetesClient()
		listOptions := metav1.ListOptions{LabelSelector: "tier=" + shardName}
		pods := getShardPods(clientset, listOptions)
		for _, pod := range pods.Items {
			if pod.Status.PodIP == viper.Get("pod_ip") {
				continue
			}
			shardPods = append(shardPods, make(chan messaging.ReplicateRequest))
			go createIntershardPodConnection(pod.Status.PodIP, shardPods[len(shardPods)-1])
		}
	} else {
		// TODO: Make more extensible for testing -- right now we assume only two replicas in our local cluster
		logger.Printf("Server %d searching for other %d servers on your local machine\n", replicaID, replicaCount)
		shardPods = append(shardPods, make(chan messaging.ReplicateRequest))
		if viper.GetString("port") == "8080" {
			go createIntershardPodConnection("0.0.0.0:8081", shardPods[len(shardPods)-1])
		} else {
			go createIntershardPodConnection("0.0.0.0:8080", shardPods[len(shardPods)-1])
		}
	}

	fs := filesystem.New("scalog-db")
	s := &dataServer{
		replicaID:     replicaID,
		replicaCount:  replicaCount,
		shardID:       shardID,
		isFinalized:   false,
		stableStorage: &fs,
		serverBuffers: make([][]Record, replicaCount),
		mu:            sync.Mutex{},
		shardServers:  shardPods,
		kubeClient:    clientset,
	}
	s.setupOrderLayerComunication()
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
			var conn *grpc.ClientConn
			var err error
			if viper.GetBool("localRun") {
				conn, err = grpc.Dial(podIP, opts...)
			} else {
				conn, err = grpc.Dial(podIP+":"+viper.GetString("port"), opts...)
			}
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

/*
	Periodically sends a vector composed of the length of this server's buffers to
	the ordering layer.
*/
func (server *dataServer) sendTentativeCutsToOrder(stream om.Order_ReportClient, ticker *time.Ticker) {
	// Keeps track of previous ordering layer reports
	sent := make([]int32, server.replicaCount)
	for i := range sent {
		sent[i] = 0
	}
	// ticker := time.NewTicker(20 * time.Millisecond) // todo remove hard-coded interval
	for range ticker.C {
		cut := make([]int32, server.replicaCount)
		for idx, buf := range server.serverBuffers {
			cut[idx] = int32(len(buf))
		}
		// If we have not received anything since the last order layer report, then don't
		// send anything.
		if sliceEq(cut, sent) {
			continue
		}
		reportReq := &om.ReportRequest{
			ShardID:      viper.GetInt32("shardID"),
			ReplicaID:    int32(server.replicaID),
			TentativeCut: cut,
		}
		sent = cut
		stream.Send(reportReq)
	}
}

func (server *dataServer) labelPodAsFinalized() {
	podClient := server.kubeClient.CoreV1().Pods("scalog")
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		pod, getErr := podClient.Get(viper.GetString("name"), metav1.GetOptions{})
		if getErr != nil {
			panic(fmt.Errorf("Failed to get kube pod: %v", getErr))
		}

		currentLabels := pod.GetLabels()
		currentLabels["status"] = "finalized"
		pod.SetLabels(currentLabels)

		_, updateErr := podClient.Update(pod)
		return updateErr
	})
	if retryErr != nil {
		panic(fmt.Errorf("Pod label update failed: %v", retryErr))
	}
}

/*
	Listens on the given stream for finalized cuts from the ordering layer. Finalized cuts
	are read into the server buffers where we record GSN's and respond to clients.
*/
func (server *dataServer) receiveFinalizedCuts(stream om.Order_ReportClient, sendTicker *time.Ticker) {
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

		// We should no longer receive or send requests to the ordering layer.
		// We should also label ourselves as finalized so that the discovery layer
		// does not advertise this pod again. No writes should be sent out.
		if in.Finalized {
			// Stop serving further Append requests
			server.isFinalized = true
			// Stop sending tentative cuts to the ordering layer
			sendTicker.Stop()
			// Prevent discovery layer from finding this pod ever again
			server.labelPodAsFinalized()
			// Upon returning, stop receiving cuts from the ordering layer
			return
		}

		gsn := int(in.StartGlobalSequenceNum)
		for idx, cut := range in.CommittedCuts {
			offset := int(in.Offsets[idx])
			numLogs := int(cut) - offset

			for i := 0; i < numLogs; i++ {
				server.serverBuffers[idx][offset+i].gsn = int32(gsn)
				server.stableStorage.WriteLog(gsn, server.serverBuffers[idx][offset+i].record)
				// If you were the one who received this client req, you should respond to it
				if idx == server.replicaID {
					server.serverBuffers[idx][offset+i].commitResp <- gsn
				}
				gsn++
			}
		}
	}
}

func (server *dataServer) setupOrderLayerComunication() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	var conn *grpc.ClientConn
	var err error
	if viper.GetBool("localRun") {
		logger.Printf("Utilizing TCP to dial Ordering Layer at 0.0.0.0:1337")
		conn, err = grpc.Dial("0.0.0.0:1337", opts...) // For local testing
	} else {
		logger.Printf("Utilizing DNS to dial Ordering Layer at scalog-order-service.scalog:" + viper.GetString("orderPort"))
		conn, err = grpc.Dial("dns:///scalog-order-service.scalog:"+viper.GetString("orderPort"), opts...)
	}
	if err != nil {
		panic(err)
	}
	client := om.NewOrderClient(conn)
	stream, err := client.Report(context.Background())
	if err != nil {
		panic(err)
	}

	// We need to stop both of these bois when finalized
	sendTicker := time.NewTicker(20 * time.Millisecond) // TODO: Allow interval to be configurable
	go server.sendTentativeCutsToOrder(stream, sendTicker)
	go server.receiveFinalizedCuts(stream, sendTicker)
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

/*
	Helper method for testing equalty of slices
*/
func sliceEq(a, b []int32) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
