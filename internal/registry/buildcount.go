package registry

import "sync/atomic"

// builtinBuilds counts every built-in index build (BuildBuiltinIndex calls).
// Only tests read it.
var builtinBuilds atomic.Int64

// BuiltinBuilds returns how many built-in index builds this process has done.
// It exists for tests; it changes no result.
func BuiltinBuilds() int64 { return builtinBuilds.Load() }
