package compile

import (
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/perfdata"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/registry"
)

// buildBuiltinForBench builds the built-in meta-schema index and compiles the
// 2020-12 meta-schema from scratch, as a one-shot run pays it. Story 202 T2
// replaces its body with a call to the unexported buildBuiltin.
func buildBuiltinForBench() (*registry.Index, error) {
	set, err := buildBuiltin()
	if err != nil {
		return nil, err
	}
	return set.idx, nil
}

// BenchmarkFirstLoadOrder times what a one-shot run pays: build the built-in
// meta-schemas from scratch, check the order schema against the 2020-12
// meta-schema and compile it.
func BenchmarkFirstLoadOrder(b *testing.B) {
	src, err := perfdata.Fixture("order_schema.json")
	if err != nil {
		b.Fatal(err)
	}
	doc, err := jsonvalue.DecodeLimited(src, jsonvalue.NumberLimits{})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx, err := buildBuiltinForBench()
		if err != nil {
			b.Fatal(err)
		}
		if _, _, err := compileDocumentOpts(doc, StdinBaseURI, idx, 0, false, false); err != nil {
			b.Fatal(err)
		}
	}
}
