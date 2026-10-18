package compile

import (
	"errors"
	"strings"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
)

// Compiling recurses only over decoded values, so the decoder's depth check
// bounds it: a schema at the default depth compiles (no stack trouble), one
// level deeper never reaches the compiler.
func TestCompileDepthBoundedByDecoder(t *testing.T) {
	n := jsonvalue.MaxDepth
	schema := strings.Repeat(`{"items":`, n-1) + `{}` + strings.Repeat("}", n-1)
	v, err := jsonvalue.DecodeLimited([]byte(schema), jsonvalue.NumberLimits{})
	if err != nil {
		t.Fatalf("decode at cap: %v", err)
	}
	if _, err := CompileDocument(v, StdinBaseURI); err != nil {
		t.Fatalf("compile at cap: %v", err)
	}
	over := `{"items":` + schema + `}`
	var le *jsonvalue.LimitError
	if _, err := jsonvalue.DecodeLimited([]byte(over), jsonvalue.NumberLimits{}); !errors.As(err, &le) || le.Max != n {
		t.Fatalf("decode over cap: %v", err)
	}
}

// A lowered depth applies to schemas the same way.
func TestCompileLoweredDepth(t *testing.T) {
	const max = 3
	lim := jsonvalue.NumberLimits{MaxDepth: max}
	at := `{"items":{"items":{}}}`
	v, err := jsonvalue.DecodeLimited([]byte(at), lim)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := CompileDocument(v, StdinBaseURI); err != nil {
		t.Fatal(err)
	}
	var le *jsonvalue.LimitError
	if _, err := jsonvalue.DecodeLimited([]byte(`{"items":`+at+`}`), lim); !errors.As(err, &le) || le.Max != max {
		t.Fatalf("over: %v", err)
	}
}
