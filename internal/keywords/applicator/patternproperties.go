package applicator

import (
	"fmt"
	"regexp"

	"github.com/neoface/jsonschema/internal/ecmaregex"
	"github.com/neoface/jsonschema/internal/keywords"
	"github.com/neoface/jsonschema/internal/keywords/validation"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "patternProperties", Vocabulary: VocabularyURI, Compile: compilePatternProperties, Evaluate: evalPatternProperties})
}

type patternEntry struct {
	re   *regexp.Regexp
	node any
}

// compileRegex: size check, ECMA-262 translation, RE2 compile. Any failure is
// a schema error. The error text names only the schema's pattern.
// ponytail: duplicates validation/pattern.go's steps; share a helper later.
func compileRegex(keyword, src string) (*regexp.Regexp, error) {
	if len(src) > validation.DefaultMaxPatternSize {
		return nil, fmt.Errorf("%s: pattern exceeds size limit of %d bytes", keyword, validation.DefaultMaxPatternSize)
	}
	re2, err := ecmaregex.Translate(src)
	if err != nil {
		return nil, fmt.Errorf("%s %q: %w", keyword, src, err)
	}
	re, err := regexp.Compile(re2)
	if err != nil {
		return nil, fmt.Errorf("%s %q: cannot be compiled: %w", keyword, src, err)
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
		re, err := compileRegex("patternProperties", k)
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
		if err := step(ctx); err != nil {
			return false, err
		}
		matched := false
		for _, p := range compiled.([]patternEntry) {
			if err := step(ctx); err != nil {
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
