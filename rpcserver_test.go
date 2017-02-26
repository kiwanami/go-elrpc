package elrpc

import (
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/k0kubun/pp"
)

func TestRpcServerStop(t *testing.T) {
	mockConn := makeMockConn()
	server := makeRPCServer("ServerStop", mockConn, nil)
	if !server.IsRunning() {
		//pp.Println(server)
		t.Error("server not running")
	}
	server.Stop()
}

func TestRpcCloseByRemote(t *testing.T) {
	mockConn := makeMockConn()
	server := makeRPCServer("CloseByRemote", mockConn, nil)
	mockConn.Close()
	time.Sleep(100 * time.Millisecond)
	if server.IsRunning() {
		t.Error("failed to shutdown server")
	} else {
		server.Stop()
	}
}

func TestRpcEcho(t *testing.T) {
	mockConn := makeMockConn()
	server := makeRPCServer("Echo1", mockConn, nil)
	//server.SetDebug(true)
	defer server.Stop()
	wt := make(chan interface{}, 1)
	go func() {
		ret, err := server.Call("echo", "test1")
		if err != nil {
			wt <- err
		} else {
			wt <- ret
		}
	}()
	time.Sleep(50 * time.Millisecond)
	// check sending message
	buf := make([]byte, 100)
	n, _ := mockConn.GetWriter(buf)
	sbuf := buf[6:n]
	callmsg := string(sbuf)
	if callmsg != fmt.Sprintf("(call %d \"echo\" (\"test1\"))", uidCounter.count) {
		pp.Println(callmsg)
		t.Error("Could not pass the echo call message.")
	}
	// check receive message
	cc := uidCounter.count
	body := fmt.Sprintf("(return %d \"test1\")", cc)
	msg := fmt.Sprintf("%06x%s", len(body), body)
	mockConn.PushReader([]byte(msg))
	ret := <-wt
	if ret.(string) != "test1" {
		pp.Println(ret)
		t.Error("Could not get the echo return message.")
	}
}

func TestRpcEcho2(t *testing.T) {
	mockConn := makeMockConn()
	ms := []*Method{
		MakeMethod("echo", func(msg string) string {
			return "echo:" + msg
		}, "echo string", "return echo string"),
	}
	server := makeRPCServer("Echo2", mockConn, ms)
	//server.SetDebug(true)
	defer server.Stop()
	time.Sleep(50 * time.Millisecond)

	cc := genuid()
	body := fmt.Sprintf("(call %d \"echo\" (\"test2\"))", cc)
	msg := fmt.Sprintf("%06x%s", len(body), body)
	buf := []byte(msg)
	mockConn.PushReader(buf)

	buf = make([]byte, 100)
	n, _ := mockConn.GetWriter(buf)
	sbuf := buf[6:n]
	ret := string(sbuf)
	//pp.Println(ret)
	if ret != fmt.Sprintf("(return %d \"echo:test2\")", uidCounter.count) {
		t.Errorf("Could not pass the echo message: %v", ret)
	}
}

func testErrorReturn(t *testing.T, conn *mockConn, name string, args interface{}, expectedf string) {
	cc := genuid()
	body := fmt.Sprintf("(call %d \"%s\" (%v))", cc, name, args)
	msg := fmt.Sprintf("%06x%s", len(body), body)
	buf := []byte(msg)
	conn.PushReader(buf)

	buf = make([]byte, 1024)
	n, _ := conn.GetWriter(buf)
	sbuf := buf[6:n]
	ret := string(sbuf)
	//pp.Println(ret)
	if ret != fmt.Sprintf(expectedf, uidCounter.count) {
		t.Errorf("Could not check the error: expected:[%v]  returned:[%v]", expectedf, ret)
	}
}

func TestRpcError1(t *testing.T) {
	mockConn := makeMockConn()
	ms := []*Method{
		MakeMethod("errorMethod", func(i int) int {
			return 10 / i
		}, "", ""),
	}
	server := makeRPCServer("Error1", mockConn, ms)
	//server.SetDebug(true)
	defer server.Stop()
	time.Sleep(50 * time.Millisecond)

	testErrorReturn(t, mockConn, "errorMethod", "0", "(return-error %d \"Go error: runtime error: integer divide by zero\")")
}

func TestRpcEpcError1(t *testing.T) {
	mockConn := makeMockConn()
	ms := []*Method{
		MakeMethod("epcerror", func(i int) interface{} {
			return MakeMethod("a", func() {}, "", "")
		}, "", ""),
	}
	server := makeRPCServer("EpcError1", mockConn, ms)
	//server.SetDebug(true)
	defer server.Stop()
	time.Sleep(50 * time.Millisecond)

	// wrong method name
	testErrorReturn(t, mockConn, "epccerror", "1", "(epc-error %d \"epc error: method not found: name=epccerror\")")

	// wrong argument number
	testErrorReturn(t, mockConn, "epcerror", "1 2 3", "(epc-error %d \"epc error: different argument length: expected 1, but received 3\")")

	// serialize error
	testErrorReturn(t, mockConn, "epcerror", "1", "(epc-error %d \"epc error: sexp encode: unsupported type: func(unsafe.Pointer, uintptr) uintptr\")")
}

func TestRpcMethods1(t *testing.T) {
	mockConn := makeMockConn()
	ms := []*Method{
		MakeMethod("echo", func(i int) interface{} {
			return i
		}, "argdoc", "docstring"),
	}
	server := makeRPCServer("Method", mockConn, ms)
	//server.SetDebug(true)
	defer server.Stop()
	time.Sleep(50 * time.Millisecond)

	cc := genuid()
	body := fmt.Sprintf("(methods %d)", cc)
	msg := fmt.Sprintf("%06x%s", len(body), body)
	buf := []byte(msg)
	mockConn.PushReader(buf)

	buf = make([]byte, 1024)
	n, _ := mockConn.GetWriter(buf)
	sbuf := buf[6:n]
	ret := string(sbuf)
	//pp.Println(ret)
	exp := Array(Array("echo", "argdoc", "docstring"))
	expstr, _ := Encode(exp)
	if ret != fmt.Sprintf("(return %d %s)", uidCounter.count, expstr) {
		t.Errorf("Could not pass the echo message: %v", ret)
	}
}

/// socket mock

type mockConn struct {
	connReader   *io.PipeReader
	connWriter   *io.PipeWriter
	clientReader *io.PipeReader
	clientWriter *io.PipeWriter
}

func makeMockConn() *mockConn {
	clr, svw := io.Pipe()
	svr, clw := io.Pipe()
	ret := &mockConn{
		connReader: svr, connWriter: svw,
		clientReader: clr, clientWriter: clw,
	}
	return ret
}

func (m *mockConn) PushReader(b []byte) (n int, err error) {
	return m.clientWriter.Write(b)
}

func (m *mockConn) GetWriter(b []byte) (n int, err error) {
	return m.clientReader.Read(b)
}

func (m *mockConn) Read(b []byte) (n int, err error) {
	return m.connReader.Read(b)
}
func (m *mockConn) Write(b []byte) (n int, err error) {
	return m.connWriter.Write(b)
}

func (m *mockConn) Close() error {
	var err error
	err = m.connWriter.Close()
	if err != nil {
		return err
	}
	err = m.connReader.Close()
	if err != nil {
		return err
	}
	return nil
}
func (m *mockConn) LocalAddr() net.Addr                { return nil }
func (m *mockConn) RemoteAddr() net.Addr               { return nil }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }
