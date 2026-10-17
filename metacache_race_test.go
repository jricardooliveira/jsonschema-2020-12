package jsonschema

import (
	"sync"
	"testing"
)

// Many goroutines compile (the first of them builds the shared built-in
// meta-schema set) and validate at once; run with -race.
func TestConcurrentCompileAndValidate(t *testing.T) {
	const schema = `{"$schema":"https://json-schema.org/draft/2020-12/schema","type":"object","properties":{"a":{"type":"integer"}},"unevaluatedProperties":false}`
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				s, err := Compile([]byte(schema))
				if err != nil {
					t.Error(err)
					return
				}
				if res, err := s.Validate([]byte(`{"a":1}`)); err != nil || !res.Valid() {
					t.Errorf("valid doc: valid=%v err=%v", res.Valid(), err)
				}
				if res, err := s.Validate([]byte(`{"a":"x"}`)); err != nil || res.Valid() {
					t.Errorf("invalid doc: valid=%v err=%v", res.Valid(), err)
				}
				if _, err := Compile([]byte(`{"type":5}`)); err == nil {
					t.Error("bad schema accepted")
				}
			}
		}()
	}
	wg.Wait()
}
