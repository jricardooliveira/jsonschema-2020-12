package formats

import "testing"

// RFC 2673 3.2.
func TestIPv4(t *testing.T) {
	table(t, "ipv4", map[string]bool{
		"192.168.0.1": true, "0.0.0.0": true, "255.255.255.255": true,
		"192.168.01.1": false, "256.1.1.1": false, "1.2.3": false, "1.2.3.4.5": false,
		"1.2.3.": false, "1..2.3": false, "a.b.c.d": false, "1.2.3.4 ": false, "": false,
	})
}
