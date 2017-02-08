package parser

import "strconv"

type SExp interface {
	IsAtom() bool
	IsCons() bool
	IsList() bool
}

type SExpAtom struct{}

func (s *SExpAtom) IsAtom() bool { return true }
func (s *SExpAtom) IsCons() bool { return false }
func (s *SExpAtom) IsList() bool { return false }

type SExpNil struct {
	*SExpAtom
}

type SExpChar struct {
	*SExpAtom
	literal string
}
type SExpString struct {
	*SExpAtom
	literal string
}
type SExpSymbol struct {
	*SExpAtom
	literal string
}

type SExpInt struct {
	*SExpAtom
	literal string
}

func (s *SExpInt) Value() int {
	i, _ := strconv.Atoi(s.literal)
	return i
}

type SExpFloat struct {
	*SExpAtom
	literal string
}

func (s *SExpFloat) Value() float64 {
	f, _ := strconv.ParseFloat(s.literal, 64)
	return f
}

type AbstSExpCons struct{}

func (s *AbstSExpCons) IsAtom() bool { return false }
func (s *AbstSExpCons) IsCons() bool { return true }
func (s *AbstSExpCons) IsList() bool { return false }

type SExpCons struct {
	*AbstSExpCons
	car, cdr SExp
}

type SExpList struct {
	*AbstSExpCons
	elements []SExp
}

type SExpListDot struct {
	*AbstSExpCons
	elements []SExp
	last     SExp
}

type SExpVector struct {
	*AbstSExpCons
	elements []SExp
}

type SExpQuoted struct {
	sexp     SExp
	function bool
}

func (s *SExpQuoted) IsAtom() bool { return false }
func (s *SExpQuoted) IsCons() bool { return false }
func (s *SExpQuoted) IsList() bool { return false }

type SExpQuasiQuoted struct {
	sexp SExp
}

func (s *SExpQuasiQuoted) IsAtom() bool { return false }
func (s *SExpQuasiQuoted) IsCons() bool { return false }
func (s *SExpQuasiQuoted) IsList() bool { return false }

type SExpUnquote struct {
	sexp   SExp
	splice bool
}

func (s *SExpUnquote) IsAtom() bool { return false }
func (s *SExpUnquote) IsCons() bool { return false }
func (s *SExpUnquote) IsList() bool { return false }
