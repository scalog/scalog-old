package golib

import (
	"github.com/scalog/scalog/logger"
	"google.golang.org/grpc"
)

/**
Attempt to connect to the URL given as a GRPC client. Returns nil during failure.
*/
func ConnectTo(url string) *grpc.ClientConn {
	var opts []grpc.DialOption
	// TODO: Use a secured connection
	opts = append(opts, grpc.WithInsecure())

	logger.Printf("Dialing " + url)
	conn, err := grpc.Dial(url, opts...)
	if err != nil {
		logger.Printf(err.Error())
		return nil
	}
	return conn
}
