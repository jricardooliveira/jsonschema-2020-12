package eval

import (
	"testing"

	"github.com/neoface/jsonschema/internal/compile"
)

// step: evaluate keyword kw (static location kwStatic) in the current node,
// then enter child node childStatic of resource res.
type step struct {
	kw, kwStatic, childStatic string
	res                       *compile.Resource
}

// eagerStrings builds the keyword and absolute locations at once, the way the
// code did before locations became lazy.
func eagerStrings(res *compile.Resource, steps []step) (kws, abs []string) {
	pf := &pathFrame{}
	r := res
	for _, s := range steps {
		kws = append(kws, pf.at(s.kw, s.kwStatic))
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

// lazyStrings renders through lazyLoc exactly as evalItem drives it.
func lazyStrings(res *compile.Resource, steps []step) (kws, abs []string) {
	pf := &pathFrame{}
	r := res
	for _, s := range steps {
		loc := &lazyLoc{pf: pf, res: r, static: s.kwStatic}
		pf.kw = s.kw
		if s.kw == "$ref" || s.kw == "$dynamicRef" {
			pf.at(s.kw, s.kwStatic) // a reference records its location now
		}
		kws = append(kws, loc.String())
		abs = append(abs, loc.Absolute())
		next := &pathFrame{static: s.childStatic, dynamic: dynamicLocation(pf, s.childStatic)}
		next.viaRef = pf.viaRef || s.kw == "$ref" || s.kw == "$dynamicRef"
		pf, r = next, s.res
	}
	return
}

func TestLazyLocEqualsEager(t *testing.T) {
	root := &compile.Resource{URI: "https://example.com/root"}
	cases := map[string]struct {
		steps   []step
		wantKw  []string
		wantAbs []string
	}{
		"ref": {[]step{
			{"$ref", "/$ref", "/$defs/x", root},
			{"properties", "/$defs/x/properties", "/$defs/x/properties/p", root},
			{"type", "/$defs/x/properties/p/type", "", root},
		}, []string{"/$ref", "/$ref/properties", "/$ref/properties/p/type"}, []string{"", "https://example.com/root#/$defs/x/properties", "https://example.com/root#/$defs/x/properties/p/type"}},
		"dynamicRef": {[]step{
			{"properties", "/properties", "/properties/n", root},
			{"$dynamicRef", "/properties/n/$dynamicRef", "/$defs/d", root},
			{"minimum", "/$defs/d/minimum", "", root},
		}, []string{"/properties", "/properties/n/$dynamicRef", "/properties/n/$dynamicRef/minimum"}, []string{"", "", "https://example.com/root#/$defs/d/minimum"}},
		"if/then/else": {[]step{
			{"if", "/if", "/if", root},
			{"properties", "/if/properties", "/if/properties/a", root},
			{"type", "/if/properties/a/type", "", root},
		}, []string{"/if", "/if/properties", "/if/properties/a/type"}, []string{"", "", ""}},
		"then": {[]step{
			{"then", "/then", "/then", root},
			{"required", "/then/required", "", root},
		}, []string{"/then", "/then/required"}, []string{"", ""}},
		"else": {[]step{
			{"else", "/else", "/else", root},
			{"minLength", "/else/minLength", "", root},
		}, []string{"/else", "/else/minLength"}, []string{"", ""}},
		"escaping": {[]step{
			{"$ref", "/$ref", "/$defs/a~1b~0c d", root},
			{"minLength", "/$defs/a~1b~0c d/minLength", "", root},
		}, []string{"/$ref", "/$ref/minLength"}, []string{"", "https://example.com/root#/$defs/a~1b~0c%20d/minLength"}},
		"empty absolute before ref": {[]step{
			{"properties", "/properties", "/properties/a", root},
			{"type", "/properties/a/type", "", root},
		}, []string{"/properties", "/properties/a/type"}, []string{"", ""}},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			wk, wa := eagerStrings(root, c.steps)
			gk, ga := lazyStrings(root, c.steps)
			for i := range wk {
				if gk[i] != wk[i] || ga[i] != wa[i] {
					t.Errorf("step %d: lazy %q %q, eager %q %q", i, gk[i], ga[i], wk[i], wa[i])
				}
			}
			// Pin a few literal values so the eager builder itself is checked.
			if name == "escaping" && ga[1] != c.wantAbs[1] {
				t.Errorf("absolute = %q, want %q", ga[1], c.wantAbs[1])
			}
			if name == "dynamicRef" && (gk[2] != c.wantKw[2] || ga[2] != c.wantAbs[2]) {
				t.Errorf("got %q %q", gk[2], ga[2])
			}
			if name == "empty absolute before ref" && (ga[0] != "" || ga[1] != "") {
				t.Errorf("absolute should be empty, got %q", ga)
			}
		})
	}
}
