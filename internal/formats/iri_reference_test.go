package formats

import "testing"

func TestIRIReference(t *testing.T) {
	table(t, "iri-reference", map[string]bool{"/ü": true, "#ü": true, "": true, "http://ƒøø.example/ü": true, "ü ": false, "/%4": false, "/\x00": false})
}
