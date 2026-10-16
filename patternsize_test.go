package jsonschema_test

import (
	"strings"
	"testing"

	"github.com/neoface/jsonschema"
	"github.com/neoface/jsonschema/internal/keywords/validation"
)

func TestPatternSizeLimit(t *testing.T) {
	at := `{"pattern":"` + strings.Repeat("a", validation.DefaultMaxPatternSize) + `"}`
	if _, err := jsonschema.Compile([]byte(at)); err != nil {
		t.Errorf("pattern at the limit: %v", err)
	}
	over := `{"pattern":"` + strings.Repeat("a", validation.DefaultMaxPatternSize+1) + `"}`
	_, err := jsonschema.Compile([]byte(over))
	if err == nil {
		t.Fatal("pattern over the limit accepted")
	}
	if strings.Contains(err.Error(), "aaaa") {
		t.Error("error echoes the pattern")
	}
}
