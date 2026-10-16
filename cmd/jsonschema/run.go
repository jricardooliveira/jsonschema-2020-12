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
	digits := intFlag("max-number-digits")
	exponent := intFlag("max-number-exponent")
	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return fail("usage: jsonschema validate <schema-file> [<document-file>|-]")
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
	rest := flags.Args()
	if len(rest) == 0 {
		return fail("missing schema argument")
	}
	if len(rest) > 2 {
		return fail("only one document is supported for now")
	}
	schemaData, err := readFile(rest[0])
	if err != nil {
		return fail(err.Error())
	}
	var docData []byte
	if len(rest) == 1 || rest[1] == "-" {
		docData, err = readCapped(stdin, "stdin")
	} else {
		docData, err = readFile(rest[1])
	}
	if err != nil {
		return fail(err.Error())
	}
	s, err := jsonschema.CompileWithOptions(schemaData, jsonschema.Options{MaxNumberDigits: *digits, MaxNumberExponent: *exponent})
	if err != nil {
		return fail(err.Error())
	}
	res, err := s.Validate(docData)
	if err != nil {
		return fail(err.Error())
	}
	out, err := json.Marshal(map[string]bool{"valid": res.Valid()})
	if err != nil {
		return fail("cannot write result")
	}
	if _, err := stdout.Write(append(out, '\n')); err != nil {
		return fail("cannot write result")
	}
	if res.Valid() {
		return exitValid
	}
	return exitInvalid
}
