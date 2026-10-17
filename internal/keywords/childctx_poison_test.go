package keywords

import "testing"

// A keyword that keeps its child context after EvalChild returns must fail.
func TestPoisonCatchesRetainedChildContext(t *testing.T) {
	var kept *EvalContext
	root := NewEvalContext(10, func(c *EvalContext, _ any, _ any) (bool, error) {
		kept = c // the deliberate retaining "keyword"
		return true, nil
	})
	if _, _, err := root.EvalChild(nil, nil, "/a"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("use of a retained child context did not panic")
		}
	}()
	_ = kept.Report("retained use")
}

// Units handed to the caller survive the poisoning.
func TestPoisonKeepsReturnedUnits(t *testing.T) {
	root := NewEvalContext(10, func(c *EvalContext, _ any, _ any) (bool, error) {
		return false, c.Report("bad")
	})
	_, units, _ := root.EvalChild(nil, nil, "/a")
	if len(units) != 1 || units[0].Error != "bad" || units[0].InstanceLocation != "/a" {
		t.Fatalf("units lost: %+v", units)
	}
}
