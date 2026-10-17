package formats

import "testing"

func TestURI(t *testing.T) {
	table(t, "uri", map[string]bool{"http://a/b": true, "a:b": true, "/a": false, "": false, "http://a b": false, "http://a/%zz": false, "http://a/%4": false, "http://a/ü": false})
}
