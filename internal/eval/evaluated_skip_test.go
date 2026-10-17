package eval

import (
	"testing"

	"github.com/neoface/jsonschema/internal/compile"
	"github.com/neoface/jsonschema/internal/keywords"
)

func TestSkippedRecordRecordsNothing(t *testing.T) {
	for _, skip := range []bool{false, true} {
		r := newEvaluatedItems(10)
		r.skip = skip
		r.RecordProperties("", "/properties", []string{"a"})
		r.RecordItems("/x", "/items")
		r.RecordPrefixItems("/x", "/prefixItems", 0, false)
		r.RecordContains("/x", "/contains", []int{1}, false)
		got := len(r.PropertiesAt("")) + len(r.At("/x"))
		if want := map[bool]int{false: 4, true: 0}[skip]; got != want {
			t.Errorf("skip=%v: recorded %d entries, want %d", skip, got, want)
		}
		// Step charging is independent of the switch.
		if err := r.Charge(10); err != nil {
			t.Errorf("skip=%v: charge: %v", skip, err)
		}
		if err := r.Step(); err == nil {
			t.Errorf("skip=%v: step cap not enforced", skip)
		}
	}
}

// probed is the record size x-eval-test-record-probe saw on its last run.
var probed = -1

func init() {
	// Runs after its siblings and reads how many evaluated-property/item
	// entries the call recorded so far.
	keywords.Register(keywords.Keyword{
		Name: "x-eval-test-record-probe", Vocabulary: "urn:test", AfterSiblings: true,
		Compile: func(*keywords.CompileContext, any) (any, error) { return nil, nil },
		Evaluate: func(ctx *keywords.EvalContext, _ any, _ any) (bool, error) {
			probed = len(ctx.Items.(*callState).entries)
			return true, nil
		},
	})
}

// End to end through the real entry points: only EvaluateValidate on a root
// marked NoUnevaluated records nothing; the tree and lazy paths, and a root
// not marked, record as before.
func TestEvaluateValidateSkipsRecordOnlyWhenMarked(t *testing.T) {
	schema := map[string]any{
		"properties":               map[string]any{"a": true},
		"x-eval-test-record-probe": true,
	}
	doc := map[string]any{"a": true}
	for _, mark := range []bool{false, true} {
		n, err := compile.Compile(schema)
		if err != nil {
			t.Fatal(err)
		}
		n.NoUnevaluated = mark // test-built node, never shared
		runs := map[string]func() (Evaluation, error){
			"validate": func() (Evaluation, error) { return EvaluateValidate(n, doc, DefaultLimits()) },
			"lazy":     func() (Evaluation, error) { return EvaluateLazy(n, doc, DefaultLimits()) },
			"tree":     func() (Evaluation, error) { return EvaluateWith(n, doc, DefaultLimits(), true) },
		}
		for name, run := range runs {
			probed = -1
			ev, err := run()
			if err != nil || !ev.Valid {
				t.Fatalf("%s mark=%v: %v %v", name, mark, ev.Valid, err)
			}
			wantEmpty := mark && name == "validate"
			if (probed == 0) != wantEmpty || probed < 0 {
				t.Errorf("%s mark=%v: %d entries recorded, want empty=%v", name, mark, probed, wantEmpty)
			}
		}
	}
}
