package jsonschema

import (
	"testing"

	"github.com/neoface/jsonschema/internal/perfdata"
)

func benchSchema(b *testing.B) (*Schema, []byte) {
	b.Helper()
	src, err := perfdata.Fixture("order_schema.json")
	if err != nil {
		b.Fatal(err)
	}
	s, err := Compile(src)
	if err != nil {
		b.Fatal(err)
	}
	return s, src
}

func BenchmarkCompileOrder(b *testing.B) {
	_, src := benchSchema(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Compile(src); err != nil {
			b.Fatal(err)
		}
	}
}

// benchValidate times parse+validate of doc, including the basic report when asked.
func benchValidate(b *testing.B, doc []byte, wantValid, basic bool) {
	s, _ := benchSchema(b)
	b.SetBytes(int64(len(doc)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := s.Validate(doc)
		if err != nil {
			b.Fatal(err)
		}
		if res.Valid() != wantValid {
			b.Fatalf("valid = %v, want %v", res.Valid(), wantValid)
		}
		if basic {
			if _, err := res.Output("basic"); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkValidateSmall(b *testing.B) {
	benchValidate(b, perfdata.ValidOrder(perfdata.SmallItems), true, false)
}

func BenchmarkValidate0_5MB(b *testing.B) {
	benchValidate(b, perfdata.ValidOrder(perfdata.MediumItems), true, false)
}

func BenchmarkValidate5_5MB(b *testing.B) {
	benchValidate(b, perfdata.ValidOrder(perfdata.LargeItems), true, false)
}

func BenchmarkValidateErrors(b *testing.B) {
	benchValidate(b, perfdata.ErrorOrder(), false, false)
}

func BenchmarkValidateErrorsBasic(b *testing.B) {
	benchValidate(b, perfdata.ErrorOrder(), false, true)
}
