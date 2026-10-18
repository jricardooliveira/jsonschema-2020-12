package eval

import "sync"

// Pending holds a Result's annotations as the evaluation left them: the
// chunks of the per-call chunk store (see keywords.chunkStore), in recording
// order. The first Sorted call flattens them into one slice, sorts it with
// the existing order (sortInPlace) and releases the chunks, all inside one
// sync.Once, so concurrent first readers never see a half-built or
// half-released state. The chunks are owned: the evaluation that filled them
// has ended and nothing else references them.
type Pending struct {
	once   sync.Once
	chunks [][]Annotation
	sorted []Annotation
}

// Sorted returns the annotations in the documented order. Treat as read only.
func (p *Pending) Sorted() []Annotation {
	if p == nil {
		return nil
	}
	p.once.Do(func() {
		n := 0
		for _, c := range p.chunks {
			n += len(c)
		}
		flat := make([]Annotation, 0, n)
		for _, c := range p.chunks {
			flat = append(flat, c...)
		}
		p.sorted, p.chunks = sortInPlace(flat), nil
	})
	return p.sorted
}
