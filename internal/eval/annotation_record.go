package eval

import (
	"sync"

	"github.com/neoface/jsonschema/internal/compile"
)

// EvaluateLazy is EvaluateWith (no tree) that leaves Annotations in recording
// order. The annotation-count cap is charged while recording, so it still
// fires here; the caller sorts on demand with Lazy.
func EvaluateLazy(n *compile.Node, instance any, lim Limits) (Evaluation, error) {
	return evaluateOpt(n, instance, lim.MaxUnits, lim.MaxAnnotations, nil, &callState{
		EvaluatedItems: newEvaluatedItems(lim.MaxSteps), active: map[frame]struct{}{},
		maxRefChain: lim.MaxRefChain, maxPath: 4 * lim.MaxDepth,
	}, true)
}

// Lazy sorts a Result's annotations once, on first use. It owns its input
// (records are copied out of per-call state at Validate), so concurrent
// readers are safe.
type Lazy struct {
	once   sync.Once
	raw    []Annotation
	sorted []Annotation
}

// NewLazy wraps annotations in recording order.
func NewLazy(raw []Annotation) *Lazy { return &Lazy{raw: raw} }

// Sorted returns the annotations in the documented order. Treat as read only.
func (l *Lazy) Sorted() []Annotation {
	if l == nil {
		return nil
	}
	l.once.Do(func() {
		l.sorted, l.raw = sortInPlace(l.raw), nil // raw is owned: no copy
	})
	return l.sorted
}
