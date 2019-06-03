package bench

import (
	"fmt"
	"time"

	"github.com/scalog/scalog/pkg/set64"

	clientlib "github.com/scalog/scalog/client"
)

type Bench struct {
	num    int32
	size   int32
	data   string
	client *clientlib.Client
}

func NewBench(num, size int32) (*Bench, error) {
	client, err := clientlib.NewClient()
	if err != nil {
		return nil, err
	}
	b := &Bench{num: num, size: size}
	b.client = client
	b.data = string(make([]byte, size))
	return b, nil
}

func (b *Bench) Start() error {
	gsns := set64.NewSet64()
	start := time.Now()
	for i := int32(0); i < b.num; i++ {
		gsn, err := b.client.Append(b.data)
		if err != nil {
			return err
		}
		gsns.Add(int64(gsn))
	}
	end := time.Now()
	if int32(gsns.Size()) != b.num {
		return fmt.Errorf("Benchmark failed: one or more records assigned identical global sequence numbers")
	}
	elapsed := end.Sub(start)
	fmt.Printf("Benchmark completed: %d append operations of %d bytes each elapsed %d ms\n", b.num, b.size, elapsed)
	return nil
}
