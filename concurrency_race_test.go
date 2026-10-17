package jsonschema

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
)

const raceSchema = `{"type":"object","properties":{"a":{"type":"integer","minimum":0},"list":{"items":{"type":"integer"}},"t":{"$ref":"#"}}}`

// The caps also apply to the schema's meta-check, so they leave it room.
var raceOpts = Options{MaxDepth: 40, MaxSteps: 5000, MaxRefChain: 20, MaxOutputUnits: 5}

var raceShapes = []string{"flag", "basic", "detailed", "verbose"}

// raceDocs maps a document to the limit it must hit ("" for none).
func raceDocs() map[string]Limit {
	return map[string]Limit{
		`{"a":1}`:  "",
		`{"a":-1}`: "",
		`{"list":` + strings.Repeat("[", 50) + strings.Repeat("]", 50) + `}`: LimitDepth,
		strings.Repeat(`{"t":`, 25) + `{}` + strings.Repeat("}", 25):         LimitRefChain,
		`{"list":[` + strings.Repeat("1,", 5000) + `1]}`:                     LimitSteps,
		`{"list":[` + strings.Repeat(`"x",`, 9) + `"x"]}`:                    LimitOutputUnits,
	}
}

// outcome is a comparable summary of Validate plus Output for every shape.
func outcome(s *Schema, doc string) string {
	res, err := s.Validate([]byte(doc))
	if err != nil {
		var le *LimitError
		if errors.As(err, &le) {
			return fmt.Sprintf("limit %s %d", le.Limit, le.Value)
		}
		return "error " + err.Error()
	}
	out := fmt.Sprint(res.Valid())
	for _, sh := range raceShapes {
		b, err := res.Output(sh)
		out += fmt.Sprintf("|%s:%s:%v", sh, b, err)
	}
	return out
}

func TestConcurrentValidateAndOutput(t *testing.T) {
	s, err := CompileWithOptions([]byte(raceSchema), raceOpts)
	if err != nil {
		t.Fatal(err)
	}
	docs := raceDocs()
	want := map[string]string{}
	for d, lim := range docs {
		want[d] = outcome(s, d)
		if lim != "" && !strings.HasPrefix(want[d], "limit "+string(lim)) {
			t.Fatalf("doc %.30s: want limit %s, got %.60s", d, lim, want[d])
		}
		if lim == "" && strings.HasPrefix(want[d], "error ") {
			t.Fatalf("doc %.30s: unexpected %s", d, want[d])
		}
	}
	var wg sync.WaitGroup
	for g := 0; g < 16; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 20; i++ {
				for d := range docs {
					if got := outcome(s, d); got != want[d] {
						t.Errorf("doc %.30s: concurrent result differs", d)
						return
					}
				}
			}
		}()
	}
	wg.Wait()
}

func TestSchemaUnchangedAfterManyValidations(t *testing.T) {
	s, err := CompileWithOptions([]byte(raceSchema), raceOpts)
	if err != nil {
		t.Fatal(err)
	}
	lim, el, max, maxAnn, maxRep, maxIn := s.lim, s.el, s.max, s.maxAnn, s.maxReport, s.maxInput
	docs := raceDocs()
	before := map[string]string{}
	for d := range docs {
		before[d] = outcome(s, d)
	}
	for i := 0; i < 1000; i++ {
		for d := range docs {
			_, _ = s.Validate([]byte(d))
		}
	}
	if s.lim != lim || s.el != el || s.max != max || s.maxAnn != maxAnn || s.maxReport != maxRep || s.maxInput != maxIn {
		t.Fatal("compiled limits changed after validations")
	}
	for d := range docs {
		if got := outcome(s, d); got != before[d] {
			t.Errorf("doc %.30s: behaviour changed after validations", d)
		}
	}
}
