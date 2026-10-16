package applicator_test

import (
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/neoface/jsonschema/internal/eval"
	"github.com/neoface/jsonschema/internal/keywords"
)

type evalResult struct {
	ok    bool
	units []keywords.Unit
	err   error
}

// TestCombinatorsConcurrentMatchSequential: one compiled schema using allOf,
// anyOf, oneOf, not and if/then/else is shared by many goroutines; every
// result and keywordLocation must equal a sequential run. Run with -race.
func TestCombinatorsConcurrentMatchSequential(t *testing.T) {
	n := mustCompile(t, `{"allOf":[
		{"anyOf":[{"type":"integer"},{"type":"string"}]},
		{"oneOf":[{"type":"integer","minimum":0},{"type":"string"}]},
		{"not":{"const":13}},
		{"if":{"type":"integer"},"then":{"maximum":100},"else":{"minLength":2}}
	]}`)
	docs := []string{`5`, `"ab"`, `-1`, `13`, `500`, `"a"`, `1.5`, `true`}
	inst := make([]any, len(docs))
	want := make([]evalResult, len(docs))
	for i, d := range docs {
		inst[i] = mustDecode(t, d)
		ok, units, err := eval.EvaluateUnits(n, inst[i], eval.DefaultMaxOutputUnits)
		if err != nil {
			t.Fatalf("%s: %v", d, err)
		}
		want[i] = evalResult{ok, units, err}
	}
	// sanity: valid and invalid both present, and nested locations are rooted
	if !want[0].ok || !want[1].ok || want[4].ok {
		t.Fatalf("sequential results: %+v", want)
	}
	var locs []string
	for _, u := range want[4].units {
		locs = append(locs, u.KeywordLocation)
	}
	if got := strings.Join(locs, ","); got != "/allOf/3/then/maximum,/allOf/3/then,/allOf" {
		t.Fatalf("rooted locations: %s", got)
	}

	const workers = 16
	// mismatch[w] holds goroutine w's first divergent result (doc index -1 if none).
	type mismatch struct {
		doc int
		got evalResult
	}
	bad := make([]mismatch, workers)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			bad[w].doc = -1
			for round := 0; round < 20; round++ {
				for i := range inst {
					j := (i + w) % len(inst) // vary order between goroutines
					ok, units, err := eval.EvaluateUnits(n, inst[j], eval.DefaultMaxOutputUnits)
					r := evalResult{ok, units, err}
					if bad[w].doc < 0 && !reflect.DeepEqual(r, want[j]) {
						bad[w] = mismatch{j, r}
					}
				}
			}
		}(w)
	}
	wg.Wait()
	for w, m := range bad {
		if m.doc >= 0 {
			t.Errorf("goroutine %d, doc %s: got %+v want %+v", w, docs[m.doc], m.got, want[m.doc])
		}
	}
}
