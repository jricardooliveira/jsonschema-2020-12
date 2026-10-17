package jsonschema_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/neoface/jsonschema"
)

// Reused child contexts are per call: concurrent Validate calls on one
// compiled schema must not share them (run with -race).
func TestConcurrentValidateReusesChildContextsPerCall(t *testing.T) {
	s, err := jsonschema.Compile([]byte(`{"type":"object","properties":{"a":{"allOf":[{"type":"array","items":{"type":"integer","maximum":5}}]}}}`))
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				bad := (g+i)%2 == 0
				v := 1
				if bad {
					v = 9
				}
				res, err := s.Validate([]byte(fmt.Sprintf(`{"a":[1,%d,2]}`, v)))
				if err != nil {
					t.Error(err)
					return
				}
				if res.Valid() == bad {
					t.Errorf("g=%d i=%d: valid=%v want %v", g, i, res.Valid(), !bad)
					return
				}
			}
		}(g)
	}
	wg.Wait()
}
