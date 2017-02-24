%{
package parser

import (
	"unicode/utf8"
)

type Token struct {
    token   int
    literal string
    pos     Pos
}

%}

%union {
    token Token
    expr  SExp
    seq   []SExp
}

%type   <expr>          sexp val nil cons list vector symbol character string quoted unquote
%type   <seq>           target sexp_seq
%token  <token>         INTEGER FLOAT SYMBOL CHARACTER STRING

%%

target
      : void
      {
          yylex.(*Lexer).result = []SExp{}
      }
      | sexp_seq
      {
          $$ = $1
          yylex.(*Lexer).result = $$
      }

void  :

sexp
      : nil | val | cons | list | vector | character
      | symbol | string | quoted | unquote

sexp_seq
      : sexp
      {
          $$ = []SExp{$1}
      }
      | sexp_seq sexp
      {
          $$ = append($1, $2)
      }

nil
      : "(" ")"
      {
          $$ = &SExpNil{}
      }

cons
      : "(" sexp_seq "." sexp ")"
      {
          sq := $2
          if len(sq) == 1 {
              $$ = &SExpCons{car:sq[0], cdr:$4}
          } else {
              $$ = &SExpListDot{elements:sq, last:$4}
          }
      }

list
      : "(" sexp_seq ")"
      {
          $$ = &SExpList{elements: $2}
      }

vector
      : "[" sexp_seq "]"
      {
          $$ = &SExpVector{elements: $2}
      }
      | "[" "]"
      {
          $$ = &SExpVector{elements: []SExp{}}
      }

val
      : INTEGER
      {
          $$ = &SExpInt{literal: $1.literal}
      }
      | FLOAT
      {
          $$ = &SExpFloat{literal: $1.literal}
      }

quoted
      : "'" sexp
      {
          $$ = &SExpQuoted{sexp: $2}
      }
      | "`" sexp
      {
          $$ = &SExpQuasiQuoted{sexp: $2}
      }
      | "#" "'" sexp
      {
          $$ = &SExpQuoted{sexp: $3, function: true}
      }

unquote
      : "," sexp
      {
          $$ = &SExpUnquote{sexp: $2, splice: false}
      }
      | "," "@" sexp
      {
          $$ = &SExpUnquote{sexp: $3, splice: true}
      }

symbol
      : SYMBOL
      {
          $$ = AstSymbol($1.literal)
      }

character
      : CHARACTER
      {
          $$ = &SExpChar{literal: $1.literal}
      }

string
      : STRING
      {
          $$ = &SExpString{literal: $1.literal}
      }


%%

func (l *Lexer) Error(e string) {
	l.error = &Error{
        Msg: e, Pos: l.lastPos,
        Line: l.lineNumber(),
        Col: l.columnNumber(),
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
        item.val = item.val[1:len(item.val)-1]
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
