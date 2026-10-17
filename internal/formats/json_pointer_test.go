package formats

import "testing"

// RFC 6901 sections 3 and 5.
func TestJSONPointer(t *testing.T) {
	table(t, "json-pointer", map[string]bool{
		"": true, "/": true, "/a~1b": true, "/~0~1": true, "/a/b//c": true, "/ü": true,
		"a": false, "/~2": false, "/~": false, "/a~": false, "#": false, "/\xff": false,
	})
}
