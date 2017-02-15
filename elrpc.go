package elrpc

import (
	"fmt"
	"net"
	"strconv"
)

/// log

type LogLevel int

const (
	_ LogLevel = iota
	LogLevelDebug
	LogLevelInfo
)

var defaultLogLevel LogLevel = LogLevelInfo

func SetDefaultLogLevel(l LogLevel) {
	defaultLogLevel = l
}

/// utility

func Array(args ...interface{}) []interface{} {
	return args
}

/// service

type Service struct {
	cmd    string
	port   uint
	client RPCServer
}

func StartServer(methods []Method) (*RPCServer, error) {
	return StartServerWithPort(methods, 0)
}

func StartServerWithPort(methods []Method, port int) (*RPCServer, error) {
	serverSocket, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	_, portStr, err := net.SplitHostPort(serverSocket.Addr().String())
	if err != nil {
		return nil, fmt.Errorf("Could not get port number: %v", serverSocket.Addr().String())
	}
	portNum, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("Could not get port number: %v", portStr)
	}
	fmt.Printf("%d\n", portNum)
	// TODO start server

	return nil, nil
}

func StartClient(port int, methods []Method, host string) {

}
