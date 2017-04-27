(require 'epc)

(setq epc (epc:start-epc (expand-file-name "./calc") nil))

(deferred:$
  (epc:call-deferred epc 'addi '(10 40))
  (deferred:nextc it 
    (lambda (x) (message "Return : %S" x))))

(deferred:$
  (epc:call-deferred epc 'adds '("AA" "BB"))
  (deferred:nextc it 
    (lambda (x) (message "Return : %S" x))))

(deferred:$
  (epc:call-deferred epc 'reducei '((1 2 3 4 5.0 6 7 8 9 10) "+"))
  (deferred:nextc it 
    (lambda (x) (message "Return : %S" x))))


(message "%S" (epc:sync epc (epc:query-methods-deferred epc)))

(epc:stop-epc epc)
