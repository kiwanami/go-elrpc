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
	if s.literal == "t" {
		return true
	}
	if s.literal == "nil" {
		return nil
	}
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

func inferArrayType(lst []SExp) string {
	if len(lst) == 0 {
		return "interface"
	}
	var ty string
	switch lst[0].(type) {
	case *SExpChar, *SExpString, *SExpSymbol:
		ty = "string"
	case *SExpFloat:
		ty = "float"
	case *SExpInt:
		ty = "int"
	default:
		return "interface"
	}
	for _, v := range lst {
		switch v.(type) {
		case *SExpChar:
			if ty != "string" {
				return "interface"
			}
		case *SExpFloat:
			if ty == "int" {
				ty = "float"
			} else if ty != "float" {
				return "interface"
			}
		case *SExpInt:
			if ty != "float" && ty != "int" {
				return "interface"
			}
		default:
			return "interface"
		}
	}
	return ty
}

func typedSlice(lst []SExp) interface{} {
	len := len(lst)
	typ := inferArrayType(lst)
	// pp.Println(lst)
	// fmt.Println("typedSlice: " + typ)
	switch typ {
	case "int":
		ret := make([]int, len)
		for i := 0; i < len; i++ {
			s := lst[i].(*SExpInt)
			v, _ := strconv.Atoi(s.literal)
			ret[i] = v
		}
		return ret
	case "float":
		ret := make([]float64, len)
		for i := 0; i < len; i++ {
			var lit string
			switch lst[i].(type) {
			case *SExpFloat:
				s := lst[i].(*SExpFloat)
				lit = s.literal
			case *SExpInt:
				s := lst[i].(*SExpInt)
				lit = s.literal
			}
			v, _ := strconv.ParseFloat(lit, 64)
			ret[i] = v
		}
		return ret
	case "string":
		ret := make([]string, len)
		for i := 0; i < len; i++ {
			s := lst[i].ToValue()
			v, _ := s.(string)
			ret[i] = v
		}
		return ret
	default:
	}
	ret := make([]interface{}, len)
	for i, e := range lst {
		ret[i] = e.ToValue()
	}
	return ret
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
	return typedSlice([]SExp{s.car, s.cdr})
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
	return typedSlice(s.elements)
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
	aa := append(s.elements, s.last)
	return typedSlice(aa)
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
	return typedSlice(s.elements)
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
