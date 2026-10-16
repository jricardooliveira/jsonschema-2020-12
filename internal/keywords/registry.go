// Package keywords is the one registration point for schema keywords. A
// keyword file calls Register from init(); the compiler and evaluator find
// keywords only through Lookup.
package keywords

// CompileContext is passed to a keyword's Compile function. It is empty for
// now; later stories add fields (base URI, limits, sub-schema compiler).
type CompileContext struct{}

// EvalContext holds per-call evaluation state. It is empty for now; later
// stories add dynamic scope, annotations and limit counters.
type EvalContext struct{}

// CompileFunc turns a keyword's schema value into compiled data. The result
// must be immutable: it is shared by every Validate call.
type CompileFunc func(ctx *CompileContext, value any) (any, error)

// EvaluateFunc checks an instance against the data CompileFunc returned.
// It reports whether the instance is valid; an error means evaluation could
// not complete and is never treated as valid.
type EvaluateFunc func(ctx *EvalContext, compiled any, instance any) (bool, error)

// Keyword describes one schema keyword.
type Keyword struct {
	Name       string
	Vocabulary string // vocabulary URI
	Compile    CompileFunc
	Evaluate   EvaluateFunc
}

// registry is written only by init() functions and read-only afterwards.
var registry = map[string]Keyword{}

// Register adds a keyword. It panics if the name is empty, a function is nil,
// or the name is already registered.
func Register(k Keyword) {
	if k.Name == "" || k.Compile == nil || k.Evaluate == nil {
		panic("keywords: incomplete registration")
	}
	if _, dup := registry[k.Name]; dup {
		panic("keywords: duplicate registration of " + k.Name)
	}
	registry[k.Name] = k
}

// Lookup returns the keyword registered under name.
func Lookup(name string) (Keyword, bool) {
	k, ok := registry[name]
	return k, ok
}
