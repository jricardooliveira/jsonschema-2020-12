package jsonschema

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
)

func TestOutputShapes(t *testing.T) {
	const secret = "s3cr3t-value"
	s, err := Compile([]byte(`{"type":"number"}`))
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.Validate([]byte(`"` + secret + `"`))
	if err != nil {
		t.Fatal(err)
	}
	for shape, want := range map[string]string{
		"flag":  `{"valid":false}`,
		"basic": `{"valid":false,"errors":[{"instanceLocation":"","keywordLocation":"/type","error":"value must be of type number"}]}`,
	} {
		b, err := r.Output(shape)
		if err != nil || string(b) != want || strings.Contains(string(b), secret) {
			t.Errorf("%s: %s %v", shape, b, err)
		}
	}
	ok, _ := s.Validate([]byte(`1`))
	if b, _ := ok.Output("basic"); string(b) != `{"valid":true}` {
		t.Errorf("valid basic: %s", b)
	}
	if _, err := r.Output("fancy"); err == nil {
		t.Error("unknown shape must error")
	}
}

func TestOutputUnitCap(t *testing.T) {
	schema := []byte(`{"type":"string","const":"x"}`)
	s, err := CompileWithOptions(schema, Options{MaxOutputUnits: 1})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.Validate([]byte(`1`))
	var le *jsonvalue.LimitError
	if !errors.As(err, &le) || le.Limit != "output units" {
		t.Fatalf("want output units LimitError, got %v", err)
	}
	s, _ = CompileWithOptions(schema, Options{MaxOutputUnits: DefaultMaxOutputUnits})
	if _, err := s.Validate([]byte(`1`)); err != nil {
		t.Fatal(err)
	}
	if _, err := CompileWithOptions(schema, Options{MaxOutputUnits: -1}); err == nil {
		t.Error("negative cap must error")
	}
}

const nestedSchema = `{"$id":"http://x/r","$ref":"#/$defs/a%20b%25%C3%A9",
 "$defs":{"a b%é":{"type":"object","properties":{"name":{"type":"string","title":"T"},"n":{"type":"number"}}}}}`

func nestedResult(t *testing.T, opts Options, doc string) Result {
	t.Helper()
	s, err := CompileWithOptions([]byte(nestedSchema), opts)
	if err != nil {
		t.Fatal(err)
	}
	res, err := s.Validate([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestNestedShapesSelectableAndSafe(t *testing.T) {
	const secret = "SECRET-VALUE-42"
	for _, shape := range []string{"detailed", "verbose"} {
		for _, tc := range []struct {
			doc   string
			valid bool
		}{
			{`{"name":"` + secret + `","n":1}`, true},
			{`{"name":"` + secret + `","n":"` + secret + `"}`, false},
		} {
			res := nestedResult(t, Options{}, tc.doc)
			out, err := res.Output(shape)
			if err != nil || !json.Valid(out) {
				t.Fatalf("%s: %s %v", shape, out, err)
			}
			var m map[string]any
			if err := json.Unmarshal(out, &m); err != nil || m["valid"] != tc.valid {
				t.Errorf("%s valid=%v: %s", shape, tc.valid, out)
			}
			if strings.Contains(string(out), secret) {
				t.Errorf("%s leaked value: %s", shape, out)
			}
			if !tc.valid && !strings.Contains(string(out), "a%20b%25%C3%A9") {
				t.Errorf("%s: no encoded absoluteKeywordLocation: %s", shape, out)
			}
		}
	}
}

func TestNestedVerboseKeepsPassingDetailedDrops(t *testing.T) {
	res := nestedResult(t, Options{}, `{"n":1}`)
	v, err := res.Output("verbose")
	if err != nil || !json.Valid(v) {
		t.Fatalf("verbose: %s %v", v, err)
	}
	d, err := res.Output("detailed")
	if err != nil || !json.Valid(d) {
		t.Fatalf("detailed: %s %v", d, err)
	}
	const passN = `"keywordLocation":"/$ref/properties/n","instanceLocation":"/n"`
	if !strings.Contains(string(v), passN) {
		t.Errorf("verbose lacks passing node %s: %s", passN, v)
	}
	if strings.Contains(string(d), `"keywordLocation":"/$ref/properties/n"`) {
		t.Errorf("detailed kept passing node: %s", d)
	}
}

// Property-name annotations (core 10.3.2) show the matched names, never the
// values, in both nested shapes.
func TestNestedPropertyNameAnnotations(t *testing.T) {
	const secret = "SECRET-VALUE-99"
	s, err := Compile([]byte(`{"properties":{"name":{}},"patternProperties":{"^p":{}},"additionalProperties":{}}`))
	if err != nil {
		t.Fatal(err)
	}
	res, err := s.Validate([]byte(`{"name":"` + secret + `1","px":"` + secret + `2","other":"` + secret + `3"}`))
	if err != nil || !res.Valid() {
		t.Fatal(res, err)
	}
	for _, shape := range []string{"detailed", "verbose"} {
		out, err := res.Output(shape)
		if err != nil || !json.Valid(out) {
			t.Fatalf("%s: %s %v", shape, out, err)
		}
		for _, want := range []string{
			`"keywordLocation":"/properties","instanceLocation":"","annotation":["name"]`,
			`"keywordLocation":"/patternProperties","instanceLocation":"","annotation":["px"]`,
			`"keywordLocation":"/additionalProperties","instanceLocation":"","annotation":["other"]`,
		} {
			if !strings.Contains(string(out), want) {
				t.Errorf("%s lacks %s: %s", shape, want, out)
			}
		}
		if strings.Contains(string(out), secret) {
			t.Errorf("%s leaked value: %s", shape, out)
		}
	}
}

func TestNestedReportCapIsLimitError(t *testing.T) {
	res := nestedResult(t, Options{MaxReportBytes: 10}, `{"n":"x"}`)
	for _, shape := range []string{"detailed", "verbose"} {
		out, err := res.Output(shape)
		var le *jsonvalue.LimitError
		if !errors.As(err, &le) || le.Limit != "report bytes" || le.Max != 10 || out != nil {
			t.Errorf("%s: out=%q err=%v", shape, out, err)
		}
	}
	// A cap above the default is accepted as a raised cap.
	res = nestedResult(t, Options{MaxReportBytes: DefaultMaxReportBytes + 1}, `{"n":"x"}`)
	if _, err := res.Output("verbose"); err != nil {
		t.Error(err)
	}
}

func TestOutputUnknownShapeStillError(t *testing.T) {
	if _, err := nestedResult(t, Options{}, `{}`).Output("nope"); err == nil {
		t.Error("unknown shape accepted")
	}
}

// Operator decision (story 194, Checkpoint 2): adding detailed and verbose
// must leave flag and basic output byte-identical to base commit c941ecb,
// and passing tree nodes must never count against MaxOutputUnits.

const unchangedSchema = `{"items":{"properties":{"a":{"type":"string"},"b":{"minimum":0}},"required":["c"]}}`

// unchangedDoc is a large failing document: 3000 items, 3 failures each.
func unchangedDoc() []byte {
	return []byte("[" + strings.TrimSuffix(strings.Repeat(`{"a":1,"b":-1},`, 3000), ",") + "]")
}

// Hashes of the flag and basic output of unchangedDoc, taken with the
// flag/basic rendering path of base commit c941ecb (internal/eval,
// internal/keywords, flag.go and basic.go are unchanged since; the schema
// has no key that the fragment encoding would touch).
var unchangedHashes = map[string]string{
	"flag":  "af5fd803088fcdc627e5cd97fb88d7fcbbb02a705f38fd48b1ab5f4ffca50ae8",
	"basic": "9352adc7389fb7162cba707b38374c0eb4bc5fd5059277e697cf65d99be4ebc6",
}

func TestFlagBasicOutputUnchanged(t *testing.T) {
	s, err := Compile([]byte(unchangedSchema))
	if err != nil {
		t.Fatal(err)
	}
	res, err := s.Validate(unchangedDoc())
	if err != nil || res.Valid() {
		t.Fatal(res.Valid(), err)
	}
	for shape, want := range unchangedHashes {
		out, err := res.Output(shape)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(out)
		if got := hex.EncodeToString(sum[:]); got != want {
			t.Errorf("%s output changed: sha256 %s, want %s", shape, got, want)
		}
	}
}

// With MaxOutputUnits set to exactly the failure count, Validate and the
// detailed and verbose renders succeed: passing nodes are not counted.
func TestNestedShapesDoNotCountPassingNodes(t *testing.T) {
	s, err := Compile([]byte(unchangedSchema))
	if err != nil {
		t.Fatal(err)
	}
	res, err := s.Validate(unchangedDoc())
	if err != nil {
		t.Fatal(err)
	}
	b, err := res.Output("basic")
	if err != nil {
		t.Fatal(err)
	}
	var rep struct{ Errors []json.RawMessage }
	if err := json.Unmarshal(b, &rep); err != nil {
		t.Fatal(err)
	}
	n := len(rep.Errors)
	if n == 0 {
		t.Fatal("no failures")
	}
	// One unit fewer is over the cap, so n is the exact failure count.
	tight, _ := CompileWithOptions([]byte(unchangedSchema), Options{MaxOutputUnits: n - 1})
	if _, err := tight.Validate(unchangedDoc()); err == nil {
		t.Fatalf("cap %d not exact", n)
	}
	exact, err := CompileWithOptions([]byte(unchangedSchema), Options{MaxOutputUnits: n})
	if err != nil {
		t.Fatal(err)
	}
	res, err = exact.Validate(unchangedDoc())
	if err != nil {
		t.Fatalf("Validate at exact cap: %v", err)
	}
	for _, shape := range []string{"detailed", "verbose"} {
		out, err := res.Output(shape)
		if err != nil || !json.Valid(out) {
			t.Errorf("%s at exact cap: %v", shape, err)
		}
	}
}
