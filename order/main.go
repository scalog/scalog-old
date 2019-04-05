package order

import (
	"fmt"
	"github.com/scalog/scalog/internal/pkg/golib"
	"log"
	"net"
	"sort"
	"time"

	"k8s.io/apimachinery/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/scalog/scalog/internal/pkg/kube"
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

	id, peers := getRaftIndexPeerUrls()
	// TODO: remove hard coded server shard count
	server := newOrderServer(golib.NewSet(), 2, nil, nil)
	raftProposeChannel, raftCommitChannel, raftErrorChannel, raftSnapshotter :=
		newRaftNode(id, peers, false, server.getSnapshot)
	server.raftProposeChannel = raftProposeChannel
	server.raftSnapshotter = <-raftSnapshotter

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
	if viper.GetBool("localRun") {
		// FOR TESTING. Single raft node
		return 1, []string{"http://0.0.0.0:9876"}
	}
	clientset, err := kube.InitKubernetesClient()
	if err != nil {
		logger.Panicf(err.Error())
	}

	listOptions := metav1.ListOptions{LabelSelector: "app=scalog-order"}
	pods, err := kube.GetShardPods(clientset, listOptions, viper.GetInt("raft_cluster_size"), viper.GetString("namespace"))
	if err != nil {
		logger.Panicf(err.Error())
	}

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

func listenForErrors(errorChannel <-chan error) {
	for err := range errorChannel {
		logger.Printf("Raft error: " + err.Error())
	}
}
