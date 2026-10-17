package output

import (
	"encoding/json"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
)

func init() {
	Register("flag", func(valid bool, _ []keywords.Unit) ([]byte, error) {
		return json.Marshal(struct {
			Valid bool `json:"valid"`
		}{valid})
	})
}
