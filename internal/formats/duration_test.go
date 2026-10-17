package formats

import "testing"

// RFC 3339 appendix A.
func TestDuration(t *testing.T) {
	table(t, "duration", map[string]bool{
		"P3D": true, "PT1H30M": true, "P2W": true, "P1Y2M3DT4H5M6S": true, "PT0S": true,
		"P1W3D": false, "P": false, "PT": false, "P1D2Y": false, "P1DT": false,
		"3D": false, "P1": false, "PT1.5S": false, "P1H": false, "": false,
		// ABNF: each unit may only be followed by the next one (dur-year [dur-month] ...).
		"P1Y2M3D": true, "PT1H2M": true, "P1DT2H": true, "P1M3D": true, "PT1M2S": true,
		"P1Y1D": false, "PT1H1S": false, "P1YT1H1S": false,
	})
}
