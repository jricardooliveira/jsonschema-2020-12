package keywords

const (
	chunkFirstCap = 16
	chunkMaxCap   = 4096
)

// chunkStore holds annotations in chunks whose capacities grow 16, 32, 64, …
// and stop growing at 4,096 (chunk k has capacity min(16<<k, 4096)). It
// replaces one ever-growing slice, whose regrowth copies were the largest
// allocations of an evaluation.
//
// Invariants: chunks before cur are full; chunks after cur are empty (kept,
// not freed, so later Adds reuse them); n is the total entry count. A store
// belongs to one evaluation and is never shared or pooled, so it needs no
// locking. Truncate clears the removed entries so their Values can be freed.
type chunkStore struct {
	chunks [][]Annotation
	cur    int
	n      int
}

// Len is the number of entries held.
func (s *chunkStore) Len() int { return s.n }

// Mark returns a position that Truncate can roll back to.
func (s *chunkStore) Mark() int { return s.n }

// Add appends one entry.
func (s *chunkStore) Add(a Annotation) {
	for s.cur < len(s.chunks) && len(s.chunks[s.cur]) == cap(s.chunks[s.cur]) {
		s.cur++
	}
	if s.cur == len(s.chunks) {
		c := chunkMaxCap
		if s.cur < 8 { // 16<<8 == 4096
			c = chunkFirstCap << s.cur
		}
		s.chunks = append(s.chunks, make([]Annotation, 0, c))
	}
	s.chunks[s.cur] = append(s.chunks[s.cur], a)
	s.n++
}

// Truncate drops every entry at or after position mark (a value from Mark).
// Emptied chunks are kept for reuse.
func (s *chunkStore) Truncate(mark int) {
	if mark >= s.n {
		return
	}
	if mark < 0 {
		mark = 0
	}
	rem := mark
	first := true
	for i := range s.chunks {
		l := len(s.chunks[i])
		if rem >= l {
			rem -= l
			continue
		}
		if first {
			s.cur = min(s.cur, i) // earlier chunks are full, this one now has room
			first = false
		}
		clear(s.chunks[i][rem:])
		s.chunks[i] = s.chunks[i][:rem]
		rem = 0
	}
	s.n = mark
}

// Flatten copies all entries, in order, into one new slice of exactly Len.
func (s *chunkStore) Flatten() []Annotation {
	out := make([]Annotation, 0, s.n)
	for _, c := range s.chunks {
		out = append(out, c...)
	}
	return out
}
