package data

import (
	"context"
	"errors"
	"io"

	"github.com/scalog/scalog/logger"

	"github.com/scalog/scalog/data/messaging"
	pb "github.com/scalog/scalog/data/messaging"
)

func (server *dataServer) Append(c context.Context, req *pb.AppendRequest) (*pb.AppendResponse, error) {
	logger.Printf("Got append request from client %d", req.Cid)
	if server.isFinalized {
		return nil, errors.New("this shard has been finalized. No further appends will be permitted")
	}

	server.mu.Lock()
	r := &Record{
		cid:        req.Cid,
		csn:        req.Csn,
		record:     req.Record,
		commitResp: make(chan int32),
	}
	server.serverBuffers[server.replicaID] = append(server.serverBuffers[server.replicaID], *r)
	server.mu.Unlock()

	replicateRequest := &messaging.ReplicateRequest{
		ServerID: int32(server.replicaID),
		Record:   req.Record,
	}
	for _, ch := range server.shardServers {
		ch <- *replicateRequest
	}
	gsn, ok := <-r.commitResp
	if !ok {
		return nil, errors.New("append request failed due to shard finalization. Please retry the append operation at a different shard")
	}
	return &pb.AppendResponse{Csn: r.csn, Gsn: gsn}, nil
}

func (server *dataServer) Replicate(stream pb.Data_ReplicateServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		server.mu.Lock()
		server.serverBuffers[in.ServerID] = append(server.serverBuffers[in.ServerID], Record{record: in.Record})
		server.mu.Unlock()
	}
}

func (server *dataServer) Subscribe(req *pb.SubscribeRequest, stream pb.Data_SubscribeServer) error {
	logger.Printf("Received SUBSCRIBE request from client starting at GSN %d", req.SubscriptionGsn)

	clientSubscription := &clientSubscription{
		state:       BEHIND,
		respChan:    make(chan messaging.SubscribeResponse),
		startGsn:    req.SubscriptionGsn,
		firstNewGsn: -1,
	}
	server.newClientSubsChan <- clientSubscription

	// Respond to client with committed records
	for resp := range clientSubscription.respChan {
		if err := stream.Send(&resp); err != nil {
			logger.Printf("Failed to respond to client for record with GSN [%d]", resp.Gsn)
			clientSubscription.state = CLOSED
			return err
		}
		logger.Printf("Responded to client subscription for record with GSN [%d]", resp.Gsn)
	}

	clientSubscription.state = CLOSED
	return nil
}

func (server *dataServer) Trim(context.Context, *pb.TrimRequest) (*pb.TrimResponse, error) {
	// TODO: Implement
	return nil, nil
}
