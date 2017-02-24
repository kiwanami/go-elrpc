package elrpc

import (
	"testing"

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
			t.Errorf("Parser Panic: " + msg + " [" + src + "]")
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
		"list1":    srcdata{"(1)", []interface{}{1}},
		"list2":    srcdata{"(1 2)", []interface{}{1, 2}},
		"nest list1": srcdata{"(1 (2 3) 4)",
			[]interface{}{1, []interface{}{2, 3}, 4}},
		"nest list2": srcdata{"(((1)))",
			[]interface{}{[]interface{}{[]interface{}{1}}}},
		"type values": srcdata{`(1 'a "b" ())`,
			[]interface{}{1, "a", "b", nil}},
		"reverse cons list": srcdata{"(((1.0) 0.2) 3.4e+4)",
			[]interface{}{[]interface{}{[]interface{}{1.0}, 0.2}, 3.4e+4}},
		"cons cell": srcdata{"(1 . 2)",
			[]interface{}{1, 2}},
		"dot list": srcdata{"(1 2 . 3)",
			[]interface{}{1, 2, 3}},
	}
	for k, v := range data {
		testDecodeObject(t, k, v.src, v.exp)
	}
}
