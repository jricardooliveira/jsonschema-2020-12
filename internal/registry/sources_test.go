package registry

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neoface/jsonschema/internal/jsonvalue"
)

const fixtures = "../../testdata/registry/"

func write(t *testing.T, path, s string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(s), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDirRecursiveIgnoresNonJSON(t *testing.T) {
	got, err := (&Loader{}).Dir(fixtures+"dir/schemas", "https://example.com/schemas")
	if err != nil {
		t.Fatal(err)
	}
	var addrs []string
	for _, l := range got {
		addrs = append(addrs, l.Address)
	}
	want := "https://example.com/schemas/common/address.json,https://example.com/schemas/root.json"
	if strings.Join(addrs, ",") != want {
		t.Fatalf("addresses %v", addrs)
	}
}

func TestDirEscapesAddressSegments(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "a#b?c%d.json"), `{}`)
	got, err := (&Loader{}).Dir(dir, "https://x.example/s/")
	if err != nil {
		t.Fatal(err)
	}
	if got[0].Address != "https://x.example/s/a%23b%3Fc%25d.json" {
		t.Fatalf("address %q", got[0].Address)
	}
}

func TestDirSymlinks(t *testing.T) {
	outside := t.TempDir()
	write(t, filepath.Join(outside, "secret.json"), `{}`)
	root := t.TempDir()
	write(t, filepath.Join(root, "real.json"), `{"a":1}`)
	if err := os.Symlink("real.json", filepath.Join(root, "inside.json")); err != nil {
		t.Skip("symlinks unavailable")
	}
	got, err := (&Loader{}).Dir(root, "https://x.example/")
	if err != nil || len(got) != 2 {
		t.Fatalf("inside link: %v %d", err, len(got))
	}
	link := filepath.Join(root, "extra.json")
	if err := os.Symlink(filepath.Join(outside, "secret.json"), link); err != nil {
		t.Fatal(err)
	}
	_, err = (&Loader{}).Dir(root, "https://x.example/")
	if err == nil || !strings.Contains(err.Error(), link) {
		t.Fatalf("escaping link: %v", err)
	}
}

func TestDirSymlinkSiblingPrefixEscapes(t *testing.T) {
	parent := t.TempDir()
	root, evil := filepath.Join(parent, "schemas"), filepath.Join(parent, "schemas-evil")
	write(t, filepath.Join(evil, "x.json"), `{}`)
	write(t, filepath.Join(root, "ok.json"), `{}`)
	if err := os.Symlink(filepath.Join(evil, "x.json"), filepath.Join(root, "x.txt")); err != nil {
		t.Skip("symlinks unavailable")
	}
	if _, err := (&Loader{}).Dir(root, "https://x.example/"); err == nil {
		t.Fatal("link to sibling folder with shared prefix must fail, whatever its extension")
	}
}

func TestDanglingSymlinkFails(t *testing.T) {
	root := t.TempDir()
	if err := os.Symlink(filepath.Join(root, "nope.json"), filepath.Join(root, "l.json")); err != nil {
		t.Skip("symlinks unavailable")
	}
	if _, err := (&Loader{}).Dir(root, "https://x.example/"); err == nil {
		t.Fatal("dangling link must be an error")
	}
}

// The walked entry and the opened handle must be the same file.
func TestOpenChecksHandleIsWalkedFile(t *testing.T) {
	dir := t.TempDir()
	a, b := filepath.Join(dir, "a.json"), filepath.Join(dir, "b.json")
	write(t, a, `{}`)
	write(t, b, `{}`)
	want, err := os.Lstat(a)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := (&Loader{}).open(b, "", want); err == nil || !strings.Contains(err.Error(), b) {
		t.Fatalf("swapped file accepted: %v", err)
	}
	if _, err := (&Loader{}).open(a, "", want); err != nil {
		t.Fatal(err)
	}
}

func TestPerFileLimit(t *testing.T) {
	p := filepath.Join(t.TempDir(), "big.json")
	write(t, p, `{"k":"0123456789"}`)
	_, err := (&Loader{PerFile: 5}).File(p, "")
	var le *jsonvalue.LimitError
	if !errors.As(err, &le) || !strings.Contains(err.Error(), p) {
		t.Fatalf("got %v", err)
	}
}

func TestTotalCap(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "a.json"), `{"k":"0123456789"}`)
	write(t, filepath.Join(dir, "b.json"), `{"k":"0123456789"}`)
	one := len(`{"k":"0123456789"}`)
	_, err := (&Loader{Total: one + 1}).Dir(dir, "https://x.example/")
	var le *jsonvalue.LimitError
	if !errors.As(err, &le) || !strings.Contains(err.Error(), "b.json") {
		t.Fatalf("got %v", err)
	}
	if _, err := (&Loader{Total: 2 * one}).Dir(dir, "https://x.example/"); err != nil {
		t.Fatal(err)
	}
}

func TestBrokenJSONNamesFile(t *testing.T) {
	p := fixtures + "broken/bad.json"
	_, err := (&Loader{}).File(p, "")
	if err == nil || !strings.Contains(err.Error(), p) {
		t.Fatalf("got %v", err)
	}
}

func TestFileMissingAndBadAddress(t *testing.T) {
	if _, err := (&Loader{}).File(fixtures+"nope.json", ""); err == nil {
		t.Fatal("missing file must fail")
	}
	for _, a := range []string{"relative.json", "https://x/a.json#frag"} {
		if _, err := (&Loader{}).File(fixtures+"dir/schemas/root.json", a); err == nil {
			t.Fatalf("address %q must fail", a)
		}
	}
	if l, err := (&Loader{}).File(fixtures+"dir/schemas/root.json", "https://x/a.json"); err != nil || l.Doc == nil {
		t.Fatal(err)
	}
}
