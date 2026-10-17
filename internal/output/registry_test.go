package output

import (
	"strings"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
)

func TestUnknownShapeIsError(t *testing.T) {
	if _, err := Render("fancy", true, nil); err == nil {
		t.Fatal("want error")
	}
}

func TestDuplicateRegistrationPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("want panic")
		}
	}()
	Register("flag", func(bool, []keywords.Unit) ([]byte, error) { return nil, nil })
}

func TestNoInstanceValueInShapes(t *testing.T) {
	u := []keywords.Unit{{InstanceLocation: "/a", KeywordLocation: "/type", Error: "value must be of type string"}}
	for _, s := range []string{"flag", "basic"} {
		b, err := Render(s, false, u)
		if err != nil || strings.Contains(string(b), "s3cr3t") {
			t.Fatalf("%s: %s %v", s, b, err)
		}
	}
}
