package applicator_test

import (
	"strings"
	"sync"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/compile"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/eval"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
	"github.com/jricardooliveira/jsonschema-2020-12/internal/keywords"
	_ "github.com/jricardooliveira/jsonschema-2020-12/internal/keywords/all"
)

// The testdata/applicator/prefixItems.json cases run through the root TestSpecCases.

func mustCompile(t *testing.T, schema string) *compile.Node {
	t.Helper()
	v, err := jsonvalue.DecodeLimited([]byte(schema), jsonvalue.NumberLimits{})
	if err != nil {
		t.Fatal(err)
	}
	n, err := compile.Compile(v)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func mustDecode(t *testing.T, doc string) any {
	t.Helper()
	v, err := jsonvalue.DecodeLimited([]byte(doc), jsonvalue.NumberLimits{})
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestPrefixItemsAnnotation(t *testing.T) {
	n := mustCompile(t, `{"prefixItems":[true,true]}`)
	cases := []struct {
		doc     string
		applied bool
		all     bool
		largest int
	}{
		{`[1]`, true, true, 0},
		{`[1,2]`, true, true, 1},
		{`[1,2,3]`, true, false, 1},
		{`[]`, false, false, 0},
		{`"x"`, false, false, 0},
		{`false`, false, false, 0},
		{`7`, false, false, 0},
	}
	for _, c := range cases {
		ok, rec, err := eval.EvaluateAnnotated(n, mustDecode(t, c.doc))
		if err != nil || !ok {
			t.Fatalf("%s: ok=%v err=%v", c.doc, ok, err)
		}
		a, got := rec.At("")["/prefixItems"]
		if got != c.applied || a.All != c.all || (c.applied && !c.all && a.Largest != c.largest) {
			t.Errorf("%s: got %+v recorded=%v", c.doc, a, got)
		}
	}
}

func TestPrefixItemsReportsEveryFailingPosition(t *testing.T) {
	n := mustCompile(t, `{"prefixItems":[{"type":"string"},{"type":"string"},{"type":"string"}]}`)
	ok, units, err := eval.EvaluateUnits(n, mustDecode(t, `[1,"secret",3]`), 100)
	if err != nil || ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	var locs, kws []string
	for _, u := range units {
		locs = append(locs, u.InstanceLocation)
		kws = append(kws, u.KeywordLocation)
		if strings.Contains(u.Error, "secret") {
			t.Errorf("unit leaks value: %q", u.Error)
		}
	}
	if strings.Join(locs, ",") != "/0,/2" || strings.Join(kws, ",") != "/prefixItems/0/type,/prefixItems/2/type" {
		t.Errorf("locs=%v kws=%v", locs, kws)
	}
}

func TestPrefixItemsFailedKeywordRecordsNothing(t *testing.T) {
	n := mustCompile(t, `{"prefixItems":[false]}`)
	ok, rec, err := eval.EvaluateAnnotated(n, mustDecode(t, `[1]`))
	if err != nil || ok || len(rec.At("")) != 0 {
		t.Errorf("ok=%v err=%v rec=%v", ok, err, rec.At(""))
	}
}

func TestPrefixItemsSchemaErrors(t *testing.T) {
	for _, s := range []string{`{"prefixItems":"x"}`, `{"prefixItems":[1]}`, `{"prefixItems":{"type":"string"}}`} {
		v := mustDecode(t, s)
		if _, err := compile.Compile(v); err == nil || strings.Contains(err.Error(), "string\"") {
			t.Errorf("%s: err=%v", s, err)
		}
	}
}

func TestPrefixItemsKeepsNestedCompileError(t *testing.T) {
	_, err := compile.Compile(mustDecode(t, `{"prefixItems":[true,{"minItems":"x"}]}`))
	if err == nil || !strings.Contains(err.Error(), "minItems: must be a number") {
		t.Errorf("err=%v", err)
	}
}

func TestPrefixItemsOutputCapPropagates(t *testing.T) {
	n := mustCompile(t, `{"prefixItems":[false,false,false]}`)
	_, _, err := eval.EvaluateUnits(n, mustDecode(t, `[1,2,3]`), 2)
	var le *jsonvalue.LimitError
	if err == nil || !errorsAs(err, &le) {
		t.Errorf("want LimitError, got %v", err)
	}
}

func errorsAs(err error, target **jsonvalue.LimitError) bool {
	le, ok := err.(*jsonvalue.LimitError)
	if ok {
		*target = le
	}
	return ok
}

func TestItemKeywordsConcurrent(t *testing.T) {
	n := mustCompile(t, `{"prefixItems":[{"type":"integer"}],"items":{"type":"string"}}`)
	doc := mustDecode(t, `[1,"a","b"]`)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if ok, _, err := eval.EvaluateAnnotated(n, doc); !ok || err != nil {
				t.Errorf("ok=%v err=%v", ok, err)
			}
		}()
	}
	wg.Wait()
}

var _ = keywords.Lookup
