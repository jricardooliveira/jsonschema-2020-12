package validation

import (
	"fmt"

	"github.com/neoface/jsonschema/internal/jsonvalue"
	"github.com/neoface/jsonschema/internal/keywords"
)

// VocabularyURI is the 2020-12 validation vocabulary.
const VocabularyURI = "https://json-schema.org/draft/2020-12/vocab/validation"

var typeNames = map[string]bool{
	"null": true, "boolean": true, "object": true, "array": true,
	"number": true, "string": true, "integer": true,
}

func init() {
	keywords.Register(keywords.Keyword{Name: "type", Vocabulary: VocabularyURI, Compile: compileType, Evaluate: evalType})
}

// compileType accepts a type name or a non-empty array of unique names.
func compileType(_ *keywords.CompileContext, value any) (any, error) {
	var names []string
	switch v := value.(type) {
	case string:
		names = []string{v}
	case []any:
		for _, e := range v {
			s, ok := e.(string)
			if !ok {
				return nil, fmt.Errorf("type: array items must be strings")
			}
			names = append(names, s)
		}
		if len(names) == 0 {
			return nil, fmt.Errorf("type: array must not be empty")
		}
	default:
		return nil, fmt.Errorf("type: must be a string or an array of strings")
	}
	seen := map[string]bool{}
	for _, n := range names {
		if !typeNames[n] {
			return nil, fmt.Errorf("type: unknown type name %q", n)
		}
		if seen[n] {
			return nil, fmt.Errorf("type: duplicate type name %q", n)
		}
		seen[n] = true
	}
	return names, nil
}

func evalType(_ *keywords.EvalContext, compiled any, instance any) (bool, error) {
	for _, n := range compiled.([]string) {
		if matchesType(n, instance) {
			return true, nil
		}
	}
	return false, nil
}

func matchesType(name string, inst any) bool {
	switch name {
	case "null":
		return inst == nil
	case "boolean":
		_, ok := inst.(bool)
		return ok
	case "object":
		_, ok := inst.(map[string]any)
		return ok
	case "array":
		_, ok := inst.([]any)
		return ok
	case "string":
		_, ok := inst.(string)
		return ok
	case "number":
		_, ok := inst.(*jsonvalue.Number)
		return ok
	case "integer":
		n, ok := inst.(*jsonvalue.Number)
		return ok && n.IsInteger()
	}
	return false
}
