package eval

import (
	"sync"
	"testing"

	"github.com/neoface/jsonschema/internal/compile"
)

// step is one hop: evaluate keyword kw (static location kwStatic) in the
// current node, then enter child node childStatic of resource res.
type step struct {
	kw, kwStatic, childStatic string
	res                       *compile.Resource
}

// eager replays the current pathFrame logic (refguard.go) for comparison.
func eager(root string, res *compile.Resource, steps []step) (kwLocs, abs []string) {
	pf := &pathFrame{static: root, dynamic: root}
	r := res
	for _, s := range steps {
		kwLocs = append(kwLocs, pf.at(s.kw, s.kwStatic))
		if pf.viaRef {
			abs = append(abs, r.AbsoluteLocation(s.kwStatic))
		} else {
			abs = append(abs, "")
		}
		next := &pathFrame{static: s.childStatic, dynamic: dynamicLocation(pf, s.childStatic)}
		next.viaRef = pf.viaRef || s.kw == "$ref" || s.kw == "$dynamicRef"
		pf, r = next, s.res
	}
	return
}

func lazy(root string, res *compile.Resource, steps []step) (kwLocs, abs []string) {
	n := &nodeLoc{static: root, res: res}
	for _, s := range steps {
		k := &kwLoc{node: n, name: s.kw, static: s.kwStatic}
		kwLocs = append(kwLocs, k.String())
		abs = append(abs, k.Absolute())
		n = k.child(s.childStatic, s.res)
	}
	return
}

func TestLazyLocationEqualsEager(t *testing.T) {
	root := &compile.Resource{URI: "https://example.com/root"}
	emb := &compile.Resource{URI: "https://example.com/emb", Pointer: "/$defs/e"}
	cases := map[string]struct {
		res   *compile.Resource
		steps []step
	}{
		"nested keywords": {root, []step{
			{"properties", "/properties", "/properties/a", root},
			{"items", "/properties/a/items", "/properties/a/items", root},
			{"type", "/properties/a/items/type", "", root},
		}},
		"escaping": {root, []step{
			{"properties", "/properties", "/properties/a~1b~0c", root},
			{"minLength", "/properties/a~1b~0c/minLength", "", root},
		}},
		"ref hop": {root, []step{
			{"$ref", "/$ref", "/$defs/x", root},
			{"properties", "/$defs/x/properties", "/$defs/x/properties/p", root},
			{"type", "/$defs/x/properties/p/type", "", root},
		}},
		"dynamicRef hop": {root, []step{
			{"properties", "/properties", "/properties/n", root},
			{"$dynamicRef", "/properties/n/$dynamicRef", "/$defs/d", root},
			{"minimum", "/$defs/d/minimum", "", root},
		}},
		"embedded resource": {root, []step{
			{"$ref", "/$ref", "/$defs/e", emb},
			{"items", "/$defs/e/items", "/$defs/e/items", emb},
			{"type", "/$defs/e/items/type", "", emb},
		}},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			wk, wa := eager("", c.res, c.steps)
			gk, ga := lazy("", c.res, c.steps)
			for i := range wk {
				if gk[i] != wk[i] || ga[i] != wa[i] {
					t.Errorf("step %d: got %q %q, want %q %q", i, gk[i], ga[i], wk[i], wa[i])
				}
			}
		})
	}
}

func TestLazyLocationParallelRender(t *testing.T) {
	res := &compile.Resource{URI: "https://example.com/r"}
	n := &nodeLoc{res: res}
	k := &kwLoc{node: n, name: "$ref", static: "/$ref"}
	k2 := &kwLoc{node: k.child("/$defs/x", res), name: "type", static: "/$defs/x/type"}
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if got := k2.String(); got != "/$ref/type" {
				t.Errorf("String = %q", got)
			}
			if got := k2.Absolute(); got != "https://example.com/r#/$defs/x/type" {
				t.Errorf("Absolute = %q", got)
			}
		}()
	}
	wg.Wait()
}
