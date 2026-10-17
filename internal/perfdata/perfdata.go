// Package perfdata builds the deterministic performance documents and reads
// the stored fixtures in testdata/perf. Standard library only.
package perfdata

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Item counts used by the benchmarks and fixture tests.
const (
	SmallItems  = 10
	MediumItems = 10000
	LargeItems  = 100000
	ErrorItems  = 10000
	ErrorEvery  = 10
)

// ValidOrder returns the valid order document with the given number of items.
func ValidOrder(items int) []byte {
	var b bytes.Buffer
	b.WriteString(`{"id": 1, "name": "order", "tags": ["a", "b", "c"], "items": [`)
	for i := 0; i < items; i++ {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, `{"sku": "ABC-%d", "qty": %d, "price": %d.99}`, i, i, i)
	}
	b.WriteString("]}")
	return b.Bytes()
}

// ErrorOrder returns the order document in which every ErrorEvery-th item
// has the sku "bad", which fails the sku pattern.
func ErrorOrder() []byte {
	var b bytes.Buffer
	b.WriteString(`{"id": 1, "items": [`)
	for i := 0; i < ErrorItems; i++ {
		if i > 0 {
			b.WriteString(", ")
		}
		if i%ErrorEvery == 0 {
			fmt.Fprintf(&b, `{"sku": "bad", "qty": %d}`, i)
		} else {
			fmt.Fprintf(&b, `{"sku": "ABC-%d", "qty": %d}`, i, i)
		}
	}
	b.WriteString("]}")
	return b.Bytes()
}

// Fixture reads testdata/perf/<name>, located relative to this source file.
func Fixture(name string) ([]byte, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("perfdata: cannot locate source file")
	}
	return os.ReadFile(filepath.Join(filepath.Dir(file), "..", "..", "testdata", "perf", name))
}
