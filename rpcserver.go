package elrpc

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"reflect"
	"strconv"
	"sync"

	"github.com/kiwanami/go-elrpc/parser"
)

var uidCounter struct {
	mu    sync.Mutex
	count int
}

func genuid() int {
	uidCounter.mu.Lock()
	defer uidCounter.mu.Unlock()
	uidCounter.count++
	return uidCounter.count
}

/// RPCServer

type socketState int

const (
	_ socketState = iota
	socketStateOpened
	socketStateClosing
	socketStateNotConnected
)

/// method and message

type Method struct {
	name      string
	mtype     reflect.Type
	mfunc     reflect.Value
	argTypes  []reflect.Type
	argdoc    string
	docstring string
}

func MakeMethod(name string, proc interface{}, argdoc string, docstring string) *Method {
	mtype := reflect.TypeOf(proc)
	if mtype.Kind() != reflect.Func {
		panic(fmt.Sprintf("not a function : %v (name=%s)", proc, name))
	}
	ats := make([]reflect.Type, mtype.NumIn())
	for i := 0; i < len(ats); i++ {
		ats[i] = mtype.In(i)
	}
	return &Method{
		name:      name,
		mtype:     mtype,
		argTypes:  ats,
		mfunc:     reflect.ValueOf(proc),
		argdoc:    argdoc,
		docstring: docstring,
	}
}

type MethodDesc struct {
	Name      string
	Argdoc    string
	Docstring string
}

type methodResult struct {
	success bool
	value   interface{}
	err     error
}

type message interface {
	msgID() int
	ToAst() (parser.SExp, error)
}

type messageCall struct {
	uid    int
	method string
	args   []interface{}
}

func (m *messageCall) msgID() int {
	return m.uid
}

func (m *messageCall) ToAst() (parser.SExp, error) {
	val, err := Encode(m.args)
	if err != nil {
		return nil, err
	}
	return parser.AstListv(
		parser.AstSymbol("call"),
		parser.AstInt(strconv.Itoa(m.uid)),
		parser.AstString(m.method), parser.AstWrapper(val),
	), nil
}

type messageMethod struct {
	uid int
}

func (m *messageMethod) msgID() int {
	return m.uid
}

func (m *messageMethod) ToAst() (parser.SExp, error) {
	return parser.AstListv(
		parser.AstSymbol("methods"),
		parser.AstInt(strconv.Itoa(m.uid)),
	), nil
}

type messageReturn struct {
	uid   int
	value interface{}
}

func (m *messageReturn) msgID() int {
	return m.uid
}

func (m *messageReturn) ToAst() (parser.SExp, error) {
	val, err := Encode(m.value)
	if err != nil {
		return nil, err
	}
	return parser.AstListv(
		parser.AstSymbol("return"),
		parser.AstInt(strconv.Itoa(m.uid)),
		parser.AstWrapper(val),
	), nil
}

type messageError struct {
	uid int
	msg string
}

func (m *messageError) msgID() int {
	return m.uid
}

func (m *messageError) ToAst() (parser.SExp, error) {
	val, err := Encode(m.msg)
	if err != nil {
		return nil, err
	}
	return parser.AstListv(
		parser.AstSymbol("return-error"),
		parser.AstInt(strconv.Itoa(m.uid)),
		parser.AstWrapper(val),
	), nil
}

type messageEpcError struct {
	uid int
	msg string
}

func (m *messageEpcError) msgID() int {
	return m.uid
}

func (m *messageEpcError) ToAst() (parser.SExp, error) {
	val, err := Encode(m.msg)
	if err != nil {
		return nil, err
	}
	return parser.AstListv(
		parser.AstSymbol("epc-error"),
		parser.AstInt(strconv.Itoa(m.uid)),
		parser.AstWrapper(val),
	), nil
}

type messageCancel struct {
	uid int
}

func (m *messageCancel) msgID() int {
	return m.uid
}

func (m *messageCancel) ToAst() (parser.SExp, error) {
	return parser.AstListv(
		parser.AstSymbol("cancel"),
		parser.AstInt(strconv.Itoa(m.uid)),
	), nil
}

/// error

type EPCRuntimeError struct {
	message   string
	backtrace string
}

func (e *EPCRuntimeError) Error() string {
	return "epc runtime error: " + e.message + "\n" + e.backtrace
}

type EPCStackError struct {
	message   string
	backtrace string
}

func (e *EPCStackError) Error() string {
	return "epc stack error: " + e.message + "\n" + e.backtrace
}

/// RPCServer

type workerMsg int

const (
	_ workerMsg = iota
	workerClose
	workerClosed
)

type serverMsgType int

const (
	_ serverMsgType = iota
	serverStop
	serverAddExitHook
)

type serverMsg struct {
	msg      serverMsgType
	response chan interface{}
}

type RPCServer struct {
	logger       *log.Logger
	debugMode    bool
	methods      map[string]*Method
	session      map[int]chan *methodResult
	sessionMutex sync.RWMutex
	socket       net.Conn
	socketOut    *bufio.Writer

	sendingQueue chan message

	socketState socketState
	user2svChan chan *serverMsg // channel from user to server
	rcv2svChan  chan workerMsg  // channel from receiver to server
	snd2svChan  chan workerMsg  // channel from sender to server
	sv2sndChan  chan workerMsg  // channel from server to sender
	exitHook    []func()        // server exit hook function
}

func (s *RPCServer) SetDebug(d bool) {
	s.debugMode = d
}

func (s *RPCServer) debugf(format string, args ...interface{}) {
	if s.debugMode {
		s.logger.Printf(format, args...)
	}
}

func makeRPCServer(name string, socket net.Conn, methods []*Method) *RPCServer {
	logger := log.New(os.Stderr, fmt.Sprintf("%s ", name), log.Ldate|log.Ltime)
	server := &RPCServer{
		logger:       logger,
		socketState:  socketStateOpened,
		socket:       socket,
		socketOut:    bufio.NewWriter(socket),
		methods:      make(map[string]*Method),
		session:      make(map[int]chan *methodResult),
		sendingQueue: make(chan message, 20),

		user2svChan: make(chan *serverMsg, 1),
		rcv2svChan:  make(chan workerMsg, 1),
		snd2svChan:  make(chan workerMsg, 1),
		sv2sndChan:  make(chan workerMsg, 1),
		exitHook:    []func(){},
	}

	if methods != nil {
		for _, m := range methods {
			server.RegisterMethod(m)
		}
	}

	go server.serverWorker()
	go server.senderWorker()
	go server.receiverWorker()

	return server
}

func (s *RPCServer) serverWorker() {
	var socketErr error
	receiverState := true
	senderState := true
	for {
		select {
		case ev := <-s.user2svChan:
			s.debugf("ServerWorker: receive comm signal [%v]", ev)
			switch ev.msg {
			case serverStop:
				if s.socketState == socketStateOpened {
					s.debugf("ServerWorker: sending stop signal")
					s.socketState = socketStateClosing
					socketErr = s.socket.Close()
					go func() { s.sv2sndChan <- workerClose }()
				}
				defer func() {
					ev.response <- socketErr
				}()
			case serverAddExitHook:
				s.exitHook = append(s.exitHook, func() {
					ev.response <- "exited"
				})
			}
		case rev := <-s.rcv2svChan:
			s.debugf("ServerWorker: receive receiver signal [%v]", rev)
			if rev == workerClosed {
				receiverState = false
				if s.socketState == socketStateOpened {
					s.debugf("ServerWorker: stop signal from receiver")
					if senderState {
						go func() { s.sv2sndChan <- workerClose }()
					}
				}
			} else {
				s.debugf("ServerWorker: invalid receiver msg: %v", rev)
			}
		case sev := <-s.snd2svChan:
			s.debugf("ServerWorker: receive sender signal [%v]", sev)
			if sev == workerClosed {
				senderState = false
			}
		}
		s.debugf("ServerWorker state: send:%v,  recv:%v", senderState, receiverState)
		if !receiverState && !senderState {
			s.socketState = socketStateNotConnected
			break
		}
	}
	close(s.snd2svChan)
	close(s.sv2sndChan)
	close(s.rcv2svChan)
	s.cleanupSessions()
	s.execExitHook()
	s.debugf("ServerWorker exited: sockerr: %v, send:%v,  recv:%v",
		socketErr, senderState, receiverState)
	close(s.user2svChan)
}

func (s *RPCServer) addExitHook(f func()) {
	msg := &serverMsg{
		msg:      serverAddExitHook,
		response: make(chan interface{}, 1),
	}
	go func() {
		s.user2svChan <- msg
		<-msg.response // wait for exit event
		f()
	}()
}

func (s *RPCServer) execExitHook() {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Printf("ExitHook: panic error : %+v\n", r)
		}
	}()
	if len(s.exitHook) > 0 {
		for _, f := range s.exitHook {
			f()
		}
	}
}

func (s *RPCServer) cleanupSessions() {
	s.sessionMutex.Lock()
	defer s.sessionMutex.Unlock()
	for k := range s.session {
		session := s.session[k]
		mresult := &methodResult{
			success: false,
			value:   nil,
			err:     fmt.Errorf("unexpected peer's shutdown"),
		}
		go func() {
			session <- mresult
		}()
	}
}

func (s *RPCServer) senderWorker() {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Printf("SenderWoker: panic error : %+v\n", r)
		}
	}()
Loop:
	for {
		select {
		case <-s.sv2sndChan:
			s.debugf("SenderWoker: received stop message.")
			break Loop
		case sndmsg := <-s.sendingQueue:
			s.debugf("SenderWoker: pop a message : %v\n", sndmsg)
			err := s.sendMessage(sndmsg)
			if err != nil {
				s.logger.Println("SenderWoker: error : " + err.Error())
				if _, ok := sndmsg.(*messageReturn); ok {
					// notify remote receiver
					errmsg := &messageEpcError{
						uid: sndmsg.msgID(),
						msg: "epc error: " + err.Error(),
					}
					go func() { s.sendingQueue <- errmsg }()
				} else {
					// notify local receiver
					_ = s.receiveReturnEpcError(Array("epc-error", sndmsg.msgID(), err.Error()))
				}
			} else {
				s.debugf("SenderWoker: sent a message. [err:%v]\n", err)
			}
			s.debugf("SenderWoker: try next message.")
		}
	}
	s.debugf("SenderWoker: exiting...")
	s.snd2svChan <- workerClosed
	s.debugf("SenderWoker: exited.")
}

func (s *RPCServer) receiverWorker() {
	lenbuf := make([]byte, 6)
	bodybuf := make([]byte, 1024)
	var bodyArr []interface{}
	var uid int
	var mtype string
	var err error
	for {
		_, err = io.ReadFull(s.socket, lenbuf)
		if err != nil {
			s.debugf("ReceiverWorker: read len error :" + err.Error())
			break
		}
		blen64, err := strconv.ParseInt(string(lenbuf), 16, 24)
		if err != nil {
			s.debugf("ReceiverWorker: read len error :" + err.Error())
			break
		}
		blen := int(blen64)
		if blen > len(bodybuf) {
			bodybuf = make([]byte, blen)
		} else {
			bodybuf = bodybuf[0:blen]
		}
		_, err = io.ReadFull(s.socket, bodybuf)
		if err != nil {
			s.logger.Println("ReceiverWorker: read body error :" + err.Error())
			break
		}
		s.debugf("[[ %v ]]", string(bodybuf))
		bodyObj, err := Decode1(string(bodybuf))
		if err != nil {
			s.logger.Println("ReceiverWorker: body parse error: " + err.Error())
			break
		}
		bodyArr = ToArray(bodyObj)
		if bodyArr == nil {
			s.logger.Println("ReceiverWorker: invalid message: not an array.")
			break
		}
		mtype, uid, err = parseMessageHeader(bodyArr)
		if err != nil {
			s.logger.Println("ReceiverWorker: invalid message header: " + err.Error())
			break
		}
		switch mtype {
		case "call":
			err = s.receiveCall(bodyArr)
			if err != nil {
				emsg := &messageEpcError{
					uid: uid,
					msg: fmt.Sprintf("epc error: %v", err),
				}
				s.sendingQueue <- emsg
				s.logger.Println("ReceiverWorker: send epc-return: " + err.Error())
				err = nil
			}
		case "cancel":
			err = s.receiveCancel(bodyArr)
		case "return":
			err = s.receiveReturn(bodyArr)
		case "return-error":
			err = s.receiveReturnError(bodyArr)
		case "epc-error":
			err = s.receiveReturnEpcError(bodyArr)
		case "methods":
			err = s.receiveMethods(bodyArr)
		default:
			err = fmt.Errorf("invalid message type: %s", mtype)
		}
		if err != nil {
			s.logger.Printf("ReceiverWorker: runtime error:  uid=%d,  mtype=%s,  err=%v\n", uid, mtype, err)
			// try to continue next message
		}
	}

	s.debugf("ReceiverWorker: exiting...")
	s.rcv2svChan <- workerClosed
	s.debugf("ReceiverWorker: exited.")
}

func parseMessageHeader(bodyArr []interface{}) (mtype string, uid int, err error) {
	var ok bool
	err = nil
	mtype, ok = bodyArr[0].(string)
	if !ok {
		err = errors.New("message type is not string")
		return
	}
	uid, ok = bodyArr[1].(int)
	if !ok {
		err = errors.New("message uid is not int")
		return
	}
	return
}

/// server functions

func (s *RPCServer) IsRunning() bool {
	return s.socketState == socketStateOpened
}

func (s *RPCServer) Stop() error {
	if !s.IsRunning() {
		return nil
	}
	s.debugf("waiting for workers shutdown: ")
	response := make(chan interface{}, 1)
	s.user2svChan <- &serverMsg{
		msg:      serverStop,
		response: response,
	}
	ret := <-response
	s.debugf("shutdown ok: %v", ret)
	if ret == nil {
		return nil
	}
	return ret.(error)
}

func (s *RPCServer) Wait() {
	w := make(chan bool, 1)
	s.addExitHook(func() {
		w <- true
	})
	<-w
}

func (s *RPCServer) WaitingSessionNum() int {
	s.sessionMutex.RLock()
	ret := len(s.session)
	s.sessionMutex.RUnlock()
	return ret
}

func (s *RPCServer) RegisterMethod(m *Method) {
	s.methods[m.name] = m
}

func (s *RPCServer) Call(name string, args ...interface{}) (interface{}, error) {
	if s.socketState != socketStateOpened {
		return nil, fmt.Errorf("epc not connected")
	}
	uid := genuid()
	msg := &messageCall{
		uid:    uid,
		method: name,
		args:   args,
	}
	rcvChan := make(chan *methodResult, 1)
	s.sessionMutex.Lock()
	s.session[uid] = rcvChan
	s.sessionMutex.Unlock()
	s.sendingQueue <- msg
	result := <-rcvChan
	if !result.success {
		return nil, result.err
	}
	return result.value, nil
}

func (s *RPCServer) CallContext(ctx context.Context, name string, args ...interface{}) (interface{}, error) {
	if s.socketState != socketStateOpened {
		return nil, fmt.Errorf("epc not connected")
	}
	uid := genuid()
	msg := &messageCall{
		uid:    uid,
		method: name,
		args:   args,
	}
	rcvChan := make(chan *methodResult, 1)
	s.sessionMutex.Lock()
	s.session[uid] = rcvChan
	s.sessionMutex.Unlock()
	s.sendingQueue <- msg
	select {
	case <-ctx.Done():
		s.sendCanceling(uid)
		return nil, fmt.Errorf("Canceled")
	case result := <-rcvChan:
		if !result.success {
			return nil, result.err
		}
		return result.value, nil
	}
}

func (s *RPCServer) QueryMethods() ([]*MethodDesc, error) {
	if s.socketState != socketStateOpened {
		return nil, fmt.Errorf("epc not connected")
	}
	uid := genuid()
	msg := &messageMethod{uid: uid}
	rcvChan := make(chan *methodResult, 1)
	s.sessionMutex.Lock()
	s.session[uid] = rcvChan
	s.sessionMutex.Unlock()
	s.sendingQueue <- msg
	result := <-rcvChan
	if !result.success {
		return nil, result.err
	}
	vs, ok := result.value.([]interface{})
	if !ok {
		val := reflect.ValueOf(result.value)
		return nil, fmt.Errorf("invalid method query result: %v %v", result.value, val.Kind().String())
	}
	ms := make([]*MethodDesc, len(vs))
	for i, vv := range vs {
		mstrs := vv.([]interface{})
		ms[i] = &MethodDesc{
			Name:      mstrs[0].(string),
			Argdoc:    mstrs[1].(string),
			Docstring: mstrs[2].(string),
		}
	}
	return ms, nil
}

func (s *RPCServer) sendCanceling(uid int) {
	msg := &messageCancel{uid: uid}
	s.sessionMutex.Lock()
	delete(s.session, uid)
	s.sessionMutex.Unlock()
	s.sendingQueue <- msg
}

func (s *RPCServer) sendMessage(m message) error {
	msgObj, err := m.ToAst()
	if err != nil {
		return err
	}
	buf := msgObj.ToSExpString()
	len := len(buf)
	_, err = io.WriteString(s.socketOut, fmt.Sprintf("%06x", len))
	if err != nil {
		return err
	}
	_, err = io.WriteString(s.socketOut, buf)
	if err != nil {
		return err
	}
	err = s.socketOut.Flush()
	if err != nil {
		return err
	}
	return nil
}

func (s *RPCServer) receiveCall(bodyArr []interface{}) (err error) {
	uid, ok := bodyArr[1].(int)
	if !ok {
		return fmt.Errorf("uid is not int [%v]", bodyArr[1])
	}
	name, ok := bodyArr[2].(string)
	if !ok {
		return fmt.Errorf("method name is not string [%v]", bodyArr[2])
	}

	argsv := reflect.ValueOf(bodyArr[3])
	if !argsv.IsValid() || argsv.IsNil() {
		argsv = reflect.ValueOf([]interface{}{})
	} else if argsv.Kind() != reflect.Slice {
		return fmt.Errorf("arguments object is not list [%v, %v]", bodyArr[3], argsv.Kind().String())
	}

	s.debugf(": called: name=%s : uid=%d", name, uid)
	method, ok := s.methods[name]
	if !ok {
		return fmt.Errorf("method not found: name=%s", name)
	}

	// argument type check
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("type invalid error : %+v", r)
		}
	}()
	argsvlen := argsv.Len()

	if argsvlen != len(method.argTypes) {
		return fmt.Errorf("different argument length: expected %d, but received %d",
			len(method.argTypes), argsvlen)
	}
	s.debugf(": extracting arguments: %v", argsvlen)
	argv := make([]reflect.Value, argsvlen)
	for i := 0; i < argsvlen; i++ {
		av := reflect.ValueOf(argsv.Index(i).Interface())
		it := method.argTypes[i]
		s.debugf("   : %v : %v -> %v", av.Interface(), av.Type().Kind(), it.Kind())
		if av.Type().Kind() != it.Kind() {
			av, err = ConvertType(it, av)
			if err != nil {
				return fmt.Errorf("can not convert type: [%v] : type[%v] -> type[%v]", av, av.Type().String(), it.String())
			}
		} else if av.Type().Kind() == reflect.Slice && it.Kind() == reflect.Slice {
			av, err = ConvertType(it, av)
			if err != nil {
				return fmt.Errorf("can not convert type: [%v] : type[%v] -> type[%v]", av, av.Type().String(), it.String())
			}
		}
		argv[i] = av
	}

	// execute function
	go func() {
		defer func() {
			rr := recover()
			if rr != nil {
				emsg := &messageError{
					uid: uid,
					msg: fmt.Sprintf("Go error: %v", rr),
				}
				s.sendingQueue <- emsg
				s.debugf(": executing DONE ERROR: name=%s : uid=%d , error=%v", name, uid, rr)
			}
		}()

		s.debugf(": executing: name=%s : uid=%d", name, uid)
		retv := method.mfunc.Call(argv)
		var vv interface{}
		if len(retv) == 0 {
			vv = nil
		} else {
			vv = retv[0].Interface()
		}
		rmsg := &messageReturn{uid: uid, value: vv}
		s.sendingQueue <- rmsg
		s.debugf(": executing DONE: name=%s : uid=%d", name, uid)
	}()

	return nil
}

func (s *RPCServer) receiveReturn(bodyArr []interface{}) (err error) {
	uid, ok := bodyArr[1].(int)
	if !ok {
		return fmt.Errorf("uid is not int [%v]", bodyArr[1])
	}
	value := bodyArr[2]
	s.debugf(": returned: uid=%d", uid)

	s.sessionMutex.RLock()
	session, ok := s.session[uid]
	s.sessionMutex.RUnlock()
	if !ok {
		return fmt.Errorf("not found a session for uid=%d", uid)
	}
	s.sessionMutex.Lock()
	delete(s.session, uid)
	s.sessionMutex.Unlock()
	mresult := &methodResult{
		success: true,
		value:   value,
	}
	go func() {
		session <- mresult
	}()

	return nil
}

func (s *RPCServer) receiveReturnError(bodyArr []interface{}) (err error) {
	uid, ok := bodyArr[1].(int)
	if !ok {
		return fmt.Errorf("uid is not int [%v]", bodyArr[1])
	}
	errval := bodyArr[2]
	s.debugf(": returned error: uid=%d  error=%v", uid, errval)

	s.sessionMutex.RLock()
	session, ok := s.session[uid]
	s.sessionMutex.RUnlock()
	if !ok {
		return fmt.Errorf("not found a session for uid=%d", uid)
	}
	s.sessionMutex.Lock()
	delete(s.session, uid)
	s.sessionMutex.Unlock()
	mresult := &methodResult{
		success: false,
		value:   nil,
		err:     fmt.Errorf("%v", errval),
	}
	go func() {
		session <- mresult
	}()
	return nil
}

func (s *RPCServer) receiveReturnEpcError(bodyArr []interface{}) (err error) {
	uid, ok := bodyArr[1].(int)
	if !ok {
		return fmt.Errorf("uid is not int [%v]", bodyArr[1])
	}
	errval := bodyArr[2]
	s.debugf(": returned epc-error: uid=%d  error=%v", uid, errval)

	s.sessionMutex.RLock()
	session, ok := s.session[uid]
	s.sessionMutex.RUnlock()
	if !ok {
		return fmt.Errorf("not found a session for uid=%d", uid)
	}
	s.sessionMutex.Lock()
	delete(s.session, uid)
	s.sessionMutex.Unlock()
	mresult := &methodResult{
		success: false,
		value:   nil,
		err:     fmt.Errorf("%v", errval),
	}
	go func() {
		session <- mresult
	}()
	return nil
}

func (s *RPCServer) receiveCancel(bodyArr []interface{}) (err error) {
	uid, ok := bodyArr[1].(int)
	if !ok {
		return fmt.Errorf("uid is not int [%v]", bodyArr[1])
	}
	s.debugf(": cancel: uid=%d", uid)
	// TODO Cancel
	// check handler function arguments have a context
	// hold the context for uid
	// send cancel event to the context
	// cleanup session info
	return nil
}

func (s *RPCServer) receiveMethods(bodyArr []interface{}) (err error) {
	uid, ok := bodyArr[1].(int)
	if !ok {
		return fmt.Errorf("uid is not int [%v]", bodyArr[1])
	}
	s.debugf(": query-methods: uid=%d", uid)

	result := make([][]string, len(s.methods))
	idx := 0
	for _, m := range s.methods {
		result[idx] = []string{
			m.name, m.argdoc, m.docstring,
		}
		idx++
	}

	rmsg := &messageReturn{
		uid:   uid,
		value: result,
	}
	s.sendingQueue <- rmsg
	s.debugf(": query-methods DONE: uid=%d", uid)

	return nil
}
