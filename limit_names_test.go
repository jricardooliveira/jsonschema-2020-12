package jsonschema

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/ecmaregex"
	"github.com/neoface/jsonschema/internal/eval"
	"github.com/neoface/jsonschema/internal/jsonvalue"
	"github.com/neoface/jsonschema/internal/output"
)

// refChainSchema is a chain of jumps $ref jumps ending in an integer check.
func refChainSchema(jumps int) string {
	var b strings.Builder
	b.WriteString(`{"$ref":"#/$defs/d0","$defs":{`)
	for i := 0; i < jumps-1; i++ {
		fmt.Fprintf(&b, `"d%d":{"$ref":"#/$defs/d%d"},`, i, i+1)
	}
	fmt.Fprintf(&b, `"d%d":{"type":"integer"}}}`, jumps-1)
	return b.String()
}

func validateWith(t *testing.T, schema, doc string, o Options) error {
	t.Helper()
	s, err := CompileWithOptions([]byte(schema), o)
	if err != nil {
		return err
	}
	res, err := s.Validate([]byte(doc))
	if err != nil {
		return err
	}
	_, err = res.Output("verbose")
	return err
}

// Every internal limit producer (decode, registry, eval, keywords, output,
// ecmaregex) names a limit from the closed set, and every public error for
// it is a top-level *LimitError with that name.
func TestEveryLimitProducerUsesClosedSet(t *testing.T) {
	dir := t.TempDir()
	reg := filepath.Join(dir, "r.json")
	if err := os.WriteFile(reg, []byte(`{"$id":"https://x.test/r","type":"string"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	regCompile := func(r Registry) error {
		_, err := CompileWithRegistry([]byte(`true`), Options{}, r)
		return err
	}
	nestedAllOf := strings.Repeat(`{"allOf":[`, 20) + `true` + strings.Repeat(`]}`, 20)
	evalDepth := func() error { // the eval-path backstop, reached directly
		v, _ := jsonvalue.DecodeLimited([]byte(nestedAllOf), jsonvalue.NumberLimits{})
		n, err := compile.CompileDocument(v, compile.StdinBaseURI)
		if err != nil {
			return err
		}
		lim := eval.DefaultLimits()
		lim.MaxDepth = 2
		_, err = eval.EvaluateWith(n, true, lim, false)
		return err
	}
	cases := map[string]struct {
		want Limit
		err  func() error
	}{
		"decode depth":    {LimitDepth, func() error { return validateWith(t, `true`, `[[[1]]]`, Options{MaxDepth: 2}) }},
		"decode digits":   {LimitNumberDigits, func() error { return validateWith(t, `true`, `123`, Options{MaxNumberDigits: 2}) }},
		"decode exponent": {LimitNumberExponent, func() error { return validateWith(t, `true`, `1e99`, Options{MaxNumberExponent: 2}) }},
		"input size":      {LimitInputSize, func() error { return validateWith(t, `true`, `  1`, Options{MaxInputBytes: 2}) }},
		"registry file":   {LimitInputSize, func() error { return regCompile(Registry{Files: []RegistryFile{{Path: reg}}, MaxFileBytes: 3}) }},
		"registry total":  {LimitRegistrySize, func() error { return regCompile(Registry{Files: []RegistryFile{{Path: reg}}, MaxTotalBytes: 3}) }},
		"eval ref chain":  {LimitRefChain, func() error { return validateWith(t, refChainSchema(50), `1`, Options{MaxRefChain: 49}) }},
		"eval steps": {LimitSteps, func() error {
			return validateWith(t, `{"items":true}`, `[`+strings.Repeat("1,", 9999)+`1]`, Options{MaxSteps: 5000})
		}},
		"eval depth":    {LimitDepth, evalDepth},
		"keyword units": {LimitOutputUnits, func() error { return validateWith(t, `{"items":{"minimum":5}}`, `[1,2]`, Options{MaxOutputUnits: 1}) }},
		"keyword annotation": {LimitAnnotations, func() error {
			return validateWith(t, `{"title":"t","description":"d"}`, `1`, Options{MaxAnnotations: 1})
		}},
		"output report": {LimitReportBytes, func() error { return validateWith(t, `{"minimum":5}`, `1`, Options{MaxReportBytes: 5}) }},
		"pattern size":  {LimitPatternSize, func() error { return validateWith(t, `{"pattern":"aaaa"}`, `"a"`, Options{MaxPatternSize: 2}) }},
		"ecmaregex direct": {LimitPatternSize, func() error {
			_, err := ecmaregex.CompileBounded("aaaa", 2)
			return err
		}},
		"output direct": {LimitReportBytes, func() error {
			_, err := output.RenderReport("basic", output.Report{Valid: true}, 1)
			return err
		}},
	}
	for name, c := range cases {
		err := c.err()
		var inner *jsonvalue.LimitError
		if !errors.As(err, &inner) || !closedSet[inner.Limit] || Limit(inner.Limit) != c.want {
			t.Errorf("%s: internal limit %v not the closed-set name %q", name, err, c.want)
			continue
		}
		if pub := asLimit(err); !isTopLimit(pub, c.want) {
			t.Errorf("%s: public error %T %v", name, pub, pub)
		}
	}
}

func isTopLimit(err error, want Limit) bool {
	le, ok := err.(*LimitError)
	return ok && le.Limit == want
}

// Options.MaxRefChain at, below and above the jumps a chain makes. The meta
// check runs under the same cap and needs far fewer jumps.
func TestMaxRefChainOption(t *testing.T) {
	const jumps = 50
	schema := refChainSchema(jumps)
	for _, c := range []struct {
		max int
		ok  bool
	}{{jumps, true}, {jumps + 1, true}, {jumps - 1, false}} {
		err := validateWith(t, schema, `1`, Options{MaxRefChain: c.max})
		var le *LimitError
		switch {
		case c.ok && err != nil:
			t.Errorf("cap %d: %v", c.max, err)
		case !c.ok && (!errors.As(err, &le) || le.Limit != LimitRefChain || le.Value != c.max):
			t.Errorf("cap %d: want reference chain limit, got %v", c.max, err)
		}
	}
	if err := validateWith(t, schema, `1`, Options{}); err != nil {
		t.Errorf("default: %v", err)
	}
}

// The schema's meta-check runs under the configured caps, not the defaults.
func TestMetaCheckUsesConfiguredLimits(t *testing.T) {
	const schema = `{"properties":{"a":{"items":{"type":"string"}}}}`
	for o, want := range map[*Options]Limit{
		{MaxSteps: 5}:    LimitSteps,
		{MaxRefChain: 1}: LimitRefChain,
	} {
		var le *LimitError
		if _, err := CompileWithOptions([]byte(schema), *o); !errors.As(err, &le) || le.Limit != want {
			t.Errorf("%+v: want %s, got %v", *o, want, err)
		}
	}
}

// Options.MaxSteps at, below and above the smallest cap a document fits in.
func TestMaxStepsOption(t *testing.T) {
	const schema = `{"items":{"type":"integer","minimum":0}}`
	doc := []byte(`[` + strings.Repeat("1,", 999) + `1]`)
	run := func(max int) error {
		s, err := CompileWithOptions([]byte(schema), Options{MaxSteps: max})
		if err != nil {
			t.Fatalf("compile at %d: %v", max, err) // the meta check fits well below
		}
		_, err = s.Validate(doc)
		return err
	}
	lo, hi := 1000, DefaultMaxSteps // run(lo) fails, run(hi) passes
	if run(lo) == nil || run(hi) != nil {
		t.Fatal("bad bounds")
	}
	for hi-lo > 1 {
		if mid := (lo + hi) / 2; run(mid) == nil {
			hi = mid
		} else {
			lo = mid
		}
	}
	if err := run(hi + 1); err != nil {
		t.Errorf("above cap: %v", err)
	}
	var le *LimitError
	if err := run(hi - 1); !errors.As(err, &le) || le.Limit != LimitSteps || le.Value != hi-1 {
		t.Errorf("below cap: %v", err)
	}
}
