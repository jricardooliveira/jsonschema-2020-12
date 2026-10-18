package validation

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/ecmaregex"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "pattern", Vocabulary: VocabularyURI, Compile: compilePattern, Evaluate: evalPattern})
}

// DefaultMaxPatternSize is the longest accepted pattern, counted in bytes of
// the schema's UTF-8 text. It is checked before translation and compilation.
// ponytail: fixed constant; story 193 turns it into a library option and CLI flag.
const DefaultMaxPatternSize = ecmaregex.DefaultMaxPatternSize

// compiledPattern is immutable: the source text and its compiled RE2 form.
type compiledPattern struct {
	src string
	re  *regexp.Regexp
}

func compilePattern(ctx *keywords.CompileContext, value any) (any, error) {
	src, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("pattern: must be a string")
	}
	re, err := ecmaregex.CompileBounded(src, ctx.MaxPatternBytes)
	if err != nil {
		var le *jsonvalue.LimitError
		if errors.As(err, &le) {
			return nil, err
		}
		return nil, fmt.Errorf("pattern %q: %w", src, err)
	}
	return &compiledPattern{src: src, re: re}, nil
}

func evalPattern(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	s, ok := instance.(string)
	p := compiled.(*compiledPattern)
	if !ok {
		return true, nil
	}
	if err := ctx.ChargeBytes(len(s)); err != nil { // RE2 matching is linear in s
		return false, err
	}
	if p.re.MatchString(s) {
		return true, nil
	}
	return false, ctx.Report(fmt.Sprintf("does not match pattern %s", p.src))
}
