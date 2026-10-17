package jsonschema

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/perfdata"
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

const firstCompileEnv = "JSONSCHEMA_FIRST_COMPILE_CHILD"

// TestFirstCompileConcurrent re-executes this test binary in a fresh process,
// where 32 goroutines race on the very first Compile (the shared meta-schema
// set is not built yet). It fails on a non-zero exit or a data-race report.
func TestFirstCompileConcurrent(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=^TestFirstCompileChild$", "-test.v")
	cmd.Env = append(os.Environ(), firstCompileEnv+"=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("child failed: %v\n%s", err, out)
	}
	if strings.Contains(string(out), "WARNING: DATA RACE") {
		t.Fatalf("child reported a data race:\n%s", out)
	}
	if !strings.Contains(string(out), "--- PASS: TestFirstCompileChild") {
		t.Fatalf("child did not run:\n%s", out)
	}
	t.Logf("child ran:\n%s", out)
}

// TestFirstCompileChild is the child half; it does nothing in a normal run.
func TestFirstCompileChild(t *testing.T) {
	if os.Getenv(firstCompileEnv) == "" {
		t.Skip("child of TestFirstCompileConcurrent only")
	}
	schema, err := perfdata.Fixture("order_schema.json")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := perfdata.Fixture("order_small.json")
	if err != nil {
		t.Fatal(err)
	}
	const n = 32
	var ready, done sync.WaitGroup
	start := make(chan struct{})
	errs := make(chan error, n)
	ready.Add(n)
	done.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer done.Done()
			ready.Done()
			<-start
			s, err := Compile(schema)
			if err != nil {
				errs <- err
				return
			}
			r, err := s.Validate(doc)
			if err != nil {
				errs <- err
				return
			}
			if !r.Valid() {
				errs <- os.ErrInvalid
			}
		}()
	}
	ready.Wait()
	close(start)
	done.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
}

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

// One Schema used from many goroutines gives the same annotations and
// report each time (run with -race).
func TestConcurrentValidateAnnotations(t *testing.T) {
	s, err := Compile([]byte(`{"title":"T","properties":{"a":{"default":7},"b":{"anyOf":[{"type":"string","x-a":1},{"x-b":[2]}]}},"items":{"examples":[1.5]}}`))
	if err != nil {
		t.Fatal(err)
	}
	docs := [][]byte{[]byte(`{"a":1,"b":2}`), []byte(`[1,2,3]`), []byte(`{"b":"s"}`)}
	type out struct {
		ann []Annotation
		rep []byte
	}
	want := make([]out, len(docs))
	for i, d := range docs {
		r, err := s.Validate(d)
		if err != nil {
			t.Fatal(err)
		}
		rep, err := r.Output("basic")
		if err != nil {
			t.Fatal(err)
		}
		want[i] = out{r.Annotations(), rep}
	}
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := 0; n < 50; n++ {
				i := n % len(docs)
				r, err := s.Validate(docs[i])
				if err != nil {
					t.Error(err)
					return
				}
				rep, err := r.Output("basic")
				if err != nil || !bytes.Equal(rep, want[i].rep) || !reflect.DeepEqual(r.Annotations(), want[i].ann) {
					t.Errorf("doc %d differs: %s %v", i, rep, err)
					return
				}
			}
		}()
	}
	wg.Wait()
}
