package main

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12"
)

// libOptions sets the Options field behind a limit flag name.
func libOptions(flag string, n int) jsonschema.Options {
	var o jsonschema.Options
	switch flag {
	case "max-depth":
		o.MaxDepth = n
	case "max-ref-chain":
		o.MaxRefChain = n
	case "max-steps":
		o.MaxSteps = n
	case "max-output-units":
		o.MaxOutputUnits = n
	case "max-report-bytes":
		o.MaxReportBytes = n
	case "max-annotations":
		o.MaxAnnotations = n
	}
	return o
}

// libRun compiles, validates and renders like the CLI does.
func libRun(c edgeCase, n int) error {
	s, err := jsonschema.CompileWithOptions([]byte(c.schema), libOptions(c.flag, n))
	if err != nil {
		return err
	}
	res, err := s.Validate([]byte(c.doc))
	if err != nil {
		return err
	}
	shape := c.shape
	if shape == "" {
		shape = "flag"
	}
	_, err = res.Output(shape)
	return err
}

// limitCases are the shared edge cases plus annotations recorded inside
// 'not', which are dropped but must still count against the cap.
func limitCases() []edgeCase {
	c := edgeCase{"annotations in not", "max-annotations",
		`{"not":{"type":"string","title":"t"},"title":"x"}`, `1`, "basic", 1, 0}
	// The compile-time meta-schema check shares the cap, so find the edge
	// from the library instead of repeating a number.
	for libRun(c, c.edge) != nil && c.edge < 10000 {
		c.edge++
	}
	return append(edgeCases(), c)
}

// TestAnnotationLimitsUnchangedThroughLibraryAndCLI: for every hostile-input
// limit, the input just under the limit passes and the input just over it is
// a limit error (library) and exit 2 (CLI). Edge values are the ones the
// code gave before annotations became lazy.
func TestAnnotationLimitsUnchangedThroughLibraryAndCLI(t *testing.T) {
	for _, c := range limitCases() {
		t.Run(c.name, func(t *testing.T) {
			if err := libRun(c, c.edge); err != nil {
				t.Errorf("library at the limit: %v", err)
			}
			err := libRun(c, c.edge-1)
			var le *jsonschema.LimitError
			if !errors.As(err, &le) {
				t.Fatalf("library over the limit: want LimitError, got %v", err)
			}
			if limitFlag[le.Limit] != c.flag {
				t.Errorf("library limit %v maps to --%s, want --%s", le.Limit, limitFlag[le.Limit], c.flag)
			}

			sch := writeTemp(t, c.schema)
			try := func(n int) (int, string, string) {
				args := []string{"validate", "--" + c.flag + "=" + strconv.Itoa(n)}
				if c.shape != "" {
					args = append(args, "-o", c.shape)
				}
				return runCLI(t, c.doc, append(args, sch)...)
			}
			if code, _, e := try(c.edge); code != c.wantCode {
				t.Errorf("CLI at the limit: code %d, want %d (%q)", code, c.wantCode, e)
			}
			code, out, e := try(c.edge - 1)
			if code != 2 || out != "" || !strings.Contains(e, "limit exceeded") {
				t.Errorf("CLI over the limit: code %d out %q stderr %q", code, out, e)
			}
		})
	}
}
