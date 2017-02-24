//line sexp.go.y:2
package parser

import __yyfmt__ "fmt"

//line sexp.go.y:2
import (
	"unicode/utf8"
)

type Token struct {
	token   int
	literal string
	pos     Pos
}

//line sexp.go.y:16
type yySymType struct {
	yys   int
	token Token
	expr  SExp
	seq   []SExp
}

const INTEGER = 57346
const FLOAT = 57347
const SYMBOL = 57348
const CHARACTER = 57349
const STRING = 57350

var yyToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"INTEGER",
	"FLOAT",
	"SYMBOL",
	"CHARACTER",
	"STRING",
	"\"(\"",
	"\")\"",
	"\".\"",
	"\"[\"",
	"\"]\"",
	"\"'\"",
	"\"`\"",
	"\"#\"",
	"\",\"",
	"\"@\"",
}
var yyStatenames = [...]string{}

const yyEofCode = 1
const yyErrCode = 2
const yyInitialStackSize = 16

//line sexp.go.y:141

func (l *Lexer) Error(e string) {
	l.error = &Error{
		Msg: e, Pos: l.lastPos,
		Line: l.lineNumber(),
		Col:  l.columnNumber(),
		Text: l.errorLineText(),
	}
}

func (l *Lexer) Lex(lval *yySymType) int {
	item := l.nextItem()
	for {
		if item.itype == itemEOF {
			return 0
		}
		if item.itype != itemSpace && item.itype != itemComment {
			break
		}
		item = l.nextItem()
	}
	//fmt.Printf("]LEX: %d - %v\n", item.itype, item.val)
	tok := -1
	switch {
	case item.itype == itemInteger:
		tok = INTEGER
	case item.itype == itemFloat:
		tok = FLOAT
	case item.itype == itemSymbol:
		tok = SYMBOL
	case item.itype == itemString:
		tok = STRING
		item.val = item.val[1 : len(item.val)-1]
	case item.itype == itemCharLit:
		tok = CHARACTER
		item.val = item.val[1:len(item.val)]
	default:
		r, _ := utf8.DecodeRuneInString(item.val)
		tok = int(r)
	}
	lval.token = Token{token: tok, literal: item.val, pos: item.pos}
	return tok
}

func Parse(str string) ([]SExp, *Error) {
	//yyErrorVerbose = true
	l := &Lexer{}
	l.Init(str)
	yyParse(l)
	if l.error == nil {
		return l.result, nil
	} else {
		return nil, l.error
	}
}

//line yacctab:1
var yyExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
}

const yyNprod = 31
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 109

var yyAct = [...]int{

	4, 33, 42, 3, 26, 16, 17, 20, 19, 21,
	15, 37, 36, 18, 2, 22, 23, 24, 25, 28,
	1, 14, 29, 31, 32, 13, 34, 12, 10, 26,
	26, 11, 9, 8, 39, 7, 40, 41, 16, 17,
	20, 19, 21, 15, 5, 6, 18, 0, 22, 23,
	24, 25, 35, 16, 17, 20, 19, 21, 15, 0,
	0, 18, 38, 22, 23, 24, 25, 16, 17, 20,
	19, 21, 15, 0, 0, 18, 30, 22, 23, 24,
	25, 16, 17, 20, 19, 21, 15, 27, 0, 18,
	0, 22, 23, 24, 25, 16, 17, 20, 19, 21,
	15, 0, 0, 18, 0, 22, 23, 24, 25,
}
var yyPact = [...]int{

	91, -1000, -1000, 91, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, 77, -1000, -1000, 63, -1000,
	-1000, -1000, 91, 91, -13, 34, -1000, -1000, 1, 49,
	-1000, -1000, -1000, 91, -1000, 91, 91, -1000, -1000, -1000,
	-1000, -8, -1000,
}
var yyPgo = [...]int{

	0, 0, 45, 44, 35, 33, 32, 31, 28, 27,
	25, 21, 20, 3, 14,
}
var yyR1 = [...]int{

	0, 12, 12, 14, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 13, 13, 3, 4, 5, 6,
	6, 2, 2, 10, 10, 10, 11, 11, 7, 8,
	9,
}
var yyR2 = [...]int{

	0, 1, 1, 0, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 2, 2, 5, 3, 3,
	2, 1, 1, 2, 2, 3, 2, 3, 1, 1,
	1,
}
var yyChk = [...]int{

	-1000, -12, -14, -13, -1, -3, -2, -4, -5, -6,
	-8, -7, -9, -10, -11, 9, 4, 5, 12, 7,
	6, 8, 14, 15, 16, 17, -1, 10, -13, -13,
	13, -1, -1, 14, -1, 18, 11, 10, 13, -1,
	-1, -1, 10,
}
var yyDef = [...]int{

	3, -2, 1, 2, 14, 4, 5, 6, 7, 8,
	9, 10, 11, 12, 13, 0, 21, 22, 0, 29,
	28, 30, 0, 0, 0, 0, 15, 16, 0, 0,
	20, 23, 24, 0, 26, 0, 0, 18, 19, 25,
	27, 0, 17,
}
var yyTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 16, 3, 3, 3, 14,
	9, 10, 3, 3, 17, 3, 11, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 18, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 12, 3, 13, 3, 3, 15,
}
var yyTok2 = [...]int{

	2, 3, 4, 5, 6, 7, 8,
}
var yyTok3 = [...]int{
	0,
}

var yyErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

//line yaccpar:1

/*	parser for yacc output	*/

var (
	yyDebug        = 0
	yyErrorVerbose = false
)

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

type yyParser interface {
	Parse(yyLexer) int
	Lookahead() int
}

type yyParserImpl struct {
	lval  yySymType
	stack [yyInitialStackSize]yySymType
	char  int
}

func (p *yyParserImpl) Lookahead() int {
	return p.char
}

func yyNewParser() yyParser {
	return &yyParserImpl{}
}

const yyFlag = -1000

func yyTokname(c int) string {
	if c >= 1 && c-1 < len(yyToknames) {
		if yyToknames[c-1] != "" {
			return yyToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func yyStatname(s int) string {
	if s >= 0 && s < len(yyStatenames) {
		if yyStatenames[s] != "" {
			return yyStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func yyErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !yyErrorVerbose {
		return "syntax error"
	}

	for _, e := range yyErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + yyTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := yyPact[state]
	for tok := TOKSTART; tok-1 < len(yyToknames); tok++ {
		if n := base + tok; n >= 0 && n < yyLast && yyChk[yyAct[n]] == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if yyDef[state] == -2 {
		i := 0
		for yyExca[i] != -1 || yyExca[i+1] != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; yyExca[i] >= 0; i += 2 {
			tok := yyExca[i]
			if tok < TOKSTART || yyExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if yyExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += yyTokname(tok)
	}
	return res
}

func yylex1(lex yyLexer, lval *yySymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = yyTok1[0]
		goto out
	}
	if char < len(yyTok1) {
		token = yyTok1[char]
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			token = yyTok2[char-yyPrivate]
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		token = yyTok3[i+0]
		if token == char {
			token = yyTok3[i+1]
			goto out
		}
	}

out:
	if token == 0 {
		token = yyTok2[1] /* unknown char */
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", yyTokname(token), uint(char))
	}
	return char, token
}

func yyParse(yylex yyLexer) int {
	return yyNewParser().Parse(yylex)
}

func (yyrcvr *yyParserImpl) Parse(yylex yyLexer) int {
	var yyn int
	var yyVAL yySymType
	var yyDollar []yySymType
	_ = yyDollar // silence set and not used
	yyS := yyrcvr.stack[:]

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yyrcvr.char = -1
	yytoken := -1 // yyrcvr.char translated into internal numbering
	defer func() {
		// Make sure we report no lookahead when not parsing.
		yystate = -1
		yyrcvr.char = -1
		yytoken = -1
	}()
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	if yyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", yyTokname(yytoken), yyStatname(yystate))
	}

	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	yyn = yyPact[yystate]
	if yyn <= yyFlag {
		goto yydefault /* simple state */
	}
	if yyrcvr.char < 0 {
		yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
	}
	yyn += yytoken
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = yyAct[yyn]
	if yyChk[yyn] == yytoken { /* valid shift */
		yyrcvr.char = -1
		yytoken = -1
		yyVAL = yyrcvr.lval
		yystate = yyn
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	}

yydefault:
	/* default state action */
	yyn = yyDef[yystate]
	if yyn == -2 {
		if yyrcvr.char < 0 {
			yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && yyExca[xi+1] == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = yyExca[xi+0]
			if yyn < 0 || yyn == yytoken {
				break
			}
		}
		yyn = yyExca[xi+1]
		if yyn < 0 {
			goto ret0
		}
	}
	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			yylex.Error(yyErrorMessage(yystate, yytoken))
			Nerrs++
			if yyDebug >= 1 {
				__yyfmt__.Printf("%s", yyStatname(yystate))
				__yyfmt__.Printf(" saw %s\n", yyTokname(yytoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				yyn = yyPact[yyS[yyp].yys] + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = yyAct[yyn] /* simulate a shift of "error" */
					if yyChk[yystate] == yyErrCode {
						goto yystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if yyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", yyS[yyp].yys)
				}
				yyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yyTokname(yytoken))
			}
			if yytoken == yyEofCode {
				goto ret1
			}
			yyrcvr.char = -1
			yytoken = -1
			goto yynewstate /* try again in the same state */
		}
	}

	/* reduction by production yyn */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", yyn, yyStatname(yystate))
	}

	yynt := yyn
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= yyR2[yyn]
	// yyp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if yyp+1 >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = yyR1[yyn]
	yyg := yyPgo[yyn]
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = yyAct[yyg]
	} else {
		yystate = yyAct[yyj]
		if yyChk[yystate] != -yyn {
			yystate = yyAct[yyg]
		}
	}
	// dummy call; replaced with literal code
	switch yynt {

	case 1:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sexp.go.y:30
		{
			yylex.(*Lexer).result = []SExp{}
		}
	case 2:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sexp.go.y:34
		{
			yyVAL.seq = yyDollar[1].seq
			yylex.(*Lexer).result = yyVAL.seq
		}
	case 14:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sexp.go.y:47
		{
			yyVAL.seq = []SExp{yyDollar[1].expr}
		}
	case 15:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sexp.go.y:51
		{
			yyVAL.seq = append(yyDollar[1].seq, yyDollar[2].expr)
		}
	case 16:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sexp.go.y:57
		{
			yyVAL.expr = &SExpNil{}
		}
	case 17:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sexp.go.y:63
		{
			sq := yyDollar[2].seq
			if len(sq) == 1 {
				yyVAL.expr = &SExpCons{car: sq[0], cdr: yyDollar[4].expr}
			} else {
				yyVAL.expr = &SExpListDot{elements: sq, last: yyDollar[4].expr}
			}
		}
	case 18:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sexp.go.y:74
		{
			yyVAL.expr = &SExpList{elements: yyDollar[2].seq}
		}
	case 19:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sexp.go.y:80
		{
			yyVAL.expr = &SExpVector{elements: yyDollar[2].seq}
		}
	case 20:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sexp.go.y:84
		{
			yyVAL.expr = &SExpVector{elements: []SExp{}}
		}
	case 21:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sexp.go.y:90
		{
			yyVAL.expr = &SExpInt{literal: yyDollar[1].token.literal}
		}
	case 22:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sexp.go.y:94
		{
			yyVAL.expr = &SExpFloat{literal: yyDollar[1].token.literal}
		}
	case 23:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sexp.go.y:100
		{
			yyVAL.expr = &SExpQuoted{sexp: yyDollar[2].expr}
		}
	case 24:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sexp.go.y:104
		{
			yyVAL.expr = &SExpQuasiQuoted{sexp: yyDollar[2].expr}
		}
	case 25:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sexp.go.y:108
		{
			yyVAL.expr = &SExpQuoted{sexp: yyDollar[3].expr, function: true}
		}
	case 26:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sexp.go.y:114
		{
			yyVAL.expr = &SExpUnquote{sexp: yyDollar[2].expr, splice: false}
		}
	case 27:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sexp.go.y:118
		{
			yyVAL.expr = &SExpUnquote{sexp: yyDollar[3].expr, splice: true}
		}
	case 28:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sexp.go.y:124
		{
			yyVAL.expr = AstSymbol(yyDollar[1].token.literal)
		}
	case 29:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sexp.go.y:130
		{
			yyVAL.expr = &SExpChar{literal: yyDollar[1].token.literal}
		}
	case 30:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sexp.go.y:136
		{
			yyVAL.expr = &SExpString{literal: yyDollar[1].token.literal}
		}
	}
	goto yystack /* stack new state and value */
}
