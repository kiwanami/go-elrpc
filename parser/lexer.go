package parser

import (
	"fmt"
	"runtime"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Pos : Offset position of source data
type Pos int

type item struct {
	itype itemType
	pos   Pos
	val   string
}

type itemType int

// Remember UPDATE stringer code
// $ stringer -type itemType
const (
	_ itemType = iota
	itemError
	itemSpace
	itemFloat
	itemInteger
	itemSymbol
	itemString
	itemDot
	itemCharLit
	itemChar
	itemComment
	itemEOF
)

// EOF : Lexer.next() method returns this code at EOF of source data
const EOF = -1

type stateFn func(*Lexer) stateFn

// Lexer : lexer object structure
type Lexer struct {
	input   string    // the string being scanned
	state   stateFn   // the next lexing function to enter
	pos     Pos       // current position in the input
	start   Pos       // start position of this item
	width   Pos       // width of last rune read from input
	lastPos Pos       // position of most recent item returned by nextItem
	items   chan item // channel of scanned items
	result  []SExp    // parser result objects
	error   *Error    // error interface
}

type Error struct {
	error
	Msg  string // error message
	Pos  Pos    // offset position
	Line int    // line position
	Col  int    // column position
	Text string // error line
}

func (s *Lexer) Init(input string) {
	s.input = input
	s.pos = 0
	s.start = 0
	s.width = 0
	s.lastPos = 0
	s.items = make(chan item, 1)
	go s.run()
}

func (s *Lexer) next() rune {
	if int(s.pos) >= len(s.input) {
		s.width = 0
		return EOF
	}
	r, w := utf8.DecodeRuneInString(s.input[s.pos:])
	s.width = Pos(w)
	s.pos += s.width
	return r
}

func (s *Lexer) peek() rune {
	r := s.next()
	s.backup()
	return r
}

func (s *Lexer) peekNext(n int) (string, bool) {
	if (s.pos + Pos(n-1)) < Pos(len(s.input)) {
		return s.input[s.pos:(s.pos + Pos(n))], true
	}
	return "", false
}

func (s *Lexer) backup() {
	s.pos -= s.width
}

func (s *Lexer) emit(t itemType) {
	v := item{t, s.start, s.input[s.start:s.pos]}
	//pp.Printf("EMIT: %v (%v)\n", v, int(s.pos))
	s.items <- v
	s.start = s.pos
}

func (s *Lexer) ignore() {
	s.start = s.pos
}

func (s *Lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, s.next()) {
		return true
	}
	s.backup()
	return false
}

func (s *Lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, s.next()) {
	}
	s.backup()
}

func (s *Lexer) lineNumber() int {
	return 1 + strings.Count(s.input[:s.lastPos], "\n")
}

func (s *Lexer) columnNumber() int {
	ps := strings.LastIndex(s.input[:s.lastPos], "\n")
	if ps == -1 {
		return int(s.lastPos) + 1
	}
	//pp.Printf("PS: %v,  lastpos: %v\n", ps, int(s.lastPos))
	return int(s.lastPos) - ps
}

func (s *Lexer) errorLineText() string {
	var ret string
	//pp.Printf("s.lastPos: %v,  s.last: %v\n", s.lastPos, len(s.input))
	ps := strings.LastIndex(s.input[:s.lastPos], "\n")
	if ps == -1 {
		ps = 0
	}
	pe := strings.Index(s.input[s.lastPos:], "\n")
	if pe == -1 {
		pe = len(s.input)
	} else {
		pe += int(s.lastPos)
	}
	//pp.Printf("start: %v,  end:%v\n", ps, pe)
	ret = s.input[ps:pe]
	ret += "\n" + strings.Repeat(" ", s.columnNumber()-1) + "^"
	return ret

}

func (s *Lexer) errorf(format string, args ...interface{}) stateFn {
	s.items <- item{itemError, s.start, fmt.Sprintf(format, args...)}
	return nil
}

func (s *Lexer) nextItem() item {
	item := <-s.items
	s.lastPos = item.pos
	return item
}

func (s *Lexer) drain() {
	for range s.items {
	}
}

func (s *Lexer) run() {
	for s.state = scanNextAction; s.state != nil; {
		s.state = s.state(s)
	}
	s.emit(itemEOF)
	close(s.items)
}

// implementation of stateFn

func scanNextAction(s *Lexer) stateFn {
	printState()

	switch r := s.next(); {
	case r == EOF:
		return nil
	case isSpace(r):
		return scanSpace
	case r == ';':
		return scanCommentRest
	case r == '?':
		return scanCharLiteral
	case r == '.' || r == '-' || r == '+' || isDigit(r):
		s.backup()
		return scanNumberOrSymbol
	case isSymbolHead(r):
		s.backup()
		return scanSymbolRest
	case r == '"':
		return scanStringRest
	default:
		s.emit(itemChar)
		return scanNextAction
	}
}

func scanCommentRest(s *Lexer) stateFn {
	printState()

Loop:
	for {
		switch s.next() {
		case EOF:
			s.emit(itemComment)
			return nil
		case '\n':
			break Loop
		}
	}
	s.emit(itemComment)
	return scanNextAction
}

func scanCharLiteral(s *Lexer) stateFn {
	printState()

	r := s.next()
	if r == EOF {
		return nil
	}
	if r == '\\' {
		rr := s.next()
		if rr == EOF {
			return nil
		}
	}

	s.emit(itemCharLit)
	return scanNextAction
}

func scanNumberOrSymbol(s *Lexer) stateFn {
	printState()

	r := s.next()

	if r == '+' || r == '-' {
		// look ahead for number token
		ns, ok1 := s.peekNext(1)
		if !ok1 {
			return nil
		}
		nr, _ := utf8.DecodeRuneInString(ns)
		if !isDigit(nr) {
			s.emit(itemSymbol) // symbol: +/-
			return scanNextAction
		}
	} else if r == '.' {
		// look ahead for number token
		ns, ok1 := s.peekNext(1)
		if !ok1 {
			return nil
		}
		nr, _ := utf8.DecodeRuneInString(ns)
		if !isDigit(nr) {
			s.emit(itemDot) // dot .
			return scanSpace
		}
	}

	s.backup()
	return scanNumber
}

func scanSpace(s *Lexer) stateFn {
	for {
		r := s.next()
		if r == EOF {
			return nil
		}
		if !isSpace(r) {
			s.backup()
			break
		}
	}
	if s.start < s.pos {
		printState()
		s.emit(itemSpace)
	}
	return scanNextAction
}

func scanNumber(s *Lexer) stateFn {
	printState()

	item := itemInteger

	s.accept("+-")
	digits := "0123456789"
	s.acceptRun(digits)
	if s.accept(".") {
		s.acceptRun(digits)
		item = itemFloat
	}
	if s.accept("eE") {
		s.accept("+-")
		s.acceptRun(digits)
		item = itemFloat
	}

	s.emit(item)
	return scanSpace
}

func scanSymbolRest(s *Lexer) stateFn {
	printState()

Loop:
	for {
		r := s.next()
		switch r {
		case '\\':
			if rr := s.next(); rr != EOF {
				break
			}
			fallthrough
		case EOF:
			break Loop
		default:
			if !isSymbolRest(r) {
				s.backup()
				break Loop
			}
		}
	}
	s.emit(itemSymbol)
	return scanSpace
}

func scanStringRest(s *Lexer) stateFn {
	printState()

Loop:
	for {
		switch s.next() {
		case '\\':
			if r := s.next(); r != EOF {
				break
			}
			fallthrough
		case EOF:
			return s.errorf("Unterminated string literal")
		case '"':
			break Loop
		}
	}
	s.emit(itemString)
	return scanSpace
}

func printState() {
	fpcs := make([]uintptr, 1)
	n := runtime.Callers(2, fpcs)
	if n == 0 {
		return
	}
	fun := runtime.FuncForPC(fpcs[0] - 1)
	if fun == nil {
		return
	}
	//pp.Printf("ST: %v\n", fun.Name())
}

func isSpace(r rune) bool {
	return r == ' ' || strings.ContainsRune("\t\n\r\f", r)
}

func isDigit(r rune) bool {
	return unicode.IsDigit(r)
}

func isAlphabet(r rune) bool {
	return unicode.IsLetter(r)
}

func isSymbolHead(r rune) bool {
	return isAlphabet(r) || strings.ContainsRune("\\+-*/_~!$%^&=:<>{}.|", r) || r > 255
}

func isSymbolRest(r rune) bool {
	return isAlphabet(r) || isDigit(r) || strings.ContainsRune("+-*/_~!$%^&=:<>{}.|", r) || r > 255
}
