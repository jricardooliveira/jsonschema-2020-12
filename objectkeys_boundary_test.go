package jsonschema

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"testing"
)

// Boundary test for story 207 (object keywords without wasted copies).
// objectKeysWant holds SHA-256 digests of the complete observable result
// (valid flag, annotations in order, flag/basic/detailed/verbose reports)
// captured by running this test on the unchanged code (a mismatch logs the
// current digest with -v).

const objectKeysSchema = `{
  "properties": {"B": {"type":"integer"}, "a": true, "é": true, "": true, "ab": true, "a1": {"type":"integer"}},
  "patternProperties": {"^x": {"type":"integer"}},
  "additionalProperties": {"type":"integer"}
}`

const objectKeysUnevaluatedSchema = `{
  "$defs": {"base": {"properties": {"B": true, "a": true, "é": true}}},
  "allOf": [{"$ref": "#/$defs/base"}, {"patternProperties": {"^x": true}}],
  "unevaluatedProperties": false
}`

var objectKeysNames = []string{"B", "a", "é", "z", "", "ab", "a1", "x1", "x2", "b", "A", "zz", "aa", "Z"}

// objectKeysDoc builds an object with n members taken from objectKeysNames
// (then generated names), all integers, or one string member when bad.
func objectKeysDoc(n int, bad bool) string {
	var parts []string
	for i := 0; i < n; i++ {
		name := fmt.Sprintf("k%03d", i)
		if i < len(objectKeysNames) {
			name = objectKeysNames[i]
		}
		val := "1"
		if bad && i == n-1 {
			val = `"s"`
		}
		parts = append(parts, fmt.Sprintf("%q:%s", name, val))
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func objectKeysDigest(t *testing.T, schema, doc string) string {
	t.Helper()
	s, err := Compile([]byte(schema))
	if err != nil {
		t.Fatal(err)
	}
	res, err := s.Validate([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "valid=%v\n", res.Valid())
	for _, a := range res.Annotations() {
		fmt.Fprintf(&b, "ann %s|%s|%s|%s|%v\n", a.InstanceLocation, a.KeywordLocation, a.AbsoluteKeywordLocation, a.Keyword, a.Value)
	}
	for _, shape := range []string{"flag", "basic", "detailed", "verbose"} {
		out, err := res.Output(shape)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(&b, "%s: %s\n", shape, out)
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

func TestObjectKeysBoundary(t *testing.T) {
	type tc struct {
		name, schema string
		n            int
		bad          bool
	}
	var cases []tc
	for _, n := range []int{0, 1, 8, 9, 14, 40} {
		for _, bad := range []bool{false, true} {
			if bad && n == 0 {
				continue
			}
			cases = append(cases,
				tc{fmt.Sprintf("main/n%d/bad=%v", n, bad), objectKeysSchema, n, bad},
				tc{fmt.Sprintf("unevaluated/n%d/bad=%v", n, bad), objectKeysUnevaluatedSchema, n, bad})
		}
	}
	got := map[string]string{}
	var names []string
	for _, c := range cases {
		got[c.name] = objectKeysDigest(t, c.schema, objectKeysDoc(c.n, c.bad))
		names = append(names, c.name)
	}
	sort.Strings(names)
	for _, n := range names {
		if got[n] != objectKeysWant[n] {
			t.Logf("\t%q: %q,", n, got[n])
			t.Errorf("%s: digest %s, want %s (output changed)", n, got[n], objectKeysWant[n])
		}
	}
}

// TestObjectKeysUnevaluatedSeesAll: unevaluatedProperties sees every property
// evaluated through $ref and allOf; an extra member is the only failure.
func TestObjectKeysUnevaluatedSeesAll(t *testing.T) {
	s, err := Compile([]byte(objectKeysUnevaluatedSchema))
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range []int{0, 1, 3} {
		res, _ := s.Validate([]byte(objectKeysDoc(n, false)))
		if !res.Valid() {
			t.Errorf("n=%d: want valid", n)
		}
	}
	res, _ := s.Validate([]byte(`{"B":1,"a":1,"zz":1}`))
	if res.Valid() {
		t.Error("extra member must fail unevaluatedProperties")
	}
}

var objectKeysWant = map[string]string{
	"main/n0/bad=false":         "8adc4fabfa6923b423726e0e40c25acd80c7a269be8bd0f10bd6892ebe8720c2",
	"main/n1/bad=false":         "8d619e0d49c0025eaa6688f2b77c1df0f27f33aa630663391d4aa638bc980dcb",
	"main/n1/bad=true":          "4e45ba81f092db57b5dfed40916412cd4fdbc38e7ba64f8ffdf8157920456d14",
	"main/n14/bad=false":        "de8facdf7e77e029699b908f6a8e76160445e207485c40bd4f3f61916609f09b",
	"main/n14/bad=true":         "1e6bb2f823add0f78e2b2d229a5383e7689e1575729ca12b5ed1ac068f1c2581",
	"main/n40/bad=false":        "46885b1b58746707638bbbddb730521fe261e4a6ef6dcff4b4893ea25e17ec91",
	"main/n40/bad=true":         "6ea890345e62283611b5f841a3da5cc9016ab8451bcf25ab6c2aba82012ef679",
	"main/n8/bad=false":         "b81f90ca3d43aa30ec8d4efb32589b3e7ce484bb01bb73f9add0eaf255b29c32",
	"main/n8/bad=true":          "2ed1bb8d734eeaacb497adb9fc3fc67a1b76858b1cd9540beebf04c744dc7d17",
	"main/n9/bad=false":         "c01ab9c3f868dd6ff14c9fce4e4aa5280f7573f5f5213b7bca1bde327df49669",
	"main/n9/bad=true":          "3619c5ff2ae8b4b5a0da1c71c5f443d76d1bd71f1557170730992eaabcd02b39",
	"unevaluated/n0/bad=false":  "5120165e88d81d344c5b41c9e741523433e5525fc190eea8d9fadeaf5a57d55b",
	"unevaluated/n1/bad=false":  "030fc6e0686565c5253326221122202a3a2a1c036dc2bf52d43184a5214e0ae3",
	"unevaluated/n1/bad=true":   "030fc6e0686565c5253326221122202a3a2a1c036dc2bf52d43184a5214e0ae3",
	"unevaluated/n14/bad=false": "798756940173bcbde06d82123b4f83b67bc376f97084cc069c4a5a1e9256c352",
	"unevaluated/n14/bad=true":  "798756940173bcbde06d82123b4f83b67bc376f97084cc069c4a5a1e9256c352",
	"unevaluated/n40/bad=false": "d2b69466e4597e3575d85c504941bc567c3658e0768a0a0f2e171f64eccdccde",
	"unevaluated/n40/bad=true":  "d2b69466e4597e3575d85c504941bc567c3658e0768a0a0f2e171f64eccdccde",
	"unevaluated/n8/bad=false":  "c73f8e9d51f01d4e7139a37dad3ed722743ef2173a54e73c2aa2689d257f5a61",
	"unevaluated/n8/bad=true":   "c73f8e9d51f01d4e7139a37dad3ed722743ef2173a54e73c2aa2689d257f5a61",
	"unevaluated/n9/bad=false":  "06fce7012bfef7abbc4ae3d73bee3a5abc3a576bb5dd633533f83a1d77b665e4",
	"unevaluated/n9/bad=true":   "06fce7012bfef7abbc4ae3d73bee3a5abc3a576bb5dd633533f83a1d77b665e4",
}
