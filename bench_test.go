package jsonschema

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/perfdata"
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

func perfSchema(t *testing.T) *Schema {
	t.Helper()
	src, err := perfdata.Fixture("order_schema.json")
	if err != nil {
		t.Fatal(err)
	}
	s, err := Compile(src)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	return s
}

func TestPerfFixtureValidDocuments(t *testing.T) {
	s := perfSchema(t)
	sizes := []int{perfdata.SmallItems, perfdata.MediumItems}
	if !testing.Short() {
		sizes = append(sizes, perfdata.LargeItems)
	}
	for _, n := range sizes {
		res, err := s.Validate(perfdata.ValidOrder(n))
		if err != nil {
			t.Fatalf("%d items: %v", n, err)
		}
		if !res.Valid() {
			t.Fatalf("%d items: want valid", n)
		}
	}
}

func TestPerfFixtureErrorDocument(t *testing.T) {
	s := perfSchema(t)
	res, err := s.Validate(perfdata.ErrorOrder())
	if err != nil {
		t.Fatal(err)
	}
	if res.Valid() {
		t.Fatal("want invalid")
	}
	out, err := res.Output("basic")
	if err != nil {
		t.Fatal(err)
	}
	var rep struct {
		Errors []struct {
			KeywordLocation string `json:"keywordLocation"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(out, &rep); err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, e := range rep.Errors {
		if strings.HasSuffix(e.KeywordLocation, "/pattern") {
			n++
		}
	}
	if want := perfdata.ErrorItems / perfdata.ErrorEvery; n != want {
		t.Fatalf("pattern errors = %d, want %d", n, want)
	}
}
