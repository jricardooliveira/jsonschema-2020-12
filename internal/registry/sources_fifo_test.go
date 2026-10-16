//go:build unix

package registry

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// within runs f and fails the test if it does not return in time.
func within(t *testing.T, what string, f func() error) error {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- f() }()
	select {
	case err := <-done:
		return err
	case <-time.After(5 * time.Second):
		t.Fatalf("%s blocked on a FIFO", what)
		return nil
	}
}

// A .json named pipe in a listed folder, a link to one, or a listed FIFO is
// an error naming it, never a hang.
func TestFIFOFailsWithoutBlocking(t *testing.T) {
	root := t.TempDir()
	fifo := filepath.Join(root, "pipe.json")
	if err := syscall.Mkfifo(fifo, 0o600); err != nil {
		t.Skip("mkfifo unavailable")
	}
	linkRoot := t.TempDir()
	if err := syscall.Mkfifo(filepath.Join(linkRoot, "p.fifo"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("p.fifo", filepath.Join(linkRoot, "l.json")); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct{ dir, name string }{{root, fifo}, {linkRoot, "l.json"}} {
		err := within(t, "Dir", func() error { _, err := (&Loader{}).Dir(c.dir, "https://x.example/"); return err })
		if err == nil || !strings.Contains(err.Error(), c.name) {
			t.Fatalf("%s: got %v", c.name, err)
		}
	}
	err := within(t, "File", func() error { _, err := (&Loader{}).File(fifo, ""); return err })
	if err == nil || !strings.Contains(err.Error(), fifo) {
		t.Fatalf("listed FIFO: got %v", err)
	}
}
