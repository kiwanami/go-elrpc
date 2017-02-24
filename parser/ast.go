package parser

import (
	"bytes"
	"strconv"
)

type SExp interface {
	ToSExpString() string // express in S-exp string
	ToValue() interface{} // transform content of this AST into Go object
}

type SExpAtom struct{}

func (s *SExpAtom) ToSExpString() string { return "--ATOM--" }

type SExpNil struct {
	*SExpAtom
}

func (s *SExpNil) ToSExpString() string {
	return "nil"
}
func (s *SExpNil) ToValue() interface{} {
	return nil
}

type SExpChar struct {
	*SExpAtom
	literal string
}

func (s *SExpChar) ToSExpString() string {
	switch s.literal {
	case " ", "?", "\\":
		return "?\\" + s.literal
	default:
		return "?" + s.literal
	}
}
func (s *SExpChar) ToValue() interface{} {
	return s.literal
}

type SExpString struct {
	*SExpAtom
	literal string
}

func (s *SExpString) ToSExpString() string {
	return StringLiteral(s.literal)
}
func (s *SExpString) ToValue() interface{} {
	return s.literal
}

type SExpSymbol struct {
	*SExpAtom
	literal string
}

func (s *SExpSymbol) ToValue() interface{} {
	return s.literal
}

func (s *SExpSymbol) ToSExpString() string {
	return SymbolLiteral(s.literal)
}

type SExpInt struct {
	*SExpAtom
	literal string
}

func (s *SExpInt) ToSExpString() string {
	return s.literal
}

func (s *SExpInt) ToValue() interface{} {
	i, _ := strconv.Atoi(s.literal)
	return i
}

type SExpFloat struct {
	*SExpAtom
	literal string
}

func (s *SExpFloat) ToSExpString() string {
	return s.literal
}

func (s *SExpFloat) ToValue() interface{} {
	f, _ := strconv.ParseFloat(s.literal, 64)
	return f
}

type AbstSExpCons struct{}

func (s *AbstSExpCons) ToSExpString() string { return "--CONS--" }

type SExpCons struct {
	*AbstSExpCons
	car, cdr SExp
}

func (s *SExpCons) ToSExpString() string {
	return "(" + s.car.ToSExpString() + " . " + s.cdr.ToSExpString() + ")"
}
func (s *SExpCons) ToValue() interface{} {
	return []interface{}{s.car.ToValue(), s.cdr.ToValue()}
}

type SExpList struct {
	*AbstSExpCons
	elements []SExp
}

func (s *SExpList) ToSExpString() string {
	buf := bytes.Buffer{}
	buf.WriteByte('(')
	first := true
	for _, e := range s.elements {
		if first {
			first = false
		} else {
			buf.WriteByte(' ')
		}
		buf.WriteString(e.ToSExpString())
	}
	buf.WriteByte(')')
	return buf.String()
}

func (s *SExpList) ToValue() interface{} {
	ret := make([]interface{}, len(s.elements))
	for i, e := range s.elements {
		ret[i] = e.ToValue()
	}
	return ret
}

type SExpListDot struct {
	*AbstSExpCons
	elements []SExp
	last     SExp
}

func (s *SExpListDot) ToSExpString() string {
	buf := bytes.Buffer{}
	buf.WriteByte('(')
	first := true
	for _, e := range s.elements {
		if first {
			first = false
		} else {
			buf.WriteByte(' ')
		}
		buf.WriteString(e.ToSExpString())
	}
	buf.WriteString(" . ")
	buf.WriteString(s.last.ToSExpString())
	buf.WriteByte(')')
	return buf.String()
}
func (s *SExpListDot) ToValue() interface{} {
	ret := make([]interface{}, len(s.elements)+1)
	for i, e := range s.elements {
		ret[i] = e.ToValue()
	}
	ret[len(s.elements)] = s.last.ToValue()
	return ret
}

type SExpVector struct {
	*AbstSExpCons
	elements []SExp
}

func (s *SExpVector) ToSExpString() string {
	buf := bytes.Buffer{}
	buf.WriteByte('[')
	first := true
	for _, e := range s.elements {
		if first {
			first = false
		} else {
			buf.WriteByte(' ')
		}
		buf.WriteString(e.ToSExpString())
	}
	buf.WriteByte(']')
	return buf.String()
}
func (s *SExpVector) ToValue() interface{} {
	ret := make([]interface{}, len(s.elements))
	for i, e := range s.elements {
		ret[i] = e.ToValue()
	}
	return ret
}

type SExpQuoted struct {
	sexp     SExp
	function bool
}

func (s *SExpQuoted) ToSExpString() string {
	ret := ""
	if s.function {
		ret = "#"
	}
	return ret + "'" + s.sexp.ToSExpString()
}
func (s *SExpQuoted) ToValue() interface{} {
	return s.sexp.ToValue()
}

type SExpQuasiQuoted struct {
	sexp SExp
}

func (s *SExpQuasiQuoted) ToSExpString() string {
	return "`" + s.sexp.ToSExpString()
}
func (s *SExpQuasiQuoted) ToValue() interface{} {
	return s.sexp.ToValue()
}

type SExpUnquote struct {
	sexp   SExp
	splice bool
}

func (s *SExpUnquote) ToSExpString() string {
	ret := ","
	if s.splice {
		ret += "@"
	}
	return ret + s.sexp.ToSExpString()
}
func (s *SExpUnquote) ToValue() interface{} {
	return s.sexp.ToValue()
}

type SExpWrapper struct {
	buf []byte
}

func (s *SExpWrapper) ToSExpString() string {
	return string(s.buf)
}
func (s *SExpWrapper) ToValue() interface{} {
	panic("BUG")
}

func AstWrapper(v []byte) *SExpWrapper {
	return &SExpWrapper{v}
}

/// AST utilities

func AstInt(v string) *SExpInt {
	return &SExpInt{literal: v}
}

func AstFloat(v string) *SExpFloat {
	return &SExpFloat{literal: v}
}

func AstSymbol(v string) SExp {
	if v == "nil" {
		return AstNil()
	}
	return &SExpSymbol{literal: v}
}

func AstChar(v string) *SExpChar {
	return &SExpChar{literal: v}
}

func AstString(v string) *SExpString {
	return &SExpString{literal: v}
}

func AstNil() *SExpNil {
	return &SExpNil{}
}

func AstCons(v1 SExp, v2 SExp) *SExpCons {
	return &SExpCons{car: v1, cdr: v2}
}

func AstList(vs []SExp) *SExpList {
	return &SExpList{elements: vs}
}

func AstListv(vs ...SExp) *SExpList {
	return &SExpList{elements: vs}
}

func AstVectorv(vs ...SExp) *SExpVector {
	return &SExpVector{elements: vs}
}

func AstDotlist(vs []SExp, v2 SExp) *SExpListDot {
	return &SExpListDot{elements: vs, last: v2}
}

func AstQ(v SExp) *SExpQuoted {
	return &SExpQuoted{sexp: v}
}

func AstQq(v SExp) *SExpQuasiQuoted {
	return &SExpQuasiQuoted{sexp: v}
}

func AstQf(v SExp) *SExpQuoted {
	return &SExpQuoted{sexp: v, function: true}
}

func AstUnq(v SExp) *SExpUnquote {
	return &SExpUnquote{sexp: v, splice: false}
}

func AstUnqs(v SExp) *SExpUnquote {
	return &SExpUnquote{sexp: v, splice: true}
}
