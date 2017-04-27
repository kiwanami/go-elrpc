# go-elrpc : EPC (RPC Stack for Emacs Lisp) for Go

EPC is an RPC stack for Emacs Lisp and `elrpc` is an implementation of EPC in Go.
Using `elrpc`, you can develop an emacs extension in Go.

- [EPC at github](https://github.com/kiwanami/emacs-epc)

## Sample Code

### Go code (server process)

This sample code defines a simple echo method.
Emacs starts this code as a child process.

`echo.go`
```go
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

	// register echo method
	s.RegisterMethod(elrpc.MakeMethod("echo", func(arg interface{}) interface{} {
		return arg
	}, "any", "return the given value"))

	// wait for client's connections
	s.Wait()
}
```

### Emacs Lisp code (client process)

This elisp code starts a child process and calls the echo method.

`echo-client.el`
```el
(require 'epc)

;; eval each s-exp one by one with `eval-last-sexp'

(setq epc (epc:start-epc (expand-file-name "./echo") nil))

(deferred:$
  (epc:call-deferred epc 'echo '(10))
  (deferred:nextc it 
    (lambda (x) (message "Return : %S" x))))

(deferred:$
  (epc:call-deferred epc 'echo '("Hello go-elrpc"))
  (deferred:nextc it 
    (lambda (x) (message "Return : %S" x))))

(epc:stop-epc epc)
```

### Go code (client process)

You can also connect between Go processes.

`echo-client.go`
```go
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
```

Here is the result.

```
$ go run echo-client.go
Echo return: 1
```

## Installation

```
$ go get github.com/kiwanami/go-elrpc
```

## License

go-elrpc is licensed under MIT.

----
(C) 2017 SAKURAI Masashi. m.sakurai at kiwanami.net
