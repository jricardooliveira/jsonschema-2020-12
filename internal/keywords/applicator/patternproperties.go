package applicator

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/ecmaregex"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "patternProperties", Vocabulary: VocabularyURI, Compile: compilePatternProperties, Evaluate: evalPatternProperties})
}

type patternEntry struct {
	re   *regexp.Regexp
	node any
}

// compileRegex re-compiles a sibling patternProperties key for
// additionalProperties. patternProperties itself enforces the size cap on the
// same keys in the same schema object, so the size check is skipped here.
// ponytail: pass ctx.MaxPatternBytes from additionalproperties.go when it is next in scope.
func compileRegex(keyword, src string) (*regexp.Regexp, error) {
	return compileRegexMax(keyword, src, len(src))
}

// compileRegexMax: bounded ECMA-262 translation and RE2 compile. Failure is a
// schema error, except an oversize pattern, which is a pattern-size limit error.
func compileRegexMax(keyword, src string, max int) (*regexp.Regexp, error) {
	re, err := ecmaregex.CompileBounded(src, max)
	if err != nil {
		var le *jsonvalue.LimitError
		if errors.As(err, &le) {
			return nil, err
		}
		return nil, fmt.Errorf("%s %q: %w", keyword, src, err)
	}
	return re, nil
}

func compilePatternProperties(ctx *keywords.CompileContext, value any) (any, error) {
	m, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("patternProperties: must be an object")
	}
	out := make([]patternEntry, 0, len(m))
	for _, k := range sortedKeys(m) {
		re, err := compileRegexMax("patternProperties", k, ctx.MaxPatternBytes)
		if err != nil {
			return nil, err
		}
		n, err := ctx.Sub(m[k], "/"+depPointerEscape.Replace(k))
		if err != nil {
			return nil, fmt.Errorf("patternProperties: %w", err)
		}
		out = append(out, patternEntry{re, n})
	}
	return out, nil
}

func evalPatternProperties(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	obj, ok := instance.(map[string]any)
	if !ok {
		return true, nil
	}
	valid := true
	var applied []string
	for _, k := range sortedKeys(obj) {
		if err := ctx.ChargeBytes(len(k)); err != nil {
			return false, err
		}
		matched := false
		for _, p := range compiled.([]patternEntry) {
			if err := ctx.ChargeBytes(len(k)); err != nil { // each RE2 match is linear in k
				return false, err
			}
			if !p.re.MatchString(k) {
				continue
			}
			matched = true
			ok, err := evalMember(ctx, p.node, obj, k)
			if err != nil {
				return false, err
			}
			valid = valid && ok
		}
		if matched {
			applied = append(applied, k)
		}
	}
	return valid, recordProps(ctx, applied, valid)
}
