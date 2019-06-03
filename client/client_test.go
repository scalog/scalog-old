package client

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	_, err := NewClient()
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func TestAppend(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf(err.Error())
	}
	gsn, err := client.Append("Hello, World!")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if gsn < 0 {
		t.Fatalf("Record assigned invalid global sequence number %d", gsn)
	}
}

func TestSubscribe(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf(err.Error())
	}
	gsn, err := client.Append("Hello, World!")
	if err != nil {
		t.Errorf(err.Error())
	}
	subscribeChan, err := client.Subscribe(gsn)
	if err != nil {
		t.Errorf(err.Error())
	}
	resp := <-subscribeChan
	if resp.Gsn != gsn {
		t.Fatalf("Expected: %d, Actual: %d", gsn, resp.Gsn)
	}
}

func TestReadRecord(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf(err.Error())
	}
	expected := "Hello, World!"
	gsn, shardID, err := client.AppendToShard(expected)
	if err != nil {
		t.Fatalf(err.Error())
	}
	actual, err := client.ReadRecord(gsn, shardID)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if expected != actual {
		t.Fatalf("Expected: %s, Actual: %s", expected, actual)
	}
}

func TestTrim(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf(err.Error())
	}
	expected := "Hello, World!"
	gsn, shardID, err := client.AppendToShard(expected)
	if err != nil {
		t.Fatalf(err.Error())
	}
	actual, err := client.ReadRecord(gsn, shardID)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if expected != actual {
		t.Fatalf("Expected: %s, Actual: %s", expected, actual)
	}
	err = client.Trim(gsn + 1)
	if err != nil {
		t.Fatalf(err.Error())
	}
}
