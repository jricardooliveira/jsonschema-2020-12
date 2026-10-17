package formats

import "testing"

func TestURIReference(t *testing.T) {
	table(t, "uri-reference", map[string]bool{"/a/b?c#d": true, "#frag": true, "": true, "a/b": true, "a b": false, "/%zz": false, "/ü": false, "[": false})
}
