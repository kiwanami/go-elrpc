package elrpc

import "github.com/kiwanami/go-elrpc/parser"

func Decode(sexp string) ([]interface{}, error) {
	sexps, err := DecodeToSExp(sexp)
	if err != nil {
		return nil, err
	}
	ret := make([]interface{}, len(sexp))
	for i, sexp := range sexps {
		ret[i] = sexp.ToValue()
	}
	return ret, nil
}

func Decode1(sexp string) (interface{}, error) {
	sexps, err := DecodeToSExp(sexp)
	if err != nil {
		return nil, err
	}
	if len(sexps) == 0 {
		return nil, nil
	}
	return sexps[0].ToValue(), nil
}

func DecodeToSExp(sexp string) ([]parser.SExp, error) {
	sexps, err := parser.Parse(sexp)
	if err != nil {
		return nil, err
	}
	return sexps, nil
}

/// utilities for the result not-typed object

func ToArray(o interface{}) []interface{} {
	arr, ok := o.([]interface{})
	if !ok {
		return nil
	}
	return arr
}
