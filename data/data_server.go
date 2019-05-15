package data

import (
	"context"
	"fmt"
	"io"
	"sort"
	"sync"
	"time"

	"github.com/scalog/scalog/internal/pkg/golib"
	"github.com/scalog/scalog/internal/pkg/kube"

	"github.com/scalog/scalog/order"

	"github.com/scalog/scalog/data/messaging"
	"github.com/scalog/scalog/data/storage"
	"github.com/scalog/scalog/logger"
	om "github.com/scalog/scalog/order/messaging"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

// Record is an internal representation of a log record
type Record struct {
	cid        int32
	csn        int32
	lsn        int64
	gsn        int32
	record     string
	commitResp chan int32 // Should send back a GSN to be forwarded to client
}

// Server's view of latest cuts in all servers of the shard
type ShardCut []int

type clientSubscriptionState int

const (
	// BEHIND indicates the client has not received a response for a record that was committed after
	// it began subscribing
	BEHIND clientSubscriptionState = 0
	// UPDATED indicates the client has received a response for a record that was committed after it
	// began subscribing
	UPDATED clientSubscriptionState = 1
	// CLOSED indiciates the gRPC connection has been closed
	CLOSED clientSubscriptionState = 2
)

type clientSubscription struct {
	// Indicates whether the client has received responses for past committed records or whether the
	// gRPC connection has been closed
	state clientSubscriptionState
	// Channel on which to send [SubscribeResponse]s to be forwarded to the client
	respChan chan messaging.SubscribeResponse
	// Global sequence number that client started subscribing from
	startGsn int32
	// The global sequence number of the first recorded that was committed after the client began
	// subscribing and has responded to the client
	firstNewGsn int32
}

type dataServer struct {
	// ID of this data server within a shard
	replicaID int32
	// Number of servers within a data shard
	replicaCount int
	// Shard ID
	shardID int32
	// lastComittedCut is the last batch recieved from the ordering layer
	lastCommittedCut order.CommittedCut
	// True if this replica has been finalized
	isFinalized bool
	// Stable storage for entries into this shard
	disk *storage.Storage
	// Main storage stacks for incoming records
	serverBuffers [][]Record
	// Protects all internal structures
	mu sync.RWMutex
	// Live connections to other servers inside the shard -- used for replication
	shardServers []chan messaging.ReplicateRequest
	// Client for API calls with kubernetes
	kubeClient *kubernetes.Clientset
	// Clients that have subscribed to this data server
	clientSubs []*clientSubscription
	// Mutex for reading and writing to clientSubs
	clientSubsMutex sync.RWMutex
	// Channel for clients that have newly subscribed to this data server
	newClientSubsChan chan *clientSubscription
	// Channel for sending committed gsn to be forwarded to subscribed clients
	clientSubsResponseChan chan int32
	// Map from gsn to committed record
	committedRecords map[int32]string
}

/*
Creates a new data server replica

Note: This is a blocking operation -- waits until all servers in a shard are up before trying to open
communication channels with them. Only then will this server become ready to take requests.
*/
func newDataServer() *dataServer {
	replicaID := viper.GetInt32("id")
	replicaCount := viper.GetInt("replica_count")
	// TODO: Refactor this out
	shardName := viper.GetString("shardGroup")
	shardID := viper.GetInt32("shardID")

	var shardPods []chan messaging.ReplicateRequest
	var clientset *kubernetes.Clientset

	// Run this only if running within a kubernetes cluster
	if !viper.GetBool("localRun") {
		logger.Printf("Server %d searching for other %d servers in %s\n", replicaID, replicaCount, shardName)
		clientset = kube.InitKubernetesClient()
		pods := kube.GetShardPods(clientset, "tier="+shardName, replicaCount, viper.GetString("namespace"))

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

	disk, err := storage.NewStorage("disk")
	if err != nil {
		// TODO handle error
		logger.Printf("Failed to initialize storage: " + err.Error())
	}
	s := &dataServer{
		replicaID:        replicaID,
		replicaCount:     replicaCount,
		shardID:          shardID,
		lastCommittedCut: make(order.CommittedCut),
		isFinalized:      false,
		disk:             disk,
		serverBuffers:    make([][]Record, replicaCount),
		mu:               sync.RWMutex{},
		shardServers:     shardPods,
		kubeClient:       clientset,
	}
	s.setupOrderLayerComunication()
	go s.respondToClientSubs()
	go s.handleNewClientSubs()
	return s
}

/*
	Periodically sends a vector composed of the length of this server's buffers to
	the ordering layer. Only sends if any new messages were received by this shard
*/
func (server *dataServer) sendTentativeCutsToOrder(stream om.Order_ReportClient, ticker *time.Ticker) {
	// Keeps track of previous ordering layer reports
	sent := make([]int32, server.replicaCount)
	for i := range sent {
		sent[i] = 0
	}
	for range ticker.C {
		cut := make([]int32, server.replicaCount)
		server.mu.RLock()
		for idx, buf := range server.serverBuffers {
			cut[idx] = int32(len(buf))
		}
		server.mu.RUnlock()
		// If we have not received anything since the last order layer report, then don't
		// send anything.
		if golib.SliceEq(cut, sent) {
			continue
		}

		reportReq := &om.ReportRequest{
			Shards: map[int32]*om.ShardView{
				server.shardID: {
					Replicas: map[int32]*om.Cut{
						server.replicaID: {
							Cut: cut,
						},
					},
				},
			},
			Finalize: nil,
			Batch:    false,
		}
		sent = cut
		stream.Send(reportReq)
	}
}

/*
	labelPodsAsFinalized adds a status:"finalized" label to the kubernetes pod object
*/
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
	Closes all channels to all records managed by this replica. If a client was still
	waiting for a request, then the shutdown of the channels will notify them that their
	append request failed.
*/
func (server *dataServer) notifyAllWaitingClients() {
	for _, record := range server.serverBuffers[server.replicaID] {
		close(record.commitResp)
	}
}

/*
	checkIfFinalized returns true if server is inside of [finalized]
*/
func (server *dataServer) checkIfFinalized(finalized []int32) bool {
	for _, sid := range finalized {
		if server.shardID == sid {
			return true
		}
	}
	return false
}

/*
	Listens on the given stream for finalized cuts from the ordering layer.

	Two actions occur upon receipt of a ordering layer report: 1) We update our cut
	2) We are finalized.
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

		if server.checkIfFinalized(in.Finalize) {
			// Stop serving further Append requests
			server.isFinalized = true
			// Stop sending tentative cuts to the ordering layer
			sendTicker.Stop()
			// Prevent discovery layer from finding this pod ever again
			server.labelPodAsFinalized()
			// Upon returning, stop receiving cuts from the ordering layer
			return
		}

		// So we have to do this in sorted order -- sort the keys and then
		// lexicographically start incrementing the gsn until we find this shard
		shardIDs := make([]int, len(in.CommitedCuts))
		i := 0
		for sid := range in.CommitedCuts {
			shardIDs[i] = int(sid)
			i++
		}
		sort.Ints(shardIDs)

		// cutGSN starts assigning GSN based on the index indicated by the fault tolerant ordering layer
		cutGSN := in.StartGSN
		for _, shardID := range shardIDs {
			currentShardCut := in.CommitedCuts[int32(shardID)].Cut
			if int(server.shardID) == shardID {
				// Last committed cut for this server
				var lastSequencedCut []int32
				// It's possible that this is the first time we are seeing this registered shard
				// come back in a cut.
				lastSequencedCutWrapper, isBound := server.lastCommittedCut[int32(shardID)]
				if isBound {
					lastSequencedCut = lastSequencedCutWrapper.Cut
				} else {
					lastSequencedCut = nil
				}
				for idx, r := range currentShardCut {
					var offset int32
					// If this is the first time we have seen this shard, initialize the previously
					// seen committed cut to be the zero vector
					if lastSequencedCut != nil {
						offset = lastSequencedCut[idx]
					} else {
						offset = 0
					}
					numLogs := int(r - offset)
					for i := 0; i < numLogs; i++ {
						server.mu.Lock()
						err := server.disk.Commit(server.serverBuffers[idx][int(offset)+i].lsn, int64(cutGSN))
						if err != nil {
							// TODO handle error
							logger.Printf(err.Error())
						}
						server.serverBuffers[idx][int(offset)+i].gsn = cutGSN
						// If you were the one who received this client req, you should respond to it
						if idx == int(server.replicaID) {
							server.serverBuffers[idx][int(offset)+i].commitResp <- cutGSN
						}
						server.mu.Unlock()
						cutGSN++
					}
				}
				continue
			}
			// Update the gsn
			for i, r := range currentShardCut {
				diff := r - server.lastCommittedCut[int32(shardID)].Cut[i]
				cutGSN += diff
			}
		}

		// update stored cut & log number
		server.lastCommittedCut = in.CommitedCuts
	}
}

/*
	Listens for new client subscriptions and assynchronously responds to the client with past committed
	records.
*/
func (server *dataServer) handleNewClientSubs() {
	for clientSub := range server.newClientSubsChan {
		server.clientSubsMutex.Lock()
		server.clientSubs = append(server.clientSubs, clientSub)
		server.clientSubsMutex.Unlock()
		go server.updateBehindClientSub(clientSub)
	}
}

/*
	Responds to a client with past committed records.
*/
func (server *dataServer) updateBehindClientSub(clientSub *clientSubscription) {
	for currGsn := clientSub.startGsn; ; currGsn++ {
		if clientSub.state == CLOSED || (clientSub.state == UPDATED && currGsn >= clientSub.firstNewGsn) {
			return
		}
		record, in := server.committedRecords[currGsn]
		if !in {
			break
		}
		resp := messaging.SubscribeResponse{
			Gsn:    currGsn,
			Record: record,
		}
		clientSub.respChan <- resp
	}
	clientSub.state = UPDATED
}

/*
	Listens for newly committed records and assynchronously responds to all active clients.
*/
func (server *dataServer) respondToClientSubs() {
	for gsn := range server.clientSubsResponseChan {
		server.clientSubsMutex.Lock()
		for _, clientSub := range server.clientSubs {
			if clientSub.state != CLOSED {
				go server.respondToClientSub(clientSub, gsn)
			}
		}
		server.clientSubsMutex.Unlock()
	}
}

/*
	Responds to a client with a newly committed record.
*/
func (server *dataServer) respondToClientSub(clientSub *clientSubscription, gsn int32) {
	if clientSub.state == BEHIND {
		clientSub.firstNewGsn = gsn
		clientSub.state = UPDATED
	}
	record := server.committedRecords[gsn]
	resp := messaging.SubscribeResponse{
		Gsn:    gsn,
		Record: record,
	}
	clientSub.respChan <- resp
}

func (server *dataServer) setupOrderLayerComunication() {
	var conn *grpc.ClientConn
	if viper.GetBool("localRun") {
		conn = golib.ConnectTo("0.0.0.0:1337")
	} else {
		conn = golib.ConnectTo("dns:///scalog-order-service.scalog:" + viper.GetString("orderPort"))
	}

	client := om.NewOrderClient(conn)
	stream, err := client.Report(context.Background())
	if err != nil {
		panic(err)
	}

	interval := time.Duration(viper.GetInt("batch_interval"))
	sendTicker := time.NewTicker(interval * time.Millisecond)
	go server.sendTentativeCutsToOrder(stream, sendTicker)
	go server.receiveFinalizedCuts(stream, sendTicker)
}

// Forms a channel which writes to a specific pod on a specific pod ip
func createIntershardPodConnection(podIP string, ch chan messaging.ReplicateRequest) {
	var conn *grpc.ClientConn
	if viper.GetBool("localRun") {
		conn = golib.ConnectTo(podIP)
	} else {
		conn = golib.ConnectTo(podIP + ":" + viper.GetString("port"))
	}

	client := messaging.NewDataClient(conn)
	stream, err := client.Replicate(context.Background())
	if err != nil {
		logger.Panicf("Couldn't create a conection with server replica")
	}
	logger.Printf("Successfully set up channels with " + podIP)

	for req := range ch {
		if err := stream.Send(&req); err != nil {
			logger.Panicf(err.Error())
		}
	}
}
