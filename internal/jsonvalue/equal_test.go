package jsonvalue

import (
	"encoding/json"
	"sync"
	"testing"
)

func dec(t *testing.T, s string) any {
	t.Helper()
	v, err := DecodeLimited([]byte(s), NumberLimits{})
	if err != nil {
		t.Fatalf("%s: %v", s, err)
	}
	return v
}

func TestEqual(t *testing.T) {
	// Instance equality is defined by 2020-12 core §4.2.2 (used by
	// validation §6.1.2 enum and §6.1.3 const); number values by RFC 8259 §6.
	const (
		kinds   = "core 4.2.2 (kinds)"
		strs    = "core 4.2.2 (strings)"
		nums    = "core 4.2.2 (numbers); RFC 8259 §6"
		arrays  = "core 4.2.2 (arrays)"
		objects = "core 4.2.2 (objects); RFC 8259 §4"
	)
	cases := []struct {
		spec string
		a, b string
		want bool
	}{
		{kinds, `null`, `null`, true}, {kinds, `null`, `false`, false}, {kinds, `null`, `0`, false},
		{kinds, `true`, `true`, true}, {kinds, `true`, `false`, false}, {kinds, `false`, `0`, false},
		{strs, `"a"`, `"a"`, true}, {strs, `"a"`, `"b"`, false}, {kinds, `"1"`, `1`, false},
		{nums, `1`, `1.0`, true}, {nums, `1`, `1e0`, true}, {nums, `0`, `-0`, true}, {nums, `0.1`, `1e-1`, true},
		{nums, `0.1`, `0.10000000000000001`, false}, {nums, `1`, `2`, false}, {nums, `1`, `-1`, false},
		{nums, `12345678901234567890`, `12345678901234567890`, true},
		{nums, `12345678901234567890`, `12345678901234567891`, false},
		{arrays, `[1,2]`, `[1.0,2]`, true}, {arrays, `[1,2]`, `[2,1]`, false}, {arrays, `[1]`, `[1,1]`, false}, {kinds, `[]`, `{}`, false},
		{objects, `{"a":1,"b":[true]}`, `{"b":[true],"a":1.0}`, true},
		{objects, `{"a":1}`, `{"a":2}`, false}, {objects, `{"a":1}`, `{"b":1}`, false}, {objects, `{"a":1}`, `{"a":1,"b":1}`, false},
		{objects, `{"a":{"b":[null]}}`, `{"a":{"b":[null]}}`, true},
		{objects, `{"a":{"b":[null]}}`, `{"a":{"b":[0]}}`, false},
	}
	for _, c := range cases {
		a, b := dec(t, c.a), dec(t, c.b)
		if got := Equal(a, b); got != c.want {
			t.Errorf("%s: Equal(%s, %s) = %v, want %v", c.spec, c.a, c.b, got, c.want)
		}
		if Equal(b, a) != c.want {
			t.Errorf("%s: Equal not symmetric for %s, %s", c.spec, c.a, c.b)
		}
	}
	// json.Number is not a model value: a different kind.
	if Equal(json.Number("1"), json.Number("1")) {
		t.Error("json.Number must not be equal")
	}
}

// Run with -race: Equal and Cmp only read the shared big.Int mantissas.
func TestEqualConcurrent(t *testing.T) {
	a, b := dec(t, `{"a":[1e5,0.5]}`), dec(t, `{"a":[100000,5e-1]}`)
	// Same adjusted exponent, different exp: Cmp must align mantissas.
	x, y, z := num(t, "1.5"), num(t, "15e-1"), num(t, "1.25")
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				if !Equal(a, b) || Cmp(x, y) != 0 || Cmp(x, z) != 1 || Cmp(z, y) != -1 {
					t.Error("concurrent Equal/Cmp gave a wrong answer")
					return
				}
			}
		}()
	}
	wg.Wait()
}
