package eval

import (
	"strings"
	"testing"

	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/jsonvalue"
	"github.com/neoface/jsonschema/internal/keywords"
)

// The verbose shape gets passing nodes from EvaluateTree, not from the
// failure units: units.go keeps only failures, so passing nodes never count
// against MaxOutputUnits.
func TestTreeKeepsPassingNodesUnitsOnlyFailures(t *testing.T) {
	const secret = "SECRET-VALUE-7731"
	s, err := jsonvalue.Decode([]byte(`{"properties":{"a":{"type":"string"},"b":{"type":"number"}}}`))
	if err != nil {
		t.Fatal(err)
	}
	n, err := compile.Compile(s)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := jsonvalue.Decode([]byte(`{"a":"` + secret + `","b":"` + secret + `"}`))
	if err != nil {
		t.Fatal(err)
	}
	_, units, err := EvaluateUnits(n, doc, DefaultMaxOutputUnits)
	if err != nil {
		t.Fatal(err)
	}
	for _, u := range units {
		if strings.HasPrefix(u.InstanceLocation, "/a") {
			t.Errorf("passing keyword recorded as unit: %+v", u)
		}
	}
	// Cap set to exactly the failure count: the tree's passing nodes must not count.
	ev, err := EvaluateTree(n, doc, len(units), 0)
	if err != nil {
		t.Fatalf("tree over unit cap: %v", err)
	}
	var passA, failB bool
	var walk func(*keywords.TreeNode)
	walk = func(t2 *keywords.TreeNode) {
		if t2.InstanceLocation == "/a" && t2.Valid {
			passA = true
		}
		if t2.InstanceLocation == "/b" && !t2.Valid {
			failB = true
		}
		for _, u := range t2.Errors {
			if strings.Contains(u.Error, secret) {
				t.Errorf("secret in error: %q", u.Error)
			}
		}
		for _, c := range t2.Children {
			walk(c)
		}
	}
	if ev.Tree == nil {
		t.Fatal("no tree")
	}
	walk(ev.Tree)
	if !passA || !failB {
		t.Errorf("passA=%v failB=%v", passA, failB)
	}
}
