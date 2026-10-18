package eval_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/compile"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/eval"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
	_ "github.com/jricardooliveira/jsonschema-2020-12/internal/keywords/all"
)

func budgetNode(t *testing.T, schema string) *compile.Node {
	t.Helper()
	s, err := jsonvalue.DecodeLimited([]byte(schema), jsonvalue.NumberLimits{})
	if err != nil {
		t.Fatal(err)
	}
	n, err := compile.CompileDocument(s, "file:///s.json")
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func budgetDoc(t *testing.T, doc string) any {
	t.Helper()
	d, err := jsonvalue.DecodeLimited([]byte(doc), jsonvalue.NumberLimits{})
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func limitName(err error) (string, int) {
	var le *jsonvalue.LimitError
	if errors.As(err, &le) {
		return le.Limit, le.Max
	}
	return "", 0
}

// A chain of $ref jumps longer than MaxRefChain stops with a reference-chain
// limit; the same chain within the cap passes.
func TestRefChainLimit(t *testing.T) {
	const jumps = 50
	var b strings.Builder
	b.WriteString(`{"$ref":"#/$defs/d0","$defs":{`)
	for i := 0; i < jumps; i++ {
		fmt.Fprintf(&b, `"d%d":{"$ref":"#/$defs/d%d"},`, i, i+1)
	}
	fmt.Fprintf(&b, `"d%d":{"type":"integer"}}}`, jumps)
	n, d := budgetNode(t, b.String()), budgetDoc(t, `1`)
	lim := eval.DefaultLimits()
	lim.MaxRefChain = jumps + 1 // the root $ref plus the chain
	if ev, err := eval.EvaluateWith(n, d, lim, false); err != nil || !ev.Valid {
		t.Fatalf("at cap: %v %v", ev.Valid, err)
	}
	lim.MaxRefChain = jumps
	_, err := eval.EvaluateWith(n, d, lim, false)
	if name, max := limitName(err); name != "reference chain" || max != jumps {
		t.Fatalf("over cap: %v", err)
	}
}

// A cycle that consumes no instance is a loop error that also matches a
// reference-chain limit error, carrying the configured cap.
func TestCycleMatchesRefChainLimit(t *testing.T) {
	n := budgetNode(t, `{"$ref":"#/$defs/a","$defs":{"a":{"$ref":"#/$defs/b","type":"object"},"b":{"$ref":"#/$defs/a","type":"object"}}}`)
	lim := eval.DefaultLimits()
	lim.MaxRefChain = 7
	_, err := eval.EvaluateWith(n, budgetDoc(t, `{}`), lim, false)
	var loop *eval.LoopError
	if !errors.As(err, &loop) {
		t.Fatalf("want loop error, got %v", err)
	}
	if name, max := limitName(err); name != "reference chain" || max != lim.MaxRefChain {
		t.Fatalf("got %q %d", name, max)
	}
}

// Deep legitimate recursion within the default chain cap passes, and each
// call starts with a fresh budget.
func TestDeepRecursionWithinDefaultsPasses(t *testing.T) {
	n := budgetNode(t, `{"items":{"$ref":"#"}}`)
	d := budgetDoc(t, strings.Repeat("[", 2000)+strings.Repeat("]", 2000))
	for i := 0; i < 2; i++ {
		if ev, err := eval.EvaluateWith(n, d, eval.DefaultLimits(), false); err != nil || !ev.Valid {
			t.Fatalf("run %d: %v %v", i, ev.Valid, err)
		}
	}
}

// Nested anyOf with $ref doubling ends in an evaluation-steps limit.
func TestExponentialBlowUpHitsSteps(t *testing.T) {
	const levels = 40
	var b strings.Builder
	b.WriteString(`{"$ref":"#/$defs/d0","$defs":{`)
	for i := 0; i < levels; i++ {
		fmt.Fprintf(&b, `"d%d":{"anyOf":[{"$ref":"#/$defs/d%d"},{"$ref":"#/$defs/d%d"}]},`, i, i+1, i+1)
	}
	fmt.Fprintf(&b, `"d%d":{"type":"string"}}}`, levels)
	lim := eval.DefaultLimits()
	lim.MaxSteps = 10_000
	_, err := eval.EvaluateWith(budgetNode(t, b.String()), budgetDoc(t, `1`), lim, false)
	if name, max := limitName(err); name != "evaluation steps" || max != lim.MaxSteps {
		t.Fatalf("got %v", err)
	}
}

// runHostile evaluates schema against doc with maxSteps and requires an
// evaluation-steps limit within a generous time bound; the same pair passes
// at the default.
func runHostile(t *testing.T, schema, doc string, maxSteps int) {
	t.Helper()
	n, d := budgetNode(t, schema), budgetDoc(t, doc)
	lim := eval.DefaultLimits()
	lim.MaxSteps = maxSteps
	start := time.Now()
	_, err := eval.EvaluateWith(n, d, lim, false)
	if name, max := limitName(err); name != "evaluation steps" || max != maxSteps {
		t.Fatalf("want evaluation steps %d, got %v", maxSteps, err)
	}
	if el := time.Since(start); el > 10*time.Second {
		t.Fatalf("took %v", el)
	}
	if ev, err := eval.EvaluateWith(n, d, eval.DefaultLimits(), false); err != nil || !ev.Valid {
		t.Fatalf("default: %v %v", ev.Valid, err)
	}
}

// uniqueItems is charged by the bytes it fingerprints, not once per element:
// the cap sits above a per-element count and below the byte-scaled one.
func TestHostileUniqueItemsHitsSteps(t *testing.T) {
	const n, size = 200, 100 * keywords.BytesPerStep
	items := make([]string, n)
	for i := range items {
		items[i] = fmt.Sprintf(`"%0*d"`, size, i) // distinct, size bytes each
	}
	maxSteps := 10*n + 50 // > n plus base steps, < n*size/BytesPerStep
	if maxSteps >= n*size/keywords.BytesPerStep {
		t.Fatal("bad test sizes")
	}
	runHostile(t, `{"uniqueItems":true}`, "["+strings.Join(items, ",")+"]", maxSteps)
}

// patternProperties x properties: each (key, pattern) match is charged by
// the key length.
func TestHostilePatternPropertiesHitsSteps(t *testing.T) {
	const k, p, l = 50, 10, 100 * keywords.BytesPerStep
	pats := make([]string, p)
	for i := range pats {
		pats[i] = fmt.Sprintf(`"^x%d":{"type":"string"}`, i) // never matches the keys
	}
	keys := make([]string, k)
	for i := range keys {
		keys[i] = fmt.Sprintf(`"%0*d":1`, l, i)
	}
	schema := `{"properties":{"a":true,"b":true},"patternProperties":{` + strings.Join(pats, ",") + `}}`
	pairs := k * p
	maxSteps := 2*(k+pairs) + 100 // > the plain count, < pairs*l/BytesPerStep
	if maxSteps >= pairs*l/keywords.BytesPerStep {
		t.Fatal("bad test sizes")
	}
	runHostile(t, schema, "{"+strings.Join(keys, ",")+"}", maxSteps)
}

// A large ordinary document checked by several keywords per element is not
// stopped at the default limits.
func TestLargeOrdinaryDocumentWithinDefaults(t *testing.T) {
	const n = 100_000
	var b strings.Builder
	b.WriteString("[")
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteString(",")
		}
		fmt.Fprintf(&b, `{"id":%d,"name":"item","tags":["a","b"]}`, i)
	}
	b.WriteString("]")
	schema := `{"type":"array","uniqueItems":true,"items":{"type":"object","required":["id","name"],` +
		`"properties":{"id":{"type":"integer","minimum":0},"name":{"type":"string","minLength":1,"maxLength":50,"pattern":"^[a-z]+$"},` +
		`"tags":{"items":{"enum":["a","b","c"]},"contains":{"const":"a"}}}}}`
	ev, err := eval.EvaluateWith(budgetNode(t, schema), budgetDoc(t, b.String()), eval.DefaultLimits(), false)
	if err != nil || !ev.Valid {
		t.Fatalf("%v %v", ev.Valid, err)
	}
}

// Charge cannot overflow: a huge charge is a limit error, not a wrap-around.
func TestChargeDoesNotOverflow(t *testing.T) {
	r := eval.NewEvaluatedItems()
	if err := r.Charge(1); err != nil {
		t.Fatal(err)
	}
	if err := r.Charge(int(^uint(0) >> 1)); err == nil {
		t.Fatal("want limit error")
	}
}
