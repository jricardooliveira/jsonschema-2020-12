package applicator

import (
	"fmt"
	"regexp"

	"github.com/neoface/jsonschema/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "additionalProperties", Vocabulary: VocabularyURI, Compile: compileAdditionalProperties, Evaluate: evalAdditionalProperties})
}

// compiledAdditional is immutable; the covered set comes from the sibling
// keywords of the same schema object, read at compile time.
type compiledAdditional struct {
	node     any
	declared map[string]bool
	patterns []*regexp.Regexp
}

func compileAdditionalProperties(ctx *keywords.CompileContext, value any) (any, error) {
	n, err := ctx.Sub(value, "")
	if err != nil {
		return nil, fmt.Errorf("additionalProperties: %w", err)
	}
	c := &compiledAdditional{node: n, declared: map[string]bool{}}
	if p, ok := ctx.Siblings["properties"].(map[string]any); ok {
		for k := range p {
			c.declared[k] = true
		}
	}
	if pp, ok := ctx.Siblings["patternProperties"].(map[string]any); ok {
		for _, k := range sortedKeys(pp) {
			re, err := compileRegex("patternProperties", k)
			if err != nil {
				return nil, err
			}
			c.patterns = append(c.patterns, re)
		}
	}
	return c, nil
}

// covers counts one step per pattern match attempt.
func (c *compiledAdditional) covers(ctx *keywords.EvalContext, name string) (bool, error) {
	if c.declared[name] {
		return true, nil
	}
	for _, re := range c.patterns {
		if err := step(ctx); err != nil {
			return false, err
		}
		if re.MatchString(name) {
			return true, nil
		}
	}
	return false, nil
}

func evalAdditionalProperties(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	obj, ok := instance.(map[string]any)
	if !ok {
		return true, nil
	}
	c := compiled.(*compiledAdditional)
	valid := true
	var applied []string
	for _, k := range sortedKeys(obj) {
		if err := step(ctx); err != nil {
			return false, err
		}
		covered, err := c.covers(ctx, k)
		if err != nil {
			return false, err
		}
		if covered {
			continue
		}
		applied = append(applied, k)
		ok, err := evalMember(ctx, c.node, obj, k)
		if err != nil {
			return false, err
		}
		valid = valid && ok
	}
	return valid, recordProps(ctx, applied, valid)
}
