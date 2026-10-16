package validation_test

// Derived from JSON Schema validation spec sections 6.2.1 to 6.2.5 and core
// spec section 12 (output): a failure carries its locations, never the value.

import (
	"strings"
	"sync"
	"testing"

	"github.com/neoface/jsonschema"
)

// No applicator exists yet, so the failure is at the root: instanceLocation "".
func TestNumericFailureBasicOutput(t *testing.T) {
	const doc = `987654321.5`
	for _, kw := range []struct{ name, value string }{
		{"multipleOf", "2"}, {"maximum", "1"}, {"exclusiveMaximum", "1"},
		{"minimum", "1e10"}, {"exclusiveMinimum", "1e10"},
	} {
		s, err := jsonschema.Compile([]byte(`{"` + kw.name + `":` + kw.value + `}`))
		if err != nil {
			t.Fatal(err)
		}
		r, err := s.Validate([]byte(doc))
		if err != nil || r.Valid() {
			t.Fatalf("%s: want invalid, err %v", kw.name, err)
		}
		b, err := r.Output("basic")
		if err != nil {
			t.Fatal(err)
		}
		out := string(b)
		want := `"instanceLocation":"","keywordLocation":"/` + kw.name + `"`
		if !strings.Contains(out, want) || strings.Contains(out, "987654321") {
			t.Errorf("%s: %s", kw.name, out)
		}
	}
}

func TestBoundsConcurrent(t *testing.T) {
	s, err := jsonschema.Compile([]byte(`{"maximum":10,"minimum":0,"exclusiveMaximum":10,"exclusiveMinimum":0}`))
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, c := range []struct {
				doc  string
				want bool
			}{{`5`, true}, {`10`, false}, {`0`, false}} {
				r, err := s.Validate([]byte(c.doc))
				if err != nil || r.Valid() != c.want {
					t.Errorf("%s: want %v", c.doc, c.want)
				}
			}
		}()
	}
	wg.Wait()
}
