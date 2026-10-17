package output

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/neoface/jsonschema/internal/keywords"
)

func TestVerboseKeepsPassingNodesDetailedDrops(t *testing.T) {
	tree := &keywords.TreeNode{Valid: true, Children: []*keywords.TreeNode{
		{Valid: true, KeywordLocation: "/properties/a", InstanceLocation: "/a"},
		{Valid: true, KeywordLocation: "/properties/b", InstanceLocation: "/b", Annotations: []keywords.Annotation{
			{KeywordLocation: "/properties/b/title", InstanceLocation: "/b", Value: "T"}}},
	}}
	rep := Report{Valid: true, Tree: tree}
	if !NeedsTree("verbose") {
		t.Fatal("verbose not a tree shape")
	}
	v, err := RenderReport("verbose", rep, 0)
	if err != nil || !json.Valid(v) || !strings.Contains(string(v), "/properties/a") || !strings.Contains(string(v), `"annotation":"T"`) {
		t.Fatalf("verbose: %s %v", v, err)
	}
	d, err := RenderReport("detailed", rep, 0)
	if err != nil || strings.Contains(string(d), "/properties/a") {
		t.Fatalf("detailed: %s %v", d, err)
	}
}

func TestVerboseNilTreeUsesReportValid(t *testing.T) {
	for _, valid := range []bool{true, false} {
		out, err := RenderReport("verbose", Report{Valid: valid}, 0)
		var m map[string]any
		if err != nil || json.Unmarshal(out, &m) != nil || m["valid"] != valid {
			t.Errorf("valid=%v: %s %v", valid, out, err)
		}
	}
}
