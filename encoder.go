package elrpc

import (
	"bytes"
	"math"
	"reflect"
	"runtime"
	"sort"
	"strconv"

	"github.com/kiwanami/go-elrpc/parser"
)

type EncodeError struct {
	Msg string
}

func (e *EncodeError) Error() string {
	return "sexp encode: " + e.Msg
}

type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return "sexp encode: unsupported type: " + e.Type.String()
}

type UnsupportedValueError struct {
	Value reflect.Value
	Str   string
}

func (e *UnsupportedValueError) Error() string {
	return "sexp encode: unsupported value: " + e.Str
}

type encodeState struct {
	bytes.Buffer
	scratch [64]byte
}

func Encode(obj interface{}) ([]byte, error) {
	s := &encodeState{}
	err := s.encode(obj)
	if err != nil {
		return nil, err
	}
	return s.Bytes(), nil
}

func (e *encodeState) error(err error) {
	panic(err)
}

func (s *encodeState) encode(v interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			if s, ok := r.(string); ok {
				panic(s)
			}
			err = r.(error)
		}
	}()
	s.reflectValue(reflect.ValueOf(v))
	return nil
}

func (e *encodeState) reflectValue(v reflect.Value) {
	valueEncoder(v)(e, v, false)
}

type encoderFunc func(e *encodeState, v reflect.Value, quoted bool)

func valueEncoder(v reflect.Value) encoderFunc {
	if !v.IsValid() {
		return invalidValueEncoder
	}
	return typeEncoder(v.Type())
}

func typeEncoder(t reflect.Type) encoderFunc {
	switch t.Kind() {
	case reflect.Bool:
		return boolEncoder
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intEncoder
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return uintEncoder
	case reflect.Float32:
		return float32Encoder
	case reflect.Float64:
		return float64Encoder
	case reflect.String:
		return stringEncoder
	case reflect.Interface:
		return interfaceEncoder
	case reflect.Struct:
		return newStructEncoder(t)
	case reflect.Map:
		return newMapEncoder(t)
	case reflect.Slice:
		return newSliceEncoder(t)
	case reflect.Array:
		return newArrayEncoder(t)
	case reflect.Ptr:
		return newPtrEncoder(t)
	default:
		return unsupportedTypeEncoder
	}
}

func invalidValueEncoder(e *encodeState, v reflect.Value, quoted bool) {
	e.WriteString("nil")
}

func boolEncoder(e *encodeState, v reflect.Value, quoted bool) {
	if quoted {
		e.WriteByte('"')
	}
	if v.Bool() {
		e.WriteString("t")
	} else {
		e.WriteString("nil")
	}
	if quoted {
		e.WriteByte('"')
	}
}

func intEncoder(e *encodeState, v reflect.Value, quoted bool) {
	b := strconv.AppendInt(e.scratch[:0], v.Int(), 10)
	if quoted {
		e.WriteByte('"')
	}
	e.Write(b)
	if quoted {
		e.WriteByte('"')
	}
}

func uintEncoder(e *encodeState, v reflect.Value, quoted bool) {
	b := strconv.AppendUint(e.scratch[:0], v.Uint(), 10)
	if quoted {
		e.WriteByte('"')
	}
	e.Write(b)
	if quoted {
		e.WriteByte('"')
	}
}

type floatEncoder int // number of bits

func (bits floatEncoder) encode(e *encodeState, v reflect.Value, quoted bool) {
	f := v.Float()
	if math.IsInf(f, 0) || math.IsNaN(f) {
		e.error(&UnsupportedValueError{v, strconv.FormatFloat(f, 'g', -1, int(bits))})
	}
	b := strconv.AppendFloat(e.scratch[:0], f, 'g', -1, int(bits))
	if quoted {
		e.WriteByte('"')
	}
	e.Write(b)
	if quoted {
		e.WriteByte('"')
	}
}

var (
	float32Encoder = (floatEncoder(32)).encode
	float64Encoder = (floatEncoder(64)).encode
)

type Number string

var numberType = reflect.TypeOf(Number(""))

func stringEncoder(e *encodeState, v reflect.Value, quoted bool) {
	if v.Type() == numberType {
		numStr := v.String()
		if numStr == "" {
			numStr = "0" // Number's zero-val
		}
		e.WriteString(numStr)
		return
	}
	if quoted {
		sb, err := Encode(v.String())
		if err != nil {
			e.error(err)
		}
		e.string(string(sb))
	} else {
		e.string(v.String())
	}
}

func interfaceEncoder(e *encodeState, v reflect.Value, quoted bool) {
	if v.IsNil() {
		e.WriteString("null")
		return
	}
	e.reflectValue(v.Elem())
}

func unsupportedTypeEncoder(e *encodeState, v reflect.Value, quoted bool) {
	e.error(&UnsupportedTypeError{v.Type()})
}

type mapEncoder struct {
	elemEnc encoderFunc
}

func (me *mapEncoder) encode(e *encodeState, v reflect.Value, _ bool) {
	if v.IsNil() {
		e.WriteString("null")
		return
	}
	e.WriteByte('(')
	var sv stringValues = v.MapKeys()
	sort.Sort(sv)
	for i, k := range sv {
		if i > 0 {
			e.WriteString(" ")
		}
		e.WriteByte('(')
		e.string(k.String())
		e.WriteString(" . ")
		me.elemEnc(e, v.MapIndex(k), false)
		e.WriteByte(')')
	}
	e.WriteByte(')')
}

func newMapEncoder(t reflect.Type) encoderFunc {
	if t.Key().Kind() != reflect.String {
		return unsupportedTypeEncoder
	}
	me := &mapEncoder{typeEncoder(t.Elem())}
	return me.encode
}

// sliceEncoder just wraps an arrayEncoder, checking to make sure the value isn't nil.
type sliceEncoder struct {
	arrayEnc encoderFunc
}

func (se *sliceEncoder) encode(e *encodeState, v reflect.Value, _ bool) {
	if v.IsNil() {
		e.WriteString("nil")
		return
	}
	se.arrayEnc(e, v, false)
}

func newSliceEncoder(t reflect.Type) encoderFunc {
	enc := &sliceEncoder{newArrayEncoder(t)}
	return enc.encode
}

type arrayEncoder struct {
	elemEnc encoderFunc
}

func (ae *arrayEncoder) encode(e *encodeState, v reflect.Value, _ bool) {
	e.WriteByte('(')
	n := v.Len()
	for i := 0; i < n; i++ {
		if i > 0 {
			e.WriteByte(' ')
		}
		ae.elemEnc(e, v.Index(i), false)
	}
	e.WriteByte(')')
}

func newArrayEncoder(t reflect.Type) encoderFunc {
	enc := &arrayEncoder{typeEncoder(t.Elem())}
	return enc.encode
}

type ptrEncoder struct {
	elemEnc encoderFunc
}

func (pe *ptrEncoder) encode(e *encodeState, v reflect.Value, quoted bool) {
	if v.IsNil() {
		e.WriteString("nil")
		return
	}
	pe.elemEnc(e, v.Elem(), quoted)
}

func newPtrEncoder(t reflect.Type) encoderFunc {
	enc := &ptrEncoder{typeEncoder(t.Elem())}
	return enc.encode
}

var hex = "0123456789abcdef"

func (e *encodeState) symbol(s string) (int, error) {
	len0 := e.Len()
	e.WriteString(parser.SymbolLiteral(s))
	return e.Len() - len0, nil
}

func (e *encodeState) string(s string) (int, error) {
	len0 := e.Len()
	e.WriteString(parser.StringLiteral(s))
	return e.Len() - len0, nil
}

// stringValues is a slice of reflect.Value holding *reflect.StringValue.
// It implements the methods to sort by string.
type stringValues []reflect.Value

func (sv stringValues) Len() int           { return len(sv) }
func (sv stringValues) Swap(i, j int)      { sv[i], sv[j] = sv[j], sv[i] }
func (sv stringValues) Less(i, j int) bool { return sv.get(i) < sv.get(j) }
func (sv stringValues) get(i int) string   { return sv[i].String() }

/// struct encoder

type structEncoder struct {
	fields []field
}

// A field represents a single field found in a struct.
type field struct {
	name    string
	typ     reflect.Type
	encoder encoderFunc
}

func newStructEncoder(t reflect.Type) encoderFunc {
	fieldDescs := make([]field, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		fieldDescs[i] = field{
			name:    sf.Name,
			typ:     sf.Type,
			encoder: typeEncoder(sf.Type),
		}
	}
	se := &structEncoder{fields: fieldDescs}
	return se.encode
}

func (se *structEncoder) encode(e *encodeState, v reflect.Value, quoted bool) {
	e.WriteByte('(')
	first := true
	for i, f := range se.fields {
		if first {
			first = false
			e.WriteByte('(')
		} else {
			e.WriteString(" (")
		}
		e.symbol(f.name)
		e.WriteString(" . ")
		f.encoder(e, v.Field(i), quoted)
		e.WriteByte(')')
	}
	e.WriteByte(')')
}
