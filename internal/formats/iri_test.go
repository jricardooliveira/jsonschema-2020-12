package formats

import "testing"

func TestIRI(t *testing.T) {
	table(t, "iri", map[string]bool{"http://ƒøø.example/ü": true, "http://a/b": true, "/ü": false, "http://a b/ü": false, "http://a/%zz": false, "http://a/": false})
}
