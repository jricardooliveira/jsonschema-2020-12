package formats_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/neoface/jsonschema"
	"github.com/neoface/jsonschema/internal/keywords"
)

// assertionSamples holds one value per format that breaks it.
var assertionSamples = map[string]string{
	"uri":                   `"/a b"`,
	"uri-reference":         `"a b"`,
	"iri":                   `"/ü "`,
	"iri-reference":         `"/%zz"`,
	"uri-template":          `"{"`,
	"json-pointer":          `"a"`,
	"relative-json-pointer": `"01"`,
	"regex":                 `"("`,
}

// validation 7.2.1/7.2.2: outside assertion mode a format is an annotation
// only; in assertion mode a broken value fails.
func TestFormatsAnnotationOnlyUnlessAsserted(t *testing.T) {
	for f, doc := range assertionSamples {
		for _, on := range []bool{false, true} {
			s, err := jsonschema.CompileWithOptions([]byte(`{"format":"`+f+`"}`), jsonschema.Options{FormatAssertion: on})
			if err != nil {
				t.Fatal(err)
			}
			r, err := s.Validate([]byte(doc))
			if err != nil {
				t.Fatal(err)
			}
			if r.Valid() == on {
				t.Errorf("format %s, assertion %v: valid = %v", f, on, r.Valid())
			}
		}
	}
}

// ADR-016 H: each check is charged per byte against the step budget.
func TestFormatsChargeSteps(t *testing.T) {
	doc := []byte(`"` + strings.Repeat("a", 1000*keywords.BytesPerStep) + `"`)
	for f := range assertionSamples {
		for _, on := range []bool{false, true} {
			s, err := jsonschema.CompileWithOptions([]byte(`{"format":"`+f+`"}`), jsonschema.Options{FormatAssertion: on, MaxSteps: 100})
			if err != nil {
				t.Fatal(err)
			}
			_, err = s.Validate(doc)
			var le *jsonschema.LimitError
			if hit := errors.As(err, &le) && le.Limit == jsonschema.LimitSteps; hit != on {
				t.Errorf("format %s, assertion %v: err %v", f, on, err)
			}
		}
	}
}
