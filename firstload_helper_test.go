package jsonschema

import (
	"os"
	"os/exec"
	"testing"
)

const freshChildEnv = "JSONSCHEMA_FRESH_CHILD"

// freshProcess runs the calling test alone in a fresh process (a re-exec of
// the test binary with an anchored -test.run). In the parent it runs the
// child, fails t if the child fails, and returns the child's combined output
// with child=false: the caller must then return. In the child it returns
// child=true and the caller runs its body, so nothing else compiled a schema
// first. Callers must measure at the very start of their body.
func freshProcess(t *testing.T) (out string, child bool) {
	t.Helper()
	if os.Getenv(freshChildEnv) == t.Name() {
		return "", true
	}
	cmd := exec.Command(os.Args[0], "-test.run=^"+t.Name()+"$", "-test.v")
	cmd.Env = append(os.Environ(), freshChildEnv+"="+t.Name())
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("fresh child failed: %v\n%s", err, b)
	}
	return string(b), false
}
