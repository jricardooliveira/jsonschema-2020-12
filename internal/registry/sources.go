package registry

import (
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/neoface/jsonschema/internal/jsonvalue"
)

// Default caps for supplied schema files.
const (
	DefaultMaxFileBytes     = 64 << 20  // per file, same as the CLI input cap
	DefaultMaxRegistryBytes = 256 << 20 // all supplied files together
)

// Loaded is one supplied schema file, read and parsed.
type Loaded struct {
	Address string // mapped URI; empty for a file found by its own $id
	Path    string // the path as listed (or found under a listed folder)
	Doc     any
}

// Loader reads exactly the files and folders it is told to, and keeps a
// running total of bytes read. It opens no other path and uses no network.
type Loader struct {
	PerFile int // 0 means DefaultMaxFileBytes
	Total   int // 0 means DefaultMaxRegistryBytes
	Numbers jsonvalue.NumberLimits
	used    int
}

func (l *Loader) perFile() int {
	if l.PerFile <= 0 {
		return DefaultMaxFileBytes
	}
	return l.PerFile
}

func (l *Loader) total() int {
	if l.Total <= 0 {
		return DefaultMaxRegistryBytes
	}
	return l.Total
}

// File loads one file the caller listed. A symbolic link is followed
// because the caller named it. address may be empty.
func (l *Loader) File(path, address string) (Loaded, error) { return l.open(path, address, nil) }

// open is File; when want is set, the opened file must be that same file
// (the entry the folder walk checked), so a swap after the walk is refused.
func (l *Loader) open(path, address string, want fs.FileInfo) (Loaded, error) {
	if address != "" {
		var err error
		if address, err = checkAbsolute(address); err != nil {
			return Loaded{}, fmt.Errorf("%s: %w", path, err)
		}
	}
	// O_NONBLOCK: opening a FIFO never waits for a writer; the checks on the
	// open handle below then reject it. Regular-file reads are unaffected.
	f, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return Loaded{}, fmt.Errorf("cannot read %s: %s", path, reason(err))
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return Loaded{}, fmt.Errorf("cannot read %s: read failed", path)
	}
	if !st.Mode().IsRegular() {
		return Loaded{}, fmt.Errorf("cannot read %s: not a regular file", path)
	}
	if want != nil && !os.SameFile(want, st) {
		return Loaded{}, fmt.Errorf("cannot read %s: changed while loading", path)
	}
	max, remaining := l.perFile(), l.total()-l.used
	if remaining < 0 {
		remaining = 0
	}
	lim := max
	if remaining < lim {
		lim = remaining
	}
	data, err := io.ReadAll(io.LimitReader(f, int64(lim)+1))
	if err != nil {
		return Loaded{}, fmt.Errorf("cannot read %s: read failed", path)
	}
	if len(data) > lim {
		if len(data) > max {
			return Loaded{}, fmt.Errorf("%s: %w", path, &jsonvalue.LimitError{Limit: "schema file bytes", Max: max})
		}
		return Loaded{}, fmt.Errorf("%s: %w", path, &jsonvalue.LimitError{Limit: "total registry bytes", Max: l.total()})
	}
	l.used += len(data)
	doc, err := jsonvalue.DecodeLimited(data, l.Numbers)
	if err != nil {
		return Loaded{}, fmt.Errorf("%s: %w", path, err)
	}
	return Loaded{Address: address, Path: path, Doc: doc}, nil
}

// Dir loads every .json file under root (recursively), addressing each one
// as base plus its slash-separated relative path. A symbolic link must
// resolve inside root, else it is an error. Links to folders inside root are
// not descended (their target is walked under its own path). Names not
// ending in .json are ignored; a .json entry that is not a regular file
// (FIFO, device, socket) is an error and is never opened. ponytail: lexical order from WalkDir, sorted again.
func (l *Loader) Dir(root, base string) ([]Loaded, error) {
	base, err := checkAbsolute(base)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", root, err)
	}
	if !strings.HasSuffix(base, "/") {
		base += "/"
	}
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, fmt.Errorf("cannot read folder %s: %s", root, reason(err))
	}
	if st, err := os.Stat(realRoot); err != nil || !st.IsDir() {
		return nil, fmt.Errorf("cannot read folder %s: not a folder", root)
	}
	type found struct {
		rel, path, real string
		info            fs.FileInfo // the checked file; the opened handle must match it
	}
	var files []found
	err = filepath.WalkDir(root, func(p string, d fs.DirEntry, werr error) error {
		if werr != nil {
			return fmt.Errorf("cannot read %s: %s", p, reason(werr))
		}
		if d.IsDir() {
			return nil
		}
		real := p
		var info fs.FileInfo
		if d.Type()&fs.ModeSymlink != 0 {
			r, err := filepath.EvalSymlinks(p)
			if err != nil {
				return fmt.Errorf("cannot resolve symbolic link %s: %s", p, reason(err))
			}
			rel, err := filepath.Rel(realRoot, r)
			if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				return fmt.Errorf("symbolic link %s leads outside folder %s", p, root)
			}
			st, err := os.Stat(r)
			if err != nil {
				return fmt.Errorf("cannot resolve symbolic link %s: %s", p, reason(err))
			}
			if st.IsDir() {
				return nil
			}
			real, info = r, st
		}
		if !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}
		if info == nil {
			if info, err = d.Info(); err != nil {
				return fmt.Errorf("cannot read %s: %s", p, reason(err))
			}
		}
		if !info.Mode().IsRegular() { // FIFO, device, socket: never opened
			return fmt.Errorf("cannot read %s: not a regular file", p)
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return fmt.Errorf("cannot read %s: bad path", p)
		}
		files = append(files, found{filepath.ToSlash(rel), p, real, info})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(files, func(i, j int) bool { return files[i].rel < files[j].rel })
	out := make([]Loaded, 0, len(files))
	for _, f := range files {
		segs := strings.Split(f.rel, "/")
		for i, s := range segs {
			segs[i] = url.PathEscape(s)
		}
		ld, err := l.open(f.real, base+strings.Join(segs, "/"), f.info)
		if err != nil {
			return nil, err
		}
		ld.Path = f.path
		out = append(out, ld)
	}
	return out, nil
}

// checkAbsolute requires an absolute URI without a fragment.
func checkAbsolute(s string) (string, error) {
	u, err := url.Parse(s)
	if err != nil || !u.IsAbs() || strings.ContainsAny(s, " \t\r\n") {
		return "", fmt.Errorf("address %q must be an absolute URI", s)
	}
	if u.Fragment != "" || strings.Contains(s, "#") {
		return "", fmt.Errorf("address %q must not have a fragment", s)
	}
	return s, nil
}

// reason is a fixed phrase for an OS error: no content, no OS text.
func reason(err error) string {
	switch {
	case os.IsNotExist(err):
		return "not found"
	case os.IsPermission(err):
		return "permission denied"
	}
	return "read failed"
}
