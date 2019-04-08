package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/scalog/scalog/data/filesystem"
	"github.com/scalog/scalog/data/messaging"
	"github.com/scalog/scalog/discovery/rpc"
	"google.golang.org/grpc"
)

func testFileSystem() {
	fmt.Println("Running RecordStorage")
	rs := filesystem.New("tmp")
	for i := 0; i < 10; i++ {
		rs.WriteLog(i, "this is "+strconv.Itoa(i))
	}
	rs.Close()
}

func testGRPC(port string) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial(port, opts...)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := messaging.NewDataClient(conn)
	appendRequest := &messaging.AppendRequest{Cid: 1, Csn: 1, Record: "Hello World"}

	_, error := client.Append(context.Background(), appendRequest)
	if error != nil {
		panic(error)
	}
}

func connectToLiveCluster(address string) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure(), grpc.WithBlock())

	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		panic(err)
	}

	client := rpc.NewDiscoveryClient(conn)
	appendRequest := &rpc.DiscoverRequest{}

	resp, error := client.DiscoverServers(context.Background(), appendRequest)
	if error != nil {
		panic(error)
	}
	for _, s := range resp.ServerAddresses {
		fmt.Println(s)
		appendToShard(fmt.Sprintf("130.127.133.24:%d", s))
	}
	conn.Close()
}

func appendToShard(address string) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure(), grpc.WithBlock())

	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		panic(err)
	}

	client := messaging.NewDataClient(conn)
	appendRequest := &messaging.AppendRequest{Cid: 32412, Csn: 1, Record: "please work :'("}

	resp, error := client.Append(context.Background(), appendRequest)
	if error != nil {
		panic(error)
	}
	fmt.Printf(fmt.Sprintf("Response: gsn -> %d, csn -> %d", resp.Gsn, resp.Csn))

	conn.Close()
}

func main() {
	fmt.Println("Test started")
	// testFileSystem()
	// testGRPC("0.0.0.0:21024")
	connectToLiveCluster("130.127.133.24:32403")
}
