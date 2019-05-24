package order

import (
	"fmt"
	"net"
	"sort"
	"time"

	"github.com/scalog/scalog/internal/pkg/golib"
	"github.com/scalog/scalog/internal/pkg/kube"
	log "github.com/scalog/scalog/logger"
	"github.com/scalog/scalog/order/messaging"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	health "google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"k8s.io/apimachinery/pkg/types"
)

func Start() {
	log.Printf("Ordering layer server starting on %d\n", viper.Get("port"))
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", viper.Get("port")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
		}),
	)

	// TODO: For more fine tuned health checking, we should pass the health server into the data serve r
	// so we can update its state with more intimiate details
	healthServer := health.NewServer()
	healthServer.Resume()
	healthgrpc.RegisterHealthServer(grpcServer, healthServer)

	id, peers := getRaftIndexPeerUrls()
	// TODO: remove hard coded server shard count
	server := newOrderServer(golib.NewSet(), viper.GetInt("replica_count"))
	rc := newRaftNode(id, peers, false, server.getSnapshot)
	server.rc = rc

	messaging.RegisterOrderServer(grpcServer, server)
	go server.proposalRaftBatch()
	go server.listenForRaftCommits()

	log.Printf("Order layer server available on port %d\n", viper.Get("port"))
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

	pods := kube.GetShardPods(kube.InitKubernetesClient(), "app=scalog-order",
		viper.GetInt("raft_cluster_size"), viper.GetString("namespace"))

	size := len(pods.Items)
	peers := make([]peerIDAndURL, size)
	for i, pod := range pods.Items {
		log.Printf("Peer ip: " + pod.Status.PodIP + ", uid: " + string(pod.UID))
		peers[i] = peerIDAndURL{
			id:  pod.UID,
			url: fmt.Sprintf("http://%s:%s", pod.Status.PodIP, viper.GetString("raftPort")),
		}
	}

	sort.Slice(peers, func(i int, j int) bool {
		return peers[i].id < peers[j].id
	})

	id := viper.GetString("uid")
	log.Printf("My uid: " + id)

	index := -1
	urls := make([]string, size)
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
