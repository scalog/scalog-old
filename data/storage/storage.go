package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/scalog/scalog/logger"
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
NewStorage creates a new directory at [storagePath] and returns a new instance
of storage for creating partitions and writing to them.
*/
func NewStorage(storagePath string) (*Storage, error) {
	storageErr := os.MkdirAll(storagePath, os.ModePerm)
	if storageErr != nil {
		logger.Printf(storageErr.Error())
		return nil, storageErr
	}
	s := &Storage{
		storagePath: storagePath,
		partitions:  make(map[int32]*partition),
	}
	_, partitionErr := s.AddPartition()
	if partitionErr != nil {
		logger.Printf(partitionErr.Error())
		return nil, partitionErr
	}
	return s, nil
}

/*
AddPartition adds a new partition to storage and returns the partition's id.
*/
func (s *Storage) AddPartition() (int32, error) {
	p := newPartition(s.storagePath, s.nextPartitionID)
	err := os.MkdirAll(p.partitionPath, os.ModePerm)
	if err != nil {
		return -1, err
	}
	s.partitions[p.partitionID] = p
	s.nextPartitionID++
	return p.partitionID, nil
}

/*
Write writes an entry to the default partition.
*/
func (s *Storage) Write(gsn int64, record string) error {
	err := s.writeToPartition(s.nextPartitionID-1, gsn, record)
	if err != nil {
		logger.Printf(err.Error())
		return err
	}
	return nil
}

func (s *Storage) Read(gsn int64) error {
	// TODO
	return nil
}

/*
writeToPartition writes an entry to partition with id [partitionID].
*/
func (s *Storage) writeToPartition(partitionID int32, gsn int64, record string) error {
	p, in := s.partitions[partitionID]
	if !in {
		return fmt.Errorf("Attempted to write to non-existant partition %d", partitionID)
	}
	err := p.checkActiveSegment()
	if err != nil {
		return err
	}
	return p.activeSegment.writeToSegment(gsn, record)
}

func (p *partition) checkActiveSegment() error {
	if p.activeSegment == nil || p.activeSegment.nextRelativeOffset >= p.maxSegmentSize {
		return p.addActiveSegment()
	}
	return nil
}

func (p *partition) addActiveSegment() error {
	if p.activeSegment != nil {
		p.nextBaseOffset += int64(p.activeSegment.nextRelativeOffset)
		p.activeSegment.log.Close()
		p.activeSegment.index.Close()
	}
	activeSegment, err := newSegment(p.partitionPath, p.nextBaseOffset)
	if err != nil {
		return err
	}
	p.segments[activeSegment.baseOffset] = activeSegment
	p.activeSegment = activeSegment
	return nil
}

func (s *segment) writeToSegment(gsn int64, record string) error {
	logEntry := newLogEntry(s.nextRelativeOffset, s.nextPosition, gsn, record)
	_, writeLogErr := s.logWriter.WriteString(logEntry.Stats())
	if writeLogErr != nil {
		return writeLogErr
	}
	s.logWriter.Flush()
	indexEntry := newIndexEntry(s.nextRelativeOffset, s.nextPosition)
	_, writeIndexErr := s.indexWriter.WriteString(indexEntry.Stats())
	if writeIndexErr != nil {
		return writeIndexErr
	}
	s.indexWriter.Flush()
	s.nextRelativeOffset++
	s.nextPosition += logEntry.payloadSize
	return nil
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

func newSegment(partitionPath string, baseOffset int64) (*segment, error) {
	log, logWriter, logErr := newLog(partitionPath, baseOffset)
	if logErr != nil {
		return nil, logErr
	}
	index, indexWriter, indexErr := newIndex(partitionPath, baseOffset)
	if indexErr != nil {
		return nil, indexErr
	}
	s := &segment{
		baseOffset:         baseOffset,
		nextRelativeOffset: 0,
		nextPosition:       0,
		log:                log,
		index:              index,
		logWriter:          logWriter,
		indexWriter:        indexWriter,
	}
	return s, nil
}

func newLog(partitionPath string, baseOffset int64) (*os.File, *bufio.Writer, error) {
	logName := fmt.Sprintf("%d.log", baseOffset)
	logPath := path.Join(partitionPath, logName)
	f, fileErr := os.Create(logPath)
	if fileErr != nil {
		return nil, nil, fileErr
	}
	w := bufio.NewWriter(f)
	_, writeErr := w.WriteString(fmt.Sprintf("baseoffset: %d\n", baseOffset))
	if writeErr != nil {
		return nil, nil, writeErr
	}
	return f, w, nil
}

func newIndex(partitionPath string, baseOffset int64) (*os.File, *bufio.Writer, error) {
	indexName := fmt.Sprintf("%d.index", baseOffset)
	indexPath := path.Join(partitionPath, indexName)
	f, err := os.Create(indexPath)
	if err != nil {
		return nil, nil, err
	}
	w := bufio.NewWriter(f)
	_, writeErr := w.WriteString(fmt.Sprintf("baseoffset: %d\n", baseOffset))
	if writeErr != nil {
		return nil, nil, writeErr
	}
	return f, w, nil
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
	if err != nil {
		logger.Printf(err.Error())
	}
	return string(out)
}

func (l logEntry) Stats() string {
	return fmt.Sprintf("relativeoffset: %d, position: %d, payloadsize: %d, payload: %s\n",
		l.relativeOffset, l.position, l.payloadSize, l.payload)
}

func (i indexEntry) Stats() string {
	return fmt.Sprintf("relativeoffset: %d, position: %d\n", i.relativeOffset, i.position)
}
