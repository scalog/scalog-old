package discovery

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/scalog/scalog/discovery/rpc"
	"github.com/scalog/scalog/internal/pkg/golib"
	"github.com/scalog/scalog/internal/pkg/kube"
	log "github.com/scalog/scalog/logger"
	om "github.com/scalog/scalog/order/messaging"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	health "google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
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

	// TODO: For more fine tuned health checking, we should pass the health server into the data serve r
	// so we can update its state with more intimiate details
	healthServer := health.NewServer()
	healthServer.Resume()
	healthgrpc.RegisterHealthServer(grpcServer, healthServer)

	rpc.RegisterDiscoveryServer(grpcServer, newDiscoveryServer())
	log.Printf("Discovery server available on %d\n", viper.Get("port"))
	grpcServer.Serve(lis)
}

type DiscoveryServer struct {
	mu     sync.RWMutex
	viewID int32
	shards map[int32]*rpc.Shard

	client       *kubernetes.Clientset
	serverLabels metav1.ListOptions
}

func (ds *DiscoveryServer) Subscribe() {
	var conn *grpc.ClientConn
	if viper.GetBool("localRun") {
		conn = golib.ConnectTo("0.0.0.0:1337")
	} else {
		conn = golib.ConnectTo("dns:///scalog-order-service.scalog:" + viper.GetString("orderPort"))
	}
	client := om.NewOrderClient(conn)
	req := &om.SubscribeMetaRequest{}
	stream, err := client.SubscribeMeta(context.Background(), req)
	if err != nil {
		log.Printf(err.Error())
	}
	for {
		meta, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Panicf(err.Error())
		}
		ds.mu.Lock()
		if meta.GetIsSnapshot() {
			ds.shards = make(map[int32]*rpc.Shard)
		}
		for _, id := range meta.GetNewShardIDs() {
			if _, ok := ds.shards[id]; ok {
				log.Printf("Trying to add existed shard: %v", id)
			} else {
				ds.shards[id] = &rpc.Shard{ShardID: id}
			}
		}
		for _, id := range meta.GetFinalizeShardIDs() {
			if _, ok := ds.shards[id]; ok {
			} else {
				log.Printf("Trying to finalize non-existed shard: %v", id)
			}
		}
		for _, id := range meta.GetRemoveShardIDs() {
			if _, ok := ds.shards[id]; ok {
				delete(ds.shards, id)
			} else {
				log.Printf("Trying to remove non-existed shard: %v", id)
			}
		}
		ds.viewID = meta.GetViewID()
		ds.mu.Unlock()
	}
}

func (ds *DiscoveryServer) DiscoverServers(ctx context.Context, req *rpc.DiscoverRequest) (*rpc.DiscoverResponse, error) {
	services, err := ds.client.CoreV1().Services(viper.GetString("namespace")).List(ds.serverLabels)
	if err != nil {
		log.Panicf(err.Error())
	}
	serversByShardID := make(map[int32][]*rpc.DataServer)
	for _, service := range services.Items {
		parts := strings.Split(service.Name, "-")
		shardID, err := strconv.ParseInt(parts[len(parts)-2], 10, 32)
		if err != nil {
			return nil, err
		}
		serverID, err := strconv.ParseInt(parts[len(parts)-1], 10, 32)
		if err != nil {
			return nil, err
		}
		if len(service.Spec.Ports) != 1 {
			return nil, errors.New("expected only a single port service for each data server")
		}
		server := &rpc.DataServer{
			ServerID: int32(serverID),
			Port:     service.Spec.Ports[0].NodePort,
			Ip:       service.Spec.ClusterIP,
		}
		serversByShardID[int32(shardID)] = append(serversByShardID[int32(shardID)], server)
	}
	shards, shardsIndex := make([]*rpc.Shard, len(serversByShardID)), 0
	for shardID, servers := range serversByShardID {
		shard := &rpc.Shard{
			ShardID: shardID,
			Servers: servers,
		}
		shards[shardsIndex] = shard
		shardsIndex++
	}
	resp := &rpc.DiscoverResponse{
		Shards: shards,
	}
	return resp, nil
}

func newDiscoveryServer() *DiscoveryServer {
	clientset := kube.InitKubernetesClient()
	listOptions := metav1.ListOptions{LabelSelector: "role=scalog-exposed-data-service"}
	ds := &DiscoveryServer{
		client:       clientset,
		serverLabels: listOptions,
	}
	return ds
}
