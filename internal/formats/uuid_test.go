package formats

import "testing"

// RFC 4122 3: shape only.
func TestUUID(t *testing.T) {
	table(t, "uuid", map[string]bool{
		"2eb8aa08-aa98-11ea-b4aa-73b441d16380": true,
		"2EB8AA08-AA98-11EA-B4AA-73B441D16380": true,
		"2eb8aa08-aa98-01ea-04aa-73b441d16380": true,
		"2eb8aa08-aa98-91ea-f4aa-73b441d16380": true,
		"2eb8aa08aa9811eab4aa73b441d16380":     false,
		"2eb8aa08-aa98-11ea-b4aa-73b441d1638":  false,
		"2eb8aa08-aa98-11ea-b4aa-73b441d1638g": false,
		"":                                     false,
	})
}
