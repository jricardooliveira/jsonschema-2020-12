package jsonschema

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/neoface/jsonschema/internal/jsonvalue"
)

func TestBrokenInputIsError(t *testing.T) {
	for _, sch := range []string{`{`, `{"a":1,"a":2}`, "\"\xff\"", `"\uD800"`, `{} {}`} {
		if _, err := Compile([]byte(sch)); err == nil {
			t.Errorf("schema %q: want error", sch)
		}
	}
	s, _ := Compile([]byte(`true`))
	for _, d := range []string{`[{"a":1,"a":2}]`, "\"\xff\"", `"\uD800"`, `1 2`} {
		if _, err := s.Validate([]byte(d)); err == nil {
			t.Errorf("doc %q: want error", d)
		}
	}
}

func TestBOMAccepted(t *testing.T) {
	s, err := Compile([]byte("\xEF\xBB\xBFtrue"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Validate([]byte("\xEF\xBB\xBF1")); err != nil {
		t.Fatal(err)
	}
}

func TestNumberCapOption(t *testing.T) {
	over := fmt.Sprintf("1e%d", DefaultMaxNumberExponent+1)
	s, _ := Compile([]byte(`true`))
	_, err := s.Validate([]byte(over))
	var le *jsonvalue.LimitError
	if !errors.As(err, &le) || strings.Contains(err.Error(), over) {
		t.Fatalf("want content-free LimitError, got %v", err)
	}
	if _, err := Compile([]byte(over)); err == nil {
		t.Error("schema over cap must error")
	}
	s, err = CompileWithOptions([]byte(`true`), Options{MaxNumberExponent: DefaultMaxNumberExponent + 1})
	if err != nil {
		t.Fatal(err)
	}
	if r, err := s.Validate([]byte(over)); err != nil || !r.Valid() {
		t.Fatalf("raised cap: %v %v", r.Valid(), err)
	}
	if _, err := CompileWithOptions([]byte(`true`), Options{MaxNumberDigits: -1}); err == nil {
		t.Error("negative option must error")
	}
}
