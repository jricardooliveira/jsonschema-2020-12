package jsonschema

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
)

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
