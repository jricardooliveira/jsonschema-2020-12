package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Documents with fast-path and big-path numbers: one below each limit's
// threshold the CLI exits 2 and names the limit; at the threshold it does
// not exit 2. The thresholds are the ones number_limits_regression_test.go
// pins in the library for the same schemas and documents.
func TestNumberDocsLimitExitCode(t *testing.T) {
	const nums = `1,1.5,9223372036854775807,9223372036854775808,12345678901234567890.5,1e2`
	flat := `[` + nums + `]`
	cases := []struct {
		flag, schema, doc string
		threshold         int
	}{
		{"max-depth", `true`, strings.Repeat("[", 12) + nums + strings.Repeat("]", 12), 12},
		{"max-input-bytes", `true`, flat, len(flat)},
		{"max-steps", `{"items":{"type":"number","multipleOf":0.5}}`, flat, 199},
		{"max-ref-chain", `{"$defs":{"a":{"$ref":"#/$defs/b"},"b":{"$ref":"#/$defs/c"},"c":{"type":"number"}},"$ref":"#/$defs/a"}`, `9223372036854775808`, 5},
		{"max-output-units", `{"items":{"multipleOf":0.3}}`, flat, 5},
		{"max-annotations", `{"items":{"title":"t"}}`, flat, 18},
		{"max-report-bytes", `{"items":{"multipleOf":0.3}}`, flat, 636},
	}
	dir := t.TempDir()
	for i, c := range cases {
		t.Run(c.flag, func(t *testing.T) {
			schemaPath := filepath.Join(dir, strconv.Itoa(i)+"s.json")
			docPath := filepath.Join(dir, strconv.Itoa(i)+"d.json")
			if err := os.WriteFile(schemaPath, []byte(c.schema), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(docPath, []byte(c.doc), 0o600); err != nil {
				t.Fatal(err)
			}
			exec := func(v int) (int, string) {
				var out, errb bytes.Buffer
				code := run([]string{"validate", "--" + c.flag + "=" + strconv.Itoa(v), schemaPath, docPath}, nil, &out, &errb)
				return code, errb.String()
			}
			code, stderr := exec(c.threshold - 1)
			if code != exitError || !strings.Contains(stderr, "--"+c.flag) {
				t.Errorf("cap %d: exit %d stderr %q, want exit %d naming --%s", c.threshold-1, code, stderr, exitError, c.flag)
			}
			if code, stderr := exec(c.threshold); code == exitError {
				t.Errorf("cap %d: exit %d stderr %q, want a valid/invalid result", c.threshold, code, stderr)
			}
		})
	}
}
