package formats

import "testing"

// draft-bhutton-relative-json-pointer-00 section 3.
func TestRelativeJSONPointer(t *testing.T) {
	table(t, "relative-json-pointer", map[string]bool{
		"0": true, "1/name": true, "0#": true, "0+1": true, "1-2/name": true, "10": true, "0+0": true, "2+3#": true, "0/": true,
		"01": false, "0+": false, "-1": false, "0##": false, "/a": false, "": false, "0-": false, "0+01": false, "0a": false, "0~": false, "0/~2": false, "#": false,
	})
}
