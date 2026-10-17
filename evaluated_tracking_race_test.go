package jsonschema

import (
	"sync"
	"testing"
)

// One compiled schema that skips recording and one that keeps it are shared
// by many goroutines asking for every report shape; each result must equal
// the serial one. Run with -race.
func TestConcurrentSkipAndKeepRecording(t *testing.T) {
	schemas := map[string]string{
		"skip": `{"type":"object","properties":{"a":{"type":"integer"},"l":{"items":{"type":"integer"}}}}`,
		"keep": `{"type":"object","properties":{"a":{"type":"integer"},"l":{"prefixItems":[{"type":"integer"}],"unevaluatedItems":false}},"unevaluatedProperties":false}`,
	}
	docs := []string{`{"a":1}`, `{"a":"x"}`, `{"l":[1,2]}`, `{"b":1,"l":[1]}`, `[]`}
	for name, src := range schemas {
		s, err := Compile([]byte(src))
		if err != nil {
			t.Fatal(err)
		}
		if want := name == "skip"; s.root.NoUnevaluated != want {
			t.Fatalf("%s: NoUnevaluated = %v", name, s.root.NoUnevaluated)
		}
		want := map[string]string{}
		for _, d := range docs {
			want[d] = outcome(s, d)
		}
		var wg sync.WaitGroup
		for g := 0; g < 16; g++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < 20; i++ {
					for _, d := range docs {
						if got := outcome(s, d); got != want[d] {
							t.Errorf("%s %s: concurrent result differs", name, d)
							return
						}
					}
				}
			}()
		}
		wg.Wait()
	}
}
