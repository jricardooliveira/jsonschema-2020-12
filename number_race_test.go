package jsonschema

import (
	"sync"
	"testing"
)

// One compiled schema validates fast-path and big-path numbers from many
// goroutines; every goroutine must see the single-threaded answer.
// Run with -race.
func TestConcurrentFastAndBigNumbers(t *testing.T) {
	const schema = `{"type":"array","uniqueItems":true,"items":{"type":"number","multipleOf":0.1,"maximum":1e30,"exclusiveMinimum":-1e30},"not":{"contains":{"const":7}}}`
	s, err := Compile([]byte(schema))
	if err != nil {
		t.Fatal(err)
	}
	docs := []string{
		`[0.1,0.3,0.7]`,
		`[0.1,0.10]`,
		`[9223372036854775807,9223372036854775808]`,
		`[12345678901234567890.5,1e2,1.0,1]`,
		`[0.30000000000000001]`,
		`[7]`,
		`[2e30]`,
	}
	want := make([]string, len(docs))
	for i, d := range docs {
		want[i] = outcome(s, d)
	}
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for r := 0; r < 50; r++ {
				for i, d := range docs {
					if got := outcome(s, d); got != want[i] {
						t.Errorf("doc %s: got %.80s want %.80s", d, got, want[i])
						return
					}
				}
			}
		}()
	}
	wg.Wait()
}
