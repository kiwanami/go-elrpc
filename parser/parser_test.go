package parser

import (
	"fmt"
	"testing"

	"github.com/k0kubun/pp"
	"github.com/sergi/go-diff/diffmatchpatch"
)

/// constructor short-cuts

func _int(v string) *SExpInt {
	return &SExpInt{literal: v}
}

func _float(v string) *SExpFloat {
	return &SExpFloat{literal: v}
}

func _symbol(v string) *SExpSymbol {
	return &SExpSymbol{literal: v}
}

func _char(v string) *SExpChar {
	return &SExpChar{literal: v}
}

func _string(v string) *SExpString {
	return &SExpString{literal: v}
}

func _nil() *SExpNil {
	return &SExpNil{}
}

func _cons(v1 SExp, v2 SExp) *SExpCons {
	return &SExpCons{car: v1, cdr: v2}
}

func _list(vs []SExp) *SExpList {
	return &SExpList{elements: vs}
}

func _listv(vs ...SExp) *SExpList {
	return &SExpList{elements: vs}
}

func _vectorv(vs ...SExp) *SExpVector {
	return &SExpVector{elements: vs}
}

func _dotlist(vs []SExp, v2 SExp) *SExpListDot {
	return &SExpListDot{elements: vs, last: v2}
}

func _q(v SExp) *SExpQuoted {
	return &SExpQuoted{sexp: v}
}

func _qq(v SExp) *SExpQuasiQuoted {
	return &SExpQuasiQuoted{sexp: v}
}

func _qf(v SExp) *SExpQuoted {
	return &SExpQuoted{sexp: v, function: true}
}

func _unq(v SExp) *SExpUnquote {
	return &SExpUnquote{sexp: v, splice: false}
}

func _unqs(v SExp) *SExpUnquote {
	return &SExpUnquote{sexp: v, splice: true}
}

/// test utils

func compareString(t *testing.T, msg string, v1 SExp, v2 SExp) {
	pp.ColoringEnabled = false
	s1 := pp.Sprintf("%v", v1)
	s2 := pp.Sprintf("%v", v2)
	if s1 != s2 {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(s1, s2, false)
		t.Errorf("%s: %s", msg, dmp.DiffPrettyText(diffs))
	}
}

func testSExp(t *testing.T, msg string, src string, exp SExp) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println("Parser Panic: " + msg + " [" + src + "]")
		}
	}()
	res, _ := Parse(src)
	compareString(t, msg, res[0], exp)
}

/// Test code

func TestNormal(t *testing.T) {
	src := "(1 2 ) (3 4)"
	res, _ := Parse(src)
	//pp.Println(res)
	exp := []SExp{
		_listv(_int("1"), _int("2")),
		_listv(_int("3"), _int("4")),
	}
	//pp.Println(exp)
	compareString(t, "Normal 1", res[0], exp[0])
	compareString(t, "Normal 2", res[1], exp[1])
}

type srcexp struct {
	src string
	exp SExp
}

func TestListStructures(t *testing.T) {
	data := map[string]srcexp{
		"nil list": srcexp{"()", _nil()},
		"list1":    srcexp{"(1)", _listv(_int("1"))},
		"list2": srcexp{"(1 2)",
			_listv(_int("1"), _int("2"))},
		"nest list1": srcexp{"(1 (2 3) 4)",
			_listv(_int("1"), _listv(_int("2"), _int("3")), _int("4"))},
		"nest list2": srcexp{"(((1)))",
			_listv(_listv(_listv(_int("1"))))},
		"type values": srcexp{`(1 'a "b" ())`,
			_listv(_int("1"), _q(_symbol("a")), _string("b"), _nil())},
		"calc terms": srcexp{"(+ 1 2 (- 2 (* 3 4)))",
			_listv(_symbol("+"), _int("1"), _int("2"),
				_listv(_symbol("-"), _int("2"),
					_listv(_symbol("*"), _int("3"), _int("4"))))},
		"reverse cons list": srcexp{"(((1.0) 0.2) 3.4e+4)",
			_listv(_listv(_listv(_float("1.0")), _float("0.2")), _float("3.4e+4"))},
		"cons cell": srcexp{"(1 . 2)",
			_cons(_int("1"), _int("2"))},
		"dot list": srcexp{"(1 2 . 3)",
			_dotlist([]SExp{_int("1"), _int("2")}, _int("3"))},
	}
	for k, v := range data {
		testSExp(t, k, v.src, v.exp)
	}
}

func TestString1(t *testing.T) {
	src := `"test
string
literal"`
	res, err := Parse(src)
	if err != nil && len(res) != 1 {
		pp.Println(err)
		t.Error("Could not parse string literal")
	} else {
		//pp.Println(res)
	}
}

func TestComment1Simple(t *testing.T) {
	src := `;comment`
	res, ok := Parse(src)
	if ok != nil || len(res) > 0 {
		pp.Println(res)
		t.Error("Could not skip comment")
	}
}

func TestComment2SkipEOLComment(t *testing.T) {
	src := `(1 2) ;comment`
	res, ok := Parse(src)
	if ok != nil || len(res) != 1 {
		pp.Println(res)
		t.Error("Could not skip comment")
	}
}

func TestComment3SkipCommentsMultiLines(t *testing.T) {
	src := `(1 2) ;comment
;;; next comment
(4 5)`
	res, ok := Parse(src)
	if ok != nil || len(res) != 2 {
		pp.Println(res)
		t.Error("Could not skip comment")
	}
}

func testError(t *testing.T, msg string, src string, eline int, ecol int) {
	res, err := Parse(src)
	if err != nil {
		line := err.Line
		col := err.Col
		if line == eline && col == ecol {
			// OK
			//fmt.Println(err.text)
		} else {
			pp.Println(err)
			t.Errorf("%s: Error should return line=%d and col=%d, but line=%d and col=%d.", msg, eline, ecol, line, col)
		}
	} else {
		pp.Println(res)
		t.Errorf("%s: Error should be returned.", msg)
	}
}

func TestError1(t *testing.T) {
	testError(t, "SyntaxError1", ")(1 2", 1, 1)
	testError(t, "SyntaxError2", "(1 2 3", 1, 7)
}

func TestError2(t *testing.T) {
	testError(t, "SyntaxError3", `(1 2 
3 4
5 )) 4`, 3, 4)
}

func TestQuasiQuote(t *testing.T) {
	data := map[string]srcexp{
		"quasiquoted list1": srcexp{"`(1 ,a)",
			_qq(_listv(_int("1"), _unq(_symbol("a"))))},
		"quasiquoted list2": srcexp{"`(1 ,@ab)",
			_qq(_listv(_int("1"), _unqs(_symbol("ab"))))},
		"quasiquoted nested list": srcexp{"`(1 ,(+ 1 2))",
			_qq(_listv(_int("1"), _unq(_listv(_symbol("+"), _int("1"), _int("2")))))},
	}
	for k, v := range data {
		testSExp(t, k, v.src, v.exp)
	}
}

func TestVectorLiteral(t *testing.T) {
	data := map[string]srcexp{
		"null vector": srcexp{"[]",
			_vectorv()},
		"vector": srcexp{"[1 2 3]",
			_vectorv(_int("1"), _int("2"), _int("3"))},
	}
	for k, v := range data {
		testSExp(t, k, v.src, v.exp)
	}
}

func TestSymbols(t *testing.T) {
	data := map[string]srcexp{
		"symbol": srcexp{`'(| (1 2) - (3 4))`,
			_q(_listv(_symbol("|"), _listv(_int("1"), _int("2")),
				_symbol("-"), _listv(_int("3"), _int("4"))))},
		"function symbol": srcexp{`(#'funcname)`,
			_listv(_qf(_symbol("funcname")))},
	}
	for k, v := range data {
		testSExp(t, k, v.src, v.exp)
	}
}

func TestCharLiteral(t *testing.T) {
	data := map[string]srcexp{
		"char normal": srcexp{"?x",
			_char("x")},
		"char space": srcexp{"? ",
			_char(" ")},
		"char escape1": srcexp{"?\\n",
			_char("\\n")},
		"char escape2": srcexp{"?\\(",
			_char("\\(")},
	}
	for k, v := range data {
		testSExp(t, k, v.src, v.exp)
	}
}
