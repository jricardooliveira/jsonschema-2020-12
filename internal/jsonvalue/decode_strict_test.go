package jsonvalue

import (
	"errors"
	"strings"
	"testing"
)

func TestStrictRejections(t *testing.T) {
	bom := "\xEF\xBB\xBF"
	cases := map[string]struct {
		in     string
		reason string
		off    int64
	}{
		"invalid utf8":   {"[\"a\xffb\"]", reasonUTF8, 3},
		"utf8 surrogate": {"\"\xed\xa0\x80\"", reasonUTF8, 1},
		"lone high":      {esc(`"@D800"`), reasonSurrogate, 1},
		"lone low":       {esc(`["@DC00"]`), reasonSurrogate, 2},
		"dup deep":       {`{"a":[{"b":1,"c":2,"b":3}]}`, "duplicate member name", 0},
		"dup escaped":    {esc(`{"a":1,"@0061":2}`), "duplicate member name", 0},
		"trailing value": {`{} []`, "trailing data after value", 0},
		"two boms":       {bom + bom + `{}`, "syntax error", 0},
		"bom offset":     {bom + "[\xff]", reasonUTF8, 4},
	}
	for name, c := range cases {
		_, err := Decode([]byte(c.in))
		var se *SyntaxError
		if !errors.As(err, &se) {
			t.Errorf("%s: want SyntaxError, got %v", name, err)
			continue
		}
		if se.Reason != c.reason {
			t.Errorf("%s: reason %q want %q", name, se.Reason, c.reason)
		}
		if c.off != 0 && se.Offset != c.off {
			t.Errorf("%s: offset %d want %d", name, se.Offset, c.off)
		}
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
