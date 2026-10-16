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
	esc(`{"@FFFD":["x` + rawFFFD + `",{"y":"@fffd@D83D@DE00"}],"z":"z"}`), `[["a"],["a"],{"a":"a"}]`,
}

// Strings: RFC 8259 §7 (escapes, surrogate pairs) and §8.1 / §8.2 (UTF-8).

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
			checkFFFDProvenance(t, data, v)
		}
	})
}

// checkFFFDProvenance: when Decode succeeds, every string and member name in
// its tree is the decoding of exactly one span scanText reported, and its
// U+FFFD count equals that span's fffd. So no replacement character appears
// that the input did not write, at any depth.
func checkFFFDProvenance(t *testing.T, data []byte, v any) {
	t.Helper()
	want := map[string]int{} // decoded span text -> remaining uses
	if _, reason := scanText(data, func(s, e, n int) {
		var str string
		if err := json.Unmarshal(data[s:e], &str); err != nil {
			t.Fatalf("span %d-%d does not decode: %v", s, e, err)
		}
		if strings.Count(str, "�") != n {
			t.Fatalf("span %d-%d: %d U+FFFD, scan said %d", s, e, strings.Count(str, "�"), n)
		}
		want[str]++
	}); reason != "" {
		t.Fatalf("Decode accepted what scanText rejected: %s", reason)
	}
	var walk func(any)
	use := func(s string) {
		if want[s] == 0 {
			t.Fatalf("decoded string %q has no matching span", s)
		}
		want[s]--
	}
	walk = func(v any) {
		switch x := v.(type) {
		case string:
			use(x)
		case []any:
			for _, e := range x {
				walk(e)
			}
		case map[string]any:
			for k, e := range x {
				use(k)
				walk(e)
			}
		}
	}
	walk(v)
	for s, n := range want {
		if n != 0 {
			t.Fatalf("span %q not in decoded tree", s)
		}
	}
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
// RFC 8259 §7, §8.2: checked for every string in the tree, not just the top.
func TestDecodedFFFDInvariant(t *testing.T) {
	decoded := 0
	for _, s := range scanSeeds {
		v, err := Decode([]byte(s))
		if err != nil {
			continue
		}
		decoded++
		checkFFFDProvenance(t, []byte(s), v)
	}
	if decoded < 10 {
		t.Fatalf("only %d seeds decoded", decoded)
	}
	if v, _ := Decode([]byte(esc(`"@FFFD` + rawFFFD + `"`))); strings.Count(v.(string), "�") != 2 {
		t.Error("expected two U+FFFD")
	}
}
