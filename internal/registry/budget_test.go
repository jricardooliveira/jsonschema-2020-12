package registry

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/neoface/jsonschema/internal/jsonvalue"
	"github.com/neoface/jsonschema/internal/limits"
)

// The running total covers files and folders together, at, below and above
// the cap, and names the closed-set limit with the value that was set.
func TestRegistryTotalBudgetAcrossFilesAndDirs(t *testing.T) {
	const body = `{"k":"0123456789"}`
	one := len(body)
	dir := t.TempDir()
	file := filepath.Join(t.TempDir(), "f.json")
	write(t, file, body)
	write(t, filepath.Join(dir, "a.json"), body)
	write(t, filepath.Join(dir, "b.json"), body)
	load := func(total int) error {
		l := &Loader{Total: total}
		if _, err := l.File(file, ""); err != nil {
			return err
		}
		_, err := l.Dir(dir, "https://x.example/")
		return err
	}
	for _, total := range []int{3 * one, 3*one + 1} {
		if err := load(total); err != nil {
			t.Fatalf("total %d: %v", total, err)
		}
	}
	var le *jsonvalue.LimitError
	if err := load(3*one - 1); !errors.As(err, &le) || le.Limit != limits.RegistrySize || le.Max != 3*one-1 {
		t.Fatalf("over total: %v", err)
	}
}

// The default total applies when unset, and a file over the per-file cap is
// an input-size error even when the total still has room.
func TestRegistryPerFileIsInputSize(t *testing.T) {
	if (&Loader{}).total() != DefaultMaxRegistryBytes || (&Loader{}).perFile() != DefaultMaxFileBytes {
		t.Fatal("zero caps must fall back to the defaults")
	}
	const body = `{"k":"0123456789"}`
	p := filepath.Join(t.TempDir(), "f.json")
	write(t, p, body)
	if _, err := (&Loader{PerFile: len(body)}).File(p, ""); err != nil {
		t.Fatalf("at cap: %v", err)
	}
	var le *jsonvalue.LimitError
	if _, err := (&Loader{PerFile: len(body) - 1}).File(p, ""); !errors.As(err, &le) || le.Limit != limits.InputSize || le.Max != len(body)-1 {
		t.Fatalf("over cap: %v", err)
	}
}

// The loader decodes with the depth it is given.
func TestRegistryFileDepth(t *testing.T) {
	p := filepath.Join(t.TempDir(), "d.json")
	write(t, p, `[[[]]]`)
	if _, err := (&Loader{Numbers: jsonvalue.NumberLimits{MaxDepth: 3}}).File(p, ""); err != nil {
		t.Fatalf("at cap: %v", err)
	}
	var le *jsonvalue.LimitError
	if _, err := (&Loader{Numbers: jsonvalue.NumberLimits{MaxDepth: 2}}).File(p, ""); !errors.As(err, &le) || le.Limit != limits.Depth || le.Max != 2 {
		t.Fatalf("over cap: %v", err)
	}
}
