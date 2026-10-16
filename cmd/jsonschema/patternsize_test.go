package main

import (
	"strings"
	"testing"

	"github.com/neoface/jsonschema/internal/keywords/validation"
)

func TestOversizePatternExitsTwo(t *testing.T) {
	s := write(t, "s.json", `{"pattern":"`+strings.Repeat("a", validation.DefaultMaxPatternSize+1)+`"}`)
	d := write(t, "d.json", `"x"`)
	code, out, errOut := do([]string{"validate", s, d}, "")
	if code != exitError || out != "" || strings.Contains(errOut, "aaaa") {
		t.Errorf("code=%d out=%q err=%q", code, out, errOut)
	}
}
