package parser

import (
	"fmt"
	"testing"

	"github.com/k0kubun/pp"
	"github.com/sergi/go-diff/diffmatchpatch"
)

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
		AstListv(AstInt("1"), AstInt("2")),
		AstListv(AstInt("3"), AstInt("4")),
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
		"nil list": srcexp{"()", AstNil()},
		"list1":    srcexp{"(1)", AstListv(AstInt("1"))},
		"list2": srcexp{"(1 2)",
			AstListv(AstInt("1"), AstInt("2"))},
		"nest list1": srcexp{"(1 (2 3) 4)",
			AstListv(AstInt("1"), AstListv(AstInt("2"), AstInt("3")), AstInt("4"))},
		"nest list2": srcexp{"(((1)))",
			AstListv(AstListv(AstListv(AstInt("1"))))},
		"type values": srcexp{`(1 'a "b" ())`,
			AstListv(AstInt("1"), AstQ(AstSymbol("a")), AstString("b"), AstNil())},
		"calc terms": srcexp{"(+ 1 2 (- 2 (* 3 4)))",
			AstListv(AstSymbol("+"), AstInt("1"), AstInt("2"),
				AstListv(AstSymbol("-"), AstInt("2"),
					AstListv(AstSymbol("*"), AstInt("3"), AstInt("4"))))},
		"reverse cons list": srcexp{"(((1.0) 0.2) 3.4e+4)",
			AstListv(AstListv(AstListv(AstFloat("1.0")), AstFloat("0.2")), AstFloat("3.4e+4"))},
		"cons cell": srcexp{"(1 . 2)",
			AstCons(AstInt("1"), AstInt("2"))},
		"dot list": srcexp{"(1 2 . 3)",
			AstDotlist([]SExp{AstInt("1"), AstInt("2")}, AstInt("3"))},
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
			AstQq(AstListv(AstInt("1"), AstUnq(AstSymbol("a"))))},
		"quasiquoted list2": srcexp{"`(1 ,@ab)",
			AstQq(AstListv(AstInt("1"), AstUnqs(AstSymbol("ab"))))},
		"quasiquoted nested list": srcexp{"`(1 ,(+ 1 2))",
			AstQq(AstListv(AstInt("1"), AstUnq(AstListv(AstSymbol("+"), AstInt("1"), AstInt("2")))))},
	}
	for k, v := range data {
		testSExp(t, k, v.src, v.exp)
	}
}

func TestVectorLiteral(t *testing.T) {
	data := map[string]srcexp{
		"null vector": srcexp{"[]",
			AstVectorv()},
		"vector": srcexp{"[1 2 3]",
			AstVectorv(AstInt("1"), AstInt("2"), AstInt("3"))},
	}
	for k, v := range data {
		testSExp(t, k, v.src, v.exp)
	}
}

func TestSymbols(t *testing.T) {
	data := map[string]srcexp{
		"symbol": srcexp{`'(| (1 2) - (3 4))`,
			AstQ(AstListv(AstSymbol("|"), AstListv(AstInt("1"), AstInt("2")),
				AstSymbol("-"), AstListv(AstInt("3"), AstInt("4"))))},
		"function symbol": srcexp{`(#'funcname)`,
			AstListv(AstQf(AstSymbol("funcname")))},
	}
	for k, v := range data {
		testSExp(t, k, v.src, v.exp)
	}
}

func TestCharLiteral(t *testing.T) {
	data := map[string]srcexp{
		"char normal": srcexp{"?x",
			AstChar("x")},
		"char space": srcexp{"? ",
			AstChar(" ")},
		"char escape1": srcexp{"?\\n",
			AstChar("\\n")},
		"char escape2": srcexp{"?\\(",
			AstChar("\\(")},
	}
	for k, v := range data {
		testSExp(t, k, v.src, v.exp)
	}
}
