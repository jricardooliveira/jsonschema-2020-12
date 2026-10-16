package jsonschema

import (
	"errors"
	"testing"

	"github.com/neoface/jsonschema/internal/keywords"
)

var (
	errRootCompile = errors.New("root test compile failure")
	errRootEval    = errors.New("root test evaluate failure")
)

// Registered once per process (init), so -count=N is safe.
func init() {
	keywords.Register(keywords.Keyword{
		Name: "x-root-test-bool", Vocabulary: "urn:test",
		Compile:  func(_ *keywords.CompileContext, v any) (any, error) { return v, nil },
		Evaluate: func(_ *keywords.EvalContext, c any, _ any) (bool, error) { return c.(bool), nil },
	})
	keywords.Register(keywords.Keyword{
		Name: "x-root-test-compile-err", Vocabulary: "urn:test",
		Compile:  func(*keywords.CompileContext, any) (any, error) { return nil, errRootCompile },
		Evaluate: func(*keywords.EvalContext, any, any) (bool, error) { return true, nil },
	})
	keywords.Register(keywords.Keyword{
		Name: "x-root-test-eval-err", Vocabulary: "urn:test",
		Compile:  func(*keywords.CompileContext, any) (any, error) { return nil, nil },
		Evaluate: func(*keywords.EvalContext, any, any) (bool, error) { return true, errRootEval },
	})
}

func TestValidateReachesRegisteredKeyword(t *testing.T) {
	cases := map[string]bool{`{"x-root-test-bool":true}`: true, `{"x-root-test-bool":false}`: false}
	for sch, want := range cases {
		s, err := Compile([]byte(sch))
		if err != nil {
			t.Fatal(sch, err)
		}
		r, err := s.Validate([]byte(`{}`))
		if err != nil || r.Valid() != want {
			t.Errorf("%s: got %v %v, want %v", sch, r.Valid(), err, want)
		}
	}
}

func TestKeywordCompileErrorIsError(t *testing.T) {
	s, err := Compile([]byte(`{"x-root-test-compile-err":1}`))
	if s != nil || !errors.Is(err, errRootCompile) {
		t.Fatalf("got %v %v", s, err)
	}
}

func TestKeywordEvaluateErrorIsError(t *testing.T) {
	s, err := Compile([]byte(`{"x-root-test-eval-err":1,"x-root-test-bool":true}`))
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.Validate([]byte(`{}`))
	if r.Valid() || !errors.Is(err, errRootEval) {
		t.Fatalf("got %v %v", r.Valid(), err)
	}
}
