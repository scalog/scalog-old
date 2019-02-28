package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/scalog/data/filesystem"
	"google.golang.org/grpc"

	"github.com/scalog/data/messaging"
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

func main() {
	fmt.Println("Test started")
	// NOTE: To test the file system, make a directory called /tmp
	// testFileSystem()
	testGRPC("0.0.0.0:21024")
}
