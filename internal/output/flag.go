package output

import (
	"encoding/json"

	"github.com/neoface/jsonschema/internal/keywords"
)

func init() {
	Register("flag", func(valid bool, _ []keywords.Unit) ([]byte, error) {
		return json.Marshal(struct {
			Valid bool `json:"valid"`
		}{valid})
	})
}
