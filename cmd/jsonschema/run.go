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
	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return fail("usage: jsonschema validate <schema-file> [<document-file>|-]")
		}
		for _, a := range args[1:] {
			if strings.HasPrefix(a, "-") && a != "-" {
				return fail("unknown flag " + a)
			}
		}
		return fail("invalid arguments")
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
	s, err := jsonschema.Compile(schemaData)
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
