package filesystem

import (
	"bufio"
	"fmt"
	"os"
)

var filePrefix = "log"
var fileSuffix = ".dat"
var recordLimit = 3

// TODO: If the number of records in a file exceeds N, close it and create a new file in
// this particular volume seamlessly.

// RecordStorage is a file writer for Scalog. v1.
type RecordStorage struct {
	volumePath         string
	file               *os.File
	writer             *bufio.Writer
	fileNumber         int
	currentRecordCount int
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

func genFilename(name string) string {
	return filePrefix + name + fileSuffix
}

// New filesystem
func New(volumePath string) RecordStorage {
	e := RecordStorage{volumePath, nil, nil, 0, 0}
	e.file = createFile(volumePath, genFilename(string(e.fileNumber)))
	e.writer = createBufferedWriter(e.file)
	return e
}

// WriteLog writes to stable storage via a buffer
func (f *RecordStorage) WriteLog(gsn int, record string) {
	_, err := f.writer.WriteString(string(gsn) + "\t" + record)
	if err != nil {
		panic(err)
	}
	f.currentRecordCount++
	if f.currentRecordCount == recordLimit {
		f.currentRecordCount = 0
		f.file.Close()
		f.fileNumber++
		f.file = createFile(f.volumePath, genFilename(string(f.fileNumber)))
		f.writer = createBufferedWriter(f.file)
	}
}

// Close terminates this file writer
func (f *RecordStorage) Close() {
	f.file.Close()
}

func main() {
	fmt.Println("Running RecordStorage")
	rs := New("/tmp")
	rs.WriteLog(1, "hello")
}
