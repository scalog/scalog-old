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

/*
partition is Scalog's unit of storage. A partition is an ordered sequence
of entries that is appended to. A partition is represented as a directory
split into segments.
*/
type partition struct {
	// path to partition directory
	partitionPath string
	// ID assigned to partition
	partitionID int64
	// max number of entries for a single segment
	maxSegmentSize int64
	// offset to be assigned to next entry
	nextOffset int64
	// position to be assigned to next entry
	nextPosition int64
	// segment that will be written to
	activeSegment *segment
	// baseOffset to segment
	segments map[int64]*segment
}

/*
segment is a continuous subsection of a partition. A segment is represented
as an index file and a log file.
*/
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

/*
logEntry is a single entry in a segment's log file.
*/
type logEntry struct {
	offset      int64
	position    int64
	payloadSize int64
	payload     string
}

/*
indexEntry is a single entry in a segment's index file.
*/
type indexEntry struct {
	offset   int64
	position int64
}

/*
payload is the payload of a logEntry.
*/
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
		nextPosition:   0,
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
		p.nextPosition = 0
	}
}

func (p *partition) writeToActiveSegment(gsn int64, record string) {
	logEntry := newLogEntry(p.nextOffset, p.nextPosition, gsn, record)
	p.activeSegment.logWriter.WriteString(logEntry.String())
	p.activeSegment.logWriter.Flush()
	indexEntry := newIndexEntry(p.nextOffset, p.nextPosition)
	p.activeSegment.indexWriter.WriteString(indexEntry.String())
	p.activeSegment.indexWriter.Flush()
	p.nextOffset++
	p.nextPosition += logEntry.payloadSize
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

func newLogEntry(offset int64, position int64, gsn int64, record string) *logEntry {
	payload := newPayload(gsn, record)
	l := &logEntry{
		offset:      offset,
		position:    position,
		payloadSize: int64(len(payload)),
		payload:     payload,
	}
	return l
}

func newIndexEntry(offset int64, position int64) *indexEntry {
	i := &indexEntry{
		offset:   offset,
		position: position,
	}
	return i
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

func (l logEntry) String() string {
	return fmt.Sprintf("offset: %d, position: %d, payloadsize: %d, payload: %s",
		l.offset, l.position, l.payloadSize, l.payload)
}

func (i indexEntry) String() string {
	return fmt.Sprintf("offset: %d, position: %d", i.offset, i.position)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
