package registry

import (
	"embed"
	"fmt"
	"io/fs"
	"sync"

	"github.com/jricardooliveira/jsonschema-2020-12/internal/jsonvalue"
)

// The official 2020-12 meta-schema and its vocabulary meta-schemas, read only
// through go:embed: no disk or network access at run time.
//
//go:embed metaschemas/schema.json metaschemas/meta/*.json
var metaFS embed.FS

var (
	builtinOnce sync.Once
	builtinDocs map[string]any
	builtinErr  error
)

// Builtins returns the decoded embedded meta-schemas by their $id. The map and
// the documents are shared and must be treated as read-only. Any failure to
// load them is returned, never skipped.
func Builtins() (map[string]any, error) {
	builtinOnce.Do(func() {
		docs := map[string]any{}
		builtinErr = fs.WalkDir(metaFS, "metaschemas", func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			data, err := metaFS.ReadFile(p)
			if err != nil {
				return err
			}
			doc, err := jsonvalue.DecodeLimited(data, jsonvalue.NumberLimits{}) // *Number leaves, like supplied schemas
			if err != nil {
				return fmt.Errorf("embedded meta-schema %s: %v", p, err)
			}
			m, _ := doc.(map[string]any)
			id, _ := m["$id"].(string)
			if id == "" {
				return fmt.Errorf("embedded meta-schema %s has no $id", p)
			}
			docs[id] = doc
			return nil
		})
		if builtinErr == nil {
			builtinDocs = docs
		}
	})
	return builtinDocs, builtinErr
}
