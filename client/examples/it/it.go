package it

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	clientlib "github.com/scalog/scalog/client"
)

type It struct {
	client *clientlib.Client
}

func NewIt() (*It, error) {
	client, err := clientlib.NewClient()
	if err != nil {
		return nil, err
	}
	it := &It{}
	it.client = client
	return it, nil
}

func (it *It) Start() error {
	regex := regexp.MustCompile(" +")
	reader := bufio.NewReader(os.Stdin)
	for {
		cmdString, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		cmdString = strings.TrimSuffix(cmdString, "\n")
		cmdString = strings.Trim(cmdString, " ")
		if cmdString == "" {
			continue
		}
		cmd := regex.Split(cmdString, -1)
		if cmd[0] == "quit" || cmd[0] == "exit" {
			break
		}
		if cmd[0] == "append" {
			if len(cmd) < 2 {
				fmt.Fprintln(os.Stderr, "Command error: missing required argument [record]")
				continue
			}
			record := strings.Join(cmd[1:], " ")
			gsn, err := it.client.Append(record)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			fmt.Fprintf(os.Stderr, "Append result: { Gsn: %d, Record: %s }\n", gsn, record)
		} else if strings.EqualFold(cmd[0], "appendToShard") {
			if len(cmd) < 2 {
				fmt.Fprintln(os.Stderr, "Command error: missing required argument [record]")
				continue
			}
			record := strings.Join(cmd[1:], " ")
			gsn, shardID, err := it.client.AppendToShard(record)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			fmt.Fprintf(os.Stderr, "AppendToShard result: { Gsn: %d, Record: %s }, shardID: %d\n", gsn, record, shardID)
		} else if cmd[0] == "subscribe" {
			if len(cmd) < 2 {
				fmt.Fprintln(os.Stderr, "Command error: missing required argument [gsn]")
				continue
			}
			gsn, err := strconv.ParseInt(cmd[1], 10, 32)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			if gsn < 1 {
				fmt.Fprintln(os.Stderr, "Command error: [gsn] must be greater than 0")
				continue
			}
			subscribeChan, err := it.client.Subscribe(int32(gsn))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			go func() {
				for committedRecord := range subscribeChan {
					fmt.Fprintf(os.Stderr, "Subscribe result: { Gsn: %d, Record: %s }\n", committedRecord.Gsn, committedRecord.Record)
				}
			}()
		} else if strings.EqualFold(cmd[0], "readRecord") {
			if len(cmd) < 3 {
				fmt.Fprintln(os.Stderr, "Command error: missing required arguments [gsn] [shardID]")
				continue
			}
			gsn, err := strconv.ParseInt(cmd[1], 10, 32)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			if gsn < 1 {
				fmt.Fprintln(os.Stderr, "Command error: [gsn] must be greater than 0")
				continue
			}
			shardID, err := strconv.ParseInt(cmd[2], 10, 32)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			if shardID < 0 {
				fmt.Fprintln(os.Stderr, "Command error: [shardID] must be greater than or equal to 0")
				continue
			}
			record, err := it.client.ReadRecord(int32(gsn), int32(shardID))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			fmt.Fprintf(os.Stderr, "ReadRecord result: { Record: %s }\n", record)
		} else if cmd[0] == "trim" {
			if len(cmd) < 2 {
				fmt.Fprintln(os.Stderr, "Command error: missing required argument [gsn]")
				continue
			}
			gsn, err := strconv.ParseInt(cmd[1], 10, 32)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			if gsn < 1 {
				fmt.Fprintln(os.Stderr, "Command error: [gsn] must be greater than 0")
				continue
			}
			err = it.client.Trim(int32(gsn))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			fmt.Fprintln(os.Stderr, "Trim result: {}")
		} else if cmd[0] == "help" {
			fmt.Fprintln(os.Stderr, "Supported commands:")
			fmt.Fprintln(os.Stderr, "    append [record]")
			fmt.Fprintln(os.Stderr, "    appendToShard [record]")
			fmt.Fprintln(os.Stderr, "    subscribe [gsn]")
			fmt.Fprintln(os.Stderr, "    readRecord [gsn] [shardID]")
			fmt.Fprintln(os.Stderr, "    trim [gsn]")
			fmt.Fprintln(os.Stderr, "    exit")
		} else {
			fmt.Fprintln(os.Stderr, "Command error: invalid command")
		}
	}
	return nil
}
