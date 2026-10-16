package jsonvalue

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestDecodeValid(t *testing.T) {
	v, err := Decode([]byte(` {"a":[1,2.50,"x",true,null],"b":{"c":{}}} `))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"a": []any{json.Number("1"), json.Number("2.50"), "x", true, nil},
		"b": map[string]any{"c": map[string]any{}},
	}
	if !reflect.DeepEqual(v, want) {
		t.Fatalf("got %#v", v)
	}
	for _, s := range []string{"true", "false", "null", "5", `"x"`, "[]", "{}"} {
		if _, err := Decode([]byte(s)); err != nil {
			t.Errorf("%s: %v", s, err)
		}
	}
}

func TestDecodeBroken(t *testing.T) {
	cases := map[string]string{
		"empty":         "",
		"blank":         "  \n",
		"syntax":        `{"a":}`,
		"truncated":     `{"a":[1,`,
		"trailing":      `{} {}`,
		"trailing bad":  `{} }`,
		"dup top":       `{"a":1,"a":2}`,
		"dup nested":    `[{"x":{"a":1,"b":2,"a":3}}]`,
		"missing comma": `[1 2]`,
		"mismatch":      `[}`,
		"stray close":   `]`,
		"key not str":   `{1:2}`,
	}
	for name, in := range cases {
		_, err := Decode([]byte(in))
		var se *SyntaxError
		if !errors.As(err, &se) {
			t.Errorf("%s: want *SyntaxError, got %v", name, err)
		}
	}
}

func TestDecodeDepthLimit(t *testing.T) {
	ok := strings.Repeat("[", MaxDepth) + strings.Repeat("]", MaxDepth)
	if _, err := Decode([]byte(ok)); err != nil {
		t.Fatalf("MaxDepth should decode: %v", err)
	}
	var le *LimitError
	for _, n := range []int{MaxDepth + 1, 100000} {
		_, err := Decode([]byte(strings.Repeat("[", n)))
		if !errors.As(err, &le) {
			t.Fatalf("depth %d: want *LimitError, got %v", n, err)
		}
	}
}

func TestErrorsQuoteNoContent(t *testing.T) {
	for _, in := range []string{`{"pw": "s3cr3t" !}`, `{"s3cr3t":1,"s3cr3t":2}`, `"s3cr3t" s3cr3t`, `{"k":s3cr3t}`} {
		_, err := Decode([]byte(in))
		if err == nil || strings.Contains(err.Error(), "s3cr3t") {
			t.Errorf("%s: bad error %v", in, err)
		}
	}
}
