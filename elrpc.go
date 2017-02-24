package elrpc

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
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

type Service interface {
	SetDebug(b bool)
	IsRunning() bool
	Stop() error
	RegisterMethod(m *Method)
	Call(name string, args ...interface{}) (interface{}, error)
	QueryMethods() ([]*MethodDesc, error)
	Wait()
}

/// Server

func queryFreePort() (int, error) {
	s, err := net.Listen("tcp", ":0")
	if err != nil {
		return -1, fmt.Errorf("could not listen TCP port 0: %v", err)
	}
	defer s.Close()
	tcpa, _ := s.Addr().(*net.TCPAddr)
	return tcpa.Port, nil
}

type serverState int

const (
	_ serverState = iota
	serverStateOpened
	serverStateClosed
)

func StartServer(methods []*Method) (*ServerService, error) {
	return StartServerWithPort(methods, 0)
}

func StartServerWithPort(methods []*Method, port int) (*ServerService, error) {
	logger := log.New(os.Stderr, fmt.Sprintf("%s ", "SS"), log.Ldate|log.Ltime)
	if port == 0 {
		nport, err := queryFreePort()
		if err != nil {
			return nil, err
		}
		port = nport
	}
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	fmt.Printf("%d\n", port)

	if methods != nil {
		methods = []*Method{}
	}
	debugMode := false
	if defaultLogLevel == LogLevelDebug {
		debugMode = true
	}
	ss := &ServerService{
		count:       0,
		logger:      logger,
		debugMode:   debugMode,
		serverState: serverStateOpened,
		listener:    ln,
		services:    []Service{},
		methods:     methods,
	}

	return ss, nil
}

type ServerService struct {
	count       int        // counter for accepted servers
	countMu     sync.Mutex // protect for count
	debugMode   bool
	logger      *log.Logger
	serverState serverState
	listener    net.Listener
	services    []Service
	methods     []*Method
}

func (ss *ServerService) incServerCount() int {
	ss.countMu.Lock()
	defer ss.countMu.Unlock()
	ss.count++
	return ss.count
}

func (ss *ServerService) SetDebug(a bool) {
	ss.debugMode = a
	for _, s := range ss.services {
		s.SetDebug(a)
	}
}

func (ss *ServerService) debugf(format string, args ...interface{}) {
	if ss.debugMode {
		ss.logger.Printf(format, args...)
	}
}

func (ss *ServerService) Close() {
	if ss.serverState == serverStateClosed {
		return
	}
	ss.listener.Close()
	ss.listener = nil
	ss.serverState = serverStateClosed
}

func (ss *ServerService) RegisterMethod(m *Method) {
	ss.methods = append(ss.methods, m)
}

func (ss *ServerService) Accept() (*RPCServer, error) {
	ss.debugf("waiting for client connection...")
	conn, err := ss.listener.Accept()
	if err != nil {
		return nil, err
	}
	ss.debugf("incoming connection.")
	s := makeRPCServer(
		fmt.Sprintf("SS%d", ss.incServerCount()),
		conn, ss.methods)
	s.SetDebug(ss.debugMode)
	ss.debugf("make a rpc server.")
	ss.services = append(ss.services, s)
	return s, nil
}

func (ss *ServerService) Wait() {
	s, err := ss.Accept()
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	s.Wait()
}

/// Client

func StartClient(port int, methods []*Method) (Service, error) {
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	cs := makeRPCServer("CL:"+addr, conn, methods)
	return cs, nil
}

type clientService struct {
	cmd  []string
	port int
	*RPCServer
}

func StartProcess(cmd []string, methods []*Method) (Service, error) {
	return StartProcessWithPort(cmd, -1, methods)
}

func StartProcessWithPort(cmd []string, port int, methods []*Method) (Service, error) {
	// start peer's process
	proc := exec.Command(cmd[0], cmd[1:]...)
	stdout, err := proc.StdoutPipe()
	if err != nil {
		return nil, err
	}
	_ = proc.Start()

	// get peer's port number
	lineBuf := bufio.NewScanner(stdout)
	if !lineBuf.Scan() {
		return nil, fmt.Errorf("could not scan port line")
	}
	line := lineBuf.Text()
	if port < 0 {
		port, err = strconv.Atoi(line)
		if err != nil {
			return nil, fmt.Errorf("could not get peer's port number: %s", line)
		}
	}

	// connect to server
	var conn net.Conn
	addr := fmt.Sprintf("localhost:%d", port)
	for times := 1; times < 10; times++ {
		conn, err = net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err != nil {
			fmt.Printf("Peer port is not opened. Try next time...: %v", err)
			continue
		}
		break
	}
	if conn == nil {
		return nil, fmt.Errorf("could not connect to the peer's port: %s", addr)
	}

	c := makeRPCServer("CL:"+addr, conn, methods)
	cs := &clientService{
		RPCServer: c,
		cmd:       cmd,
		port:      port,
	}

	return cs, nil
}
