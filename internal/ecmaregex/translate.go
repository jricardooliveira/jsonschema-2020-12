// Package ecmaregex translates ECMA-262 (Unicode mode) regular expressions into
// RE2 syntax for Go's regexp package. It fails closed: any construct it does not
// explicitly handle is an *Error, never passed through with its RE2 meaning.
package ecmaregex

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/neoface/jsonschema/internal/jsonvalue"
	"github.com/neoface/jsonschema/internal/limits"
)

// DefaultMaxPatternSize is the longest accepted pattern, in bytes of UTF-8.
const DefaultMaxPatternSize = 64 << 10

// CompileBounded translates and compiles src. A source longer than max bytes
// (0 means DefaultMaxPatternSize) is a *jsonvalue.LimitError named "pattern
// size", checked before any translation. Other errors never carry a limit.
func CompileBounded(src string, max int) (*regexp.Regexp, error) {
	if max <= 0 {
		max = DefaultMaxPatternSize
	}
	if len(src) > max {
		return nil, &jsonvalue.LimitError{Limit: limits.PatternSize, Max: max}
	}
	re2, err := Translate(src)
	if err != nil {
		return nil, err
	}
	re, err := regexp.Compile(re2)
	if err != nil {
		return nil, fmt.Errorf("cannot be compiled: %w", err)
	}
	return re, nil
}

// Error reports a pattern construct that cannot be translated.
type Error struct {
	Offset int // rune offset into the pattern
	Reason string
}

func (e *Error) Error() string {
	return fmt.Sprintf("unsupported regular expression: %s at position %d", e.Reason, e.Offset)
}

type rng struct{ lo, hi rune }

const maxRune = 0x10FFFF

var (
	digitSet = []rng{{'0', '9'}}
	wordSet  = []rng{{'0', '9'}, {'A', 'Z'}, {'_', '_'}, {'a', 'z'}}
	// ECMA WhiteSpace + LineTerminator.
	spaceSet = []rng{{9, 13}, {' ', ' '}, {0xA0, 0xA0}, {0x1680, 0x1680}, {0x2000, 0x200A},
		{0x2028, 0x2029}, {0x202F, 0x202F}, {0x205F, 0x205F}, {0x3000, 0x3000}, {0xFEFF, 0xFEFF}}
	dotSet = complement([]rng{{'\n', '\n'}, {'\r', '\r'}, {0x2028, 0x2029}})
)

// complement returns the sorted, non-overlapping ranges not covered by rs.
func complement(rs []rng) []rng {
	var out []rng
	next := rune(0)
	for _, r := range rs {
		if r.lo > next {
			out = append(out, rng{next, r.lo - 1})
		}
		next = r.hi + 1
	}
	if next <= maxRune {
		out = append(out, rng{next, maxRune})
	}
	return out
}

func body(rs []rng) string {
	var b strings.Builder
	for _, r := range rs {
		fmt.Fprintf(&b, `\x{%X}`, r.lo)
		if r.hi != r.lo {
			fmt.Fprintf(&b, `-\x{%X}`, r.hi)
		}
	}
	return b.String()
}

// lit renders one literal rune safely in or out of a class.
func lit(r rune) string {
	if r >= 0x80 || r >= '0' && r <= '9' || r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' {
		return string(r)
	}
	return fmt.Sprintf(`\x{%X}`, r)
}

type scanner struct {
	p   []rune
	i   int
	out strings.Builder
}

// atom is the result of parsing one escape or class member.
type atom struct {
	single bool
	r      rune   // when single
	text   string // when not single: ready-to-emit RE2
}

// Translate converts an ECMA-262 pattern (Unicode mode, no flags) into an
// equivalent RE2 pattern, or returns an *Error.
func Translate(pattern string) (string, error) {
	s := &scanner{p: []rune(pattern)}
	for s.i < len(s.p) {
		var err error
		switch c := s.p[s.i]; c {
		case '.':
			s.i++
			s.out.WriteString("[" + body(dotSet) + "]")
		case '\\':
			var a atom
			if a, err = s.escape(false); err == nil {
				if a.single {
					s.out.WriteString(lit(a.r))
				} else {
					s.out.WriteString(a.text)
				}
			}
		case '[':
			err = s.class()
		case '(':
			err = s.group()
		case '{':
			s.brace()
		case '}', ']':
			s.i++
			s.out.WriteString(lit(c))
		default: // ) | * + ? ^ $ and ordinary characters
			s.i++
			s.out.WriteRune(c)
		}
		if err != nil {
			return "", err
		}
	}
	return s.out.String(), nil
}

func (s *scanner) err(at int, format string, a ...any) error {
	return &Error{Offset: at, Reason: fmt.Sprintf(format, a...)}
}

func (s *scanner) peek(n int) rune {
	if s.i+n < len(s.p) {
		return s.p[s.i+n]
	}
	return -1
}

func isDigit(r rune) bool { return r >= '0' && r <= '9' }

// brace handles '{': a well-formed quantifier is kept, anything else is a literal.
func (s *scanner) brace() {
	j := s.i + 1
	start := j
	for j < len(s.p) && isDigit(s.p[j]) {
		j++
	}
	if j > start {
		if j < len(s.p) && s.p[j] == ',' {
			j++
			for j < len(s.p) && isDigit(s.p[j]) {
				j++
			}
		}
		if j < len(s.p) && s.p[j] == '}' {
			s.out.WriteString(string(s.p[s.i : j+1]))
			s.i = j + 1
			return
		}
	}
	s.i++
	s.out.WriteString(lit('{'))
}

func (s *scanner) group() error {
	start := s.i
	if s.peek(1) != '?' {
		s.i++
		s.out.WriteByte('(')
		return nil
	}
	switch s.peek(2) {
	case ':':
		s.i += 3
		s.out.WriteString("(?:")
	case '=', '!':
		return s.err(start, "lookahead is not supported")
	case '<':
		if c := s.peek(3); c == '=' || c == '!' {
			return s.err(start, "lookbehind is not supported")
		}
		j := s.i + 3
		for j < len(s.p) && s.p[j] != '>' {
			j++
		}
		if j >= len(s.p) {
			return s.err(start, "unterminated group name")
		}
		name := s.p[s.i+3 : j]
		if len(name) == 0 || isDigit(name[0]) {
			return s.err(start, "invalid group name")
		}
		for _, c := range name {
			if !(isDigit(c) || c == '_' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z') {
				return s.err(start, "invalid group name (only ASCII letters, digits and _ are supported)")
			}
		}
		s.out.WriteString("(?P<" + string(name) + ">")
		s.i = j + 1
	default:
		return s.err(start, "unsupported group syntax (inline flags and RE2 groups are not ECMA-262)")
	}
	return nil
}

func (s *scanner) class() error {
	start := s.i
	s.i++
	neg := false
	if s.peek(0) == '^' {
		neg = true
		s.i++
	}
	if s.peek(0) == ']' { // ECMA: [] matches nothing, [^] matches anything
		s.i++
		if neg {
			s.out.WriteString(`[\x{0}-\x{10FFFF}]`)
		} else {
			s.out.WriteString(`[^\x{0}-\x{10FFFF}]`)
		}
		return nil
	}
	var b strings.Builder
	for {
		if s.i >= len(s.p) {
			return s.err(start, "unterminated character class")
		}
		if s.p[s.i] == ']' {
			s.i++
			break
		}
		at := s.i
		a, err := s.classAtom()
		if err != nil {
			return err
		}
		if s.peek(0) == '-' && s.peek(1) != ']' && s.peek(1) != -1 {
			if !a.single {
				return s.err(at, "class escape cannot be a range endpoint")
			}
			s.i++
			at2 := s.i
			z, err := s.classAtom()
			if err != nil {
				return err
			}
			if !z.single {
				return s.err(at2, "class escape cannot be a range endpoint")
			}
			if z.r < a.r {
				return s.err(at, "character class range out of order")
			}
			b.WriteString(lit(a.r) + "-" + lit(z.r))
			continue
		}
		if a.single {
			b.WriteString(lit(a.r))
		} else {
			b.WriteString(a.text)
		}
	}
	if neg {
		s.out.WriteString("[^" + b.String() + "]")
	} else {
		s.out.WriteString("[" + b.String() + "]")
	}
	return nil
}

func (s *scanner) classAtom() (atom, error) {
	if s.p[s.i] == '\\' {
		return s.escape(true)
	}
	r := s.p[s.i]
	s.i++
	return atom{single: true, r: r}, nil
}

func (s *scanner) hex(n int) (rune, bool) {
	if s.i+n > len(s.p) {
		return 0, false
	}
	var v rune
	for _, c := range s.p[s.i : s.i+n] {
		switch {
		case isDigit(c):
			v = v<<4 | (c - '0')
		case c >= 'a' && c <= 'f':
			v = v<<4 | (c - 'a' + 10)
		case c >= 'A' && c <= 'F':
			v = v<<4 | (c - 'A' + 10)
		default:
			return 0, false
		}
	}
	s.i += n
	return v, true
}

func isSurrogate(r rune) bool { return r >= 0xD800 && r <= 0xDFFF }

// escape parses the escape starting at s.p[s.i] == '\\'.
func (s *scanner) escape(inClass bool) (atom, error) {
	start := s.i
	s.i++
	if s.i >= len(s.p) {
		return atom{}, s.err(start, "trailing backslash")
	}
	c := s.p[s.i]
	s.i++
	set := func(rs []rng) (atom, error) {
		if inClass {
			return atom{text: body(rs)}, nil
		}
		return atom{text: "[" + body(rs) + "]"}, nil
	}
	one := func(r rune) (atom, error) { return atom{single: true, r: r}, nil }
	switch c {
	case 'd':
		return set(digitSet)
	case 'D':
		return set(complement(digitSet))
	case 'w':
		return set(wordSet)
	case 'W':
		return set(complement(wordSet))
	case 's':
		return set(spaceSet)
	case 'S':
		return set(complement(spaceSet))
	case 'p', 'P':
		if s.peek(0) != '{' {
			return atom{}, s.err(start, `\%c must be followed by {property}`, c)
		}
		j := s.i + 1
		for j < len(s.p) && s.p[j] != '}' {
			j++
		}
		if j >= len(s.p) {
			return atom{}, s.err(start, "unterminated Unicode property escape")
		}
		name, err := resolveProperty(string(s.p[s.i+1 : j]))
		if err != nil {
			return atom{}, s.err(start, "%s", err.Error())
		}
		s.i = j + 1
		return atom{text: `\` + string(c) + `{` + name + `}`}, nil
	case 'u':
		r, err := s.unicode(start)
		if err != nil {
			return atom{}, err
		}
		return one(r)
	case 'x':
		r, ok := s.hex(2)
		if !ok {
			return atom{}, s.err(start, `\x must be followed by two hex digits`)
		}
		return one(r)
	case 'c':
		l := s.peek(0)
		if !(l >= 'A' && l <= 'Z' || l >= 'a' && l <= 'z') {
			return atom{}, s.err(start, `\c must be followed by an ASCII letter`)
		}
		s.i++
		return one(l % 32)
	case '0':
		if isDigit(s.peek(0)) {
			return atom{}, s.err(start, "octal escapes are not supported")
		}
		return one(0)
	case 'k':
		return atom{}, s.err(start, "named backreferences are not supported")
	case 'b':
		if inClass {
			return one(8)
		}
		return atom{text: `\b`}, nil
	case 'B':
		if inClass {
			return atom{}, s.err(start, `\B is not valid in a character class`)
		}
		return atom{text: `\B`}, nil
	case 'f':
		return one(12)
	case 'n':
		return one('\n')
	case 'r':
		return one('\r')
	case 't':
		return one('\t')
	case 'v':
		return one(11)
	}
	if c >= '1' && c <= '9' {
		return atom{}, s.err(start, "backreferences are not supported")
	}
	// Identity escape of ASCII punctuation (\/ \_ \- \. \( ...).
	if c < 0x80 && !isDigit(c) && !(c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z') {
		return one(c)
	}
	return atom{}, s.err(start, "unknown escape")
}

// unicode parses after "\u": XXXX, a surrogate pair of two \u escapes, or {X...}.
func (s *scanner) unicode(start int) (rune, error) {
	if s.peek(0) == '{' {
		j := s.i + 1
		var v rune
		for ; j < len(s.p) && s.p[j] != '}'; j++ {
			save := s.i
			s.i = j
			d, ok := s.hex(1)
			s.i = save
			if !ok || j-s.i > 6 {
				return 0, s.err(start, `invalid \u{...} escape`)
			}
			v = v<<4 | d
			if v > maxRune {
				return 0, s.err(start, `\u{...} escape out of range`)
			}
		}
		if j >= len(s.p) || j == s.i+1 {
			return 0, s.err(start, `invalid \u{...} escape`)
		}
		if isSurrogate(v) {
			return 0, s.err(start, "lone surrogate escape")
		}
		s.i = j + 1
		return v, nil
	}
	v, ok := s.hex(4)
	if !ok {
		return 0, s.err(start, `\u must be followed by four hex digits`)
	}
	if !isSurrogate(v) {
		return v, nil
	}
	if v < 0xDC00 && s.peek(0) == '\\' && s.peek(1) == 'u' {
		save := s.i
		s.i += 2
		if lo, ok := s.hex(4); ok && lo >= 0xDC00 && lo <= 0xDFFF {
			return 0x10000 + (v-0xD800)<<10 + (lo - 0xDC00), nil
		}
		s.i = save
	}
	return 0, s.err(start, "lone surrogate escape")
}
