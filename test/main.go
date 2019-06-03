package test

import (
	"context"
	"fmt"
	"strconv"

	"github.com/scalog/scalog/data/datapb"
	"github.com/scalog/scalog/data/storage"
	"github.com/scalog/scalog/discovery/discpb"
	"google.golang.org/grpc"
)

func testStorage() {
	fmt.Println("Running RecordStorage")
	rs, err := storage.NewStorage("tmp")
	if err != nil {
		panic(err)
	}
	for i := 0; i < 10; i++ {
		rs.Write("this is " + strconv.Itoa(i))
	}
	rs.Destroy()
}

func testGRPC(port string) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial(port, opts...)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := datapb.NewDataClient(conn)
	appendRequest := &datapb.AppendRequest{Cid: 1, Csn: 1, Record: "Hello World"}

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

	client := discpb.NewDiscoveryClient(conn)
	appendRequest := &discpb.DiscoverRequest{}

	resp, error := client.DiscoverServers(context.Background(), appendRequest)
	if error != nil {
		panic(error)
	}
	for _, s := range resp.Shards {
		// Append to first replica within a shard
		appendToShard(fmt.Sprintf("130.127.133.35:%d", s.Servers[0].Port))
		break
	}
	conn.Close()
}

func appendToShard(address string) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure(), grpc.WithBlock())

	fmt.Println("Dialing..." + address)
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		panic(err)
	}
	fmt.Println("Finished dialing.")

	client := datapb.NewDataClient(conn)

	for i := 1; i < 100; i++ {
		appendRequest := &datapb.AppendRequest{Cid: 53325, Csn: int32(i), Record: "please work :'("}
		resp, error := client.Append(context.Background(), appendRequest)
		if error != nil {
			panic(error)
		}
		fmt.Printf(fmt.Sprintf("Response: gsn -> %d, csn -> %d", resp.Gsn, resp.Csn))
	}

	conn.Close()
}

func main() {
	fmt.Println("Test started")
	// testStorage()
	// testGRPC("0.0.0.0:21024")
	connectToLiveCluster("130.127.133.35:30823")
}
