package parser

import (
	"testing"

	"github.com/k0kubun/pp"
)

func runScan(v string) item {
	s := Lexer{}
	s.Init(v)
	return s.nextItem()
}

func runSeq(v string) []item {
	s := Lexer{}
	s.Init(v)
	items := make([]item, 0)
	for {
		i := s.nextItem()
		if i.itype == itemEOF {
			break
		}
		items = append(items, i)
		if i.itype == itemError {
			break
		}
	}
	return items
}

func checkSeq(t *testing.T, v string, is []item, ts []itemType) {
	for i := 0; i < len(is); i++ {
		if is[i].itype != ts[i] {
			t.Errorf("[%v]:%d  itemType: %s is not %s", v, i, is[i].itype, ts[i])
		}
	}
}

func testItem(t *testing.T, i item, itype itemType, val string, pos Pos) {
	if i.itype != itype {
		pp.Println(i)
		t.Errorf("itemType: %s is not %s", i.itype.String(), itype)
	}
	if i.val != val {
		t.Errorf("itemVal: [%s] is not [%s]", i.val, val)
	}
	if i.pos != Pos(0) {
		t.Errorf("itemPos: %d is not %d", i.pos, pos)
	}
}

func TestLexerInit(t *testing.T) {
	s := Lexer{}
	s.Init("ABCD")
	s.nextItem()
}

func TestInteger(t *testing.T) {
	i := runScan("1234")
	testItem(t, i, itemInteger, "1234", 0)
	i = runScan("1234")
	testItem(t, i, itemInteger, "1234", 1)
	i = runScan("1")
	testItem(t, i, itemInteger, "1", 1)
	i = runScan("-1")
	testItem(t, i, itemInteger, "-1", 1)
	i = runScan("+1")
	testItem(t, i, itemInteger, "+1", 0)
}

func TestFloat(t *testing.T) {
	i := runScan("123.4")
	testItem(t, i, itemFloat, "123.4", 0)
	i = runScan("0.12")
	testItem(t, i, itemFloat, "0.12", 0)
	i = runScan(".12")
	testItem(t, i, itemFloat, ".12", 0)
	i = runScan("1.12e-3")
	testItem(t, i, itemFloat, "1.12e-3", 0)
	i = runScan("1.12e3")
	testItem(t, i, itemFloat, "1.12e3", 0)
	i = runScan("-1.12")
	testItem(t, i, itemFloat, "-1.12", 0)
}

func TestSymbol(t *testing.T) {
	i := runScan(`abcd`)
	testItem(t, i, itemSymbol, `abcd`, 0)
	i = runScan(`a:b2-c/d`)
	testItem(t, i, itemSymbol, `a:b2-c/d`, 0)
	i = runScan(`\.file`)
	testItem(t, i, itemSymbol, `\.file`, 0)
}

func TestString(t *testing.T) {
	i := runScan(`"abcd"`)
	testItem(t, i, itemString, `"abcd"`, 0)
	i = runScan(`"aa\naa\"bb\\cc"`)
	testItem(t, i, itemString, `"aa\naa\"bb\\cc"`, 0)
}

func TestChars(t *testing.T) {
	i := runScan("()")
	testItem(t, i, itemChar, "(", 0)
	i = runScan(`'`)
	testItem(t, i, itemChar, `'`, 0)
}

func TestQuotes(t *testing.T) {
	v := `'a '(b)`
	items := runSeq(v)
	checkSeq(t, v, items, []itemType{
		itemChar, itemSymbol, itemSpace,
		itemChar, itemChar, itemSymbol, itemChar,
	})
}

func TestSequence(t *testing.T) {
	v := `(1 2 aa "str")`
	items := runSeq(v)
	//pp.Println(items)
	checkSeq(t, v, items, []itemType{
		itemChar, itemInteger, itemSpace, itemInteger, itemSpace,
		itemSymbol, itemSpace, itemString, itemChar,
	})
}
