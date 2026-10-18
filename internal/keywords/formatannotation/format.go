package formatannotation

import (
	"github.com/jricardooliveira/jsonschema-2020-12/internal/formats"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
)

// VocabularyURI is the 2020-12 format-annotation vocabulary.
const VocabularyURI = "https://json-schema.org/draft/2020-12/vocab/format-annotation"

// compiled is the immutable compiled form of a format keyword.
type compiled struct {
	value any               // the schema value, recorded as the annotation
	name  string            // format name; "" when value is not a string
	check func(string) bool // nil unless the mode is on and the format is known
}

// format annotates its value (validation 7.2.1). It fails only in assertion
// mode (validation 7.2.2), for a known format and a string instance; unknown
// formats and other instance types always pass.
func init() {
	keywords.Register(keywords.Keyword{
		Name:       "format",
		Vocabulary: VocabularyURI,
		Compile: func(ctx *keywords.CompileContext, value any) (any, error) {
			c := &compiled{value: value}
			if name, ok := value.(string); ok && ctx.FormatAssert {
				c.name = name
				c.check, _ = formats.Lookup(name)
			}
			return c, nil
		},
		Evaluate: func(ctx *keywords.EvalContext, data any, instance any) (bool, error) {
			c := data.(*compiled)
			if err := ctx.Annotate(c.value); err != nil {
				return false, err
			}
			s, ok := instance.(string)
			if c.check == nil || !ok {
				return true, nil
			}
			if err := ctx.ChargeBytes(len(s)); err != nil {
				return false, err
			}
			if c.check(s) {
				return true, nil
			}
			return false, ctx.Report("value does not match format " + c.name)
		},
	})
}
