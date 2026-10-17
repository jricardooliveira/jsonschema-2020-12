package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/jricardooliveira/jsonschema-2020-12"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/output"
)

// Exit codes: the CLI contract.
const (
	exitValid   = 0
	exitInvalid = 1
	exitError   = 2
)

// readCapped reads all of r, failing with a limit error when it holds more
// than max bytes.
func readCapped(r io.Reader, name string, max int) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, int64(max)+1))
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: read failed", name)
	}
	if len(data) > max {
		return nil, &jsonschema.LimitError{Limit: jsonschema.LimitInputSize, Value: max}
	}
	return data, nil
}

func readFile(path string, max int) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("cannot read %s: file not found", path)
		}
		return nil, fmt.Errorf("cannot read %s: read failed", path)
	}
	defer f.Close()
	return readCapped(f, path, max)
}

// oneLine keeps stderr to a single line.
func oneLine(s string) string { return strings.NewReplacer("\n", " ", "\r", " ").Replace(s) }

// run is the whole CLI; it returns the process exit code.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fail := func(msg string) int {
		fmt.Fprintln(stderr, oneLine(msg))
		return exitError
	}
	if len(args) == 0 || args[0] != "validate" {
		return fail("unknown subcommand")
	}
	flags := flag.NewFlagSet("validate", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	outShort := flags.String("o", "basic", "")
	outLong := flags.String("output", "", "")
	baseURI := flags.String("base-uri", "", "")
	formatAssert := flags.Bool("format-assertion", false, "")
	lf := registerLimitFlags(flags)
	var refs, refDirs listFlag
	flags.Var(&refs, "ref", "file: "+refHelp)
	flags.Var(&refDirs, "ref-dir", "folder, needs a mapping: "+refHelp)
	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return fail("usage: jsonschema validate [-o flag|basic|detailed|verbose] <schema-file> [<document-file>|-]...")
		}
		if lf.bad != "" {
			return fail("invalid value for flag -" + lf.bad)
		}
		// ponytail: the flag package's message holds only the name here, never a value.
		if name, ok := strings.CutPrefix(err.Error(), "flag provided but not defined: -"); ok {
			return fail("unknown flag -" + name)
		}
		return fail("invalid arguments")
	}
	if bad := lf.notPositive(flags); bad != "" {
		return fail("flag -" + bad + " must be greater than 0")
	}
	shape := *outShort
	if *outLong != "" {
		shape = *outLong
	}
	if _, err := output.Render(shape, true, nil); err != nil {
		return fail(err.Error())
	}
	rest := flags.Args()
	if len(rest) == 0 {
		return fail("missing schema argument")
	}
	reg := jsonschema.Registry{MaxFileBytes: *lf.vals["max-ref-file-bytes"], MaxTotalBytes: *lf.vals["max-ref-total-bytes"]}
	for _, v := range refs {
		u, p, _, err := parseRef(v)
		if err != nil {
			return fail("invalid value for flag -ref: " + err.Error())
		}
		reg.Files = append(reg.Files, jsonschema.RegistryFile{URI: u, Path: p})
	}
	for _, v := range refDirs {
		u, p, mapped, err := parseRef(v)
		if err != nil || !mapped {
			return fail("invalid value for flag -ref-dir: expected <uri>=<folder>")
		}
		reg.Dirs = append(reg.Dirs, jsonschema.RegistryDir{BaseURI: u, Path: p})
	}
	inCap := lf.inputCap()
	schemaData, err := readFile(rest[0], inCap)
	if err != nil {
		return fail(limitLine(err, false))
	}
	opts := lf.options()
	opts.BaseURI, opts.SchemaPath = *baseURI, rest[0]
	opts.FormatAssertion = *formatAssert
	s, err := jsonschema.CompileWithRegistry(schemaData, opts, reg)
	if err != nil {
		return fail(limitLine(err, true))
	}
	docs := rest[1:]
	if len(docs) == 0 {
		docs = []string{"-"}
	}
	// check returns the rendered report and exit code for one document.
	check := func(path string) ([]byte, int, error) {
		var data []byte
		var err error
		if path == "-" {
			data, err = readCapped(stdin, "stdin", inCap)
		} else {
			data, err = readFile(path, inCap)
		}
		if err != nil {
			return nil, exitError, err
		}
		res, err := s.Validate(data)
		if err != nil {
			return nil, exitError, err
		}
		out, err := res.Output(shape)
		if err != nil {
			var le *jsonschema.LimitError
			if errors.As(err, &le) { // limit name and value only, no content
				return nil, exitError, le
			}
			return nil, exitError, errors.New("cannot render result")
		}
		if res.Valid() {
			return out, exitValid, nil
		}
		return out, exitInvalid, nil
	}
	write := func(b []byte) bool {
		_, err := stdout.Write(append(b, '\n'))
		return err == nil
	}
	if len(docs) == 1 {
		out, c, err := check(docs[0])
		if err != nil {
			return fail(limitLine(err, false))
		}
		if !write(out) {
			return fail("cannot write result")
		}
		return c
	}
	code := exitValid
	for _, d := range docs {
		out, c, err := check(d)
		if c > code {
			code = c
		}
		var line []byte
		var merr error
		if err != nil {
			line, merr = json.Marshal(struct {
				Document string `json:"document"`
				Error    string `json:"error"`
			}{d, oneLine(limitLine(err, false))})
		} else {
			line, merr = json.Marshal(struct {
				Document string          `json:"document"`
				Output   json.RawMessage `json:"output"`
			}{d, out})
		}
		if merr != nil || !write(line) {
			return fail("cannot write result")
		}
	}
	return code
}
