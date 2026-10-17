package jsonvalue

import (
	"runtime"
	"testing"

	"github.com/neoface/jsonschema/internal/perfdata"
)

// heapAlloc returns HeapAlloc after several GCs so the reading is stable.
func heapAlloc() uint64 {
	var m runtime.MemStats
	for i := 0; i < 3; i++ {
		runtime.GC()
	}
	runtime.ReadMemStats(&m)
	return m.HeapAlloc
}

// BenchmarkRetainedHeap reports the heap a decoded valid order document keeps
// alive, in the validator's own decoded form, as the "retained-B" metric.
func BenchmarkRetainedHeap(b *testing.B) {
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
				v, err := Decode(doc)
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
