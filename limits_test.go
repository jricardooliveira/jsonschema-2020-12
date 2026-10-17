package jsonschema

import (
	"errors"
	"strings"
	"testing"
)

func TestDefaultLimitValues(t *testing.T) {
	if DefaultMaxRefChain != 2*DefaultMaxDepth {
		t.Errorf("ref chain default %d is not twice depth %d", DefaultMaxRefChain, DefaultMaxDepth)
	}
	if DefaultMaxInputBytes != 64<<20 || DefaultMaxSteps != 100_000_000 {
		t.Errorf("unexpected defaults %d %d", DefaultMaxInputBytes, DefaultMaxSteps)
	}
}

func TestNegativeLimitOptionsRejected(t *testing.T) {
	for name, o := range map[string]Options{
		"MaxDepth":       {MaxDepth: -1},
		"MaxInputBytes":  {MaxInputBytes: -1},
		"MaxRefChain":    {MaxRefChain: -1},
		"MaxSteps":       {MaxSteps: -1},
		"MaxPatternSize": {MaxPatternSize: -1},
	} {
		_, err := CompileWithOptions([]byte(`true`), o)
		if err == nil || !strings.Contains(err.Error(), name) {
			t.Errorf("%s: want error naming it, got %v", name, err)
		}
	}
}

func TestInputSizeLimit(t *testing.T) {
	const max = 8
	at := strings.Repeat(" ", max-4) + `true`
	s, err := CompileWithOptions([]byte(at), Options{MaxInputBytes: max})
	if err != nil {
		t.Fatalf("at cap: %v", err)
	}
	var le *LimitError
	if _, err := CompileWithOptions([]byte(at+" "), Options{MaxInputBytes: max}); !errors.As(err, &le) || le.Limit != LimitInputSize || le.Value != max {
		t.Fatalf("schema over cap: %v", err)
	}
	if _, err := s.Validate([]byte(at)); err != nil {
		t.Fatalf("document at cap: %v", err)
	}
	le = nil
	if _, err := s.Validate([]byte(at + " ")); !errors.As(err, &le) || le.Limit != LimitInputSize || le.Value != max {
		t.Fatalf("document over cap: %v", err)
	}
	if want := "limit exceeded: input size 8"; le.Error() != want {
		t.Errorf("message %q, want %q", le.Error(), want)
	}
}

func TestExistingLimitErrorsAreTopLevel(t *testing.T) {
	s, err := CompileWithOptions([]byte(`true`), Options{MaxNumberDigits: 2})
	if err != nil {
		t.Fatal(err)
	}
	var le *LimitError
	if _, err := s.Validate([]byte(`123`)); !errors.As(err, &le) || le.Limit != LimitNumberDigits || le.Value != 2 {
		t.Fatalf("got %v", err)
	}
}

func TestPatternSizeOption(t *testing.T) {
	const max = 8
	at := `{"pattern":"` + strings.Repeat("a", max) + `"}`
	if _, err := CompileWithOptions([]byte(at), Options{MaxPatternSize: max}); err != nil {
		t.Fatalf("at cap: %v", err)
	}
	for _, s := range []string{
		`{"pattern":"` + strings.Repeat("a", max+1) + `"}`,
		`{"patternProperties":{"` + strings.Repeat("a", max+1) + `":true}}`,
		`{"additionalProperties":false,"patternProperties":{"` + strings.Repeat("a", max+1) + `":true}}`,
	} {
		var le *LimitError
		_, err := CompileWithOptions([]byte(s), Options{MaxPatternSize: max})
		if !errors.As(err, &le) || le.Limit != LimitPatternSize || le.Value != max {
			t.Errorf("%.40s: got %v", s, err)
		}
	}
	big := `{"pattern":"` + strings.Repeat("a", DefaultMaxPatternSize+1) + `"}`
	if _, err := CompileWithOptions([]byte(big), Options{MaxPatternSize: DefaultMaxPatternSize + 1}); err != nil {
		t.Errorf("raised cap: %v", err)
	}
	var le *LimitError
	if _, err := Compile([]byte(big)); !errors.As(err, &le) || le.Value != DefaultMaxPatternSize {
		t.Errorf("default cap: %v", err)
	}
}

func TestOutputUnitsAndAnnotationsLimitErrors(t *testing.T) {
	s, err := CompileWithOptions([]byte(`{"items":{"minimum":5}}`), Options{MaxOutputUnits: 1})
	if err != nil {
		t.Fatal(err)
	}
	var le *LimitError
	if _, err := s.Validate([]byte(`[1,2,3]`)); !errors.As(err, &le) || le.Limit != LimitOutputUnits || le.Value != 1 {
		t.Errorf("units: %v", err)
	}
	// The meta-check of the schema records annotations under the same cap.
	le = nil
	_, err = CompileWithOptions([]byte(`{"title":"t","description":"d"}`), Options{MaxAnnotations: 1})
	if !errors.As(err, &le) || le.Limit != LimitAnnotations || le.Value != 1 {
		t.Errorf("annotations: %v", err)
	}
}

func TestNumberExponentLimitError(t *testing.T) {
	s, err := CompileWithOptions([]byte(`true`), Options{MaxNumberExponent: 2})
	if err != nil {
		t.Fatal(err)
	}
	var le *LimitError
	if _, err := s.Validate([]byte(`1e999`)); !errors.As(err, &le) || le.Limit != LimitNumberExponent || le.Value != 2 {
		t.Errorf("got %v", err)
	}
}
