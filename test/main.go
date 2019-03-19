package test

import (
	"context"
	"fmt"
	"strconv"

	"github.com/scalog/scalog/data/filesystem"
	"github.com/scalog/scalog/data/messaging"
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

func main() {
	fmt.Println("Test started")
	// testFileSystem()
	testGRPC("0.0.0.0:21024")
}
