package output

import (
	"encoding/json"

	"github.com/neoface/jsonschema/internal/keywords"
)

// basicUnit is one error unit. absoluteKeywordLocation is left out until an
// absolute base URI is known (stories 186/187).
type basicUnit struct {
	InstanceLocation string `json:"instanceLocation"`
	KeywordLocation  string `json:"keywordLocation"`
	Error            string `json:"error"`
}

func init() {
	Register("basic", func(valid bool, units []keywords.Unit) ([]byte, error) {
		if valid {
			return json.Marshal(struct {
				Valid bool `json:"valid"`
			}{true})
		}
		errs := make([]basicUnit, len(units))
		for i, u := range units {
			errs[i] = basicUnit{u.InstanceLocation, u.KeywordLocation, u.Error}
		}
		return json.Marshal(struct {
			Valid  bool        `json:"valid"`
			Errors []basicUnit `json:"errors"`
		}{false, errs})
	})
}
