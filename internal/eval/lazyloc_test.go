package eval

import (
	"testing"

	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/jsonvalue"
	_ "github.com/neoface/jsonschema/internal/keywords/all"
)

// locs is one unit's keyword location and absolute keyword location.
type locs struct{ kw, abs string }

// TestLazyLocEqualsEager runs the real evaluator (evalNode drives lazyLoc)
// and compares every unit's rendered locations with the strings the old
// eager builder produced for the same schema, written out literally per core
// 12.3.1 (path along the evaluation, a $ref target starting at the $ref) and
// 12.3.2 (absolute = resource URI + '#' + percent-encoded pointer, empty
// before any reference is crossed). No production location code is reused
// to build the reference.
func TestLazyLocEqualsEager(t *testing.T) {
	const base = "file:///s.json"
	cases := map[string]struct {
		schema, doc string
		want        []locs
	}{
		"through $ref": {
			`{"$ref":"#/$defs/x","$defs":{"x":{"properties":{"p":{"type":"string"}}}}}`, `{"p":1}`,
			[]locs{
				{"/$ref/properties/p/type", base + "#/$defs/x/properties/p/type"},
				{"/$ref", ""},
			},
		},
		"through $dynamicRef": {
			`{"$defs":{"d":{"$dynamicAnchor":"d","minimum":5}},"properties":{"n":{"$dynamicRef":"#d"}}}`, `{"n":1}`,
			[]locs{
				{"/properties/n/$dynamicRef/minimum", base + "#/$defs/d/minimum"},
				{"/properties/n/$dynamicRef", ""},
			},
		},
		"inside then": {
			`{"if":{"properties":{"a":{"type":"string"}}},"then":{"required":["b"]},"else":{"type":"string"}}`, `{"a":"x"}`,
			[]locs{{"/then/required", ""}, {"/then", ""}},
		},
		"inside else": {
			`{"if":{"properties":{"a":{"type":"string"}}},"then":{"required":["b"]},"else":{"type":"string"}}`, `{"a":1}`,
			[]locs{{"/else/type", ""}, {"/else", ""}},
		},
		"escaping and percent-encoding": {
			`{"$ref":"#/$defs/a~1b~0c%20d","$defs":{"a/b~c d":{"minLength":3}}}`, `"x"`,
			[]locs{{"/$ref/minLength", base + "#/$defs/a~1b~0c%20d/minLength"}, {"/$ref", ""}},
		},
		"empty absolute before any reference": {
			`{"properties":{"a":{"type":"string"}}}`, `{"a":1}`,
			[]locs{{"/properties/a/type", ""}},
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			s, err := jsonvalue.DecodeLimited([]byte(c.schema), jsonvalue.NumberLimits{})
			if err != nil {
				t.Fatal(err)
			}
			d, err := jsonvalue.DecodeLimited([]byte(c.doc), jsonvalue.NumberLimits{})
			if err != nil {
				t.Fatal(err)
			}
			n, err := compile.CompileDocument(s, base)
			if err != nil {
				t.Fatal(err)
			}
			ev, err := EvaluateLazy(n, d, DefaultLimits())
			if err != nil {
				t.Fatal(err)
			}
			var got []locs
			for _, u := range ev.Units {
				got = append(got, locs{u.KeywordLocation, u.AbsoluteKeywordLocation})
			}
			if ev.Valid || len(got) != len(c.want) {
				t.Fatalf("valid %v, units %q, want %q", ev.Valid, got, c.want)
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Errorf("unit %d: got %q, want %q", i, got[i], c.want[i])
				}
			}
		})
	}
}
