package jsonschema

import (
	"errors"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
)

var docs = []string{`null`, `true`, `1`, `"s"`, `[1,2]`, `{"a":{}}`}

func TestValidateSchemas(t *testing.T) {
	cases := map[string]bool{`true`: true, `false`: false, `{}`: true, `{"x-unknown":"string"}`: true, `{"foo":1}`: true}
	for sch, want := range cases {
		s, err := Compile([]byte(sch))
		if err != nil {
			t.Fatal(sch, err)
		}
		for _, d := range docs {
			r, err := s.Validate([]byte(d))
			if err != nil || r.Valid() != want {
				t.Errorf("%s vs %s: %v %v", sch, d, r.Valid(), err)
			}
		}
	}
}

func TestCompileErrors(t *testing.T) {
	for _, sch := range []string{`5`, `"x"`, `null`, `[]`, ``, `{`, `{} {}`, `{"a":1,"a":2}`} {
		if _, err := Compile([]byte(sch)); err == nil {
			t.Errorf("%q: want error", sch)
		}
	}
}

func TestValidateBrokenDocument(t *testing.T) {
	s, _ := Compile([]byte(`true`))
	for _, d := range []string{``, `{`, `{} {}`, `{"a":1,"x":{"a":1,"a":2}}`} {
		if _, err := s.Validate([]byte(d)); err == nil {
			t.Errorf("%q: want error", d)
		}
	}
	var se *jsonvalue.SyntaxError
	_, err := s.Validate([]byte(`{`))
	if !errors.As(err, &se) {
		t.Errorf("want SyntaxError, got %v", err)
	}
}

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
