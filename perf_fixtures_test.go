package jsonschema

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/neoface/jsonschema/internal/perfdata"
)

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
