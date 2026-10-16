package output

import (
	"testing"

	"github.com/neoface/jsonschema/internal/keywords"
)

func TestBasicExact(t *testing.T) {
	b, err := Render("basic", true, nil)
	if err != nil || string(b) != `{"valid":true}` {
		t.Fatalf("valid: %s %v", b, err)
	}
	u := []keywords.Unit{
		{InstanceLocation: "/a", KeywordLocation: "/properties/a/type", Error: "value must be of type string"},
		{InstanceLocation: "", KeywordLocation: "/properties", Error: "rule properties failed"},
	}
	want := `{"valid":false,"errors":[{"instanceLocation":"/a","keywordLocation":"/properties/a/type","error":"value must be of type string"},{"instanceLocation":"","keywordLocation":"/properties","error":"rule properties failed"}]}`
	b, err = Render("basic", false, u)
	if err != nil || string(b) != want {
		t.Fatalf("invalid: %s %v", b, err)
	}
	b, _ = Render("basic", false, nil)
	if string(b) != `{"valid":false,"errors":[]}` {
		t.Fatalf("empty: %s", b)
	}
}
