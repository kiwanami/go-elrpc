package elrpc

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

func withProcess(args []string, f func(cmd *exec.Cmd) error) error {
	// start with port num
	args = append([]string{"run"}, args...)
	cmd := exec.Command("go", args...)
	// create a new process group for our child by setting the Setpgid field
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	defer func() {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGINT)
		_ = cmd.Wait()
	}()
	return f(cmd)
}

func TestStartProcess1(t *testing.T) {
	err := withProcess([]string{"testcs/test-server.go", "-port", "8888"}, func(cmd *exec.Cmd) error {
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			t.Error(err.Error())
		}
		lineBuf := bufio.NewScanner(stdout)
		_ = cmd.Start()
		if !lineBuf.Scan() {
			t.Error("could not scan port line")
		}
		line := lineBuf.Text()
		pn, err := strconv.Atoi(line)
		if err != nil || pn != 8888 {
			t.Errorf("wrong port number : %v / %v", line, err)
		}
		return nil
	})
	if err != nil {
		t.Error(err.Error())
	}
}

func TestStartProcess2(t *testing.T) {
	err := withProcess([]string{"testcs/test-server.go"}, func(cmd *exec.Cmd) error {
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			t.Error(err.Error())
		}
		lineBuf := bufio.NewScanner(stdout)
		_ = cmd.Start()
		if !lineBuf.Scan() {
			t.Error("could not scan port line")
		}
		line := lineBuf.Text()
		pn, err := strconv.Atoi(line)
		if err != nil || pn < 1024 {
			t.Errorf("wrong port number : %v [%d] / %v", line, pn, err)
		}
		return nil
	})
	if err != nil {
		t.Error(err.Error())
	}
}

func withEPC(progname string, debug bool, f func(cl Service) error) error {
	cmds := []string{progname}
	if debug {
		cmds = append(cmds, "-debug")
	}
	return withProcess(cmds, func(cmd *exec.Cmd) error {
		//fmt.Println("## with epc ")
		// relay stdout
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}
		lineBuf := bufio.NewScanner(stdout)

		// relay stderr
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return err
		}
		errBuf := bufio.NewScanner(stderr)

		_ = cmd.Start()
		if !lineBuf.Scan() {
			return fmt.Errorf("could not scan port line")
		}
		// get peer's port number
		line := lineBuf.Text()
		//fmt.Println("## port number " + line)
		pn, err := strconv.Atoi(line)
		if err != nil {
			return fmt.Errorf("wrong port number : %v [%d] / %v", line, pn, err)
		}
		go func() {
			//fmt.Println("SVOUT: start")
			for lineBuf.Scan() {
				a := lineBuf.Text()
				if debug {
					fmt.Printf("SVOUT: %s\n", a)
				}
			}
			//fmt.Println("SVOUT: exited.")
		}()
		go func() {
			//fmt.Println("SVERR: start")
			for errBuf.Scan() {
				a := errBuf.Text()
				if debug {
					fmt.Printf("SVERR: %s\n", a)
				}
			}
			//fmt.Println("SVERR: exited.")
		}()
		// start client
		time.Sleep(10 * time.Millisecond)
		//fmt.Println("## start client")
		cl, err := StartClient(pn, nil)
		if err != nil {
			return err
		}
		cl.SetDebug(debug)
		defer cl.Stop()
		if debug {
			cl.SetDebug(true)
		}
		// eval test code
		//fmt.Println("## start test")
		err = f(cl)
		if err != nil {
			return err
		}
		//fmt.Println("## done test")
		return nil
	})
}

func TestEpcHello1(t *testing.T) {
	err := withEPC("testcs/test-server.go", false, func(cl Service) error {
		// runtime error
		_, err := cl.Call("hello")
		if err != nil {
			t.Errorf(err.Error())
		}
		return nil
	})
	if err != nil {
		t.Error(err.Error())
	}
}

func TestEpcEcho1(t *testing.T) {
	err := withEPC("testcs/test-server.go", false, func(cl Service) error {
		ret, err := cl.Call("echo", "hello")
		if err != nil {
			return err
		}
		str := ret.(string)
		if str != "hello" {
			t.Errorf("expected[%s] but returned [%s]", "hello", str)
		}
		ret, err = cl.Call("echo", 12345)
		if err != nil {
			return err
		}
		i := ret.(int)
		if i != 12345 {
			t.Errorf("expected[%v] but returned [%v]", 12345, i)
		}
		return nil
	})
	if err != nil {
		t.Error(err.Error())
	}
}

func TestEpcAdd1(t *testing.T) {
	err := withEPC("testcs/test-server.go", false, func(cl Service) error {
		// add integers
		ret, err := cl.Call("addi", 2, 3)
		if err != nil {
			return err
		}
		reti := ret.(int)
		if reti != 5 {
			t.Errorf("expected[%v] but returned [%v]", 5, reti)
		}
		// add strings
		ret, err = cl.Call("adds", "A", "B")
		if err != nil {
			return err
		}
		rets := ret.(string)
		if rets != "AB" {
			t.Errorf("expected[%v] but returned [%v]", "AB", rets)
		}
		return nil
	})
	if err != nil {
		t.Error(err.Error())
	}
}

func TestManyType1(t *testing.T) {
	err := withEPC("testcs/test-server.go", false, func(cl Service) error {
		ret, err := cl.Call("format", ">> %d %0.2f", 3, 3.14)
		if err != nil {
			return err
		}
		exp := ">> 3 3.14"
		reti := ret.(string)
		if reti != exp {
			t.Errorf("expected[%v] but returned [%v]", exp, reti)
		}
		return nil
	})
	if err != nil {
		t.Error(err.Error())
	}
}

func TestMap1(t *testing.T) {
	err := withEPC("testcs/test-server.go", false, func(cl Service) error {
		ret, err := cl.Call("mapi", []int{1, 2, 3}, 10)
		if err != nil {
			return err
		}
		exp := []int{10, 20, 30}
		reti := ret.([]int)
		if !reflect.DeepEqual(reti, exp) {
			t.Errorf("expected[%v] but returned [%v]", exp, reti)
		}
		return nil
	})
	if err != nil {
		t.Error(err.Error())
	}
}

func TestFlatmap1(t *testing.T) {
	err := withEPC("testcs/test-server.go", false, func(cl Service) error {
		arg := [][]int{
			[]int{1, 2, 3},
			[]int{4, 5, 6, 7, 8},
			[]int{9, 10},
		}
		ret, err := cl.Call("flatmapi", arg, 10.0)
		if err != nil {
			return err
		}
		exp := []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
		reti := ret.([]int)
		if !reflect.DeepEqual(reti, exp) {
			t.Errorf("expected[%v] but returned [%v]", exp, reti)
		}
		return nil
	})
	if err != nil {
		t.Error(err.Error())
	}
}

func TestEpcError1(t *testing.T) {
	err := withEPC("testcs/test-server.go", false, func(cl Service) error {
		// runtime error
		_, err := cl.Call("num-error", 0.0)
		if err != nil {
			s := err.Error()
			if !strings.Contains(s, "runtime error: integer divide by zero") {
				t.Errorf("dividing by zero -> %v", s)
			}
		} else {
			t.Error("error should be returned")
		}

		// panic error
		_, err = cl.Call("panic-error")
		if err != nil {
			s := err.Error()
			if !strings.Contains(s, "!! panic error !!") {
				t.Errorf("panic -> %v", s)
			}
		} else {
			t.Error("error should be returned")
		}

		// serialize error (client side)
		_, err = cl.Call("echo", make(chan int, 1))
		if err != nil {
			s := err.Error()
			if !strings.Contains(s, "unsupported type: chan int") {
				t.Errorf("serialize -> %v", s)
			}
		} else {
			t.Error("error should be returned")
		}

		// serialize error (peer side)
		_, err = cl.Call("serialize-error")
		if err != nil {
			s := err.Error()
			if !strings.Contains(s, "unsupported type: chan string") {
				t.Errorf("serialize -> %v", s)
			}
		} else {
			t.Error("error should be returned")
		}

		// unexpected peer's shutdown
		_, err = cl.Call("killme")
		if err != nil {
			s := err.Error()
			if !strings.Contains(s, "unexpected peer's shutdown") {
				t.Errorf("shutdown error -> %v", s)
			}
		} else {
			t.Error("error should be returned")
		}

		// peer has gone
		_, err = cl.Call("echo", 1)
		if err != nil {
			s := err.Error()
			if !strings.Contains(s, "epc not connected") {
				t.Errorf("connection closed -> %v", s)
			}
		} else {
			t.Error("error should be returned")
		}

		return nil
	})
	if err != nil {
		t.Error(err.Error())
	}
}

func TestEpcQuery1(t *testing.T) {
	err := withEPC("testcs/test-server.go", false, func(cl Service) error {
		ms, err := cl.QueryMethods()
		if err != nil {
			return err
		}
		if len(ms) != 12 {
			t.Errorf("expected[%d] but returned [%d]", 12, len(ms))
		}
		mdm := make(map[string]*MethodDesc)
		for _, md := range ms {
			mdm[md.Name] = md
		}
		if md := mdm["addi"]; md.Argdoc != "int, int" {
			t.Errorf("expected[%v] but returned [%v]", "int, int", md.Argdoc)
		}
		if md := mdm["adds"]; md.Docstring != "concat string" {
			t.Errorf("expected[%v] but returned [%v]", "concat string", md.Docstring)
		}

		return nil
	})
	if err != nil {
		t.Error(err.Error())
	}
}

func TestEpcConcurrency(t *testing.T) {
	err := withEPC("testcs/test-server.go", false, func(cl Service) error {
		rand.Seed(1)
		loopnum := 10000
		var errors = []error{}
		var wg sync.WaitGroup
		for i := 0; i < loopnum; i++ {
			wg.Add(1)
			go func() {
				dur := rand.Intn(100) // msec
				ret, err := cl.Call("sleep", dur)
				if err != nil {
					errors = append(errors, err)
				}
				if ret != dur {
					t.Errorf("Not same result: %v  ->  %v", ret, dur)
				}
				wg.Done()
			}()
		}
		wg.Wait()
		return nil
	})
	if err != nil {
		t.Error(err.Error())
	}
}

func TestEpcEcho2(t *testing.T) {
	err := withEPC("testcs/test-server.go", false, func(cl Service) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		ret, err := cl.CallContext(ctx, "echo", "hello")
		if err != nil {
			return err
		}
		str := ret.(string)
		if str != "hello" {
			t.Errorf("expected[%s] but returned [%s]", "hello", str)
		}
		ret, err = cl.Call("echo", 12345)
		if err != nil {
			return err
		}
		i := ret.(int)
		if i != 12345 {
			t.Errorf("expected[%v] but returned [%v]", 12345, i)
		}
		return nil
	})
	if err != nil {
		t.Error(err.Error())
	}
}

func TestEpcCancel(t *testing.T) {
	err := withEPC("testcs/test-server.go", false, func(cl Service) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		back := make(chan string, 1)
		go func() {
			_, err := cl.CallContext(ctx, "sleep", 2000)
			if err.Error() == "Canceled" {
				back <- "OK"
			} else {
				back <- "NG: Not canceled"
			}
		}()
		for i := 0; i < 30; i++ {
			time.Sleep(time.Duration(20) * time.Millisecond)
			wn := cl.WaitingSessionNum()
			if wn == 1 {
				break
			}
			if i > 20 {
				return fmt.Errorf("Wrong waiting sessions: %v != 1", wn)
			}
		}
		cancel()
		time.Sleep(time.Duration(100) * time.Millisecond) // wait for cancel msg
		ret := <-back
		if ret != "OK" {
			return fmt.Errorf("Could not canceled")
		}
		if wn := cl.WaitingSessionNum(); wn != 0 {
			return fmt.Errorf("Wrong waiting sessions: %v != 0", wn)
		}
		return nil
	})
	if err != nil {
		t.Error(err.Error())
	}
}
