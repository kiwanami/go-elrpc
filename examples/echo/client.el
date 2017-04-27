(require 'epc)

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
