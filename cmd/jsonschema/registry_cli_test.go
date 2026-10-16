package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const regFixtures = "../../testdata/registry/"

func TestParseRef(t *testing.T) {
	cases := []struct {
		in, uri, path string
		mapped, bad   bool
	}{
		{in: "./a=b.json", path: "./a=b.json"},
		{in: `C:\x.json`, path: `C:\x.json`},
		{in: "a.json", path: "a.json"},
		{in: "https://x/a.json=./a.json", uri: "https://x/a.json", path: "./a.json", mapped: true},
		{in: "https://x/a.json=", bad: true},
		{in: "", bad: true},
	}
	for _, c := range cases {
		u, p, m, err := parseRef(c.in)
		if (err != nil) != c.bad || u != c.uri || p != c.path || m != c.mapped {
			t.Errorf("%q: got %q %q %v %v", c.in, u, p, m, err)
		}
	}
}

// --max-ref-file-bytes and --max-ref-total-bytes reach the library: over the
// limit is exit 2 naming the file, a raised limit passes, and 0 or less is a
// usage error.
func TestRegistryLimitFlags(t *testing.T) {
	doc := write(t, "d.json", `{"name":"a"}`)
	main := regFixtures + "file/main.json"
	personPath := regFixtures + "file/person.json"
	person := "https://example.com/person.json=" + personPath
	info, err := os.Stat(personPath)
	if err != nil {
		t.Fatal(err)
	}
	size := strconv.Itoa(int(info.Size()))
	under := strconv.Itoa(int(info.Size()) - 1)
	run := func(extra ...string) (int, string, string) {
		args := append([]string{"validate", "--ref", person, "--ref", regFixtures + "file/self_id.json"}, extra...)
		return do(append(args, main, doc), "")
	}
	for _, c := range []struct {
		name  string
		flags []string
		want  string // text the one-line error must hold
	}{
		{"per-file over", []string{"--max-ref-file-bytes", under}, "person.json"},
		{"total over", []string{"--max-ref-total-bytes", size}, "self_id.json"},
		{"per-file zero", []string{"--max-ref-file-bytes", "0"}, "max-ref-file-bytes"},
		{"total negative", []string{"--max-ref-total-bytes", "-1"}, "max-ref-total-bytes"},
	} {
		code, out, errOut := run(c.flags...)
		wantErr(t, c.name, code, out, errOut)
		if !strings.Contains(errOut, c.want) {
			t.Errorf("%s: error %q does not name %q", c.name, errOut, c.want)
		}
	}
	code, out, errOut := run("--max-ref-file-bytes", size, "--max-ref-total-bytes", strconv.Itoa(1<<20))
	if code != exitValid || strings.Count(out, "\n") != 1 || errOut != "" {
		t.Errorf("raised limits: code=%d out=%q err=%q", code, out, errOut)
	}
}

func TestRegistryFlagsExitCodes(t *testing.T) {
	doc := write(t, "d.json", `{"name":"a"}`)
	bad := write(t, "bad.json", `{}`)
	person := "https://example.com/person.json=" + regFixtures + "file/person.json"
	root := t.TempDir()
	outside := t.TempDir()
	os.WriteFile(filepath.Join(outside, "x.json"), []byte(`{}`), 0o600)
	os.WriteFile(filepath.Join(root, "ok.json"), []byte(`{}`), 0o600)
	linked := os.Symlink(filepath.Join(outside, "x.json"), filepath.Join(root, "l.json")) == nil

	cases := []struct {
		name string
		args []string
		doc  string
		code int
	}{
		{"file mapping valid", []string{"--ref", person, regFixtures + "file/main.json"}, doc, exitValid},
		{"file mapping invalid", []string{"--ref", person, regFixtures + "file/main.json"}, bad, exitInvalid},
		{"address-less ref", []string{"--ref", regFixtures + "file/self_id.json", "--base-uri", "https://z.example/m.json", write(t, "s.json", `{"$ref":"https://other.org/self.json"}`)}, write(t, "i.json", `1`), exitValid},
		{"ref-dir", []string{"--ref-dir", "https://example.com/schemas/=" + regFixtures + "dir/schemas", regFixtures + "dir/main.json"}, write(t, "a.json", `{"address":{"street":"x"}}`), exitValid},
		{"base-uri", []string{"--base-uri", "https://example.com/a/s.json", "--ref", "https://example.com/a/person.json=" + regFixtures + "file/person.json", write(t, "r.json", `{"$ref":"person.json"}`)}, doc, exitValid},
		{"unresolved", []string{regFixtures + "unresolved/main.json"}, doc, exitError},
		{"bad ref-dir syntax", []string{"--ref-dir", "plain", regFixtures + "file/main.json"}, doc, exitError},
		{"empty mapping path", []string{"--ref", "https://x/a.json=", regFixtures + "file/main.json"}, doc, exitError},
		{"duplicate address", []string{"--ref", regFixtures + "conflict/a.json", "--ref", regFixtures + "conflict/b.json", regFixtures + "file/main.json"}, doc, exitError},
		{"broken supplied file", []string{"--ref", regFixtures + "broken/bad.json", regFixtures + "file/main.json"}, doc, exitError},
		{"missing supplied file", []string{"--ref", "nope.json", regFixtures + "file/main.json"}, doc, exitError},
	}
	if linked {
		cases = append(cases, struct {
			name string
			args []string
			doc  string
			code int
		}{"escaping symlink", []string{"--ref-dir", "https://x.example/=" + root, regFixtures + "file/main.json"}, doc, exitError})
	}
	for _, c := range cases {
		args := append([]string{"validate"}, c.args...)
		args = append(args, c.doc)
		code, out, errOut := do(args, "")
		if c.code == exitError {
			wantErr(t, c.name, code, out, errOut)
			continue
		}
		if code != c.code || strings.Count(out, "\n") != 1 || errOut != "" {
			t.Errorf("%s: code=%d out=%q err=%q", c.name, code, out, errOut)
		}
	}
}
