// Package jsonvalue decodes JSON strictly: no duplicate member names, no
// trailing data, bounded nesting. Errors never quote the input.
package jsonvalue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// MaxDepth is the deepest array/object nesting Decode accepts.
const MaxDepth = 10000

// SyntaxError reports broken JSON by location only.
type SyntaxError struct {
	Offset       int64
	Line, Column int
	Reason       string // fixed phrase, never input content
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("invalid JSON at line %d, column %d (offset %d): %s", e.Line, e.Column, e.Offset, e.Reason)
}

// LimitError reports that a built-in limit was exceeded.
type LimitError struct {
	Limit  string
	Max    int
	Offset int64
}

func (e *LimitError) Error() string {
	return fmt.Sprintf("%s exceeds limit %d", e.Limit, e.Max)
}

func syntaxErr(data []byte, off int64, reason string) *SyntaxError {
	if off < 0 {
		off = 0
	}
	if off > int64(len(data)) {
		off = int64(len(data))
	}
	head := data[:off]
	line := 1 + bytes.Count(head, []byte("\n"))
	col := int(off) - (bytes.LastIndexByte(head, '\n') + 1) + 1
	return &SyntaxError{Offset: off, Line: line, Column: col, Reason: reason}
}

type frame struct {
	obj  map[string]any
	arr  []any
	key  string
	keys bool // true when the next token is a member name (or closing brace)
}

// Decode parses exactly one JSON value into map[string]any, []any,
// json.Number, string, bool or nil, with the default number caps.
func Decode(data []byte) (any, error) { return decode(data, NumberLimits{}, false) }

// DecodeLimited is Decode with *Number leaves and the given number caps.
func DecodeLimited(data []byte, lim NumberLimits) (any, error) { return decode(data, lim, true) }

var bom = []byte{0xEF, 0xBB, 0xBF}

func decode(orig []byte, lim NumberLimits, exact bool) (any, error) {
	lim = lim.resolved()
	data := bytes.TrimPrefix(orig, bom) // at most one BOM
	base := int64(len(orig) - len(data))
	syntaxErr := func(_ []byte, off int64, reason string) *SyntaxError {
		return syntaxErr(orig, off+base, reason)
	}
	if off, reason := scanText(data, nil); reason != "" {
		return nil, syntaxErr(data, int64(off), reason)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var stack []*frame
	var result any
	done := false

	// put attaches a finished value to its parent, or makes it the result.
	put := func(v any) {
		if len(stack) == 0 {
			result, done = v, true
			return
		}
		f := stack[len(stack)-1]
		if f.obj != nil {
			f.obj[f.key] = v
			f.keys = true
		} else {
			f.arr = append(f.arr, v)
		}
	}

	for !done {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				if len(stack) == 0 {
					return nil, syntaxErr(data, int64(len(data)), "empty input")
				}
				return nil, syntaxErr(data, int64(len(data)), "unexpected end of input")
			}
			// encoding/json has its own depth cap (same value) and fails first
			// when the opener is the one that would exceed it.
			if rest := bytes.TrimLeft(data[min(dec.InputOffset(), int64(len(data))):], " \t\r\n,:"); len(stack) >= MaxDepth && len(rest) > 0 && (rest[0] == '[' || rest[0] == '{') {
				return nil, &LimitError{Limit: "nesting depth", Max: MaxDepth, Offset: dec.InputOffset() + base}
			}
			return nil, syntaxErr(data, dec.InputOffset(), "syntax error")
		}
		if len(stack) > 0 {
			if f := stack[len(stack)-1]; f.keys {
				if d, ok := tok.(json.Delim); ok && d == '}' {
					stack = stack[:len(stack)-1]
					put(f.obj)
					continue
				}
				k, ok := tok.(string)
				if !ok {
					return nil, syntaxErr(data, dec.InputOffset(), "syntax error")
				}
				if _, dup := f.obj[k]; dup {
					return nil, syntaxErr(data, dec.InputOffset(), "duplicate member name")
				}
				f.key, f.keys = k, false
				f.obj[k] = nil // reserve the name
				continue
			}
		}
		switch t := tok.(type) {
		case json.Delim:
			switch t {
			case '{', '[':
				if len(stack)+1 > MaxDepth {
					return nil, &LimitError{Limit: "nesting depth", Max: MaxDepth, Offset: dec.InputOffset() + base}
				}
				if t == '{' {
					stack = append(stack, &frame{obj: map[string]any{}, keys: true})
				} else {
					stack = append(stack, &frame{arr: []any{}})
				}
			case ']':
				if len(stack) == 0 || stack[len(stack)-1].obj != nil {
					return nil, syntaxErr(data, dec.InputOffset(), "syntax error")
				}
				f := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				put(f.arr)
			default:
				return nil, syntaxErr(data, dec.InputOffset(), "syntax error")
			}
		case json.Number:
			off := dec.InputOffset() - int64(len(t)) + base
			e, err := checkNumber(string(t), lim, off)
			if err != nil {
				return nil, err
			}
			if !exact {
				put(t)
				break
			}
			n, err := newNumber(string(t), e, off)
			if err != nil {
				return nil, err
			}
			put(n)
		case string, bool, nil:
			put(t)
		default:
			return nil, syntaxErr(data, dec.InputOffset(), "syntax error")
		}
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, syntaxErr(data, dec.InputOffset(), "trailing data after value")
	}
	return result, nil
}
