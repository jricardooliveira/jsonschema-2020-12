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
	cases := []struct {
		a, b string
		want bool
	}{
		{`null`, `null`, true}, {`null`, `false`, false}, {`null`, `0`, false},
		{`true`, `true`, true}, {`true`, `false`, false}, {`false`, `0`, false},
		{`"a"`, `"a"`, true}, {`"a"`, `"b"`, false}, {`"1"`, `1`, false},
		{`1`, `1.0`, true}, {`1`, `1e0`, true}, {`0`, `-0`, true}, {`0.1`, `1e-1`, true},
		{`0.1`, `0.10000000000000001`, false}, {`1`, `2`, false}, {`1`, `-1`, false},
		{`12345678901234567890`, `12345678901234567890`, true},
		{`12345678901234567890`, `12345678901234567891`, false},
		{`[1,2]`, `[1.0,2]`, true}, {`[1,2]`, `[2,1]`, false}, {`[1]`, `[1,1]`, false}, {`[]`, `{}`, false},
		{`{"a":1,"b":[true]}`, `{"b":[true],"a":1.0}`, true},
		{`{"a":1}`, `{"a":2}`, false}, {`{"a":1}`, `{"b":1}`, false}, {`{"a":1}`, `{"a":1,"b":1}`, false},
		{`{"a":{"b":[null]}}`, `{"a":{"b":[null]}}`, true},
		{`{"a":{"b":[null]}}`, `{"a":{"b":[0]}}`, false},
	}
	for _, c := range cases {
		a, b := dec(t, c.a), dec(t, c.b)
		if got := Equal(a, b); got != c.want {
			t.Errorf("Equal(%s, %s) = %v, want %v", c.a, c.b, got, c.want)
		}
		if Equal(b, a) != c.want {
			t.Errorf("Equal not symmetric for %s, %s", c.a, c.b)
		}
	}
	// json.Number is not a model value: a different kind.
	if Equal(json.Number("1"), json.Number("1")) {
		t.Error("json.Number must not be equal")
	}
}

func TestEqualConcurrent(t *testing.T) {
	a, b := dec(t, `{"a":[1e5,0.5]}`), dec(t, `{"a":[100000,5e-1]}`)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				if !Equal(a, b) {
					t.Error("not equal")
					return
				}
			}
		}()
	}
	wg.Wait()
}
