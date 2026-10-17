package formats

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// validECMAPattern reports whether s is a valid ECMA-262 pattern under the
// strict (u flag) grammar: no Annex B leniencies, so "\a", a lone "{", "}" or
// "]", quantified lookaheads and out-of-range back-references are errors.
// JavaScript-only constructs that RE2 cannot run (lookahead, lookbehind,
// back-references, named groups) are VALID here: the verdict is by syntax only.
//
// It is one forward scan with an explicit heap slice as the group stack and
// no recursion, so it is linear in len(s) and needs no depth cap. Back
// references are resolved after the scan against the collected sets.
// ponytail: partly checked: group names accept ID_Start/ID_Continue as
// approximated by Go's unicode tables; the ES2025 "(?i:...)" modifier groups
// are rejected (fail closed); \p{...} names come from ecma_unicode_props.go.
func validECMAPattern(s string) bool {
	if !utf8.ValidString(s) {
		return false
	}
	r := []rune(s)
	n := len(r)
	type frame struct {
		quantifiable bool // may the closed group be quantified?
		all, cur     map[string]bool
	}
	stack := []frame{{}} // stack[0] is the root, never popped
	declared := map[string]bool{}
	var refs []string
	groups, maxRef := 0, 0
	canQuant := false
	for i := 0; i < n; {
		switch c := r[i]; c {
		case '\\':
			if i+1 >= n {
				return false
			}
			e := parseEscape(r, i+1, false)
			if !e.ok {
				return false
			}
			i = e.next
			canQuant = !e.assertion
			if e.backref > maxRef {
				maxRef = e.backref
			}
			if e.kref != "" {
				refs = append(refs, e.kref)
			}
		case '[':
			j, ok := scanClass(r, i+1)
			if !ok {
				return false
			}
			i, canQuant = j, true
		case '(':
			f := frame{quantifiable: true}
			j := i + 1
			if j < n && r[j] == '?' {
				switch {
				case j+1 < n && r[j+1] == ':':
					j += 2
				case j+1 < n && (r[j+1] == '=' || r[j+1] == '!'):
					f.quantifiable = false
					j += 2
				case j+1 < n && r[j+1] == '<' && j+2 < n && (r[j+2] == '=' || r[j+2] == '!'):
					f.quantifiable = false
					j += 3
				case j+1 < n && r[j+1] == '<':
					name, next, ok := parseGroupName(r, j+2)
					if !ok {
						return false
					}
					top := &stack[len(stack)-1]
					if top.cur[name] {
						return false
					}
					if top.cur == nil {
						top.cur = map[string]bool{}
					}
					top.cur[name] = true
					declared[name] = true
					groups++
					j = next
				default:
					return false
				}
			} else {
				groups++
			}
			stack = append(stack, f)
			i, canQuant = j, false
		case ')':
			if len(stack) == 1 {
				return false
			}
			f := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			all, _ := unionNames(f.all, f.cur, false)
			parent := &stack[len(stack)-1]
			var ok bool
			if parent.cur, ok = unionNames(parent.cur, all, true); !ok {
				return false
			}
			i, canQuant = i+1, f.quantifiable
		case '|':
			top := &stack[len(stack)-1]
			top.all, _ = unionNames(top.all, top.cur, false)
			top.cur = nil
			i, canQuant = i+1, false
		case '^', '$':
			i, canQuant = i+1, false
		case '*', '+', '?':
			if !canQuant {
				return false
			}
			i++
			if i < n && r[i] == '?' {
				i++
			}
			canQuant = false
		case '{':
			j, ok := scanBraceQuantifier(r, i)
			if !ok || !canQuant {
				return false
			}
			i, canQuant = j, false
		case '}', ']':
			return false
		default:
			i, canQuant = i+1, true
		}
	}
	if len(stack) != 1 || maxRef > groups {
		return false
	}
	for _, name := range refs {
		if !declared[name] {
			return false
		}
	}
	return true
}

// unionNames merges b into a, always iterating the smaller set so repeated
// merges stay O(n log n). With check, a shared name makes ok false.
func unionNames(a, b map[string]bool, check bool) (map[string]bool, bool) {
	if len(a) < len(b) {
		a, b = b, a
	}
	for k := range b {
		if check && a[k] {
			return a, false
		}
		a[k] = true
	}
	return a, true
}

// scanBraceQuantifier parses {n}, {n,} or {n,m}[?] at r[i]=='{' and returns
// the index after it. Bounds are compared as digit strings, so no overflow.
func scanBraceQuantifier(r []rune, i int) (int, bool) {
	j := i + 1
	digits := func() string {
		st := j
		for j < len(r) && r[j] >= '0' && r[j] <= '9' {
			j++
		}
		return string(r[st:j])
	}
	lo := digits()
	if lo == "" {
		return 0, false
	}
	if j < len(r) && r[j] == ',' {
		j++
		if hi := digits(); hi != "" && cmpDigits(hi, lo) < 0 {
			return 0, false
		}
	}
	if j >= len(r) || r[j] != '}' {
		return 0, false
	}
	j++
	if j < len(r) && r[j] == '?' {
		j++
	}
	return j, true
}

// cmpDigits compares two non-negative decimal strings of any length.
func cmpDigits(a, b string) int {
	a, b = strings.TrimLeft(a, "0"), strings.TrimLeft(b, "0")
	if len(a) != len(b) {
		return len(a) - len(b)
	}
	return strings.Compare(a, b)
}

// scanClass scans a character class whose "[" is just before i and returns
// the index after the closing "]". Classes do not nest under the u flag.
func scanClass(r []rune, i int) (int, bool) {
	if i < len(r) && r[i] == '^' {
		i++
	}
	for i < len(r) && r[i] != ']' {
		a, j, ok := classAtom(r, i)
		if !ok {
			return 0, false
		}
		i = j
		if i+1 < len(r) && r[i] == '-' && r[i+1] != ']' {
			b, j, ok := classAtom(r, i+1)
			if !ok || a.set || b.set || a.val > b.val {
				return 0, false
			}
			i = j
		}
	}
	if i >= len(r) {
		return 0, false
	}
	return i + 1, true
}

type escape struct {
	val       rune
	set       bool   // \d \w \p{..} and friends: a set, not one character
	assertion bool   // \b \B
	backref   int    // \N, saturated
	kref      string // \k<name>
	next      int
	ok        bool
}

func classAtom(r []rune, i int) (escape, int, bool) {
	if r[i] != '\\' {
		return escape{val: r[i]}, i + 1, true
	}
	if i+1 >= len(r) {
		return escape{}, 0, false
	}
	e := parseEscape(r, i+1, true)
	return e, e.next, e.ok
}

func hexVal(c rune) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	}
	return -1
}

// hexN reads exactly n hex digits at r[i:].
func hexN(r []rune, i, n int) (rune, bool) {
	if i+n > len(r) {
		return 0, false
	}
	v := 0
	for _, c := range r[i : i+n] {
		h := hexVal(c)
		if h < 0 {
			return 0, false
		}
		v = v*16 + h
	}
	return rune(v), true
}

// parseUnicodeEscape reads what follows "\u": XXXX or {X...}.
func parseUnicodeEscape(r []rune, i int) (rune, int, bool) {
	if i < len(r) && r[i] == '{' {
		v, j := 0, i+1
		for j < len(r) && hexVal(r[j]) >= 0 {
			if v = v*16 + hexVal(r[j]); v > unicode.MaxRune {
				return 0, 0, false
			}
			j++
		}
		if j == i+1 || j >= len(r) || r[j] != '}' {
			return 0, 0, false
		}
		return rune(v), j + 1, true
	}
	v, ok := hexN(r, i, 4)
	return v, i + 4, ok
}

// parseGroupName reads an identifier up to ">" starting at r[i].
func parseGroupName(r []rune, i int) (string, int, bool) {
	var sb strings.Builder
	for first := true; ; first = false {
		if i >= len(r) {
			return "", 0, false
		}
		c := r[i]
		if c == '>' {
			return sb.String(), i + 1, !first
		}
		i++
		if c == '\\' {
			if i >= len(r) || r[i] != 'u' {
				return "", 0, false
			}
			var ok bool
			if c, i, ok = parseUnicodeEscape(r, i+1); !ok {
				return "", 0, false
			}
		}
		start := c == '$' || c == '_' || unicode.IsLetter(c) || unicode.Is(unicode.Nl, c)
		if !start && (first || !(unicode.In(c, unicode.Mn, unicode.Mc, unicode.Nd, unicode.Pc) || c == 0x200C || c == 0x200D)) {
			return "", 0, false
		}
		sb.WriteRune(c)
	}
}

// parseEscape parses the escape whose first rune after the backslash is r[i].
func parseEscape(r []rune, i int, inClass bool) escape {
	c := r[i]
	e := escape{next: i + 1, ok: true}
	switch c {
	case 'd', 'D', 's', 'S', 'w', 'W':
		e.set = true
	case 'p', 'P':
		if i+1 >= len(r) || r[i+1] != '{' {
			return escape{}
		}
		end := i + 2
		for end < len(r) && r[end] != '}' {
			end++
		}
		if end >= len(r) || !validECMAUnicodeProperty(string(r[i+2:end])) {
			return escape{}
		}
		e.set, e.next = true, end+1
	case 'b':
		if inClass {
			e.val = 8
		} else {
			e.assertion = true
		}
	case 'B':
		if inClass {
			return escape{}
		}
		e.assertion = true
	case 'f':
		e.val = '\f'
	case 'n':
		e.val = '\n'
	case 'r':
		e.val = '\r'
	case 't':
		e.val = '\t'
	case 'v':
		e.val = '\v'
	case 'c':
		if i+1 >= len(r) || !(r[i+1]|0x20 >= 'a' && r[i+1]|0x20 <= 'z' && r[i+1] < 0x80) {
			return escape{}
		}
		e.val, e.next = r[i+1]%32, i+2
	case '0':
		if i+1 < len(r) && r[i+1] >= '0' && r[i+1] <= '9' {
			return escape{}
		}
	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		if inClass {
			return escape{}
		}
		j, v := i, 0
		for ; j < len(r) && r[j] >= '0' && r[j] <= '9'; j++ {
			if v < 1<<30 {
				v = v*10 + int(r[j]-'0')
			}
		}
		e.backref, e.next = v, j
	case 'k':
		if inClass || i+1 >= len(r) || r[i+1] != '<' {
			return escape{}
		}
		name, j, ok := parseGroupName(r, i+2)
		if !ok {
			return escape{}
		}
		e.kref, e.next = name, j
	case 'x':
		v, ok := hexN(r, i+1, 2)
		if !ok {
			return escape{}
		}
		e.val, e.next = v, i+3
	case 'u':
		v, j, ok := parseUnicodeEscape(r, i+1)
		if !ok {
			return escape{}
		}
		e.val, e.next = v, j
	case '-':
		if !inClass {
			return escape{}
		}
		e.val = '-'
	default:
		if !strings.ContainsRune("^$\\.*+?()[]{}|/", c) {
			return escape{}
		}
		e.val = c
	}
	return e
}
