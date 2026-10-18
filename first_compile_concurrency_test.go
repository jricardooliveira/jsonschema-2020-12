package jsonschema

import (
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/perfdata"
)

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
