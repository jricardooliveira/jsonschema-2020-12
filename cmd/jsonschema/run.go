package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strconv"
	"strings"

	"github.com/neoface/jsonschema"
	"github.com/neoface/jsonschema/internal/output"
)

// Exit codes: the CLI contract.
const (
	exitValid   = 0
	exitInvalid = 1
	exitError   = 2
)

// maxInputBytes caps every input (schema file, document file, stdin).
const maxInputBytes = 64 << 20

// readCapped reads all of r, failing when it holds more than maxInputBytes.
func readCapped(r io.Reader, name string) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxInputBytes+1))
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: read failed", name)
	}
	if len(data) > maxInputBytes {
		return nil, fmt.Errorf("cannot read %s: input exceeds limit %d bytes", name, maxInputBytes)
	}
	return data, nil
}

func readFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("cannot read %s: file not found", path)
		}
		return nil, fmt.Errorf("cannot read %s: read failed", path)
	}
	defer f.Close()
	return readCapped(f, path)
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
	badValue := "" // name of the flag whose value failed to parse
	intFlag := func(name string) *int {
		p := new(int)
		flags.Func(name, "", func(s string) error {
			n, err := strconv.ParseInt(s, 0, strconv.IntSize)
			if err != nil {
				badValue = name
				return err
			}
			*p = int(n)
			return nil
		})
		return p
	}
	outShort := flags.String("o", "basic", "")
	outLong := flags.String("output", "", "")
	baseURI := flags.String("base-uri", "", "")
	digits := intFlag("max-number-digits")
	exponent := intFlag("max-number-exponent")
	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return fail("usage: jsonschema validate [-o flag|basic] <schema-file> [<document-file>|-]...")
		}
		if badValue != "" {
			return fail("invalid value for flag -" + badValue)
		}
		// ponytail: the flag package's message holds only the name here, never a value.
		if name, ok := strings.CutPrefix(err.Error(), "flag provided but not defined: -"); ok {
			return fail("unknown flag -" + name)
		}
		return fail("invalid arguments")
	}
	badFlag := ""
	flags.Visit(func(f *flag.Flag) {
		if (f.Name == "max-number-digits" && *digits <= 0) || (f.Name == "max-number-exponent" && *exponent <= 0) {
			badFlag = f.Name
		}
	})
	if badFlag != "" {
		return fail("flag -" + badFlag + " must be greater than 0")
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
	schemaData, err := readFile(rest[0])
	if err != nil {
		return fail(err.Error())
	}
	s, err := jsonschema.CompileWithOptions(schemaData, jsonschema.Options{MaxNumberDigits: *digits, MaxNumberExponent: *exponent, BaseURI: *baseURI, SchemaPath: rest[0]})
	if err != nil {
		return fail(err.Error())
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
			data, err = readCapped(stdin, "stdin")
		} else {
			data, err = readFile(path)
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
			return fail(err.Error())
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
			}{d, oneLine(err.Error())})
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
