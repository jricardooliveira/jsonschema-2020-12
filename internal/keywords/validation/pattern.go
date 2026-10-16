package validation

import (
	"fmt"
	"regexp"

	"github.com/neoface/jsonschema/internal/ecmaregex"
	"github.com/neoface/jsonschema/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "pattern", Vocabulary: VocabularyURI, Compile: compilePattern, Evaluate: evalPattern})
}

// DefaultMaxPatternSize is the longest accepted pattern, counted in bytes of
// the schema's UTF-8 text. It is checked before translation and compilation.
// ponytail: fixed constant; story 193 turns it into a library option and CLI flag.
const DefaultMaxPatternSize = 64 << 10

// compiledPattern is immutable: the source text and its compiled RE2 form.
type compiledPattern struct {
	src string
	re  *regexp.Regexp
}

func compilePattern(_ *keywords.CompileContext, value any) (any, error) {
	src, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("pattern: must be a string")
	}
	if len(src) > DefaultMaxPatternSize {
		return nil, fmt.Errorf("pattern: exceeds size limit of %d bytes", DefaultMaxPatternSize)
	}
	re2, err := ecmaregex.Translate(src)
	if err != nil {
		return nil, fmt.Errorf("pattern %q: %w", src, err)
	}
	re, err := regexp.Compile(re2)
	if err != nil {
		return nil, fmt.Errorf("pattern %q: cannot be compiled: %w", src, err)
	}
	return &compiledPattern{src: src, re: re}, nil
}

func evalPattern(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	s, ok := instance.(string)
	p := compiled.(*compiledPattern)
	if !ok || p.re.MatchString(s) {
		return true, nil
	}
	return false, ctx.Report(fmt.Sprintf("does not match pattern %s", p.src))
}
