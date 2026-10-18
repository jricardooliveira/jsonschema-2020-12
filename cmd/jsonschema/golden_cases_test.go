package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/perfdata"
)

// Golden CLI cases: shared by the capture (golden_capture_test.go) and the
// comparison test. A case is one CLI process: arguments, optional stdin and
// optional files written fresh to a temp folder. Cases with files run from
// that folder, all others from the project folder; paths are always relative,
// so the output is the same on every machine.

const (
	goldenRoot       = "../.."
	goldenLargeBytes = 64 * 1024 // outputs over this are stored as size + SHA-256
	goldenBaseURI    = "https://golden.example/schema.json"
	goldenProvenance = "captured from base d7198b7, production code identical to v1.1 f35d9e4"
)

var goldenShapes = []string{"flag", "basic", "detailed", "verbose"}

// goldenCase is one CLI call before the output shape is applied.
type goldenCase struct {
	ID    string
	Args  []string          // after "validate -o <shape>"
	Files map[string]string // when non-nil: written to a fresh folder, which is the working directory
	Stdin string            // "", "file:<project path>" or "gen:<name>"
}

// goldenOut is one captured stream: full text, or size plus SHA-256.
type goldenOut struct {
	Text   *string `json:"text,omitempty"`
	Size   int     `json:"size"`
	SHA256 string  `json:"sha256,omitempty"`
}

// goldenRecord is one golden entry.
type goldenRecord struct {
	ID          string    `json:"id"`
	Shape       string    `json:"shape"`
	Args        []string  `json:"args"`
	Stdin       string    `json:"stdin,omitempty"`
	StdinSize   int       `json:"stdinSize,omitempty"`
	StdinSHA256 string    `json:"stdinSha256,omitempty"`
	Exit        int       `json:"exit"`
	Stdout      goldenOut `json:"stdout"`
	Stderr      goldenOut `json:"stderr"`
}

// goldenFile is the layout of testdata/golden/*.json.
type goldenFile struct {
	Header struct {
		Provenance string `json:"provenance"`
		GoVersion  string `json:"goVersion"`
		Cases      int    `json:"cases"`
	} `json:"header"`
	Cases []goldenRecord `json:"cases"`
}

func sum(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

// goldenStream makes the golden form of one output stream.
func goldenStream(b []byte) goldenOut {
	if len(b) > goldenLargeBytes || !utf8.Valid(b) {
		return goldenOut{Size: len(b), SHA256: sum(b)}
	}
	s := string(b)
	return goldenOut{Text: &s, Size: len(b)}
}

// isLarge: generated stdin or a hashed stream; these go to the large file.
func (r goldenRecord) isLarge() bool {
	return strings.HasPrefix(r.Stdin, "gen:") || r.Stdout.Text == nil || r.Stderr.Text == nil
}

// goldenStdin returns the stdin bytes of a case.
func goldenStdin(spec string) ([]byte, error) {
	switch {
	case spec == "":
		return nil, nil
	case strings.HasPrefix(spec, "file:"):
		return os.ReadFile(filepath.Join(goldenRoot, strings.TrimPrefix(spec, "file:")))
	case spec == "gen:valid-0.5MB":
		return perfdata.ValidOrder(perfdata.MediumItems), nil
	case spec == "gen:valid-5.5MB":
		return perfdata.ValidOrder(perfdata.LargeItems), nil
	case spec == "gen:errors-0.33MB":
		return perfdata.ErrorOrder(), nil
	}
	return nil, fmt.Errorf("unknown stdin source %q", spec)
}

// goldenBuild builds the real CLI into dir and returns its path.
func goldenBuild(dir string) (string, error) {
	bin := filepath.Join(dir, "jsonschema")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/jsonschema")
	cmd.Dir = goldenRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("build: %v: %s", err, out)
	}
	return bin, nil
}

// goldenRun runs one case in one shape as a separate process.
func goldenRun(bin string, c goldenCase, shape string) (goldenRecord, error) {
	// The CLI turns a schema path into an absolute file:// base URI, which shows
	// in absoluteKeywordLocation and differs per machine; a fixed --base-uri
	// keeps every output machine-independent.
	args := append([]string{"--base-uri", goldenBaseURI}, c.Args...)
	rec := goldenRecord{ID: c.ID + "|" + shape, Shape: shape, Args: args, Stdin: c.Stdin}
	in, err := goldenStdin(c.Stdin)
	if err != nil {
		return rec, err
	}
	if in != nil {
		rec.StdinSize, rec.StdinSHA256 = len(in), sum(in)
	}
	absBin, err := filepath.Abs(bin)
	if err != nil {
		return rec, err
	}
	dir, err := filepath.Abs(goldenRoot)
	if err != nil {
		return rec, err
	}
	if c.Files != nil {
		tmp, err := os.MkdirTemp("", "golden-case-")
		if err != nil {
			return rec, err
		}
		defer os.RemoveAll(tmp)
		for name, content := range c.Files {
			p := filepath.Join(tmp, name)
			if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
				return rec, err
			}
			if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
				return rec, err
			}
		}
		dir = tmp
	}
	cmd := exec.Command(absBin, append([]string{"validate", "-o", shape}, args...)...)
	cmd.Dir = dir
	cmd.Stdin = bytes.NewReader(in)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err = cmd.Run()
	var ee *exec.ExitError
	switch {
	case err == nil:
	case errors.As(err, &ee):
		rec.Exit = ee.ExitCode()
	default:
		return rec, err
	}
	// Machine-dependent text (absolute paths) would make the golden fail elsewhere.
	if tmp := os.TempDir(); bytes.Contains(stdout.Bytes(), []byte(dir)) || bytes.Contains(stderr.Bytes(), []byte(dir)) ||
		bytes.Contains(stdout.Bytes(), []byte(tmp)) || bytes.Contains(stderr.Bytes(), []byte(tmp)) {
		return rec, errors.New("output holds an absolute path of this machine")
	}
	rec.Stdout, rec.Stderr = goldenStream(stdout.Bytes()), goldenStream(stderr.Bytes())
	return rec, nil
}

func readRoot(rel string) (string, error) {
	b, err := os.ReadFile(filepath.Join(goldenRoot, rel))
	return string(b), err
}

// goldenCases enumerates every case (shape not yet applied).
func goldenCases() ([]goldenCase, error) {
	var cases []goldenCase
	add := func(c goldenCase) { cases = append(cases, c) }
	const (
		perfSchema = "testdata/perf/order_schema.json"
		perfSmall  = "testdata/perf/order_small.json"
		ex         = "docs/examples/"
	)
	// perf fixtures, docs/examples, stdin and several documents
	add(goldenCase{ID: "perf/small", Args: []string{perfSchema, perfSmall}})
	add(goldenCase{ID: "perf/small-stdin", Args: []string{perfSchema}, Stdin: "file:" + perfSmall})
	add(goldenCase{ID: "perf/small-dash-twice", Args: []string{perfSchema, perfSmall, "-"}, Stdin: "file:" + perfSmall})
	add(goldenCase{ID: "perf/schema-as-doc", Args: []string{perfSchema, perfSchema}})
	add(goldenCase{ID: "perf/gen-0.5MB-stdin", Args: []string{perfSchema}, Stdin: "gen:valid-0.5MB"})
	add(goldenCase{ID: "perf/gen-errors-stdin", Args: []string{perfSchema}, Stdin: "gen:errors-0.33MB"})
	add(goldenCase{ID: "perf/gen-5.5MB-stdin", Args: []string{perfSchema}, Stdin: "gen:valid-5.5MB"})
	add(goldenCase{ID: "examples/valid", Args: []string{ex + "schema.json", ex + "valid.json"}})
	add(goldenCase{ID: "examples/invalid", Args: []string{ex + "schema.json", ex + "invalid.json"}})
	add(goldenCase{ID: "examples/bad-schema", Args: []string{ex + "bad-schema.json", ex + "valid.json"}})
	add(goldenCase{ID: "examples/multi", Args: []string{ex + "schema.json", ex + "valid.json", ex + "invalid.json", ex + "valid.json"}})
	add(goldenCase{ID: "examples/multi-stdin", Args: []string{ex + "schema.json", ex + "valid.json", "-"}, Stdin: "file:" + ex + "invalid.json"})
	add(goldenCase{ID: "examples/stdin-valid", Args: []string{ex + "schema.json"}, Stdin: "file:" + ex + "valid.json"})
	add(goldenCase{ID: "examples/missing-doc", Args: []string{ex + "schema.json", "no-such.json"}})
	add(goldenCase{ID: "examples/missing-schema", Args: []string{"no-such.json", ex + "valid.json"}})
	add(goldenCase{ID: "examples/not-json-doc", Args: []string{ex + "schema.json", "README.md"}})
	// the registry and metaschema fixture folders
	reg := "testdata/registry/"
	person := "https://example.com/person.json=" + reg + "file/person.json"
	add(goldenCase{ID: "registry/file", Args: []string{"--ref", person, "--ref", reg + "file/self_id.json", reg + "file/main.json", ex + "valid.json"}})
	add(goldenCase{ID: "registry/file-nodoc", Args: []string{"--ref", person, reg + "file/main.json", perfSmall}})
	add(goldenCase{ID: "registry/dir", Args: []string{"--ref-dir", "https://example.com/schemas/=" + reg + "dir/schemas", reg + "dir/main.json", ex + "valid.json"}})
	add(goldenCase{ID: "registry/unresolved", Args: []string{reg + "unresolved/main.json", ex + "valid.json"}})
	add(goldenCase{ID: "registry/conflict", Args: []string{"--ref", reg + "conflict/a.json", "--ref", reg + "conflict/b.json", reg + "file/main.json", ex + "valid.json"}})
	add(goldenCase{ID: "registry/broken", Args: []string{"--ref", reg + "broken/bad.json", reg + "file/main.json", ex + "valid.json"}})
	cli, err := filepath.Glob(filepath.Join(goldenRoot, "testdata/metaschema/cli/*.json"))
	if err != nil {
		return nil, err
	}
	for _, f := range cli {
		rel := "testdata/metaschema/cli/" + filepath.Base(f)
		if filepath.Base(f) == "document.json" {
			continue
		}
		add(goldenCase{ID: "metaschema-cli/" + filepath.Base(f), Args: []string{rel, "testdata/metaschema/cli/document.json"}})
	}

	// spec-derived groups: one schema file and one call per document
	for _, dir := range []string{"annotations", "applicator", "core", "unevaluated", "validation"} {
		files, err := filepath.Glob(filepath.Join(goldenRoot, "testdata", dir, "*.json"))
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			raw, err := os.ReadFile(f)
			if err != nil {
				return nil, err
			}
			var groups []struct {
				Schema json.RawMessage `json:"schema"`
				Tests  []struct {
					Data json.RawMessage `json:"data"`
				} `json:"tests"`
			}
			if err := json.Unmarshal(raw, &groups); err != nil {
				return nil, fmt.Errorf("%s: %v", f, err)
			}
			for gi, g := range groups {
				for ti, t := range g.Tests {
					add(goldenCase{
						ID:    fmt.Sprintf("testdata/%s/%s#g%d.t%d", dir, filepath.Base(f), gi, ti),
						Args:  []string{"schema.json", "doc.json"},
						Files: map[string]string{"schema.json": string(g.Schema), "doc.json": string(t.Data)},
					})
				}
			}
		}
	}

	// metaschema cases: the schema, with its registry files, against {}
	metaFiles, err := filepath.Glob(filepath.Join(goldenRoot, "testdata/metaschema/*.json"))
	if err != nil {
		return nil, err
	}
	for _, f := range metaFiles {
		raw, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		var mcs []struct {
			Name     string          `json:"name"`
			Schema   json.RawMessage `json:"schema"`
			Registry []struct {
				URI    string          `json:"uri"`
				Doc    json.RawMessage `json:"doc"`
				CopyOf string          `json:"copyOf"`
			} `json:"registry"`
		}
		if err := json.Unmarshal(raw, &mcs); err != nil {
			return nil, fmt.Errorf("%s: %v", f, err)
		}
		for i, c := range mcs {
			gc := goldenCase{
				ID:    fmt.Sprintf("testdata/metaschema/%s#%d", filepath.Base(f), i),
				Args:  []string{},
				Files: map[string]string{"doc.json": "{}", "schema.json": string(c.Schema)},
			}
			for _, r := range c.Registry {
				data := string(r.Doc)
				if r.CopyOf != "" {
					if data, err = readRoot(r.CopyOf); err != nil {
						return nil, err
					}
				}
				name := "reg/" + r.URI[strings.LastIndex(r.URI, "/")+1:] + ".json"
				gc.Files[name] = data
				gc.Args = append(gc.Args, "--ref", r.URI+"="+name)
			}
			gc.Args = append(gc.Args, "schema.json", "doc.json")
			add(gc)
		}
	}

	// vocabulary cases: custom meta-schema through --ref, one call per document
	vocFiles, err := filepath.Glob(filepath.Join(goldenRoot, "testdata/vocabulary/*.json"))
	if err != nil {
		return nil, err
	}
	for _, f := range vocFiles {
		raw, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		var vcs []struct {
			Schema          json.RawMessage `json:"schema"`
			Meta            string          `json:"meta"`
			FormatAssertion bool            `json:"formatAssertion"`
			Docs            []struct {
				Doc json.RawMessage `json:"doc"`
			} `json:"docs"`
		}
		if err := json.Unmarshal(raw, &vcs); err != nil {
			return nil, fmt.Errorf("%s: %v", f, err)
		}
		for i, c := range vcs {
			docs := []string{"{}"}
			if len(c.Docs) > 0 {
				docs = docs[:0]
				for _, d := range c.Docs {
					docs = append(docs, string(d.Doc))
				}
			}
			for di, d := range docs {
				gc := goldenCase{
					ID:    fmt.Sprintf("testdata/vocabulary/%s#%d.d%d", filepath.Base(f), i, di),
					Files: map[string]string{"schema.json": string(c.Schema), "doc.json": d},
				}
				if c.Meta != "" {
					m, err := readRoot("testdata/vocabulary/metaschemas/" + c.Meta)
					if err != nil {
						return nil, err
					}
					gc.Files["metaschemas/"+c.Meta] = m
					gc.Args = append(gc.Args, "--ref", "./metaschemas/"+c.Meta)
				}
				if c.FormatAssertion {
					gc.Args = append(gc.Args, "--format-assertion")
				}
				gc.Args = append(gc.Args, "schema.json", "doc.json")
				add(gc)
			}
		}
	}
	seen := map[string]bool{}
	for _, c := range cases {
		if seen[c.ID] {
			return nil, errors.New("duplicate golden case id " + strconv.Quote(c.ID))
		}
		seen[c.ID] = true
	}
	return cases, nil
}
