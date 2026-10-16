// Package keywords is the one registration point for schema keywords. A
// keyword file calls Register from init(); the compiler and evaluator find
// keywords only through Lookup.
package keywords

import "github.com/neoface/jsonschema/internal/jsonvalue"

// CompileContext is passed to a keyword's Compile function.
type CompileContext struct {
	// Location is the JSON Pointer of the keyword being compiled.
	Location string
	// Siblings are the raw values of every keyword in the same schema object.
	// Read only.
	Siblings map[string]any
	// Sub compiles a sub-schema found at Location+rel (rel is "" or a JSON
	// Pointer suffix such as "/0"), so its units carry the right keyword
	// location. The result is a compiled node to pass to EvalChild.
	Sub func(value any, rel string) (any, error)
}

// ItemRecorder is what keywords need from the per-evaluation item record.
type ItemRecorder interface {
	RecordPrefixItems(instLoc, kwLoc string, largest int, all bool)
	RecordItems(instLoc, kwLoc string)
	RecordContains(instLoc, kwLoc string, indices []int, all bool)
}

// Unit is one failure record: where it failed and which rule, with a
// sentence that never contains an instance value.
type Unit struct {
	InstanceLocation string
	KeywordLocation  string
	Error            string
}

// EvalContext holds per-call evaluation state: the current locations, the
// units reported at this level, and the output-unit counter shared with every
// child context. Create one with NewEvalContext.
type EvalContext struct {
	InstanceLocation string // JSON Pointer of the instance being checked
	KeywordLocation  string // JSON Pointer of the keyword being evaluated
	// Items is the per-evaluation record of item annotations (an
	// *eval.EvaluatedItems, which implements ItemRecorder), shared with every
	// child context. It is typed any so this package need not import eval.
	Items    any
	units    []Unit
	count    *int
	max      int
	evalNode func(ctx *EvalContext, node any, instance any) (bool, error)
}

// NewEvalContext returns a root context. evalNode evaluates a compiled child
// schema into the context it is given. max is the output-unit cap.
func NewEvalContext(max int, evalNode func(ctx *EvalContext, node any, instance any) (bool, error)) *EvalContext {
	return &EvalContext{count: new(int), max: max, evalNode: evalNode}
}

// Report appends a failure unit at the current locations and counts it. It
// returns a LimitError once the count of every appended unit exceeds the cap.
// msg must not contain instance values.
func (c *EvalContext) Report(msg string) error {
	*c.count++
	if *c.count > c.max {
		return &jsonvalue.LimitError{Limit: "output units", Max: c.max}
	}
	c.units = append(c.units, Unit{c.InstanceLocation, c.KeywordLocation, msg})
	return nil
}

// Units returns the units reported so far, in order.
func (c *EvalContext) Units() []Unit { return c.units }

// Keep adds a child's units (already counted) to this context's list.
func (c *EvalContext) Keep(units []Unit) { c.units = append(c.units, units...) }

// EvalChild evaluates a compiled child schema against instance, located at
// instLoc. It returns the child's units; the caller calls Keep to retain them
// or drops them. Every unit is counted when appended, kept or not.
func (c *EvalContext) EvalChild(node any, instance any, instLoc string) (bool, []Unit, error) {
	child := &EvalContext{InstanceLocation: instLoc, count: c.count, max: c.max, evalNode: c.evalNode, Items: c.Items}
	ok, err := c.evalNode(child, node, instance)
	return ok, child.units, err
}

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
