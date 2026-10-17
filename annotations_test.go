package jsonschema

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
)

func TestAnnotationsPassingOrdered(t *testing.T) {
	s, err := Compile([]byte(`{"properties":{"b":{"title":"B"},"a":{"title":"A"}},"description":"n"}`))
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.Validate([]byte(`{"a":1,"b":2}`))
	if err != nil || !r.Valid() {
		t.Fatal(r, err)
	}
	got := r.Annotations()
	if len(got) < 3 || got[0].InstanceLocation != "" {
		t.Fatalf("got %+v", got)
	}
	for i := 1; i < len(got); i++ {
		if got[i-1].InstanceLocation > got[i].InstanceLocation {
			t.Fatalf("unordered: %+v", got)
		}
	}
	got[0].Keyword = "mutated"
	if r.Annotations()[0].Keyword == "mutated" {
		t.Fatal("accessor must return a copy")
	}
}

func TestAnnotationsEmptyOnFailure(t *testing.T) {
	s, _ := Compile([]byte(`{"title":"t","type":"string"}`))
	r, err := s.Validate([]byte(`1`))
	if err != nil || r.Valid() || len(r.Annotations()) != 0 {
		t.Fatal(r, err)
	}
}

// plain turns schema values (exact numbers) into encoding/json values.
func plain(v any) any {
	switch x := v.(type) {
	case *jsonvalue.Number:
		return json.Number(x.String())
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = plain(e)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, e := range x {
			out[k] = plain(e)
		}
		return out
	}
	return v
}

// TestAnnotationSpecCases runs testdata/annotations: each test with an
// "annotations" field must produce exactly that ordered list (empty when invalid).
func TestAnnotationSpecCases(t *testing.T) {
	files, _ := filepath.Glob("testdata/annotations/*.json")
	if len(files) == 0 {
		t.Fatal("no annotation cases")
	}
	type want struct {
		InstanceLocation string `json:"instanceLocation"`
		KeywordLocation  string `json:"keywordLocation"`
		Keyword          string `json:"keyword"`
		Value            any    `json:"value"`
	}
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		var groups []struct {
			Spec, Description string
			Schema            json.RawMessage
			Tests             []struct {
				Description string
				Data        json.RawMessage
				Valid       bool
				Annotations *[]want
			}
		}
		if err := json.Unmarshal(raw, &groups); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		for _, g := range groups {
			if g.Spec == "" {
				t.Errorf("%s: %q cites no spec section", f, g.Description)
			}
			s, err := Compile(g.Schema)
			if err != nil {
				t.Errorf("%s: %s: %v", f, g.Description, err)
				continue
			}
			for _, c := range g.Tests {
				if c.Annotations == nil {
					continue
				}
				t.Run(filepath.Base(f)+"/"+g.Description+"/"+c.Description, func(t *testing.T) {
					r, err := s.Validate(c.Data)
					if err != nil || r.Valid() != c.Valid {
						t.Fatalf("valid=%v err=%v want %v (%s)", r.Valid(), err, c.Valid, g.Spec)
					}
					got := []want{}
					for _, a := range r.Annotations() {
						got = append(got, want{a.InstanceLocation, a.KeywordLocation, a.Keyword, plain(a.Value)})
					}
					// compare through JSON so number types match
					gb, _ := json.Marshal(got)
					wb, _ := json.Marshal(*c.Annotations)
					var gv, wv any
					json.Unmarshal(gb, &gv)
					json.Unmarshal(wb, &wv)
					if !reflect.DeepEqual(gv, wv) {
						t.Errorf("got %s\nwant %s (%s)", gb, wb, g.Spec)
					}
				})
			}
		}
	}
}

func TestAnnotationLimit(t *testing.T) {
	schema := []byte(`{"title":"t","description":"d"}`)
	// The cap also covers the load-time meta-schema check (core 8.1.1), so
	// it must leave room for that; the document then goes over it.
	s, err := CompileWithOptions([]byte(`{"items":{"title":"t"}}`), Options{MaxAnnotations: 100})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.Validate([]byte(`[` + strings.Repeat(`1,`, 200) + `1]`))
	var le *jsonvalue.LimitError
	if !errors.As(err, &le) || le.Limit != "annotations" {
		t.Fatalf("want annotations LimitError, got %v", err)
	}
	if _, err := CompileWithOptions(schema, Options{MaxAnnotations: -1}); err == nil {
		t.Fatal("negative must be an error")
	}
	s, _ = CompileWithOptions(schema, Options{})
	if r, err := s.Validate([]byte(`1`)); err != nil || len(r.Annotations()) != 2 || DefaultMaxAnnotations <= 0 {
		t.Fatal(r, err)
	}
}
