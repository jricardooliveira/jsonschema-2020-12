package jsonschema

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/neoface/jsonschema/internal/jsonvalue"
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
	v, _ := res.Output("verbose")
	d, _ := res.Output("detailed")
	if !strings.Contains(string(v), `"/n"`) && !strings.Contains(string(v), "/properties/n") {
		t.Errorf("verbose lacks passing node: %s", v)
	}
	if strings.Contains(string(d), "/properties/n") {
		t.Errorf("detailed kept passing node: %s", d)
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
