package main

import (
	"fmt"

	"github.com/kiwanami/go-elrpc"
)

func main() {
	// construct epc server
	s, err := elrpc.StartServer(nil)
	if err != nil {
		fmt.Printf(err.Error())
		return
	}

	// register echo method
	s.RegisterMethod(elrpc.MakeMethod("echo", func(arg interface{}) interface{} {
		return arg
	}, "any", "return the given value"))

	// wait for client's connections
	s.Wait()
}
