package jsonvalue

import (
	"runtime"
	"testing"

	"github.com/neoface/jsonschema/internal/perfdata"
)

// BenchmarkRetainedHeapExact is BenchmarkRetainedHeap on the decode path the
// validator uses (DecodeLimited, *Number leaves), so the "retained-B" metric
// shows how numbers are stored. BenchmarkRetainedHeap decodes into
// json.Number and never builds a *Number.
func BenchmarkRetainedHeapExact(b *testing.B) {
	for _, c := range []struct {
		name  string
		items int
	}{{"0.5MB", perfdata.MediumItems}, {"5.5MB", perfdata.LargeItems}} {
		b.Run(c.name, func(b *testing.B) {
			doc := perfdata.ValidOrder(c.items)
			var total float64
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				before := heapAlloc()
				b.StartTimer()
				v, err := DecodeLimited(doc, NumberLimits{})
				if err != nil {
					b.Fatal(err)
				}
				b.StopTimer()
				after := heapAlloc()
				runtime.KeepAlive(v)
				if after > before {
					total += float64(after - before)
				}
				b.StartTimer()
			}
			b.ReportMetric(total/float64(b.N), "retained-B")
		})
	}
}
