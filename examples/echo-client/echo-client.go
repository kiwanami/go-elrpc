package main

import (
	"fmt"

	"github.com/kiwanami/go-elrpc"
)

func main() {
	sv, err := elrpc.StartProcess([]string{"./echo"}, nil)
	if err != nil {
		fmt.Println("Could not start epc process")
		return
	}
	defer sv.Stop()

	ret, err := sv.Call("echo", 1)
	if err != nil {
		fmt.Printf("Error: %v\n", err.Error())
		return
	}

	fmt.Printf("Echo return: %v\n", ret)
}
