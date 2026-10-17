package jsonschema

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/compile"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/ecmaregex"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/eval"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/output"
)

func TestDefaultLimitValues(t *testing.T) {
	if DefaultMaxRefChain != 2*DefaultMaxDepth {
		t.Errorf("ref chain default %d is not twice depth %d", DefaultMaxRefChain, DefaultMaxDepth)
	}
	if DefaultMaxInputBytes != 64<<20 || DefaultMaxSteps != 100_000_000 {
		t.Errorf("unexpected defaults %d %d", DefaultMaxInputBytes, DefaultMaxSteps)
	}
}

func TestNegativeLimitOptionsRejected(t *testing.T) {
	for name, o := range map[string]Options{
		"MaxDepth":       {MaxDepth: -1},
		"MaxInputBytes":  {MaxInputBytes: -1},
		"MaxRefChain":    {MaxRefChain: -1},
		"MaxSteps":       {MaxSteps: -1},
		"MaxPatternSize": {MaxPatternSize: -1},
	} {
		_, err := CompileWithOptions([]byte(`true`), o)
		if err == nil || !strings.Contains(err.Error(), name) {
			t.Errorf("%s: want error naming it, got %v", name, err)
		}
	}
}

func TestInputSizeLimit(t *testing.T) {
	const max = 8
	at := strings.Repeat(" ", max-4) + `true`
	s, err := CompileWithOptions([]byte(at), Options{MaxInputBytes: max})
	if err != nil {
		t.Fatalf("at cap: %v", err)
	}
	var le *LimitError
	if _, err := CompileWithOptions([]byte(at+" "), Options{MaxInputBytes: max}); !errors.As(err, &le) || le.Limit != LimitInputSize || le.Value != max {
		t.Fatalf("schema over cap: %v", err)
	}
	if _, err := s.Validate([]byte(at)); err != nil {
		t.Fatalf("document at cap: %v", err)
	}
	le = nil
	if _, err := s.Validate([]byte(at + " ")); !errors.As(err, &le) || le.Limit != LimitInputSize || le.Value != max {
		t.Fatalf("document over cap: %v", err)
	}
	if want := "limit exceeded: input size 8"; le.Error() != want {
		t.Errorf("message %q, want %q", le.Error(), want)
	}
}

func TestExistingLimitErrorsAreTopLevel(t *testing.T) {
	s, err := CompileWithOptions([]byte(`true`), Options{MaxNumberDigits: 2})
	if err != nil {
		t.Fatal(err)
	}
	var le *LimitError
	if _, err := s.Validate([]byte(`123`)); !errors.As(err, &le) || le.Limit != LimitNumberDigits || le.Value != 2 {
		t.Fatalf("got %v", err)
	}
}

func TestPatternSizeOption(t *testing.T) {
	const max = 8
	at := `{"pattern":"` + strings.Repeat("a", max) + `"}`
	if _, err := CompileWithOptions([]byte(at), Options{MaxPatternSize: max}); err != nil {
		t.Fatalf("at cap: %v", err)
	}
	for _, s := range []string{
		`{"pattern":"` + strings.Repeat("a", max+1) + `"}`,
		`{"patternProperties":{"` + strings.Repeat("a", max+1) + `":true}}`,
		`{"additionalProperties":false,"patternProperties":{"` + strings.Repeat("a", max+1) + `":true}}`,
	} {
		var le *LimitError
		_, err := CompileWithOptions([]byte(s), Options{MaxPatternSize: max})
		if !errors.As(err, &le) || le.Limit != LimitPatternSize || le.Value != max {
			t.Errorf("%.40s: got %v", s, err)
		}
	}
	big := `{"pattern":"` + strings.Repeat("a", DefaultMaxPatternSize+1) + `"}`
	if _, err := CompileWithOptions([]byte(big), Options{MaxPatternSize: DefaultMaxPatternSize + 1}); err != nil {
		t.Errorf("raised cap: %v", err)
	}
	var le *LimitError
	if _, err := Compile([]byte(big)); !errors.As(err, &le) || le.Value != DefaultMaxPatternSize {
		t.Errorf("default cap: %v", err)
	}
}

func TestOutputUnitsAndAnnotationsLimitErrors(t *testing.T) {
	s, err := CompileWithOptions([]byte(`{"items":{"minimum":5}}`), Options{MaxOutputUnits: 1})
	if err != nil {
		t.Fatal(err)
	}
	var le *LimitError
	if _, err := s.Validate([]byte(`[1,2,3]`)); !errors.As(err, &le) || le.Limit != LimitOutputUnits || le.Value != 1 {
		t.Errorf("units: %v", err)
	}
	// The meta-check of the schema records annotations under the same cap.
	le = nil
	_, err = CompileWithOptions([]byte(`{"title":"t","description":"d"}`), Options{MaxAnnotations: 1})
	if !errors.As(err, &le) || le.Limit != LimitAnnotations || le.Value != 1 {
		t.Errorf("annotations: %v", err)
	}
}

func TestNumberExponentLimitError(t *testing.T) {
	s, err := CompileWithOptions([]byte(`true`), Options{MaxNumberExponent: 2})
	if err != nil {
		t.Fatal(err)
	}
	var le *LimitError
	if _, err := s.Validate([]byte(`1e999`)); !errors.As(err, &le) || le.Limit != LimitNumberExponent || le.Value != 2 {
		t.Errorf("got %v", err)
	}
}

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

func nestArr(n int) string { return strings.Repeat("[", n) + strings.Repeat("]", n) }

func TestDepthLimitSchemaAndDocument(t *testing.T) {
	const max = 5
	o := Options{MaxDepth: max}
	// schema: depth max is fine, max+1 is a LimitError
	if _, err := CompileWithOptions([]byte(`{"enum":[`+nestArr(max-2)+`]}`), o); err != nil {
		t.Fatalf("schema at cap: %v", err)
	}
	var le *LimitError
	_, err := CompileWithOptions([]byte(`{"enum":[`+nestArr(max-1)+`]}`), o)
	if !errors.As(err, &le) || le.Limit != LimitDepth || le.Value != max {
		t.Fatalf("schema over cap: %v", err)
	}
	s, err := CompileWithOptions([]byte(`true`), o)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Validate([]byte(nestArr(max))); err != nil {
		t.Fatalf("doc at cap: %v", err)
	}
	if _, err := s.Validate([]byte(nestArr(max - 1))); err != nil {
		t.Fatalf("doc below cap: %v", err)
	}
	le = nil
	if _, err := s.Validate([]byte(nestArr(max + 1))); !errors.As(err, &le) || le.Limit != LimitDepth || le.Value != max {
		t.Fatalf("doc over cap: %v", err)
	}
}

// encoding/json stops at DefaultMaxDepth, so a higher setting never turns
// into a "syntax error" or a valid result: it is still a LimitError.
func TestDepthAboveStdlibCeilingIsLimitError(t *testing.T) {
	s, err := CompileWithOptions([]byte(`true`), Options{MaxDepth: DefaultMaxDepth * 2})
	if err != nil {
		t.Fatal(err)
	}
	var le *LimitError
	if _, err := s.Validate([]byte(nestArr(DefaultMaxDepth + 1))); !errors.As(err, &le) || le.Limit != LimitDepth {
		t.Fatalf("got %v", err)
	}
}

func TestRegistryTotalAndFileLimits(t *testing.T) {
	dir := t.TempDir()
	body := `{"$id":"https://x.test/%s","type":"string"}`
	var files []RegistryFile
	size := 0
	for _, n := range []string{"a", "b"} {
		b := strings.Replace(body, "%s", n, 1)
		size = len(b)
		p := filepath.Join(dir, n+".json")
		if err := os.WriteFile(p, []byte(b), 0o600); err != nil {
			t.Fatal(err)
		}
		files = append(files, RegistryFile{Path: p})
	}
	compile := func(r Registry) error {
		_, err := CompileWithRegistry([]byte(`true`), Options{}, r)
		return err
	}
	if err := compile(Registry{Files: files, MaxTotalBytes: 2 * size}); err != nil {
		t.Fatalf("total at cap: %v", err)
	}
	var le *LimitError
	err := compile(Registry{Files: files, MaxTotalBytes: 2*size - 1})
	if !errors.As(err, &le) || le.Limit != LimitRegistrySize || le.Value != 2*size-1 {
		t.Fatalf("total over cap: %v", err)
	}
	// per file: explicit cap, and the default follows MaxInputBytes
	if err := compile(Registry{Files: files, MaxFileBytes: size}); err != nil {
		t.Fatalf("file at cap: %v", err)
	}
	le = nil
	err = compile(Registry{Files: files, MaxFileBytes: size - 1})
	if !errors.As(err, &le) || le.Limit != LimitInputSize {
		t.Fatalf("file over cap: %v", err)
	}
	le = nil
	_, err = CompileWithRegistry([]byte(`true`), Options{MaxInputBytes: size - 1 + 8}, Registry{Files: files})
	if err != nil {
		t.Fatalf("schema smaller than cap, files within: %v", err)
	}
	_, err = CompileWithRegistry([]byte(`true`), Options{MaxInputBytes: size - 1}, Registry{Files: files})
	if !errors.As(err, &le) || le.Limit != LimitInputSize {
		t.Fatalf("file over MaxInputBytes: %v", err)
	}
	if err := compile(Registry{Files: files, MaxTotalBytes: -1}); err == nil {
		t.Fatal("negative total accepted")
	}
}

func TestNumberCapOption(t *testing.T) {
	over := fmt.Sprintf("1e%d", DefaultMaxNumberExponent+1)
	s, _ := Compile([]byte(`true`))
	_, err := s.Validate([]byte(over))
	var le *jsonvalue.LimitError
	if !errors.As(err, &le) || strings.Contains(err.Error(), over) {
		t.Fatalf("want content-free LimitError, got %v", err)
	}
	overSchema := []byte(`{"const":` + over + `}`)
	if _, err := Compile(overSchema); !errors.As(err, &le) || le.Limit != "number exponent" {
		t.Errorf("schema over cap: want number exponent LimitError, got %v", err)
	}
	if _, err := CompileWithOptions(overSchema, Options{MaxNumberExponent: DefaultMaxNumberExponent + 1}); err != nil {
		t.Errorf("raised cap must compile the schema: %v", err)
	}
	s, err = CompileWithOptions([]byte(`true`), Options{MaxNumberExponent: DefaultMaxNumberExponent + 1})
	if err != nil {
		t.Fatal(err)
	}
	if r, err := s.Validate([]byte(over)); err != nil || !r.Valid() {
		t.Fatalf("raised cap: %v %v", r.Valid(), err)
	}
	// F4: a normalised exponent of MinInt64 is a limit error, never a value.
	s, err = CompileWithOptions([]byte(`true`), Options{MaxNumberExponent: math.MaxInt})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Validate([]byte("1.5e-9223372036854775807")); !errors.As(err, &le) || le.Limit != "number exponent" {
		t.Errorf("min exponent: want number exponent LimitError, got %v", err)
	}
	if r, err := s.Validate([]byte("15e-9223372036854775807")); err != nil || !r.Valid() {
		t.Errorf("neighbour that fits: %v %v", r.Valid(), err)
	}
	if _, err := CompileWithOptions([]byte(`true`), Options{MaxNumberDigits: -1}); err == nil {
		t.Error("negative option must error")
	}
}

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

// trackSchemas returns two schemas with the same shape and cost: skip has no
// unevaluated* keyword, so recording is off; keep names one inside an unknown
// keyword, which the cautious load-time scan counts, so recording stays on.
func trackSchemas(body string) (skip, keep string) {
	tpl := `{%s,"x-note":{"%s":true}}`
	return fmt.Sprintf(tpl, body, "unevaluatedNot"), fmt.Sprintf(tpl, body, "unevaluatedItems")
}

// limitProbe compiles with the limit set to n, then validates (and renders).
type limitProbe func(n int) (compileErr, runErr error)

// boundary binary-searches the smallest n in (lo, hi] for which the run
// succeeds, then checks that n-1 compiles and fails with the wanted limit.
// It returns n and the error text at n-1.
func boundary(t *testing.T, want Limit, lo, hi int, probe limitProbe) (int, string) {
	t.Helper()
	ok := func(n int) bool { c, r := probe(n); return c == nil && r == nil }
	if ok(lo) || !ok(hi) {
		t.Fatalf("limit %s: bounds %d..%d do not bracket the boundary", want, lo, hi)
	}
	n := lo + 1 + sort.Search(hi-lo, func(i int) bool { return ok(lo + 1 + i) })
	c, r := probe(n - 1)
	var le *LimitError
	if c != nil || !errors.As(r, &le) || le.Limit != want {
		t.Fatalf("limit %s at %d: compile %v, run %v", want, n-1, c, r)
	}
	return n, r.Error()
}

func TestEvaluatedTrackingLimitsUnchanged(t *testing.T) {
	deep := strings.Repeat("[", 20) + strings.Repeat("]", 20)
	many := func(item string, n int) string {
		return "[" + strings.TrimSuffix(strings.Repeat(item+",", n), ",") + "]"
	}
	var chain strings.Builder
	chain.WriteString(`"$ref":"#/$defs/r0","$defs":{`)
	for i := 0; i < 30; i++ {
		fmt.Fprintf(&chain, `"r%d":{"$ref":"#/$defs/r%d"},`, i, i+1)
	}
	chain.WriteString(`"r30":true}`)

	set := func(f func(*Options, int)) func(string, string) limitProbe {
		return func(schema, doc string) limitProbe {
			return func(n int) (error, error) {
				var o Options
				f(&o, n)
				s, err := CompileWithOptions([]byte(schema), o)
				if err != nil {
					return err, nil
				}
				_, err = s.Validate([]byte(doc))
				return nil, err
			}
		}
	}
	render := func(shape string) func(string, string) limitProbe {
		return func(schema, doc string) limitProbe {
			return func(n int) (error, error) {
				s, err := CompileWithOptions([]byte(schema), Options{MaxReportBytes: n})
				if err != nil {
					return err, nil
				}
				res, err := s.Validate([]byte(doc))
				if err != nil {
					return err, nil
				}
				_, err = res.Output(shape)
				return nil, err
			}
		}
	}
	failing := many(`1`, 20)
	cases := []struct {
		name, body, doc string
		want            Limit
		lo, hi          int
		probe           func(string, string) limitProbe
	}{
		{"depth", `"type":"array"`, deep, LimitDepth, 4, 100, set(func(o *Options, n int) { o.MaxDepth = n })},
		{"ref chain", chain.String(), `1`, LimitRefChain, 20, 80, set(func(o *Options, n int) { o.MaxRefChain = n })},
		{"steps", `"items":{"type":"integer"}`, many(`1`, 500), LimitSteps, 50, 100000, set(func(o *Options, n int) { o.MaxSteps = n })},
		{"output units", `"items":{"type":"string"}`, failing, LimitOutputUnits, 1, 100, set(func(o *Options, n int) { o.MaxOutputUnits = n })},
		{"annotations", `"items":{"title":"t"}`, many(`1`, 50), LimitAnnotations, 1, 1000, set(func(o *Options, n int) { o.MaxAnnotations = n })},
		{"report bytes basic", `"items":{"type":"string"}`, failing, LimitReportBytes, 1, 100000, render("basic")},
		{"report bytes verbose", `"items":{"type":"string"}`, failing, LimitReportBytes, 1, 1000000, render("verbose")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			skip, keep := trackSchemas(c.body)
			for schema, wantSkip := range map[string]bool{skip: true, keep: false} {
				s, err := Compile([]byte(schema))
				if err != nil {
					t.Fatal(err)
				}
				if s.root.NoUnevaluated != wantSkip {
					t.Fatalf("NoUnevaluated = %v, want %v", s.root.NoUnevaluated, wantSkip)
				}
			}
			ns, es := boundary(t, c.want, c.lo, c.hi, c.probe(skip, c.doc))
			nk, ek := boundary(t, c.want, c.lo, c.hi, c.probe(keep, c.doc))
			if ns != nk || es != ek {
				t.Errorf("skipping boundary %d (%q), recording boundary %d (%q)", ns, es, nk, ek)
			}
		})
	}
}

// Story 202: caching the built-in meta-schema must not move any hostile-input
// limit. Each case is the smallest input from cmd/jsonschema's
// TestLimitsAtTheEdge (measured before this story): at the edge the run gives
// its normal answer, one below it fails with the same limit error. The library
// is checked on a repeated Compile (the shared set is built by then) and the
// built CLI binary, a fresh process whose Compile is the first, must exit 2.

type cacheLimitCase struct {
	name, schema, doc, shape string
	limit                    Limit
	edge                     int
	set                      func(*Options, int)
	flag                     string
	wantCode                 int // CLI exit code at the edge
}

func cacheLimitCases() []cacheLimitCase {
	arr := func(n int, v string) string {
		return "[" + strings.TrimSuffix(strings.Repeat(v+",", n), ",") + "]"
	}
	var defs []string
	for i := 0; i < 4; i++ {
		defs = append(defs, fmt.Sprintf(`"r%d":{"$ref":"#/$defs/r%d"}`, i, i+1))
	}
	defs = append(defs, `"r4":true`)
	refSchema := `{"$defs":{` + strings.Join(defs, ",") + `},"$ref":"#/$defs/r0"}`
	return []cacheLimitCase{
		{"depth", `true`, `[[[[1]]]]`, "", LimitDepth, 4, func(o *Options, n int) { o.MaxDepth = n }, "max-depth", 0},
		{"ref chain", refSchema, `1`, "", LimitRefChain, 5, func(o *Options, n int) { o.MaxRefChain = n }, "max-ref-chain", 0},
		{"steps", `{"items":{"type":"integer"}}`, arr(5, "1"), "", LimitSteps, 186, func(o *Options, n int) { o.MaxSteps = n }, "max-steps", 0},
		{"output units", `{"items":{"type":"string"}}`, arr(4, "1"), "flag", LimitOutputUnits, 4, func(o *Options, n int) { o.MaxOutputUnits = n }, "max-output-units", 1},
		{"report bytes", `{"items":{"type":"string"}}`, arr(4, "1"), "basic", LimitReportBytes, 414, func(o *Options, n int) { o.MaxReportBytes = n }, "max-report-bytes", 1},
		{"annotations", `{"allOf":[{"title":"a"},{"title":"b"},{"title":"c"}]}`, `1`, "basic", LimitAnnotations, 37, func(o *Options, n int) { o.MaxAnnotations = n }, "max-annotations", 0},
	}
}

// runLib compiles with the limit set to n, validates and renders the shape.
func (c cacheLimitCase) runLib(n int) error {
	var o Options
	c.set(&o, n)
	s, err := CompileWithOptions([]byte(c.schema), o)
	if err != nil {
		return err
	}
	res, err := s.Validate([]byte(c.doc))
	if err != nil {
		return err
	}
	if c.shape != "" {
		_, err = res.Output(c.shape)
	}
	return err
}

func TestCacheLimitsLibrary(t *testing.T) {
	for _, c := range cacheLimitCases() {
		t.Run(c.name, func(t *testing.T) {
			for round := 0; round < 3; round++ { // later rounds reuse the cached meta-schema
				if err := c.runLib(c.edge); err != nil {
					t.Fatalf("round %d at the edge: %v", round, err)
				}
				err := c.runLib(c.edge - 1)
				var le *LimitError
				if !errors.As(err, &le) || le.Limit != c.limit {
					t.Fatalf("round %d one below: got %v, want limit %s", round, err, c.limit)
				}
			}
		})
	}
}

func TestCacheLimitsCLIFreshProcess(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "jsonschema")
	if out, err := exec.Command("go", "build", "-o", bin, "./cmd/jsonschema").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	for _, c := range cacheLimitCases() {
		t.Run(c.name, func(t *testing.T) {
			sch := filepath.Join(t.TempDir(), "schema.json")
			if err := os.WriteFile(sch, []byte(c.schema), 0o600); err != nil {
				t.Fatal(err)
			}
			try := func(n int) (int, string, string) {
				args := []string{"validate", "--" + c.flag + "=" + strconv.Itoa(n)}
				if c.shape != "" {
					args = append(args, "-o", c.shape)
				}
				cmd := exec.Command(bin, append(args, sch)...)
				cmd.Stdin = strings.NewReader(c.doc)
				var so, se bytes.Buffer
				cmd.Stdout, cmd.Stderr = &so, &se
				code := 0
				if err := cmd.Run(); err != nil {
					var ee *exec.ExitError
					if !errors.As(err, &ee) {
						t.Fatal(err)
					}
					code = ee.ExitCode()
				}
				return code, so.String(), se.String()
			}
			if code, _, e := try(c.edge); code != c.wantCode {
				t.Errorf("at the edge: code %d, want %d (stderr %q)", code, c.wantCode, e)
			}
			code, out, e := try(c.edge - 1)
			if code != 2 || out != "" || !strings.Contains(e, "limit exceeded") || !strings.Contains(e, "--"+c.flag) {
				t.Errorf("one below: code %d out %q stderr %q", code, out, e)
			}
		})
	}
}

// A large default under items, applied to many items, overruns a lowered
// report cap; the same input renders under DefaultMaxReportBytes.
func TestReportBytesCap(t *testing.T) {
	schema := []byte(`{"items":{"default":"` + strings.Repeat("d", 1000) + `"}}`)
	doc := []byte("[" + strings.TrimSuffix(strings.Repeat("0,", 200), ",") + "]")
	s, err := CompileWithOptions(schema, Options{MaxReportBytes: 50_000})
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.Validate(doc)
	if err != nil || !r.Valid() {
		t.Fatal(r, err)
	}
	b, err := r.Output("basic")
	var le *jsonvalue.LimitError
	if !errors.As(err, &le) || le.Limit != "report bytes" || le.Max != 50_000 || b != nil {
		t.Fatalf("want report bytes LimitError and nil bytes, got %d bytes, %v", len(b), err)
	}
	s, _ = Compile(schema)
	r, _ = s.Validate(doc)
	if b, err := r.Output("basic"); err != nil || len(b) <= 50_000 || len(b) > DefaultMaxReportBytes {
		t.Fatalf("default cap: %d bytes, %v", len(b), err)
	}
	if _, err := CompileWithOptions(schema, Options{MaxReportBytes: -1}); err == nil || err.Error() != "options: MaxReportBytes must not be negative" {
		t.Fatalf("negative MaxReportBytes: %v", err)
	}
	// A zero-value Result falls back to the default cap, never unlimited.
	if b, err := (Result{}).Output("flag"); err != nil || string(b) != `{"valid":false}` {
		t.Fatalf("zero Result: %s %v", b, err)
	}
}
