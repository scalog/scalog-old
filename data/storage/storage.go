package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path"
)

/*
Storage handles all I/O operations to disk for a data server replica.
Storage's use of partitions and segments is modeled closely after Kafka's
storage system.
*/
type Storage struct {
	// path to storage directory
	storagePath string
	// ID to be assigned to next partition added
	nextPartitionID int64
	// partitionID to partition
	partitions map[int64]*partition
}

type partition struct {
	// path to partition directory
	partitionPath string
	// ID assigned to partition
	partitionID int64
	// max number of entries for a single segment
	maxSegmentSize int64
	// offset to be assigned to next entry
	nextOffset int64
	// segment that will be written to
	activeSegment *segment
	// baseOffset to segment
	segments map[int64]*segment
}

type segment struct {
	// offset of first entry in segment
	baseOffset int64
	// log file
	log *os.File
	// index file
	index *os.File
	// writer for log file
	logWriter *bufio.Writer
	// writer for index file
	indexWriter *bufio.Writer
}

type logEntry struct {
	offset      int64
	position    int64
	payloadSize int64
	payload     string
}

type indexEntry struct {
	offset   int64
	position int64
}

type payload struct {
	gsn    int64
	record string
}

/*
NewStorage TODO
*/
func NewStorage(storagePath string) *Storage {
	err := os.MkdirAll(storagePath, os.ModePerm)
	check(err)
	s := &Storage{
		storagePath: storagePath,
		partitions:  make(map[int64]*partition),
	}
	return s
}

/*
AddPartition TODO
*/
func (s *Storage) AddPartition() int64 {
	partitionPath := path.Join(s.storagePath, fmt.Sprintf("partition%d", s.nextPartitionID))
	p := &partition{
		partitionPath:  partitionPath,
		partitionID:    s.nextPartitionID,
		maxSegmentSize: 1024,
		nextOffset:     0,
		activeSegment:  nil,
		segments:       make(map[int64]*segment),
	}
	s.partitions[p.partitionID] = p
	s.nextPartitionID++
	return p.partitionID
}

/*
Write TODO
*/
func (s *Storage) Write(gsn int64, record string) {
	if len(s.partitions) == 0 {
		panic(fmt.Sprintf("Attempted to write to storage with no partitions"))
	}
	s.WriteToPartition(0, gsn, record)
}

/*
WriteToPartition TODO
*/
func (s *Storage) WriteToPartition(partitionID int64, gsn int64, record string) {
	p, in := s.partitions[partitionID]
	if !in {
		panic(fmt.Sprintf("Attempted to write to non-existant partition %d", partitionID))
	}
	p.checkActiveSegment()
	p.writeToActiveSegment(gsn, record)
}

func (p *partition) checkActiveSegment() {
	if p.nextOffset%p.maxSegmentSize == 0 {
		if p.activeSegment != nil {
			p.activeSegment.log.Close()
			p.activeSegment.index.Close()
		}
		p.segments[p.nextOffset] = newSegment(p.partitionPath, p.nextOffset)
		p.activeSegment = p.segments[p.nextOffset]
	}
}

func (p *partition) writeToActiveSegment(gsn int64, record string) {
	// TODO
	// Write log entry
	p.activeSegment.logWriter.Flush()
	// Write index entry
	p.activeSegment.indexWriter.Flush()
	p.nextOffset++
}

func newSegment(partitionPath string, baseOffset int64) *segment {
	log := newLog(partitionPath, baseOffset)
	index := newIndex(partitionPath, baseOffset)
	s := &segment{
		baseOffset:  baseOffset,
		log:         log,
		index:       index,
		logWriter:   bufio.NewWriter(log),
		indexWriter: bufio.NewWriter(index),
	}
	return s
}

func newLog(partitionPath string, baseOffset int64) *os.File {
	logName := fmt.Sprintf("%d.log", baseOffset)
	logPath := path.Join(partitionPath, logName)
	f, err := os.Create(logPath)
	check(err)
	return f
}

func newIndex(partitionPath string, baseOffset int64) *os.File {
	indexName := fmt.Sprintf("%d.index", baseOffset)
	indexPath := path.Join(partitionPath, indexName)
	f, err := os.Create(indexPath)
	check(err)
	return f
}

func newLogEntry(offset int64, gsn int64, record string) *logEntry {
	// TODO
	return nil
}

func newIndexEntry(offset int64, position int64) *indexEntry {
	// TODO
	return nil
}

func newPayload(gsn int64, record string) string {
	payload := &payload{
		gsn:    gsn,
		record: record,
	}
	out, err := json.Marshal(payload)
	check(err)
	return string(out)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
