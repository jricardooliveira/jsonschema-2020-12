package jsonschema

import (
	"errors"
	"strings"
	"testing"
)

// Hostile-input limits must trigger at exactly the inputs they did before the
// compact number representation. Each case scans the cap upward and pins the
// smallest cap that lets the input through; below it the error must be the
// named *LimitError carrying the cap that was set. Documents mix fast-path
// and big-path numbers, so number storage is on every path.

const bigNums = `1,1.5,9223372036854775807,9223372036854775808,12345678901234567890.5,1e2`

// threshold returns the smallest cap in [1, hi] for which run returns nil.
func threshold(t *testing.T, want Limit, hi int, run func(cap int) error) int {
	t.Helper()
	for c := 1; c <= hi; c++ {
		err := run(c)
		if err == nil {
			return c
		}
		var le *LimitError
		if !errors.As(err, &le) || le.Limit != want || le.Value != c {
			t.Fatalf("cap %d: want %s limit error with value %d, got %v", c, want, c, err)
		}
	}
	t.Fatalf("%s: no cap up to %d lets the input through", want, hi)
	return 0
}

func numValidateErr(schema, doc string, o Options) error {
	s, err := CompileWithOptions([]byte(schema), o)
	if err != nil {
		return err
	}
	_, err = s.Validate([]byte(doc))
	return err
}

func TestNumberDocsHitSameLimits(t *testing.T) {
	const arr = `{"items":{"type":"number","multipleOf":0.5}}`
	nest := strings.Repeat("[", 12) + bigNums + strings.Repeat("]", 12)
	cases := []struct {
		name   string
		limit  Limit
		schema string
		doc    string
		hi     int
		opts   func(c int) Options
		want   int
	}{
		{"depth", LimitDepth, `true`, nest, 40, func(c int) Options { return Options{MaxDepth: c} }, 12},
		{"input size", LimitInputSize, `true`, `[` + bigNums + `]`, 200, func(c int) Options { return Options{MaxInputBytes: c} }, len(`[` + bigNums + `]`)},
		{"steps", LimitSteps, arr, `[` + bigNums + `]`, 400, func(c int) Options { return Options{MaxSteps: c} }, 199},
		{"ref chain", LimitRefChain, `{"$defs":{"a":{"$ref":"#/$defs/b"},"b":{"$ref":"#/$defs/c"},"c":{"type":"number"}},"$ref":"#/$defs/a"}`, `9223372036854775808`, 20, func(c int) Options { return Options{MaxRefChain: c} }, 5},
		{"output units", LimitOutputUnits, `{"items":{"multipleOf":0.3}}`, `[` + bigNums + `]`, 40, func(c int) Options { return Options{MaxOutputUnits: c} }, 5},
		{"annotations", LimitAnnotations, `{"items":{"title":"t"}}`, `[` + bigNums + `]`, 40, func(c int) Options { return Options{MaxAnnotations: c} }, 18},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := threshold(t, c.limit, c.hi, func(n int) error { return numValidateErr(c.schema, c.doc, c.opts(n)) })
			if got != c.want {
				t.Errorf("threshold %d, want %d", got, c.want)
			}
		})
	}
}

func TestNumberReportBytesLimit(t *testing.T) {
	got := threshold(t, LimitReportBytes, 4000, func(n int) error {
		s, err := CompileWithOptions([]byte(`{"items":{"multipleOf":0.3}}`), Options{MaxReportBytes: n})
		if err != nil {
			return err
		}
		r, err := s.Validate([]byte(`[` + bigNums + `]`))
		if err != nil {
			return err
		}
		_, err = r.Output("basic")
		return err
	})
	if got != 636 {
		t.Errorf("report bytes threshold %d, want 636", got)
	}
}
