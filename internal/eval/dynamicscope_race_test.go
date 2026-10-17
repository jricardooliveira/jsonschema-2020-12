package eval_test

import (
	"sync"
	"testing"

	"github.com/neoface/jsonschema"
)

// core 8.2.3.2: the dynamic scope is per call, so parallel validations with
// one compiled schema must not see each other (run with -race).
func TestDynamicScopeConcurrentValidation(t *testing.T) {
	const schema = `{"$id":"https://t.example/strict","$dynamicAnchor":"node","$ref":"tree","required":["data"],
 "$defs":{"tree":{"$id":"https://t.example/tree","$dynamicAnchor":"node","type":"object",
  "properties":{"data":true,"children":{"type":"array","items":{"$dynamicRef":"#node"}}}}}}`
	s, err := jsonschema.Compile([]byte(schema))
	if err != nil {
		t.Fatal(err)
	}
	docs := map[string]bool{
		`{"data":1,"children":[{"data":2,"children":[{"data":3}]}]}`:      true,
		`{"data":1,"children":[{"data":2,"children":[{"children":[]}]}]}`: false,
	}
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		for doc, want := range docs {
			wg.Add(1)
			go func() {
				defer wg.Done()
				r, err := s.Validate([]byte(doc))
				if err != nil || r.Valid() != want {
					t.Errorf("%s: valid=%v err=%v want %v", doc, r.Valid(), err, want)
				}
			}()
		}
	}
	wg.Wait()
}
