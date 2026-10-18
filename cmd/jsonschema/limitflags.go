package main

import (
	"errors"
	"flag"
	"strconv"

	"github.com/jricardooliveira/jsonschema-2020-12"
)

// maxInputBytes is the default per-input cap (schema, document, stdin).
const maxInputBytes = jsonschema.DefaultMaxInputBytes

// limitFlagNames lists every limit flag. An explicit value of 0 or less is a
// usage error: a limit can be raised or lowered, never switched off.
var limitFlagNames = []string{
	"max-depth", "max-input-bytes", "max-ref-chain", "max-steps", "max-pattern-size",
	"max-output-units", "max-report-bytes", "max-number-digits", "max-number-exponent",
	"max-annotations", "max-ref-file-bytes", "max-ref-total-bytes",
}

// limitFlag maps each limit to the flag that raises it. Input size is
// handled in limitLine, since it depends on where the read happened.
var limitFlag = map[jsonschema.Limit]string{
	jsonschema.LimitDepth:          "max-depth",
	jsonschema.LimitInputSize:      "max-input-bytes",
	jsonschema.LimitRegistrySize:   "max-ref-total-bytes",
	jsonschema.LimitRefChain:       "max-ref-chain",
	jsonschema.LimitSteps:          "max-steps",
	jsonschema.LimitNumberDigits:   "max-number-digits",
	jsonschema.LimitNumberExponent: "max-number-exponent",
	jsonschema.LimitPatternSize:    "max-pattern-size",
	jsonschema.LimitOutputUnits:    "max-output-units",
	jsonschema.LimitAnnotations:    "max-annotations",
	jsonschema.LimitReportBytes:    "max-report-bytes",
}

// limitFlags holds the parsed limit flag values.
type limitFlags struct {
	vals map[string]*int
	bad  string // name of the flag whose value failed to parse
}

// registerLimitFlags adds every limit flag to fs.
func registerLimitFlags(fs *flag.FlagSet) *limitFlags {
	lf := &limitFlags{vals: map[string]*int{}}
	for _, name := range limitFlagNames {
		p := new(int)
		lf.vals[name] = p
		fs.Func(name, "", func(s string) error {
			n, err := strconv.ParseInt(s, 0, strconv.IntSize)
			if err != nil {
				lf.bad = name
				return err
			}
			*p = int(n)
			return nil
		})
	}
	return lf
}

// notPositive returns the name of a limit flag given an explicit value <= 0.
func (lf *limitFlags) notPositive(fs *flag.FlagSet) string {
	bad := ""
	fs.Visit(func(f *flag.Flag) {
		if p, ok := lf.vals[f.Name]; ok && *p <= 0 {
			bad = f.Name
		}
	})
	return bad
}

// inputCap is the per-input byte cap the CLI applies to its own reads.
func (lf *limitFlags) inputCap() int {
	if n := *lf.vals["max-input-bytes"]; n > 0 {
		return n
	}
	return maxInputBytes
}

func (lf *limitFlags) options() jsonschema.Options {
	v := func(n string) int { return *lf.vals[n] }
	return jsonschema.Options{
		MaxDepth: v("max-depth"), MaxInputBytes: v("max-input-bytes"), MaxRefChain: v("max-ref-chain"),
		MaxSteps: v("max-steps"), MaxPatternSize: v("max-pattern-size"), MaxOutputUnits: v("max-output-units"),
		MaxReportBytes: v("max-report-bytes"), MaxNumberDigits: v("max-number-digits"),
		MaxNumberExponent: v("max-number-exponent"), MaxAnnotations: v("max-annotations"),
	}
}

// limitLine is the stderr text for err. A limit error becomes
// "limit exceeded: <limit> <value> (raise with --<flag>)". duringCompile
// says the failing read happened inside compilation, where the only input
// the library reads is a registry file (the CLI reads the schema itself).
func limitLine(err error, duringCompile bool) string {
	var le *jsonschema.LimitError
	if !errors.As(err, &le) {
		return err.Error()
	}
	name := limitFlag[le.Limit]
	if le.Limit == jsonschema.LimitInputSize && duringCompile {
		name = "max-ref-file-bytes"
	}
	if name == "" {
		return le.Error()
	}
	return le.Error() + " (raise with --" + name + ")"
}
