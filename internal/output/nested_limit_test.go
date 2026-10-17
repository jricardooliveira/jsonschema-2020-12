package output

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/neoface/jsonschema/internal/jsonvalue"
	"github.com/neoface/jsonschema/internal/keywords"
)

func init() {
	RegisterTreeReport("test-nested-detailed", func(r Report, w *Budget) error { return renderTree(r, w, false) })
	RegisterTreeReport("test-nested-verbose", func(r Report, w *Budget) error { return renderTree(r, w, true) })
}

// bigTree is a failing root with n failing children, each with one error.
func bigTree(n int) *keywords.TreeNode {
	root := &keywords.TreeNode{Valid: false}
	for i := 0; i < n; i++ {
		root.Children = append(root.Children, &keywords.TreeNode{
			KeywordLocation: "/allOf/0", InstanceLocation: "/a",
			Errors: []keywords.Unit{{KeywordLocation: "/allOf/0/type", InstanceLocation: "/a", Error: "wrong type"}},
		})
	}
	return root
}

func TestNestedOverCapIsLimitError(t *testing.T) {
	rep := Report{Tree: bigTree(DefaultMaxReportBytes/100 + 1)}
	for _, shape := range []string{"test-nested-detailed", "test-nested-verbose"} {
		out, err := RenderReport(shape, rep, DefaultMaxReportBytes)
		var le *jsonvalue.LimitError
		if !errors.As(err, &le) || le.Limit != "report bytes" || le.Max != DefaultMaxReportBytes || out != nil {
			t.Errorf("%s: out=%d err=%v", shape, len(out), err)
		}
	}
}

func TestNestedRaisedCapPasses(t *testing.T) {
	rep := Report{Tree: bigTree(DefaultMaxReportBytes/100 + 1)}
	out, err := RenderReport("test-nested-verbose", rep, 4*DefaultMaxReportBytes)
	if err != nil || !json.Valid(out) {
		t.Fatalf("err=%v valid=%v", err, json.Valid(out))
	}
}

func TestNestedSmallCapStopsEarly(t *testing.T) {
	_, err := RenderReport("test-nested-verbose", Report{Tree: bigTree(50)}, 100)
	var le *jsonvalue.LimitError
	if !errors.As(err, &le) {
		t.Fatalf("err=%v", err)
	}
}

func TestNestedShapes(t *testing.T) {
	pass := &keywords.TreeNode{Valid: true, Children: []*keywords.TreeNode{
		{Valid: true, KeywordLocation: "/properties/a", InstanceLocation: "/a"},
		{Valid: true, KeywordLocation: "/properties/b", InstanceLocation: "/b", Annotations: []keywords.Annotation{
			{KeywordLocation: "/properties/b/title", InstanceLocation: "/b", Value: "T"}}},
	}}
	det, err := RenderReport("test-nested-detailed", Report{Valid: true, Tree: pass}, 0)
	if err != nil || strings.Contains(string(det), "/properties/a") || !strings.Contains(string(det), `"annotation":"T"`) {
		t.Errorf("detailed: %s %v", det, err)
	}
	verb, err := RenderReport("test-nested-verbose", Report{Valid: true, Tree: pass}, 0)
	if err != nil || !strings.Contains(string(verb), "/properties/a") || !json.Valid(verb) {
		t.Errorf("verbose: %s %v", verb, err)
	}
	for _, valid := range []bool{true, false} {
		out, _ := RenderReport("test-nested-verbose", Report{Valid: valid}, 0)
		var m map[string]any
		if json.Unmarshal(out, &m) != nil || m["valid"] != valid {
			t.Errorf("nil tree valid=%v: %s", valid, out)
		}
	}
}
