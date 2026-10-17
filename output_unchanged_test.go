package jsonschema

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
)

// Operator decision (story 194, Checkpoint 2): adding detailed and verbose
// must leave flag and basic output byte-identical to base commit c941ecb,
// and passing tree nodes must never count against MaxOutputUnits.

const unchangedSchema = `{"items":{"properties":{"a":{"type":"string"},"b":{"minimum":0}},"required":["c"]}}`

// unchangedDoc is a large failing document: 3000 items, 3 failures each.
func unchangedDoc() []byte {
	return []byte("[" + strings.TrimSuffix(strings.Repeat(`{"a":1,"b":-1},`, 3000), ",") + "]")
}

// Hashes of the flag and basic output of unchangedDoc, taken with the
// flag/basic rendering path of base commit c941ecb (internal/eval,
// internal/keywords, flag.go and basic.go are unchanged since; the schema
// has no key that the fragment encoding would touch).
var unchangedHashes = map[string]string{
	"flag":  "af5fd803088fcdc627e5cd97fb88d7fcbbb02a705f38fd48b1ab5f4ffca50ae8",
	"basic": "9352adc7389fb7162cba707b38374c0eb4bc5fd5059277e697cf65d99be4ebc6",
}

func TestFlagBasicOutputUnchanged(t *testing.T) {
	s, err := Compile([]byte(unchangedSchema))
	if err != nil {
		t.Fatal(err)
	}
	res, err := s.Validate(unchangedDoc())
	if err != nil || res.Valid() {
		t.Fatal(res.Valid(), err)
	}
	for shape, want := range unchangedHashes {
		out, err := res.Output(shape)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(out)
		if got := hex.EncodeToString(sum[:]); got != want {
			t.Errorf("%s output changed: sha256 %s, want %s", shape, got, want)
		}
	}
}

// With MaxOutputUnits set to exactly the failure count, Validate and the
// detailed and verbose renders succeed: passing nodes are not counted.
func TestNestedShapesDoNotCountPassingNodes(t *testing.T) {
	s, err := Compile([]byte(unchangedSchema))
	if err != nil {
		t.Fatal(err)
	}
	res, err := s.Validate(unchangedDoc())
	if err != nil {
		t.Fatal(err)
	}
	b, err := res.Output("basic")
	if err != nil {
		t.Fatal(err)
	}
	var rep struct{ Errors []json.RawMessage }
	if err := json.Unmarshal(b, &rep); err != nil {
		t.Fatal(err)
	}
	n := len(rep.Errors)
	if n == 0 {
		t.Fatal("no failures")
	}
	// One unit fewer is over the cap, so n is the exact failure count.
	tight, _ := CompileWithOptions([]byte(unchangedSchema), Options{MaxOutputUnits: n - 1})
	if _, err := tight.Validate(unchangedDoc()); err == nil {
		t.Fatalf("cap %d not exact", n)
	}
	exact, err := CompileWithOptions([]byte(unchangedSchema), Options{MaxOutputUnits: n})
	if err != nil {
		t.Fatal(err)
	}
	res, err = exact.Validate(unchangedDoc())
	if err != nil {
		t.Fatalf("Validate at exact cap: %v", err)
	}
	for _, shape := range []string{"detailed", "verbose"} {
		out, err := res.Output(shape)
		if err != nil || !json.Valid(out) {
			t.Errorf("%s at exact cap: %v", shape, err)
		}
	}
}
