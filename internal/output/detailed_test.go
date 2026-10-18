package output

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
)

func TestDetailedRegisteredNeedsTree(t *testing.T) {
	if !NeedsTree("detailed") {
		t.Fatal("detailed not registered as tree shape")
	}
}

func TestDetailedNilTreeUsesReportValid(t *testing.T) {
	for _, valid := range []bool{true, false} {
		out, err := RenderReport("detailed", Report{Valid: valid}, 0)
		var m map[string]any
		if err != nil || json.Unmarshal(out, &m) != nil || m["valid"] != valid {
			t.Errorf("valid=%v: %s %v", valid, out, err)
		}
	}
}

func TestDetailedCondensing(t *testing.T) {
	const secret = "SECRET-VALUE"
	fail := &keywords.TreeNode{Valid: false, Children: []*keywords.TreeNode{
		{Valid: true, KeywordLocation: "/properties/ok", InstanceLocation: "/ok"},
		{Valid: false, KeywordLocation: "/properties/a", InstanceLocation: "/a", Children: []*keywords.TreeNode{
			{Valid: false, KeywordLocation: "/properties/a/allOf/0", InstanceLocation: "/a",
				Errors: []keywords.Unit{{KeywordLocation: "/properties/a/allOf/0/type", AbsoluteKeywordLocation: "http://x/r#/properties/a/allOf/0/type", InstanceLocation: "/a", Error: "wrong type"}}},
		}},
	}}
	out, err := RenderReport("detailed", Report{Tree: fail}, 0)
	s := string(out)
	if err != nil || !json.Valid(out) {
		t.Fatalf("%s %v", s, err)
	}
	if strings.Contains(s, "/properties/ok") {
		t.Errorf("passing node kept: %s", s)
	}
	if strings.Contains(s, `"keywordLocation":"/properties/a"`) {
		t.Errorf("single-child wrapper not collapsed: %s", s)
	}
	for _, want := range []string{`"valid":false`, `"errors":[`, `"error":"wrong type"`, `"absoluteKeywordLocation":"http://x/r#/properties/a/allOf/0/type"`, `"instanceLocation":"/a"`} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %s in %s", want, s)
		}
	}
	if strings.Contains(s, secret) {
		t.Errorf("secret leaked: %s", s)
	}

	// Annotations under a failing node are dropped (core 12.4.2), even when
	// they sit on a passing child.
	failAnn := &keywords.TreeNode{Valid: false, Children: []*keywords.TreeNode{
		{Valid: true, KeywordLocation: "/properties/b", InstanceLocation: "/b", Annotations: []keywords.Annotation{
			{KeywordLocation: "/properties/b/title", InstanceLocation: "/b", Value: "T"}}},
		{Valid: false, KeywordLocation: "/properties/a", InstanceLocation: "/a",
			Errors: []keywords.Unit{{KeywordLocation: "/properties/a/type", InstanceLocation: "/a", Error: "wrong type"}}},
	}}
	out, err = RenderReport("detailed", Report{Tree: failAnn}, 0)
	s = string(out)
	if err != nil || !json.Valid(out) || strings.Contains(s, "annotation") || strings.Contains(s, "/properties/b") || !strings.Contains(s, `"error":"wrong type"`) {
		t.Errorf("annotated passing child kept under failing node: %s %v", s, err)
	}

	pass := &keywords.TreeNode{Valid: true, Children: []*keywords.TreeNode{
		{Valid: true, KeywordLocation: "/properties/b", InstanceLocation: "/b", Annotations: []keywords.Annotation{
			{KeywordLocation: "/properties/b/title", InstanceLocation: "/b", Value: "T"}}},
		{Valid: true, KeywordLocation: "/properties/c", InstanceLocation: "/c"},
	}}
	out, err = RenderReport("detailed", Report{Valid: true, Tree: pass}, 0)
	s = string(out)
	if err != nil || !strings.Contains(s, `"annotation":"T"`) || strings.Contains(s, "/properties/c") {
		t.Errorf("%s %v", s, err)
	}
}
