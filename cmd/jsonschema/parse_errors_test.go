package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neoface/jsonschema"
)

const secret = "s3cr3tvalue"

func runCLI(t *testing.T, stdin string, args ...string) (int, string, string) {
	t.Helper()
	var out, errb bytes.Buffer
	code := run(args, strings.NewReader(stdin), &out, &errb)
	if strings.Contains(out.String()+errb.String(), secret) {
		t.Errorf("output leaks secret: %q %q", out.String(), errb.String())
	}
	return code, out.String(), errb.String()
}

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "f.json")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestBrokenJSONExitsTwo(t *testing.T) {
	good := writeTemp(t, `true`)
	bad := []string{
		`{`, `{"a":1,"x":{"` + secret + `":1,"` + secret + `":2}}`, "\"\xff" + secret + "\"",
		`"\uD800` + secret + `"`, `1 ` + secret,
	}
	for _, b := range bad {
		if c, out, _ := runCLI(t, "", "validate", writeTemp(t, b), good); c != 2 || out != "" {
			t.Errorf("schema %q: code %d out %q", b, c, out)
		}
		if c, out, _ := runCLI(t, "", "validate", good, writeTemp(t, b)); c != 2 || out != "" {
			t.Errorf("doc file %q: code %d out %q", b, c, out)
		}
		if c, out, _ := runCLI(t, b, "validate", good); c != 2 || out != "" {
			t.Errorf("stdin %q: code %d out %q", b, c, out)
		}
	}
}

// Files go through readCapped too, not only stdin (ADR-001 input-size limit).
func TestOversizeFilesExitTwo(t *testing.T) {
	good := writeTemp(t, `true`)
	big := writeTemp(t, strings.Repeat(" ", maxInputBytes+1))
	for _, args := range [][]string{{"validate", big, good}, {"validate", good, big}} {
		if c, out, e := runCLI(t, "", args...); c != 2 || out != "" || !strings.Contains(e, "limit exceeded: input size") {
			t.Errorf("%v: code %d out %q err %q", args, c, out, e)
		}
	}
}

func TestNumberCapFlags(t *testing.T) {
	sch := writeTemp(t, `true`)
	over := fmt.Sprintf("1e%d", jsonschema.DefaultMaxNumberExponent+1)
	if c, out, _ := runCLI(t, over, "validate", sch); c != 2 || out != "" {
		t.Errorf("over cap: %d %q", c, out)
	}
	if c, _, _ := runCLI(t, over, "validate", "--max-number-exponent", fmt.Sprint(jsonschema.DefaultMaxNumberExponent+1), sch); c != 0 {
		t.Errorf("raised cap: %d", c)
	}
	long := strings.Repeat("1", jsonschema.DefaultMaxNumberDigits+1)
	if c, _, _ := runCLI(t, long, "validate", fmt.Sprintf("-max-number-digits=%d", jsonschema.DefaultMaxNumberDigits+1), sch); c != 0 {
		t.Errorf("raised digits: %d", c)
	}
	for _, v := range []string{"0", "-1"} {
		c, out, e := runCLI(t, "1", "validate", "--max-number-digits="+v, sch)
		if c != 2 || out != "" || e == "" || strings.Count(strings.TrimSpace(e), "\n") != 0 {
			t.Errorf("value %s: %d %q %q", v, c, out, e)
		}
	}
	c, _, e := runCLI(t, "1", "validate", "--max-number-digits=abc", sch)
	if c != 2 || strings.Contains(e, "unknown flag") {
		t.Errorf("bad value: %d %q", c, e)
	}
}

// The usage line names the flag whose parsing failed, never its value.
func TestFlagErrorNamesFailingFlag(t *testing.T) {
	sch := writeTemp(t, `true`)
	c, out, e := runCLI(t, "1", "validate", "--max-number-digits=10", "--max-number-exponent="+secret, sch)
	if c != 2 || out != "" || strings.TrimSpace(e) != "invalid value for flag -max-number-exponent" {
		t.Errorf("two flags: %d %q %q", c, out, e)
	}
	c, out, e = runCLI(t, "1", "validate", "--max-number-digits=10", "--bogus="+secret, sch)
	if c != 2 || out != "" || strings.TrimSpace(e) != "unknown flag -bogus" {
		t.Errorf("unknown name=value: %d %q %q", c, out, e)
	}
}
