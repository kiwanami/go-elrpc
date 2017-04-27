package main

import (
	"fmt"

	"github.com/kiwanami/go-elrpc"
)

func main() {
	// construct epc server
	s, err := elrpc.StartServerWithPort(nil, 0)
	if err != nil {
		fmt.Printf(err.Error())
		return
	}

	// register methods
	s.RegisterMethod(elrpc.MakeMethod("addi", func(a int, b int) int {
		return a + b
	}, "int, int", "add integers"))

	s.RegisterMethod(elrpc.MakeMethod("adds", func(a string, b string) string {
		return a + b
	}, "string, string", "concat strings"))

	s.RegisterMethod(elrpc.MakeMethod("reducei", func(ls []int, op string) int64 {
		ret := (int64)(0)
		switch op {
		case "+":
			for _, v := range ls {
				ret += (int64)(v)
			}
		case "*":
			for _, v := range ls {
				ret *= (int64)(v)
			}
		default:
			for _, v := range ls {
				ret += (int64)(v)
			}
		}
		return ret
	}, "(list []int, op string) -> int", "calculate and reduce list elements to a result value"))

	// wait for client connections
	s.Wait()
}
