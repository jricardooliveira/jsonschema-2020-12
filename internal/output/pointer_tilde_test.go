package output

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/eval"
)

// oldUnescape is the decode used before the '~' fast path (RFC 6901 §4:
// "~1" first, then "~0").
var oldUnescape = strings.NewReplacer("~1", "/", "~0", "~")

func TestDecodeSegmentMatchesReplacer(t *testing.T) {
	for _, s := range []string{"", "a", "~", "~0", "~1", "~01", "~10", "~2", "a~1b~0c", "~~0", "~~1", "é~1", "a/b"} {
		if got, want := eval.DecodeSegment(s), oldUnescape.Replace(s); got != want {
			t.Errorf("DecodeSegment(%q) = %q, want %q", s, got, want)
		}
	}
	if got := eval.DecodeSegment("~01"); got != "~1" {
		t.Errorf("~01 decoded to %q, want ~1", got)
	}
	r := rand.New(rand.NewSource(1))
	alphabet := []string{"~", "0", "1", "a", "/", "é"}
	for i := 0; i < 20000; i++ {
		var b strings.Builder
		for n := r.Intn(8); n > 0; n-- {
			b.WriteString(alphabet[r.Intn(len(alphabet))])
		}
		s := b.String()
		if got, want := eval.DecodeSegment(s), oldUnescape.Replace(s); got != want {
			t.Fatalf("DecodeSegment(%q) = %q, want %q", s, got, want)
		}
	}
}

func TestDecodeSegmentNoTildeAllocatesNothing(t *testing.T) {
	seg := strings.Repeat("abc", 5)
	var sink string
	if n := testing.AllocsPerRun(100, func() { sink = eval.DecodeSegment(seg) }); n != 0 {
		t.Errorf("allocs = %v, want 0", n)
	}
	_ = sink
}
