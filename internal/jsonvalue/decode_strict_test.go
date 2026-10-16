package jsonvalue

import (
	"errors"
	"strings"
	"testing"
)

func TestStrictRejections(t *testing.T) {
	bom := "\xEF\xBB\xBF"
	// off is the byte offset of the offending text in the original input.
	cases := map[string]struct {
		spec   string
		in     string
		reason string
		off    int64
	}{
		"invalid utf8":   {"RFC 8259 §8.1", "[\"a\xffb\"]", reasonUTF8, 3},
		"utf8 surrogate": {"RFC 8259 §8.1", "\"\xed\xa0\x80\"", reasonUTF8, 1},
		"lone high":      {"RFC 8259 §7, §8.2", esc(`"@D800"`), reasonSurrogate, 1},
		"lone low":       {"RFC 8259 §7, §8.2", esc(`["@DC00"]`), reasonSurrogate, 2},
		"dup top":        {"RFC 8259 §4", `{"age":5,"age":9}`, "duplicate member name", 9},
		"dup deep":       {"RFC 8259 §4", `{"a":[{"b":1,"c":2,"b":3}]}`, "duplicate member name", 19},
		"dup spaced":     {"RFC 8259 §4", "{\"a\":1 ,\n \"a\":2}", "duplicate member name", 10},
		"dup escaped":    {"RFC 8259 §4, §7", esc(`{"a":1,"@0061":2}`), "duplicate member name", 7},
		"dup after bom":  {"RFC 8259 §4, §8.1", bom + `{"a":1,"a":2}`, "duplicate member name", 10},
		"trailing value": {"RFC 8259 §2", `{} []`, "trailing data after value", 3},
		"two boms":       {"RFC 8259 §8.1", bom + bom + `{}`, "syntax error", 3},
		"bom offset":     {"RFC 8259 §8.1", bom + "[\xff]", reasonUTF8, 4},
	}
	for name, c := range cases {
		_, err := Decode([]byte(c.in))
		var se *SyntaxError
		if !errors.As(err, &se) {
			t.Errorf("%s (%s): want SyntaxError, got %v", name, c.spec, err)
			continue
		}
		if se.Reason != c.reason {
			t.Errorf("%s (%s): reason %q want %q", name, c.spec, se.Reason, c.reason)
		}
		if se.Offset != c.off {
			t.Errorf("%s (%s): offset %d want %d", name, c.spec, se.Offset, c.off)
		}
	}
}

// The duplicate error names the line and column of the repeated name.
func TestDuplicateLineColumn(t *testing.T) { // RFC 8259 §4
	_, err := Decode([]byte("{\n  \"age\": 5,\n  \"age\": 9\n}"))
	var se *SyntaxError
	if !errors.As(err, &se) || se.Line != 3 || se.Column != 3 {
		t.Fatalf("got %+v", se)
	}
}

func TestStrictAcceptances(t *testing.T) {
	v, err := DecodeLimited([]byte("\xEF\xBB\xBF \n"+esc(`["@D83D@DE00", "@00e9", 12345678901234567890, 0.1, 1.0]`)+"\t"), NumberLimits{})
	if err != nil {
		t.Fatal(err)
	}
	a := v.([]any)
	if a[0] != "\U0001F600" || a[1] != "é" {
		t.Errorf("strings: %q %q", a[0], a[1])
	}
	if !a[4].(*Number).IsInteger() || a[3].(*Number).IsInteger() {
		t.Error("integer detection wrong")
	}
	if a[2].(*Number).Rat().String() != "12345678901234567890/1" {
		t.Error("big integer lost precision")
	}
}

func TestDuplicateAndLimitErrorsQuoteNoContent(t *testing.T) {
	for _, in := range []string{`{"s3cr3t":1,"s3cr3t":2}`, "[\"s3cr3t\xff\"]", esc(`"s3cr3t@D800"`)} {
		_, err := Decode([]byte(in))
		if err == nil || strings.Contains(err.Error(), "s3cr3t") {
			t.Errorf("%q: bad error %v", in, err)
		}
	}
}

func TestDepthCheckBeforeDescent(t *testing.T) {
	_, err := Decode([]byte(strings.Repeat("[", MaxDepth+1)))
	var le *LimitError
	if !errors.As(err, &le) || le.Limit != "nesting depth" || le.Max != MaxDepth {
		t.Fatalf("got %v", err)
	}
}
