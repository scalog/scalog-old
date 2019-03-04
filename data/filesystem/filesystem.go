package filesystem

import (
	"bufio"
	"os"
	"strconv"

	"github.com/scalog/scalog/logger"
)

var filePrefix = "log"
var fileSuffix = ".dat"
var recordLimit = 3

// RecordStorage is a struct managing information required to write
// information to stable storage in Scalog using files. Written files
// are of the format log#.dat. New files will be created when the
// presently open file has been written to [recordLimit] times
type RecordStorage struct {
	volumePath         string
	file               *os.File
	writer             *bufio.Writer
	fileNumber         int
	currentRecordCount int
}

// New filesystem
func New(volumePath string) RecordStorage {
	// Ensure volume path exists
	err := os.MkdirAll(volumePath, os.ModePerm)
	if err != nil {
		logger.Panicf(err.Error())
	}

	l := RecordStorage{volumePath, nil, nil, 1, 0}
	l.file = createFile(volumePath, genFilename(l.fileNumber))
	l.writer = createBufferedWriter(l.file)
	return l
}

// WriteLog writes to stable storage via a buffer
func (f *RecordStorage) WriteLog(gsn int, record string) {
	_, err := f.writer.WriteString(strconv.Itoa(gsn) + "\t" + record + "\n")
	if err != nil {
		panic(err)
	}
	// Flush to force write to stable storage
	f.writer.Flush()
	f.currentRecordCount++
	if f.currentRecordCount == recordLimit {
		f.currentRecordCount = 0
		f.file.Close()
		f.fileNumber++
		f.file = createFile(f.volumePath, genFilename(f.fileNumber))
		f.writer = createBufferedWriter(f.file)
	}
}

// Close terminates this file writer
func (f *RecordStorage) Close() {
	f.file.Close()
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func createFile(volume string, filename string) *os.File {
	f, err := os.Create(volume + "/" + filename)
	check(err)
	return f
}

func createBufferedWriter(file *os.File) *bufio.Writer {
	w := bufio.NewWriter(file)
	return w
}

func genFilename(fileNum int) string {
	return filePrefix + strconv.Itoa(fileNum) + fileSuffix
}
