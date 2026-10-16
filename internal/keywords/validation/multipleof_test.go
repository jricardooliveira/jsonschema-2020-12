package validation_test

// Derived from JSON Schema validation spec section 6.2.1 (multipleOf).

import (
	"math/big"
	"strconv"
	"sync"
	"testing"

	"github.com/neoface/jsonschema"
)

func multipleOfValid(t *testing.T, schema, doc string, opts jsonschema.Options) bool {
	t.Helper()
	s, err := jsonschema.CompileWithOptions([]byte(schema), opts)
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.Validate([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	return r.Valid()
}

func TestMultipleOfSchemaErrors(t *testing.T) {
	for _, s := range []string{`{"multipleOf":0}`, `{"multipleOf":-1}`, `{"multipleOf":"2"}`, `{"multipleOf":null}`} {
		if _, err := jsonschema.Compile([]byte(s)); err == nil {
			t.Errorf("%s: want schema error", s)
		}
	}
}

func TestMultipleOfMatchesRat(t *testing.T) {
	vals := []string{"0.1", "0.3", "0.35", "0.5", "2.5", "0.125", "3", "1.5", "7", "12", "1e2", "0.0625", "10", "4.5", "0"}
	for _, a := range vals {
		for _, b := range vals[:len(vals)-1] {
			ra, _ := new(big.Rat).SetString(a)
			rb, _ := new(big.Rat).SetString(b)
			want := new(big.Rat).Quo(ra, rb).IsInt()
			if got := multipleOfValid(t, `{"multipleOf":`+b+`}`, a, jsonschema.Options{}); got != want {
				t.Errorf("%s / %s: got %v want %v", a, b, got, want)
			}
		}
	}
}

func TestMultipleOfHostileScale(t *testing.T) {
	e := strconv.Itoa(1_000_000_000)
	o := jsonschema.Options{MaxNumberExponent: 2_000_000_000}
	cases := []struct {
		schema, doc string
		want        bool
	}{
		{`{"multipleOf":0.1}`, `1e` + e, true},
		{`{"multipleOf":0.1}`, `1e-` + e, false},
		{`{"multipleOf":3}`, `1e` + e, false},
		{`{"multipleOf":3}`, `3e` + e, true},
		{`{"multipleOf":1e-` + e + `}`, `1e` + e, true},
		{`{"multipleOf":1e` + e + `}`, `1e-` + e, false},
	}
	for _, c := range cases {
		if got := multipleOfValid(t, c.schema, c.doc, o); got != c.want {
			t.Errorf("%s vs %s: got %v", c.schema, c.doc, got)
		}
	}
}

func TestMultipleOfConcurrent(t *testing.T) {
	s, err := jsonschema.Compile([]byte(`{"multipleOf":0.1}`))
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := s.Validate([]byte(`0.3`))
			if err != nil || !r.Valid() {
				t.Error("want valid")
			}
		}()
	}
	wg.Wait()
}
