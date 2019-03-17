package order

import (
	"fmt"
	"log"
	"net"
	"sort"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/scalog/scalog/logger"
	"github.com/scalog/scalog/order/messaging"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func Start() {
	logger.Printf("Ordering layer server starting on %d\n", viper.Get("port"))
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", viper.Get("port")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
		}),
	)

	var server *orderServer
	id, peers := getRaftIndexPeerUrls()
	raftProposeChannel, raftCommitChannel, raftErrorChannel, raftSnapshotter :=
		newRaftNode(id, peers, false, server.getSnapshot)
	server = newOrderServer(make([]int, 1, 1), 2, raftProposeChannel, <-raftSnapshotter)

	messaging.RegisterOrderServer(grpcServer, server)
	go server.respondToDataLayer()
	go server.listenForRaftCommits(raftCommitChannel)
	go listenForErrors(raftErrorChannel)

	logger.Printf("Order layer server available on port %d\n", viper.Get("port"))
	//Blocking, must be last step
	grpcServer.Serve(lis)
}

// combine peer Raft node ID & URL for ease of sorting by ID
type peerIDAndURL struct {
	id  types.UID
	url string
}

/**
Returns own ID (as index into array of IP addresses) and an array of IP addresses of Raft peers, sorted by increasing UID.
The IP addresses should include the port number.
*/
func getRaftIndexPeerUrls() (int, []string) {
	clientset := initKubernetesClient()
	listOptions := metav1.ListOptions{
		LabelSelector: "app=scalog-order",
	}
	pods := getShardPods(clientset, listOptions)
	size := len(pods.Items)

	peers := make([]peerIDAndURL, size, size)

	for i, pod := range pods.Items {
		logger.Printf("Peer ip: " + pod.Status.PodIP + ", uid: " + string(pod.UID))
		peers[i] = peerIDAndURL{
			id:  pod.UID,
			url: fmt.Sprintf("http://%s:%s", pod.Status.PodIP, viper.GetString("raftPort")),
		}
	}

	sort.Slice(peers, func(i int, j int) bool {
		return peers[i].id < peers[j].id
	})

	id := viper.GetString("uid")
	logger.Printf("My uid: " + id)

	index := -1
	urls := make([]string, size, size)
	for i, idURL := range peers {
		if string(idURL.id) == id {
			index = i
		}
		urls[i] = idURL.url
	}
	// ID's in raft should be one indexed
	index++
	return index, urls
}

//TODO copy pasted from data/main.go. Move into kubernetes util package
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

func getShardPods(clientset *kubernetes.Clientset, listOptions metav1.ListOptions) *v1.PodList {
	for {
		pods, err := clientset.CoreV1().Pods(viper.GetString("namespace")).List(listOptions)
		if err != nil {
			panic(err)
		}
		fmt.Println(fmt.Sprintf("looking for %d, but have %d", viper.GetInt("raft_cluster_size"), len(pods.Items)))
		if len(pods.Items) == viper.GetInt("raft_cluster_size") && allPodsAreRunning(pods.Items) {
			return pods
		}
		//wait for pods to start up
		time.Sleep(1000 * time.Millisecond)
	}
}

func allPodsAreRunning(pods []v1.Pod) bool {
	for _, pod := range pods {
		if pod.Status.Phase != v1.PodRunning {
			return false
		}
	}
	return true
}

//TODO end copy pasta

func listenForErrors(errorChannel <-chan error) {
	for err := range errorChannel {
		logger.Printf("Raft error: " + err.Error())
	}
}
