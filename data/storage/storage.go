package storage

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	// next local sequence number to be assigned to record
	nextLSN int64
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
	// segment that will be written to
	activeSegment *segment
	// baseOffset to segment
	segments map[int64]*segment
	// global index file that will be written to
	activeGlobalIndex *globalIndex
	// first gsn in global index file to global index file
	globalIndexes map[int64]*globalIndex
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
	// local index file
	localIndex *os.File
	// writer for log file
	logWriter *bufio.Writer
	// writer for local index file
	localIndexWriter *bufio.Writer
}

/*
globalIndex is an index mapping an entry's global sequence number to its
position in a segment's log file.
*/
type globalIndex struct {
	// first gsn in global index file
	startGsn          int64
	size              int32
	globalIndex       *os.File
	globalIndexWriter *bufio.Writer
}

/*
logEntry is a single entry in a segment's log file.
*/
type logEntry struct {
	Length int32
	Record string
}

const logSuffix = ".log"
const localIndexSuffix = ".local"
const globalIndexSuffix = ".global"

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
		storagePath:     storagePath,
		nextLSN:         0,
		nextPartitionID: 0,
		partitions:      make(map[int32]*partition),
	}
	_, partitionErr := s.addPartition()
	if partitionErr != nil {
		logger.Printf(partitionErr.Error())
		return nil, partitionErr
	}
	return s, nil
}

/*
Write writes an entry to the default partition and returns the local sequence number.
*/
func (s *Storage) Write(record string) (int64, error) {
	lsn := s.nextLSN
	err := s.writeToPartition(s.nextPartitionID-1, lsn, record)
	if err != nil {
		logger.Printf(err.Error())
		return -1, err
	}
	s.nextLSN++
	return lsn, err
}

/*
Read reads an entry from the default partition.
*/
func (s *Storage) Read(lsn int64) (string, error) {
	record, err := s.readFromPartition(s.nextPartitionID-1, lsn)
	if err != nil {
		logger.Printf(err.Error())
		return "", err
	}
	return record, nil
}

/*
Commit writes an entry with local sequence number [lsn] and global sequence number
[gsn] to the appropriate global index file.
*/
func (s *Storage) Commit(lsn int64, gsn int64) error {
	err := s.commitToPartition(s.nextPartitionID-1, lsn, gsn)
	if err != nil {
		logger.Printf(err.Error())
		return err
	}
	return nil
}

/*
Sync commits the storage's in-memory copy of recently written files to disk.
*/
func (s *Storage) Sync() error {
	for _, p := range s.partitions {
		if p.activeSegment != nil {
			segmentErr := p.activeSegment.syncSegment()
			if segmentErr != nil {
				logger.Printf(segmentErr.Error())
				return segmentErr
			}
		}
		if p.activeGlobalIndex != nil {
			globalIndexErr := p.activeGlobalIndex.syncGlobalIndex()
			if globalIndexErr != nil {
				logger.Printf(globalIndexErr.Error())
				return globalIndexErr
			}
		}
	}
	return nil
}

/*
addPartition adds a new partition to storage and returns the partition's id.
*/
func (s *Storage) addPartition() (int32, error) {
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
writeToPartition writes an entry to partition with id [partitionID].
*/
func (s *Storage) writeToPartition(partitionID int32, lsn int64, record string) error {
	p, in := s.partitions[partitionID]
	if !in {
		return fmt.Errorf("Attempted to write to non-existant partition %d", partitionID)
	}
	return p.writeToActiveSegment(lsn, record)
}

func (p *partition) writeToActiveSegment(lsn int64, record string) error {
	if p.activeSegment == nil || p.activeSegment.nextRelativeOffset >= p.maxSegmentSize ||
		lsn > p.activeSegment.baseOffset+int64(p.maxSegmentSize) {
		err := p.addActiveSegment(lsn)
		if err != nil {
			return err
		}
	}
	return p.activeSegment.writeToSegment(lsn, record)
}

func (p *partition) addActiveSegment(lsn int64) error {
	if p.activeSegment != nil {
		err := p.activeSegment.finalizeSegment()
		if err != nil {
			return err
		}

	}
	activeSegment, err := newSegment(p.partitionPath, lsn)
	if err != nil {
		return err
	}
	p.segments[activeSegment.baseOffset] = activeSegment
	p.activeSegment = activeSegment
	return nil
}

func (s *segment) writeToSegment(lsn int64, record string) error {
	logEntry := newLogEntry(record)
	bytesWritten, writeLogErr := s.logWriter.WriteString(logEntry + "\n")
	if writeLogErr != nil {
		return writeLogErr
	}
	buffer := make([]byte, 8)
	binary.LittleEndian.PutUint32(buffer[0:], uint32(s.nextRelativeOffset))
	binary.LittleEndian.PutUint32(buffer[4:], uint32(s.nextPosition))
	for _, b := range buffer {
		writeIndexErr := s.localIndexWriter.WriteByte(b)
		if writeIndexErr != nil {
			return writeIndexErr
		}
	}
	s.nextRelativeOffset++
	s.nextPosition += int32(bytesWritten)
	return nil
}

func (s *Storage) readFromPartition(partitionID int32, lsn int64) (string, error) {
	p, in := s.partitions[partitionID]
	if !in {
		return "", fmt.Errorf("Attempted to read from non-existant partition %d", partitionID)
	}
	segment, err := p.getSegmentContainingLSN(lsn)
	if err != nil {
		return "", err
	}
	return segment.readFromSegment(int32(lsn - segment.baseOffset))
}

func (p *partition) getSegmentContainingLSN(lsn int64) (*segment, error) {
	for baseOffset := range p.segments {
		if lsn >= baseOffset && lsn < baseOffset+int64(p.maxSegmentSize) {
			return p.segments[baseOffset], nil
		}
	}
	return nil, fmt.Errorf("Failed to find segment containing entry with lsn %d", lsn)
}

func (s *segment) readFromSegment(relativeOffset int32) (string, error) {
	position, indexErr := getPositionOfRelativeOffset(s.localIndex.Name(), relativeOffset)
	if indexErr != nil {
		return "", indexErr
	}
	record, logErr := getRecordAtPosition(s.log.Name(), position)
	if logErr != nil {
		return "", logErr
	}
	return record, nil
}

func getPositionOfRelativeOffset(indexPath string, relativeOffset int32) (int32, error) {
	buffer, err := ioutil.ReadFile(indexPath)
	if err != nil {
		return -1, err
	}
	left := 0
	right := len(buffer) / 8
	for left < right {
		target := left + ((right - left) / 2)
		targetOffset := int32(binary.LittleEndian.Uint32(buffer[target*8:]))

		if relativeOffset == targetOffset {
			return int32(binary.LittleEndian.Uint32(buffer[target*8+4:])), nil
		} else if relativeOffset > targetOffset {
			left = target
		} else {
			right = target
		}
	}
	return -1, fmt.Errorf("Failed to find entry with relative offset %d", relativeOffset)
}

func getRecordAtPosition(logPath string, position int32) (string, error) {
	log, osErr := os.Open(logPath)
	if osErr != nil {
		return "", osErr
	}
	logReader := bufio.NewReader(log)
	_, discardErr := logReader.Discard(int(position))
	if discardErr != nil {
		return "", discardErr
	}
	line, _, readErr := logReader.ReadLine()
	if readErr != nil {
		return "", readErr
	}
	logEntry := logEntry{}
	jsonErr := json.Unmarshal(line, &logEntry)
	if jsonErr != nil {
		return "", jsonErr
	}
	return logEntry.Record, nil
}

func (s *Storage) commitToPartition(partitionID int32, lsn int64, gsn int64) error {
	p, in := s.partitions[partitionID]
	if !in {
		return fmt.Errorf("Attempted to commit to non-existant partition %d", partitionID)
	}
	return p.commitToActiveGlobalIndex(lsn, gsn)
}

func (p *partition) commitToActiveGlobalIndex(lsn int64, gsn int64) error {
	if p.activeGlobalIndex == nil || p.activeGlobalIndex.size >= p.maxSegmentSize ||
		gsn > p.activeGlobalIndex.startGsn+int64(p.maxSegmentSize) {
		err := p.addActiveGlobalIndex(gsn)
		if err != nil {
			return err
		}
	}
	segment, logErr := p.getSegmentContainingLSN(lsn)
	if logErr != nil {
		return logErr
	}
	position, indexErr := getPositionOfRelativeOffset(segment.localIndex.Name(), int32(lsn-segment.baseOffset))
	if indexErr != nil {
		return indexErr
	}
	return p.activeGlobalIndex.commitToGlobalIndex(gsn, segment.baseOffset, position)
}

func (p *partition) addActiveGlobalIndex(gsn int64) error {
	if p.activeGlobalIndex != nil {
		err := p.activeGlobalIndex.finalizeGlobalIndex()
		if err != nil {
			return err
		}
	}
	activeGlobalIndex, err := newGlobalIndex(p.partitionPath, gsn)
	if err != nil {
		return err
	}
	p.globalIndexes[activeGlobalIndex.startGsn] = activeGlobalIndex
	p.activeGlobalIndex = activeGlobalIndex
	return nil
}

func (g *globalIndex) commitToGlobalIndex(gsn int64, baseOffset int64, position int32) error {
	buffer := make([]byte, 20)
	binary.LittleEndian.PutUint64(buffer[0:], uint64(gsn))
	binary.LittleEndian.PutUint64(buffer[8:], uint64(baseOffset))
	binary.LittleEndian.PutUint32(buffer[16:], uint32(position))
	for _, b := range buffer {
		writeIndexErr := g.globalIndexWriter.WriteByte(b)
		if writeIndexErr != nil {
			return writeIndexErr
		}
	}
	g.size++
	return nil
}

func (g *globalIndex) finalizeGlobalIndex() error {
	syncErr := g.syncGlobalIndex()
	if syncErr != nil {
		return syncErr
	}
	closeErr := g.globalIndex.Close()
	if closeErr != nil {
		return closeErr
	}
	return nil
}

func (g *globalIndex) syncGlobalIndex() error {
	flushErr := g.globalIndexWriter.Flush()
	if flushErr != nil {
		logger.Printf(flushErr.Error())
		return flushErr
	}
	syncErr := g.globalIndex.Sync()
	if syncErr != nil {
		logger.Printf(syncErr.Error())
		return syncErr
	}
	return nil
}

func (s *segment) finalizeSegment() error {
	syncErr := s.syncSegment()
	if syncErr != nil {
		return syncErr
	}
	closeLogErr := s.log.Close()
	if closeLogErr != nil {
		return closeLogErr
	}
	closeIndexErr := s.localIndex.Close()
	if closeIndexErr != nil {
		return closeIndexErr
	}
	return nil
}

func (s *segment) syncSegment() error {
	flushLogErr := s.logWriter.Flush()
	if flushLogErr != nil {
		return flushLogErr
	}
	syncLogErr := s.log.Sync()
	if syncLogErr != nil {
		return syncLogErr
	}
	flushIndexErr := s.localIndexWriter.Flush()
	if flushIndexErr != nil {
		return flushIndexErr
	}
	syncIndexErr := s.localIndex.Sync()
	if syncIndexErr != nil {
		return syncIndexErr
	}
	return nil
}

func newPartition(storagePath string, partitionID int32) *partition {
	partitionPath := path.Join(storagePath, fmt.Sprintf("partition%d", partitionID))
	p := &partition{
		partitionPath:     partitionPath,
		partitionID:       partitionID,
		maxSegmentSize:    1024,
		activeSegment:     nil,
		segments:          make(map[int64]*segment),
		activeGlobalIndex: nil,
		globalIndexes:     make(map[int64]*globalIndex),
	}
	return p
}

func newGlobalIndex(partitionPath string, startGsn int64) (*globalIndex, error) {
	globalIndexName := getGlobalIndexName(startGsn)
	globalIndexPath := path.Join(partitionPath, globalIndexName)
	f, err := os.Create(globalIndexPath)
	if err != nil {
		return nil, err
	}
	g := &globalIndex{
		startGsn:          startGsn,
		size:              0,
		globalIndex:       f,
		globalIndexWriter: bufio.NewWriter(f),
	}
	return g, nil
}

func newSegment(partitionPath string, baseOffset int64) (*segment, error) {
	log, logWriter, logErr := newLog(partitionPath, baseOffset)
	if logErr != nil {
		return nil, logErr
	}
	index, indexWriter, indexErr := newLocalIndex(partitionPath, baseOffset)
	if indexErr != nil {
		return nil, indexErr
	}
	s := &segment{
		baseOffset:         baseOffset,
		nextRelativeOffset: 0,
		nextPosition:       0,
		log:                log,
		localIndex:         index,
		logWriter:          logWriter,
		localIndexWriter:   indexWriter,
	}
	return s, nil
}

func newLog(partitionPath string, baseOffset int64) (*os.File, *bufio.Writer, error) {
	logName := getLogName(baseOffset)
	logPath := path.Join(partitionPath, logName)
	f, fileErr := os.Create(logPath)
	if fileErr != nil {
		return nil, nil, fileErr
	}
	w := bufio.NewWriter(f)
	return f, w, nil
}

func newLocalIndex(partitionPath string, baseOffset int64) (*os.File, *bufio.Writer, error) {
	localIndexName := getLocalIndexName(baseOffset)
	localIndexPath := path.Join(partitionPath, localIndexName)
	f, err := os.Create(localIndexPath)
	if err != nil {
		return nil, nil, err
	}
	w := bufio.NewWriter(f)
	return f, w, nil
}

func newLogEntry(record string) string {
	l := &logEntry{
		Length: int32(len(record)),
		Record: record,
	}
	out, err := json.Marshal(l)
	if err != nil {
		logger.Printf(err.Error())
	}
	return string(out)
}

func getGlobalIndexName(startGsn int64) string {
	return fmt.Sprintf("%019d%s", startGsn, globalIndexSuffix)
}

func getLogName(baseOffset int64) string {
	return fmt.Sprintf("%019d%s", baseOffset, logSuffix)
}

func getLocalIndexName(baseOffset int64) string {
	return fmt.Sprintf("%019d%s", baseOffset, localIndexSuffix)
}
