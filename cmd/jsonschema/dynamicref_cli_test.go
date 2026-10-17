package main

import (
	"strings"
	"testing"
	"time"
)

const (
	dynDeepBad = `{"data":1,"children":[{"data":2,"children":[{"children":[]}]}]}`
	dynDeepOK  = `{"data":1,"children":[{"data":2,"children":[{"data":3}]}]}`
)

// core 8.2.3.2 recursive extension, through the CLI exit codes.
func TestDynamicRefCLIRecursiveExtension(t *testing.T) {
	for name, c := range map[string]struct {
		schema, doc string
		want        int
	}{
		"plain tree, missing data ok":  {"testdata/dynamicref/tree.json", dynDeepBad, exitValid},
		"extension, all levels":        {"testdata/dynamicref/strict-tree.json", dynDeepOK, exitValid},
		"extension, deep missing data": {"testdata/dynamicref/strict-tree.json", dynDeepBad, exitInvalid},
	} {
		code, _, errOut := do([]string{"validate", c.schema, write(t, "d.json", c.doc)}, "")
		if code != c.want {
			t.Errorf("%s: code=%d want %d err=%q", name, code, c.want, errOut)
		}
	}
}

// core 8.2.2 / 8.2.3.2: broken schemas and loops are exit 2, and never hang.
func TestDynamicRefCLIErrorsExit2(t *testing.T) {
	for _, name := range []string{"duplicate-anchor", "unresolved", "cycle"} {
		done := make(chan struct{})
		go func() {
			defer close(done)
			code, out, errOut := do([]string{"validate", "testdata/dynamicref/" + name + ".json", write(t, "d.json", `{}`)}, "")
			wantErr(t, name, code, out, errOut)
		}()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Fatalf("%s: did not terminate", name)
		}
	}
}

// output 12.3: the basic report names /$dynamicRef and never the value.
func TestDynamicRefCLIBasicLocation(t *testing.T) {
	s := write(t, "s.json", `{"$defs":{"a":{"$anchor":"a","type":"number"}},"$dynamicRef":"#a"}`)
	code, out, _ := do([]string{"validate", s, write(t, "d.json", `"secret-value-123"`)}, "")
	if code != exitInvalid {
		t.Fatalf("code=%d", code)
	}
	if !strings.Contains(out, "/$dynamicRef") || strings.Contains(out, "secret-value-123") {
		t.Errorf("bad report: %s", out)
	}
}
