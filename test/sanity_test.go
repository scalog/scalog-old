package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/scalog/scalog/data/messaging"
	"google.golang.org/grpc"
)

// Returns a client for communicating with a Data node running on the specified port
func getDataClient(port string) messaging.DataClient {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial(port, opts...)
	if err != nil {
		panic(err)
	}
	client := messaging.NewDataClient(conn)
	return client
}

// Failure in the underlying channel
func checkError(err error, t *testing.T) {
	if err != nil {
		t.Errorf(err.Error())
	}
}

// Tests a put operation
func TestPut(t *testing.T) {
	// These cancel the scalog instances when this test ends
	// dataOneContext, cancel1 := context.WithCancel(context.Background())
	// defer cancel1()
	// dataTwoContext, cancel2 := context.WithCancel(context.Background())
	// defer cancel2()
	// orderContext, cancel3 := context.WithCancel(context.Background())
	// defer cancel3()

	// if err := exec.CommandContext(dataOneContext, "../../scalog", "data", "--localRun", "--port", "8081").Run(); err != nil {
	// 	t.Errorf(err.Error())
	// }
	// if err := exec.CommandContext(dataTwoContext, "../../scalog", "data", "--localRun", "--port", "8080").Run(); err != nil {
	// 	t.Errorf(err.Error())
	// }
	// if err := exec.CommandContext(orderContext, "../../scalog", "order", "--localRun", "--port", "1337").Run(); err != nil {
	// 	t.Errorf(err.Error())
	// }

	// // Wait for services to start
	// time.Sleep(100 * time.Millisecond)

	// Create two clients for communicating with the scalog data servers
	replicaOne := getDataClient("0.0.0.0:8081")
	replicaTwo := getDataClient("0.0.0.0:8080")

	appendRequest := messaging.AppendRequest{Cid: 1, Csn: 1, Record: "Hello World"}
	resp, err := replicaOne.Append(context.Background(), &appendRequest)
	checkError(err, t)

	if resp.Csn != 1 {
		t.Errorf(fmt.Sprintf("Returned client sequence number is not 1. It is %d", resp.Csn))
	}

	if resp.Gsn != 0 {
		t.Errorf(fmt.Sprintf("Returned global sequence number is not 0. It is %d", resp.Gsn))
	}

	appendRequest2 := messaging.AppendRequest{Cid: 2, Csn: 1, Record: "i love distributed systems"}
	resp2, err := replicaTwo.Append(context.Background(), &appendRequest2)
	checkError(err, t)

	if resp2.Csn != 1 {
		t.Errorf(fmt.Sprintf("Returned client sequence number is not 1. It is %d", resp2.Csn))
	}

	if resp2.Gsn != 1 {
		t.Errorf(fmt.Sprintf("Returned global sequence number is not 1. It is %d", resp2.Gsn))
	}
}

/*
TestPutStress makes many requests to a single shard from a single client. All of its requests should
be eventually served.
*/
func TestSinglePutStress(t *testing.T) {

}

/*
TestMultiPutStress makes many requests to a single shard from a single client. All of its requests should
be eventually served.
*/
func TestMultiPutStress(t *testing.T) {

}

func TestSum(t *testing.T) {
	total := 10
	if total != 10 {
		t.Errorf("Sum was incorrect, got: %d, want: %d.", total, 10)
	}
}
