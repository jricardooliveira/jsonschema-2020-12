package formats

import "testing"

// RFC 4291 2.2.
func TestIPv6(t *testing.T) {
	table(t, "ipv6", map[string]bool{
		"::1": true, "::": true, "1::": true, "2001:db8::8a2e:370:7334": true,
		"1:2:3:4:5:6:7:8": true, "::ffff:192.168.0.1": true, "1:2:3:4:5:6:1.2.3.4": true,
		"fe80::1%eth0": false, "1:2:3:4:5:6:7": false, "1:2:3:4:5:6:7:8:9": false,
		"1::2::3": false, "12345::1": false, "g::1": false, ":1": false, "1:": false,
		"::1.2.3": false, "::01.2.3.4": false, "1.2.3.4::1": false, "1:2:3:4:5:6:7::8": false, "": false,
	})
}
