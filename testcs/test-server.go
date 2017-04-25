package main

import (
	"flag"
	"fmt"

	"os"

	"github.com/kiwanami/go-elrpc"
)

func main() {
	port := flag.Int("port", 0, "port number")
	debug := flag.Bool("debug", false, "debug")
	flag.Parse()

	// construct epc server
	s, err := elrpc.StartServerWithPort(nil, *port)
	if *debug {
		s.SetDebug(true)
	}
	if err != nil {
		fmt.Printf(err.Error())
	}

	// register methods

	// no-arg-return method
	s.RegisterMethod(elrpc.MakeMethod("hello", func() {
		fmt.Println("hello!")
	}, "", "print hello"))

	// echo method
	s.RegisterMethod(elrpc.MakeMethod("echo", func(arg interface{}) interface{} {
		return arg
	}, "any", "return the given value"))

	// add methods
	s.RegisterMethod(elrpc.MakeMethod("addi", func(a int, b int) int {
		return a + b
	}, "int, int", "add integers"))
	s.RegisterMethod(elrpc.MakeMethod("adds", func(a string, b string) string {
		return a + b
	}, "string, string", "concat string"))

	// many argument types
	s.RegisterMethod(elrpc.MakeMethod("format", func(format string, a int, b float64) string {
		return fmt.Sprintf(format, a, b)
	}, "string, int, float", "format values"))

	// array types
	s.RegisterMethod(elrpc.MakeMethod("mapi", func(lst []int, sc int) []int {
		ret := make([]int, len(lst))
		for i, v := range lst {
			ret[i] = v * sc
		}
		return ret
	}, "[]int, int -> []int", "multiply over int array"))
	s.RegisterMethod(elrpc.MakeMethod("flatmapi", func(lst [][]int, sc float64) []float64 {
		ret := make([]float64, 0)
		for i, vs := range lst {
			for j := range vs {
				ret = append(ret, (float64)(lst[i][j])*sc)
			}
		}
		return ret
	}, "[][]int, float -> []float", "flatmap"))

	// error test methods
	s.RegisterMethod(elrpc.MakeMethod("num-error", func(a int) {
		fmt.Printf("%v\n", 1.0/a)
	}, "float", "raise div by zero error"))
	s.RegisterMethod(elrpc.MakeMethod("panic-error", func() {
		panic("!! panic error !!")
	}, "", "panic error"))
	s.RegisterMethod(elrpc.MakeMethod("serialize-error", func() interface{} {
		return make(chan string, 1)
	}, "", "panic error"))
	s.RegisterMethod(elrpc.MakeMethod("killme", func() {
		os.Exit(0)
	}, "", "exit server"))

	// accept for peer's connection
	if *debug {
		fmt.Println("Server started.")
	}
	s.Wait()
	if *debug {
		fmt.Print("OK.")
	}
}
