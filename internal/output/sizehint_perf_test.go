package output_test

import (
	"encoding/json"
	"testing"

	"github.com/neoface/jsonschema"
	"github.com/neoface/jsonschema/internal/keywords"
	"github.com/neoface/jsonschema/internal/output"
	"github.com/neoface/jsonschema/internal/perfdata"
)

// The hint covers the whole rendered basic report for the perf documents, so
// Budget.Write never regrows the buffer; it is only a capacity.
func TestSizeHintCoversRenderedPerfReports(t *testing.T) {
	src, err := perfdata.Fixture("order_schema.json")
	if err != nil {
		t.Fatal(err)
	}
	s, err := jsonschema.Compile(src)
	if err != nil {
		t.Fatal(err)
	}
	docs := map[string][]byte{
		"medium": perfdata.ValidOrder(perfdata.MediumItems),
		"errors": perfdata.ErrorOrder(),
	}
	for name, doc := range docs {
		res, err := s.Validate(doc)
		if err != nil {
			t.Fatal(err)
		}
		out, err := res.Output("basic")
		if err != nil {
			t.Fatal(err)
		}
		rep := output.Report{Valid: res.Valid()}
		if res.Valid() {
			for _, a := range res.Annotations() {
				rep.Annotations = append(rep.Annotations, keywords.Annotation{
					InstanceLocation: a.InstanceLocation, KeywordLocation: a.KeywordLocation,
					AbsoluteKeywordLocation: a.AbsoluteKeywordLocation, Keyword: a.Keyword, Value: a.Value})
			}
		} else {
			var parsed struct {
				Errors []struct {
					InstanceLocation        string `json:"instanceLocation"`
					KeywordLocation         string `json:"keywordLocation"`
					AbsoluteKeywordLocation string `json:"absoluteKeywordLocation"`
					Error                   string `json:"error"`
				} `json:"errors"`
			}
			if err := json.Unmarshal(out, &parsed); err != nil {
				t.Fatal(err)
			}
			for _, u := range parsed.Errors {
				rep.Units = append(rep.Units, keywords.Unit{InstanceLocation: u.InstanceLocation,
					KeywordLocation: u.KeywordLocation, AbsoluteKeywordLocation: u.AbsoluteKeywordLocation, Error: u.Error})
			}
		}
		if h := output.SizeHint(rep, output.DefaultMaxReportBytes); h < len(out) {
			t.Errorf("%s: hint %d below rendered size %d", name, h, len(out))
		}
	}
}
