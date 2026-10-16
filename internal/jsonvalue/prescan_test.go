package jsonvalue

import (
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf8"
)

// esc turns "@" into a backslash-u escape introducer, so test text stays
// readable and survives tooling that rewrites escape sequences.
func esc(s string) string { return strings.ReplaceAll(s, "@", `\u`) }

const rawFFFD = "\xEF\xBF\xBD" // U+FFFD written literally in UTF-8

var scanSeeds = []string{
	`"ok"`, esc(`"@D83D@DE00"`), esc(`"@D800"`), esc(`"@DC00"`), esc(`"@D800@0041"`), esc(`"@D800\\@DC00"`),
	esc(`"@D800\@DC00"`), `"\\"`, `"\\\\"`, `"\\\""`, "\"\xff\"", "\xff", "\"\xed\xa0\x80\"",
	esc(`"@D83D`), esc(`"@D83D@`), `"` + rawFFFD + `"`, esc(`"@FFFD"`), esc(`"@fffd@FFFD"`),
	`["a","b` + "\xc3\xa9" + `"]`, `{"k":"v"}`, esc(`"@12`),
}

func TestScanTextTable(t *testing.T) {
	bad := map[string]string{
		esc(`"@D800"`): reasonSurrogate, esc(`"@DC00"`): reasonSurrogate, esc(`"@D800@0041"`): reasonSurrogate,
		esc(`"@D800\\@DC00"`): reasonSurrogate, "\"\xff\"": reasonUTF8, "\xff": reasonUTF8,
		"\"\xed\xa0\x80\"": reasonUTF8, esc(`"@D83D`): reasonSurrogate, esc(`"@D83D@`): reasonSurrogate,
	}
	for in, want := range bad {
		if _, got := scanText([]byte(in), nil); got != want {
			t.Errorf("%q: got %q want %q", in, got, want)
		}
	}
	good := []string{esc(`"@D83D@DE00"`), `"\\"`, `"\\\\"`, `"\\\""`, `"\\uD800"`, `"` + rawFFFD + `"`, esc(`"@12`), `"a\`}
	for _, in := range good {
		if off, got := scanText([]byte(in), nil); got != "" {
			t.Errorf("%q: unexpected %q at %d", in, got, off)
		}
	}
}

func FuzzScanTextAgreesWithEncodingJSON(f *testing.F) {
	for _, s := range scanSeeds {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		type span struct{ start, end, fffd int }
		var spans []span
		_, reason := scanText(data, func(s, e, n int) { spans = append(spans, span{s, e, n}) })
		if reason == "" && json.Valid(data) {
			var toks []string
			dec := json.NewDecoder(strings.NewReader(string(data)))
			for {
				tok, err := dec.Token()
				if err != nil {
					break
				}
				if s, ok := tok.(string); ok {
					toks = append(toks, s)
				}
			}
			if len(toks) != len(spans) {
				t.Fatalf("string count: %d vs %d", len(toks), len(spans))
			}
			for i, sp := range spans {
				var s string
				if err := json.Unmarshal(data[sp.start:sp.end], &s); err != nil {
					t.Fatal(err)
				}
				if s != toks[i] || !utf8.ValidString(s) || strings.Count(s, "�") != sp.fffd {
					t.Fatalf("span %d mismatch", i)
				}
			}
		}
		if v, err := Decode(data); err == nil {
			checkStrings(t, v)
		}
	})
}

// checkStrings: every decoded string and member name is valid UTF-8.
func checkStrings(t *testing.T, v any) {
	switch x := v.(type) {
	case string:
		if !utf8.ValidString(x) {
			t.Fatal("invalid UTF-8 in decoded string")
		}
	case []any:
		for _, e := range x {
			checkStrings(t, e)
		}
	case map[string]any:
		for k, e := range x {
			checkStrings(t, k)
			checkStrings(t, e)
		}
	}
}

// The U+FFFD count of a decoded string equals what scanText reported, so a
// replacement character never appears that the input did not write.
func TestDecodedFFFDInvariant(t *testing.T) {
	for _, s := range scanSeeds {
		var want []int
		if off, reason := scanText([]byte(s), func(_, _, n int) { want = append(want, n) }); reason != "" {
			_ = off
			continue
		}
		v, err := Decode([]byte(s))
		if err != nil {
			continue
		}
		if str, ok := v.(string); ok && (len(want) != 1 || strings.Count(str, "�") != want[0]) {
			t.Errorf("%q: %d FFFD, scan said %v", s, strings.Count(str, "�"), want)
		}
	}
	if v, _ := Decode([]byte(esc(`"@FFFD` + rawFFFD + `"`))); strings.Count(v.(string), "�") != 2 {
		t.Error("expected two U+FFFD")
	}
}
