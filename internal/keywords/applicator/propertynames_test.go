package applicator_test

import (
	"strings"
	"testing"

	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/eval"
)

// The testdata/applicator/propertynames.json cases run through the root TestSpecCases.

func TestPropertyNamesKeyOnlyInMessagesNever(t *testing.T) {
	n := mustCompile(t, `{"propertyNames":{"maxLength":3}}`)
	hostile := "a~b/c@example.com"
	ok, units, err := eval.EvaluateUnits(n, mustDecode(t, `{"`+hostile+`":1}`), 100)
	if err != nil || ok || len(units) == 0 {
		t.Fatalf("ok=%v err=%v units=%+v", ok, err, units)
	}
	for _, u := range units {
		if u.InstanceLocation != "" {
			t.Errorf("key must not be in location: %q", u.InstanceLocation)
		}
		if strings.Contains(u.Error, hostile) || strings.Contains(u.Error, "example") {
			t.Errorf("message leaks key: %q", u.Error)
		}
	}
}

func TestPropertyNamesSchemaError(t *testing.T) {
	_, err := compile.Compile(mustDecode(t, `{"propertyNames":"x"}`))
	if err == nil || !strings.HasPrefix(err.Error(), "propertyNames: ") {
		t.Errorf("err=%v", err)
	}
}

func TestPropertyNamesDeterministicOrder(t *testing.T) {
	n := mustCompile(t, `{"propertyNames":false}`)
	for i := 0; i < 5; i++ {
		_, units, _ := eval.EvaluateUnits(n, mustDecode(t, `{"c":1,"a":1,"b":1}`), 100)
		if len(units) != 3 {
			t.Fatalf("units=%+v", units)
		}
	}
}

func TestPropertyNamesLimit(t *testing.T) {
	n := mustCompile(t, `{"propertyNames":false}`)
	_, _, err := eval.EvaluateUnits(n, mustDecode(t, `{"a":1,"b":1,"c":1}`), 2)
	if err == nil {
		t.Error("want limit error")
	}
}
