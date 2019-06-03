package logger

import (
	"log"
	"os"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "[scalog]", log.Ldate|log.Lmicroseconds)
}

func Printf(format string, v ...interface{}) {
	logger.Printf(format, v...)
}

func Fatalf(format string, v ...interface{}) {
	logger.Fatalf(format, v...)
}

func Panicf(format string, v ...interface{}) {
	logger.Panicf(format, v...)
}
