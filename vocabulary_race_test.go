package jsonschema_test

import (
	"sync"
	"testing"
)

// core 8.1.2: vocabulary sets are fixed at Compile and only read during
// evaluation, so one schema with a restricted-vocabulary meta-schema and an
// embedded 2020-12 resource gives identical results from many goroutines
// (run with -race).
func TestVocabularyConcurrentValidate(t *testing.T) {
	v := `{"` + vocabBase + `core":true,"` + vocabBase + `applicator":true}`
	s := compileWithMeta(t, v, `{"$schema":"https://example.com/meta","minimum":5,"properties":{
		"a":{"$id":"https://example.com/in","$schema":"https://json-schema.org/draft/2020-12/schema","minimum":5}}}`)
	docs := map[string]bool{`3`: true, `{"a":3}`: false, `{"a":7}`: true}
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 50 {
				for d, want := range docs {
					r, err := s.Validate([]byte(d))
					if err != nil {
						t.Error(err)
						return
					}
					if r.Valid() != want {
						t.Errorf("doc %s: valid = %v, want %v", d, r.Valid(), want)
					}
				}
			}
		}()
	}
	wg.Wait()
}
