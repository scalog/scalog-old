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
	nextPartitionID int32
	// partitionID to partition
	partitions map[int32]*partition
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
	partitionID int32
	// max number of entries for a single segment
	maxSegmentSize int32
	// base offset to be assigned to next segment
	nextBaseOffset int64
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
	// relativeOffset to be assigned to next entry
	nextRelativeOffset int32
	// position to be assigned to next entry
	nextPosition int32
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
	relativeOffset int32
	position       int32
	payloadSize    int32
	payload        string
}

/*
indexEntry is a single entry in a segment's index file.
*/
type indexEntry struct {
	relativeOffset int32
	position       int32
}

/*
payload is the payload of a logEntry.
*/
type payload struct {
	Gsn    int64
	Record string
}

/*
NewStorage TODO
*/
func NewStorage(storagePath string) *Storage {
	err := os.MkdirAll(storagePath, os.ModePerm)
	check(err)
	s := &Storage{
		storagePath: storagePath,
		partitions:  make(map[int32]*partition),
	}
	return s
}

/*
AddPartition TODO
*/
func (s *Storage) AddPartition() int32 {
	p := newPartition(s.storagePath, s.nextPartitionID)
	err := os.MkdirAll(p.partitionPath, os.ModePerm)
	check(err)
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
func (s *Storage) WriteToPartition(partitionID int32, gsn int64, record string) {
	p, in := s.partitions[partitionID]
	if !in {
		panic(fmt.Sprintf("Attempted to write to non-existant partition %d", partitionID))
	}
	p.checkActiveSegment()
	p.activeSegment.writeToSegment(gsn, record)
}

func (p *partition) checkActiveSegment() {
	if p.activeSegment == nil || p.activeSegment.nextRelativeOffset >= p.maxSegmentSize {
		p.addActiveSegment()
	}
}

func (p *partition) addActiveSegment() {
	if p.activeSegment != nil {
		p.nextBaseOffset += int64(p.activeSegment.nextRelativeOffset)
		p.activeSegment.log.Close()
		p.activeSegment.index.Close()
	}
	activeSegment := newSegment(p.partitionPath, p.nextBaseOffset)
	p.segments[activeSegment.baseOffset] = activeSegment
	p.activeSegment = activeSegment
}

func (s *segment) writeToSegment(gsn int64, record string) {
	logEntry := newLogEntry(s.nextRelativeOffset, s.nextPosition, gsn, record)
	s.logWriter.WriteString(logEntry.String())
	s.logWriter.Flush()
	indexEntry := newIndexEntry(s.nextRelativeOffset, s.nextPosition)
	s.indexWriter.WriteString(indexEntry.String())
	s.indexWriter.Flush()
	s.nextRelativeOffset++
	s.nextPosition += logEntry.payloadSize
}

func newPartition(storagePath string, partitionID int32) *partition {
	partitionPath := path.Join(storagePath, fmt.Sprintf("partition%d", partitionID))
	p := &partition{
		partitionPath:  partitionPath,
		partitionID:    partitionID,
		maxSegmentSize: 1024,
		nextBaseOffset: 0,
		activeSegment:  nil,
		segments:       make(map[int64]*segment),
	}
	return p
}

func newSegment(partitionPath string, baseOffset int64) *segment {
	log, logWriter := newLog(partitionPath, baseOffset)
	index, indexWriter := newIndex(partitionPath, baseOffset)
	s := &segment{
		baseOffset:         baseOffset,
		nextRelativeOffset: 0,
		nextPosition:       0,
		log:                log,
		index:              index,
		logWriter:          logWriter,
		indexWriter:        indexWriter,
	}
	return s
}

func newLog(partitionPath string, baseOffset int64) (*os.File, *bufio.Writer) {
	logName := fmt.Sprintf("%d.log", baseOffset)
	logPath := path.Join(partitionPath, logName)
	f, err := os.Create(logPath)
	check(err)
	w := bufio.NewWriter(f)
	w.WriteString(fmt.Sprintf("baseoffset: %d\n", baseOffset))
	return f, w
}

func newIndex(partitionPath string, baseOffset int64) (*os.File, *bufio.Writer) {
	indexName := fmt.Sprintf("%d.index", baseOffset)
	indexPath := path.Join(partitionPath, indexName)
	f, err := os.Create(indexPath)
	check(err)
	w := bufio.NewWriter(f)
	w.WriteString(fmt.Sprintf("baseoffset: %d\n", baseOffset))
	return f, w
}

func newLogEntry(relativeOffset int32, position int32, gsn int64, record string) *logEntry {
	payload := newPayload(gsn, record)
	l := &logEntry{
		relativeOffset: relativeOffset,
		position:       position,
		payloadSize:    int32(len(payload)),
		payload:        payload,
	}
	return l
}

func newIndexEntry(relativeOffset int32, position int32) *indexEntry {
	i := &indexEntry{
		relativeOffset: relativeOffset,
		position:       position,
	}
	return i
}

func newPayload(gsn int64, record string) string {
	payload := &payload{
		Gsn:    gsn,
		Record: record,
	}
	out, err := json.Marshal(payload)
	check(err)
	return string(out)
}

func (l logEntry) String() string {
	return fmt.Sprintf("relativeoffset: %d, position: %d, payloadsize: %d, payload: %s\n",
		l.relativeOffset, l.position, l.payloadSize, l.payload)
}

func (i indexEntry) String() string {
	return fmt.Sprintf("relativeoffset: %d, position: %d\n", i.relativeOffset, i.position)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
