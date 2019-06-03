package test

import (
	"fmt"

	clientlib "github.com/scalog/scalog/client"
)

type Test struct {
	client *clientlib.Client
}

func NewTest() (*Test, error) {
	client, err := clientlib.NewClient()
	if err != nil {
		return nil, err
	}
	t := &Test{}
	t.client = client
	return t, err
}

func (t *Test) Start() error {
	num := 32
	minGsn := int32(-1)
	maxGsn := int32(-1)
	gsnToRecord := make(map[int32]string, num)
	gsnToShardID := make(map[int32]int32, num)
	for i := 0; i < num; i++ {
		record := fmt.Sprintf("Record %d", i)
		gsn, shardID, err := t.client.AppendToShard(record)
		if err != nil {
			return err
		}
		if _, in := gsnToRecord[gsn]; in {
			return fmt.Errorf("Identical global sequence number assigned to multiple records")
		}
		if gsn < maxGsn {
			return fmt.Errorf("Order of global sequence numbers not strictly increasing")
		}
		if i == 0 {
			minGsn = gsn
		}
		maxGsn = gsn
		gsnToRecord[gsn] = record
		gsnToShardID[gsn] = shardID
	}
	subscribeChan, err := t.client.Subscribe(minGsn)
	if err != nil {
		return err
	}
	for i := 0; i < num; i++ {
		committedRecord := <-subscribeChan
		if committedRecord.Record != gsnToRecord[committedRecord.Gsn] {
			return fmt.Errorf("Subscribe result inconsistent with append result")
		}
	}
	for gsn, shardID := range gsnToShardID {
		record, err := t.client.ReadRecord(gsn, shardID)
		if err != nil {
			return err
		}
		if record != gsnToRecord[gsn] {
			return fmt.Errorf("ReadRecord result inconsistent with append result")
		}
	}
	fmt.Println("Test completed successfully")
	return nil
}
