package parser

import (
	"bytes"
	"fmt"
	"unicode/utf8"
)

// convert string literal
func StringLiteral(content string) string {
	buf := bytes.Buffer{}
	buf.WriteByte('"')
	start := 0
	for i := 0; i < len(content); {
		if b := content[i]; b < utf8.RuneSelf {
			if 0x20 <= b && b != '\\' && b != '"' {
				i++
				continue
			}
			if start < i {
				buf.WriteString(content[start:i])
			}
			switch b {
			case '\\':
				fallthrough
			case '"':
				buf.WriteByte('\\')
				buf.WriteByte(b)
			case '\n':
				buf.WriteByte('\\')
				buf.WriteByte('n')
			case '\r':
				buf.WriteByte('\\')
				buf.WriteByte('r')
			case '\t':
				buf.WriteByte('\\')
				buf.WriteByte('t')
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRuneInString(content[i:])
		if c == utf8.RuneError && size == 1 {
			if start < i {
				buf.WriteString(content[start:i])
			}
			buf.WriteString(`\ufffd`)
			i += size
			start = i
			continue
		}
		if c != utf8.RuneError && size >= 4 {
			if start < i {
				buf.WriteString(content[start:i])
			}
			buf.WriteString(fmt.Sprintf("\\U000%x", c))
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(content) {
		buf.WriteString(content[start:])
	}
	buf.WriteByte('"')
	return buf.String()
}

// convert symbol literal
func SymbolLiteral(symbolName string) string {
	buf := bytes.Buffer{}
	start := 0
	for i := 0; i < len(symbolName); {
		if b := symbolName[i]; b < utf8.RuneSelf {
			if 0x20 <= b && b != '\\' && b != '"' && b != '?' && b != ' ' {
				i++
				continue
			}
			if start < i {
				buf.WriteString(symbolName[start:i])
			}
			switch b {
			case '\\', '"', '?', ' ':
				buf.WriteByte('\\')
				buf.WriteByte(b)
			default:
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRuneInString(symbolName[i:])
		if c == utf8.RuneError && size == 1 {
			if start < i {
				buf.WriteString(symbolName[start:i])
			}
			buf.WriteString(`\ufffd`)
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(symbolName) {
		buf.WriteString(symbolName[start:])
	}
	return buf.String()
}
