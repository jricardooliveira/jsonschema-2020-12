package main

import (
	"bytes"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// docExample is one CLI example: a ```sh command, the ```json block that
// follows it and the "Exit code: N" line after that.
type docExample struct {
	file, cmd, out string
	code           int
}

var exitLine = regexp.MustCompile(`^Exit code: (\d+)$`)

// parseExamples extracts the examples from one markdown page. A sh block that
// starts with "jsonschema validate" must be followed by its json block and
// exit code line, or the page is malformed.
func parseExamples(t *testing.T, file, text string) []docExample {
	t.Helper()
	lines := strings.Split(text, "\n")
	var out []docExample
	block := func(i int) ([]string, int) { // lines of the fence opened at i
		var body []string
		for i++; i < len(lines) && lines[i] != "```"; i++ {
			body = append(body, lines[i])
		}
		return body, i
	}
	next := func(i int) int { // next non-blank line after i
		for i++; i < len(lines) && strings.TrimSpace(lines[i]) == ""; i++ {
		}
		return i
	}
	for i := 0; i < len(lines); i++ {
		if lines[i] != "```sh" {
			continue
		}
		body, end := block(i)
		i = end
		if len(body) == 0 || !strings.HasPrefix(body[0], "jsonschema validate") {
			continue
		}
		j := next(i)
		if j >= len(lines) || lines[j] != "```json" {
			t.Fatalf("%s: example %q has no json block", file, body[0])
		}
		got, end := block(j)
		k := next(end)
		if k >= len(lines) || !exitLine.MatchString(lines[k]) {
			t.Fatalf("%s: example %q has no exit code line", file, body[0])
		}
		code, _ := strconv.Atoi(exitLine.FindStringSubmatch(lines[k])[1])
		out = append(out, docExample{file, strings.Join(body, " "), strings.Join(got, "\n"), code})
		i = k
	}
	return out
}

func TestDocsCLIExamples(t *testing.T) {
	t.Chdir("../../") // examples use paths relative to the repository root
	files, err := filepath.Glob("docs/*.md")
	if err != nil {
		t.Fatal(err)
	}
	files = append(files, "README.md")
	n := 0
	for _, f := range files {
		data, err := readFile(f, 1<<24)
		if err != nil {
			t.Fatal(err)
		}
		for _, ex := range parseExamples(t, f, string(data)) {
			n++
			t.Run(ex.file+"/"+ex.cmd, func(t *testing.T) {
				args := strings.Fields(ex.cmd)[1:] // drop "jsonschema"
				var stdout, stderr bytes.Buffer
				code := run(args, strings.NewReader(""), &stdout, &stderr)
				if code != ex.code {
					t.Errorf("exit code %d, page shows %d (stderr %q)", code, ex.code, stderr.String())
				}
				if got := strings.TrimRight(stdout.String(), "\n"); got != ex.out {
					t.Errorf("stdout differs from page\n got: %s\nwant: %s", got, ex.out)
				}
			})
		}
	}
	if n == 0 {
		t.Fatal("no CLI examples found in the docs")
	}
}
