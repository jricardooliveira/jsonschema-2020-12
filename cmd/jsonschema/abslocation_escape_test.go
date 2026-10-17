package main

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"
)

func TestBasicAbsoluteKeywordLocationIsEncoded(t *testing.T) {
	schema := write(t, "s.json", `{"$id":"http://x/r","$ref":"#/$defs/a%20b%25%C3%A9","$defs":{"a b%é":{"type":"number"}}}`)
	doc := write(t, "d.json", `"x"`)
	code, out, _ := do([]string{"validate", schema, doc}, "")
	if code != exitInvalid {
		t.Fatalf("code=%d out=%q", code, out)
	}
	var rep struct {
		Errors []struct {
			Abs string `json:"absoluteKeywordLocation"`
		}
	}
	if err := json.Unmarshal([]byte(out), &rep); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, e := range rep.Errors {
		if strings.Contains(e.Abs, "a%20b%25%C3%A9") {
			found = true
		}
		if _, err := url.Parse(e.Abs); err != nil {
			t.Errorf("not a URI: %q", e.Abs)
		}
	}
	if !found {
		t.Errorf("no encoded location in %s", out)
	}
}
