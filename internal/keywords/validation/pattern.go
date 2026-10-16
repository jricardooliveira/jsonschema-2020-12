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
