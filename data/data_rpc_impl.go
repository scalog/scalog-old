package data

import (
	"context"
	"errors"
	"io"

	"github.com/scalog/scalog/data/datapb"
	log "github.com/scalog/scalog/logger"
)

func (server *dataServer) Append(c context.Context, req *datapb.AppendRequest) (*datapb.AppendResponse, error) {
	log.Printf("Got append request from client %d", req.Cid)
	if server.isFinalized {
		return nil, errors.New("this shard has been finalized. No further appends will be permitted")
	}

	lsn, err := server.disk.Write(req.Record)
	if err != nil {
		return nil, err
	}
	r := &Record{
		cid:        req.Cid,
		csn:        req.Csn,
		lsn:        lsn,
		record:     req.Record,
		commitResp: make(chan int32),
	}
	server.mu.Lock()
	server.serverBuffers[server.replicaID] = append(server.serverBuffers[server.replicaID], *r)
	server.mu.Unlock()

	replicateRequest := &datapb.ReplicateRequest{
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
	server.viewMu.RLock()
	resp := &datapb.AppendResponse{
		Csn:    r.csn,
		Gsn:    gsn,
		ViewID: server.viewID,
	}
	server.viewMu.RUnlock()
	return resp, nil
}

func (server *dataServer) Replicate(stream datapb.Data_ReplicateServer) error {
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

func (server *dataServer) Subscribe(req *datapb.SubscribeRequest, stream datapb.Data_SubscribeServer) error {
	log.Printf("Received SUBSCRIBE request from client starting at GSN %d", req.SubscriptionGsn)

	clientSub := &clientSubscription{
		state:       BEHIND,
		respChan:    make(chan datapb.SubscribeResponse),
		startGsn:    req.SubscriptionGsn,
		firstNewGsn: -1,
	}
	server.newClientSubsChan <- clientSub

	// Respond to client with committed records
	for resp := range clientSub.respChan {
		if err := stream.Send(&resp); err != nil {
			log.Printf("Failed to respond to client for record with GSN [%d]", resp.Gsn)
			clientSub.state = CLOSED
			return err
		}
		log.Printf("Responded to client subscription for record with GSN [%d]", resp.Gsn)
	}

	clientSub.state = CLOSED
	return nil
}

func (server *dataServer) Trim(c context.Context, req *datapb.TrimRequest) (*datapb.TrimResponse, error) {
	log.Printf("Received TRIM request from client starting at GSN %d", req.Gsn)

	err := server.disk.Delete(int64(req.Gsn))
	if err != nil {
		log.Printf("Failed to trim starting at GSN %d", req.Gsn)
		return nil, err
	}
	server.viewMu.RLock()
	resp := &datapb.TrimResponse{ViewID: server.viewID}
	server.viewMu.RUnlock()
	return resp, nil
}

func (server *dataServer) Read(c context.Context, req *datapb.ReadRequest) (*datapb.ReadResponse, error) {
	if record, in := server.committedRecords[req.Gsn]; in {
		return &datapb.ReadResponse{Record: record}, nil
	}
	record, err := server.disk.ReadGSN(int64(req.Gsn))
	if err != nil {
		log.Printf("Failed to read record with GSN %d", req.Gsn)
		return nil, err
	}
	server.viewMu.RLock()
	resp := &datapb.ReadResponse{
		Record: record,
		ViewID: server.viewID,
	}
	server.viewMu.RUnlock()
	return resp, nil
}
