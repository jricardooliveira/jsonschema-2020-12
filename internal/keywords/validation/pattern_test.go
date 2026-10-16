package validation_test

import (
	"strings"
	"sync"
	"testing"

	"github.com/neoface/jsonschema"
)

func TestPatternSchemaErrors(t *testing.T) {
	for _, p := range []string{`(?=a)`, `(?<=a)`, `(a)\\1`, `\\k<n>`, `\\a`, `\\z`, `(?i)a`, `\\ud800`, `a{2000}`, `\\p{scx=Greek}`} {
		if _, err := jsonschema.Compile([]byte(`{"pattern":"` + p + `"}`)); err == nil {
			t.Errorf("%s: want schema error", p)
		}
	}
	if _, err := jsonschema.Compile([]byte(`{"pattern":5}`)); err == nil {
		t.Error("non-string pattern accepted")
	}
}

func TestPatternMessageShowsPatternNotInstance(t *testing.T) {
	const secret = "S3CRET"
	s, err := jsonschema.Compile([]byte(`{"pattern":"^[a-z]+$"}`))
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.Validate([]byte(`"` + secret + `"`))
	if err != nil || r.Valid() {
		t.Fatalf("want invalid: %v %v", r.Valid(), err)
	}
	out, err := r.Output("basic")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "does not match pattern ^[a-z]+$") || strings.Contains(string(out), secret) {
		t.Errorf("bad output: %s", out)
	}
}

func TestPatternConcurrent(t *testing.T) {
	s, err := jsonschema.Compile([]byte(`{"pattern":"^\\w+$"}`))
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				if r, err := s.Validate([]byte(`"abc"`)); err != nil || !r.Valid() {
					t.Error("want valid")
				}
				if r, err := s.Validate([]byte(`"a b"`)); err != nil || r.Valid() {
					t.Error("want invalid")
				}
			}
		}()
	}
	wg.Wait()
}
