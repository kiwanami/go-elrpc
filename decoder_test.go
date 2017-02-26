package elrpc

import (
	"fmt"
	"testing"

	"reflect"

	"github.com/k0kubun/pp"
	ps "github.com/kiwanami/go-elrpc/parser"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func testCompareSEexp(t *testing.T, src string, expected ps.SExp) {
	sexps, err := DecodeToSExp(src)
	if err != (error)(nil) {
		pp.Println(err)
		t.Error(err)
		return
	}

	//pp.Println(sexps)

	tsrc := sexps[0].ToSExpString()
	esrc := expected.ToSExpString()
	if tsrc != esrc {
		t.Errorf("Not equal exp:[%s] -> result:[%s]", esrc, tsrc)
	}
}

func TestDecodeSExp1(t *testing.T) {
	src := "(+ 1 2 (- 2 (* 3 4)))"

	exp := ps.AstListv(ps.AstSymbol("+"), ps.AstInt("1"), ps.AstInt("2"),
		ps.AstListv(ps.AstSymbol("-"), ps.AstInt("2"),
			ps.AstListv(ps.AstSymbol("*"), ps.AstInt("3"), ps.AstInt("4"))))

	testCompareSEexp(t, src, exp)
}

type srcdata struct {
	src string
	exp interface{}
}

func compareObjectString(t *testing.T, msg string, v1 interface{}, v2 interface{}) {
	pp.ColoringEnabled = false
	s1 := pp.Sprintf("%v", v1)
	s2 := pp.Sprintf("%v", v2)
	if s1 != s2 {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(s1, s2, false)
		t.Errorf("%s: %s", msg, dmp.DiffPrettyText(diffs))
	}
}

func testDecodeObject(t *testing.T, msg string, src string, exp interface{}) {
	defer func() {
		if e := recover(); e != nil {
			pp.Println(e)
			t.Errorf("Parser Panic: %s [%s] , %v", msg, src, e)
		}
	}()
	res, _ := Decode1(src)
	//aa, _ := DecodeToSExp(src)
	//pp.Println(aa[0])
	compareObjectString(t, msg, res, exp)
}

func TestPrimitives1(t *testing.T) {
	data := map[string]srcdata{
		"nil":      srcdata{"nil", nil},
		"nil list": srcdata{"()", nil},
		"list1":    srcdata{"(1)", []int{1}},
		"list2":    srcdata{"(1 2)", []int{1, 2}},
		"nest list1": srcdata{"(1 (2 3) 4)",
			[]interface{}{1, []int{2, 3}, 4}},
		"nest list2": srcdata{"(((1)))",
			[]interface{}{[]interface{}{[]int{1}}}},
		"type values": srcdata{`(1 'a "b" ())`,
			[]interface{}{1, "a", "b", nil}},
		"reverse cons list": srcdata{"(((1.0) 0.2) 3.4e+4)",
			[]interface{}{[]interface{}{[]float64{1.0}, 0.2}, 3.4e+4}},
		"cons cell": srcdata{"(1 . 2)",
			[]int{1, 2}},
		"dot list": srcdata{"(1 2 . 3)",
			[]int{1, 2, 3}},
	}
	for k, v := range data {
		testDecodeObject(t, k, v.src, v.exp)
	}
}

/// type conversion

type srcconv struct {
	src interface{}
	exp reflect.Type
}

func makeConv(a interface{}, prot interface{}) srcconv {
	return srcconv{a, reflect.TypeOf(prot)}
}

func TestConvert1(t *testing.T) {
	data := map[string]srcconv{
		"int->int":     makeConv(1, 1),
		"int->int8":    makeConv(1, (int8)(1)),
		"int->int16":   makeConv(1, (int16)(1)),
		"int->int64":   makeConv(1, (int64)(1)),
		"int->uint8":   makeConv(1, (uint8)(1)),
		"int->uint16":  makeConv(1, (uint16)(1)),
		"int->uint64":  makeConv(1, (uint64)(1)),
		"int->float32": makeConv(1, (float32)(1)),
		"int->float64": makeConv(1, (float64)(1)),
		"float32->int": makeConv((float32)(1), 1),
		"float64->int": makeConv((float64)(1), 1),
	}
	for k, v := range data {
		testConvertObject(t, k, v.src, v.exp)
	}
}

func testConvertObject(t *testing.T, key string, src interface{}, exp reflect.Type) {
	val := reflect.ValueOf(src)
	cval := val.Convert(exp)
	sval := fmt.Sprintf("%v", val.Interface())
	scval := fmt.Sprintf("%v", cval.Interface())
	if sval != scval {
		t.Errorf("wrong convert [%s]: %v -> %v", key, sval, scval)
	}
}

func TestConvertArray1(t *testing.T) {
	srcVal1 := []interface{}{1, 2, 3}
	targetType := reflect.TypeOf([]int{})

	cval2, err := ConvertArrayType(targetType, reflect.ValueOf(srcVal1))
	if err != nil {
		t.Errorf("wrong array convert: %v -> %v", srcVal1, err)
	}
	for i := 0; i < len(srcVal1); i++ {
		sval1 := fmt.Sprintf("%v", srcVal1[i])
		sval2 := fmt.Sprintf("%v", cval2.Index(i))
		if sval1 != sval2 {
			t.Errorf("wrong array element convert [%d]: %v -> %v", i, sval1, sval2)
		}
	}
}

func TestConvertDArray2(t *testing.T) {
	srcVal1 := []interface{}{
		[]int{1, 2, 3},
		[]int{4, 5, 6},
	}
	targetType := reflect.TypeOf([][]int{})

	cval2, err := ConvertType(targetType, reflect.ValueOf(srcVal1))
	if err != nil {
		t.Errorf("wrong array convert: %v -> %v", srcVal1, err)
	}
	for i := 0; i < len(srcVal1); i++ {
		sval1 := fmt.Sprintf("%v", srcVal1[i])
		sval2 := fmt.Sprintf("%v", cval2.Index(i))
		if sval1 != sval2 {
			t.Errorf("wrong array element convert [%d]: %v -> %v", i, sval1, sval2)
		}
	}
}
