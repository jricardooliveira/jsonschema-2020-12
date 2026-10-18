package output

import (
	"errors"
	"strings"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
)

func num(t *testing.T, s string) any {
	t.Helper()
	v, err := jsonvalue.DecodeLimited([]byte(s), jsonvalue.NumberLimits{MaxDigits: 1000, MaxExponent: 1000})
	if err != nil {
		t.Fatal(err)
	}
	return v
}

// Schema numbers are *jsonvalue.Number; they must render as written, nested too.
func TestBasicRendersNumericAnnotations(t *testing.T) {
	r := Report{Valid: true, Annotations: []keywords.Annotation{
		{KeywordLocation: "/default", Value: num(t, `5`)},
		{KeywordLocation: "/examples", Value: num(t, `[1, 2.50, {"x": -1e3}]`)},
	}}
	b, err := RenderReport("basic", r, 0)
	want := `{"valid":true,"annotations":[` +
		`{"valid":true,"instanceLocation":"","keywordLocation":"/default","annotation":5},` +
		`{"valid":true,"instanceLocation":"","keywordLocation":"/examples","annotation":[1,2.50,{"x":-1e3}]}]}`
	if err != nil || string(b) != want {
		t.Fatalf("got %s %v", b, err)
	}
}

func TestBudgetRefusesPastCap(t *testing.T) {
	w := &Budget{max: 4}
	if _, err := w.Write([]byte("abcd")); err != nil {
		t.Fatal(err)
	}
	_, err := w.Write([]byte("e"))
	var le *jsonvalue.LimitError
	if !errors.As(err, &le) || le.Limit != "report bytes" || le.Max != 4 || w.buf.String() != "abcd" {
		t.Fatalf("got %v, buf %q", err, w.buf.String())
	}
}

func TestRenderReportCap(t *testing.T) {
	ann := []keywords.Annotation{{KeywordLocation: "/title", Value: strings.Repeat("t", 100)}}
	r := Report{Valid: true, Annotations: ann}
	full, err := RenderReport("basic", r, 0) // 0 means the default, never unlimited
	if err != nil {
		t.Fatal(err)
	}
	if b, err := RenderReport("basic", r, len(full)); err != nil || string(b) != string(full) {
		t.Fatalf("at cap: %s %v", b, err)
	}
	for _, shape := range []string{"basic", "flag"} { // flag goes through the wrapped renderer
		b, err := RenderReport(shape, Report{Valid: false, Units: []keywords.Unit{{Error: "e"}}}, 5)
		var le *jsonvalue.LimitError
		if !errors.As(err, &le) || le.Limit != "report bytes" || b != nil {
			t.Fatalf("%s: got %q %v", shape, b, err)
		}
	}
	if b, err := RenderReport("basic", r, len(full)-1); err == nil || b != nil {
		t.Fatalf("over cap: %q %v", b, err)
	}
	if b, err := RenderReport("basic", r, -1); err != nil || string(b) != string(full) || DefaultMaxReportBytes <= len(full) {
		t.Fatalf("negative cap must mean default: %q %v", b, err)
	}
}
