package output

import (
	"strings"
	"testing"

	"github.com/neoface/jsonschema/internal/keywords"
)

func hintUnits(n int) []keywords.Unit {
	us := make([]keywords.Unit, n)
	for i := range us {
		us[i] = keywords.Unit{InstanceLocation: "/a/b", KeywordLocation: "/properties/a", Error: "bad"}
	}
	return us
}

func TestSizeHintIsLocationSizePlusNewline(t *testing.T) {
	r := Report{Units: hintUnits(3)}
	per := errUnitOverhead + len("/a/b") + len("/properties/a") + len("bad")
	if got, want := sizeHint(r, 1<<30), 3*per+1; got != want {
		t.Errorf("hint = %d, want %d", got, want)
	}
}

func TestSizeHintCappedAtRemainingLimit(t *testing.T) {
	r := Report{Units: hintUnits(100)}
	if got := sizeHint(r, 10); got != 11 {
		t.Errorf("hint = %d, want 11", got)
	}
	if got := sizeHint(r, -5); got != 1 {
		t.Errorf("hint = %d, want 1", got)
	}
}

func TestSizeHintNotBelowReportAndBufferHoldsNewline(t *testing.T) {
	r := Report{Units: hintUnits(50)}
	out, err := RenderReport("basic", r, 0)
	if err != nil {
		t.Fatal(err)
	}
	if h := sizeHint(r, DefaultMaxReportBytes); h > len(out)+1 {
		t.Errorf("hint %d larger than report+newline %d", h, len(out)+1)
	}
	if cap(out) < len(out)+1 {
		t.Errorf("cap %d leaves no room for newline (len %d)", cap(out), len(out))
	}
}

func TestBasicLimitUnchangedByHint(t *testing.T) {
	r := Report{Units: hintUnits(5)}
	full, err := RenderReport("basic", r, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := RenderReport("basic", r, len(full)); err != nil {
		t.Errorf("exact limit failed: %v", err)
	}
	_, err = RenderReport("basic", r, len(full)-1)
	if err == nil || !strings.Contains(err.Error(), "report bytes") {
		t.Errorf("limit-1 err = %v, want report bytes", err)
	}
}
