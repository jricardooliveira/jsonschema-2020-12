package compile

import (
	"strings"
	"testing"

	"github.com/neoface/jsonschema/internal/jsonvalue"
	_ "github.com/neoface/jsonschema/internal/keywords/all"
)

func load(t *testing.T, src string) (*Node, error) {
	t.Helper()
	v, err := jsonvalue.DecodeLimited([]byte(src), jsonvalue.NumberLimits{})
	if err != nil {
		t.Fatal(err)
	}
	return CompileDocument(v, "file:///s.json")
}

func TestResolves(t *testing.T) { // core 8.2.3.1, 8.2.4, 9.4
	for _, src := range []string{
		`{"$defs":{"a":{"type":"string"}},"$ref":"#/$defs/a"}`,
		`{"$defs":{"a/b":true},"$ref":"#/$defs/a~1b"}`,
		`{"$defs":{"a%b":true},"$ref":"#/$defs/a%25b"}`,
		`{"$defs":{"a":{"$anchor":"x"}},"$ref":"#x"}`,
		`{"$ref":"s.json#/$defs/a","$defs":{"a":false}}`,
		`{"$id":"http://Ex.com/r","$defs":{"e":{"$id":"e.json","$ref":"r#/$defs/z"},"z":true},"$ref":"http://ex.com/e.json"}`,
		`{"definitions":{"x":{"type":"string"}},"$ref":"#/definitions/x"}`,
		`{"definitions":{"x":true},"$ref":"#/definitions/x"}`,
		`{"properties":{"next":{"$ref":"#"}}}`, // recursion that descends
	} {
		if _, err := load(t, src); err != nil {
			t.Errorf("%s: %v", src, err)
		}
	}
}

func TestBrokenRefs(t *testing.T) {
	for src, want := range map[string]string{
		`{"properties":{"a":{"$ref":"#/nope"}}}`:                   "/properties/a/$ref", // unreached branch
		`{"$defs":{"d":{"$ref":"#/nope"}}}`:                        "/$defs/d/$ref",      // inside $defs
		`{"$ref":"#/a%zz"}`:                                        "percent",
		`{"$ref":"other.json"}`:                                    "unresolved",
		`{"$ref":"file:///etc/passwd"}`:                            "unresolved",
		`{"$ref":"#/enum/0","enum":[1]}`:                           "not an object or boolean",
		`{"$ref":"#/x","x":"s"}`:                                   "not an object or boolean",
		`{"$ref":"#/x","x":[{}]}`:                                  "not an object or boolean",
		`{"$ref":"#/x","x":null}`:                                  "not an object or boolean",
		`{"$ref":"#nope"}`:                                         "no anchor",
		`{"$defs":{"A":{"$id":"http://x/A"}},"$ref":"http://x/a"}`: "unresolved", // path case differs
		`{"$ref":"#/$defs/a","$defs":{"a":{"$ref":"#/$defs/b"},"b":{"$ref":"#/$defs/a"}}}`: "cycle",
		`{"$ref":"#"}`: "cycle",
	} {
		_, err := load(t, src)
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: err=%v want %q", src, err, want)
		}
	}
}

// Load time stays linear on a hostile schema with a $ref at every level:
// each schema object is compiled once, never once per $ref above it.
func TestDeepRefChainCompilesEachNodeOnce(t *testing.T) {
	n := jsonvalue.MaxDepth - 1
	src := strings.Repeat(`{"$ref":"#","not":`, n) + `{"$ref":"#"}` + strings.Repeat(`}`, n)
	v, err := jsonvalue.DecodeLimited([]byte(src), jsonvalue.NumberLimits{})
	if err != nil {
		t.Fatal(err)
	}
	r, _, err := compileDocument(v, "file:///s.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(r.nodes) != n+1 || len(r.refs) != n+1 {
		t.Errorf("nodes=%d refs=%d, want %d each", len(r.nodes), len(r.refs), n+1)
	}
}

// Targets under unknown keywords, referenced deepest first, are compiled once.
func TestDeepUnknownKeywordTargetsCompileOnce(t *testing.T) {
	const n = 200
	var refs []string
	for d := n; d >= 1; d-- {
		refs = append(refs, `{"$ref":"#/definitions/x`+strings.Repeat("/not", d)+`"}`)
	}
	src := `{"allOf":[` + strings.Join(refs, ",") + `],"definitions":{"x":` +
		strings.Repeat(`{"not":`, n) + `true` + strings.Repeat(`}`, n) + `}}`
	v, err := jsonvalue.DecodeLimited([]byte(src), jsonvalue.NumberLimits{})
	if err != nil {
		t.Fatal(err)
	}
	r, _, err := compileDocument(v, "file:///s.json")
	if err != nil {
		t.Fatal(err)
	}
	if want := 1 + n + n; len(r.nodes) != want { // root, allOf items, x's chain below /not
		t.Errorf("nodes=%d want %d", len(r.nodes), want)
	}
}

// core 8.2.2 with RFC 3986: a percent-encoded plain-name fragment is decoded.
func TestPercentEncodedAnchor(t *testing.T) {
	if _, err := load(t, `{"$defs":{"a":{"$anchor":"A"}},"$ref":"#%41"}`); err != nil {
		t.Error(err)
	}
}

func TestNotACycle(t *testing.T) {
	for _, src := range []string{
		`{"$defs":{"a":{"$ref":"#/$defs/b"},"b":{"type":"string"}},"$ref":"#/$defs/a"}`,
		`{"$ref":"#","type":"object"}`,
	} {
		if _, err := load(t, src); err != nil {
			t.Errorf("%s: %v", src, err)
		}
	}
}

func TestNormalizedIDMatch(t *testing.T) { // core 8.2.1 with RFC 3986 6.2.2
	_, err := load(t, `{"$defs":{"e":{"$id":"https://Example.com/a%7Eb.json"}},"$ref":"https://example.com/a~b.json"}`)
	if err != nil {
		t.Error(err)
	}
}

func TestRootBaseURI(t *testing.T) { // core 9.1
	if got, err := RootBaseURI("", ""); err != nil || got != StdinBaseURI {
		t.Errorf("stdin: %q %v", got, err)
	}
	if got, err := RootBaseURI("", "/home/me/my schema.json"); err != nil || got != "file:///home/me/my%20schema.json" {
		t.Errorf("file: %q %v", got, err)
	}
	if got, err := RootBaseURI("https://X.com/a/#", "/f.json"); err != nil || got != "https://x.com/a/" {
		t.Errorf("override: %q %v", got, err)
	}
	for _, bad := range []string{"rel/path", "http://x/a#frag", "http://x/%zz"} {
		if _, err := RootBaseURI(bad, ""); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
	// the base URI changes how a relative reference resolves
	v, _ := jsonvalue.DecodeLimited([]byte(`{"$defs":{"e":{"$id":"http://x/y/e.json"}},"$ref":"e.json"}`), jsonvalue.NumberLimits{})
	if _, err := CompileDocument(v, "http://x/y/s.json"); err != nil {
		t.Error(err)
	}
	if _, err := CompileDocument(v, "http://x/z/s.json"); err == nil {
		t.Error("expected unresolved")
	}
}
