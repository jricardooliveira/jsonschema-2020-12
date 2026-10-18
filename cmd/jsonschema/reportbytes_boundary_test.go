package main

import (
	"strconv"
	"strings"
	"testing"
)

// checkReportBoundary runs the basic report unlimited, then with the limit set
// to its exact size (must pass, same bytes) and one byte less (must fail with
// exit code 2 and a report-bytes limit error).
func checkReportBoundary(t *testing.T, schema, doc string) {
	t.Helper()
	code, full, _ := runCLI(t, "", "validate", "-o", "basic", schema, doc)
	if code != exitInvalid {
		t.Fatalf("unlimited run exit = %d, want %d", code, exitInvalid)
	}
	size := len(strings.TrimSuffix(full, "\n"))
	limit := func(n int) (int, string, string) {
		return runCLI(t, "", "validate", "-o", "basic", "-max-report-bytes", strconv.Itoa(n), schema, doc)
	}
	c, out, _ := limit(size)
	if c != exitInvalid || out != full {
		t.Errorf("limit == size: exit %d, same output %v", c, out == full)
	}
	c, out, errs := limit(size - 1)
	if c != exitError || out != "" || !strings.Contains(errs, "report") {
		t.Errorf("limit == size-1: exit %d, stdout %q, stderr %q", c, out, errs)
	}
	// the failure is the same error whatever the limit below the size
	_, _, errs2 := limit(size - 2)
	if size > 2 && errs2 == "" {
		t.Error("limit == size-2 gave no error")
	}
}

func TestReportBytesBoundaryBasic(t *testing.T) {
	schema := writeTemp(t, `{"properties":{"a":{"type":"string"},"b":{"type":"string"}}}`)
	doc := writeTemp(t, `{"a":1,"b":2}`)
	checkReportBoundary(t, schema, doc)
}

func TestReportBytesBoundaryTilde(t *testing.T) {
	checkReportBoundary(t, "../../testdata/reportbytes/tilde_schema.json", "../../testdata/reportbytes/tilde_document.json")
	_, out, _ := runCLI(t, "", "validate", "../../testdata/reportbytes/tilde_schema.json", "../../testdata/reportbytes/tilde_document.json")
	if !strings.Contains(out, "a~0b") {
		t.Errorf("fixture does not exercise '~': %s", out)
	}
}
