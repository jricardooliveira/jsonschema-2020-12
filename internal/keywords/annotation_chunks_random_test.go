package keywords

import (
	"math/rand"
	"testing"
)

const chunkRandomSeed = 208

// TestChunkStoreRandomAgainstSlice runs random add/mark/truncate/flatten
// steps against a plain slice. The seed is fixed and printed on failure.
func TestChunkStoreRandomAgainstSlice(t *testing.T) {
	rng := rand.New(rand.NewSource(chunkRandomSeed))
	var s chunkStore
	var ref []Annotation
	var marks []int
	seq := 0
	crossed, capped := false, false

	check := func(step int) {
		got := s.Flatten()
		if len(got) != len(ref) || s.Len() != len(ref) {
			t.Fatalf("seed %d step %d: len %d/%d, want %d", chunkRandomSeed, step, len(got), s.Len(), len(ref))
		}
		for i := range ref {
			if got[i].Seq != ref[i].Seq {
				t.Fatalf("seed %d step %d: entry %d = %d, want %d", chunkRandomSeed, step, i, got[i].Seq, ref[i].Seq)
			}
		}
	}
	truncate := func(step, m int) {
		before := chunkIndex(&s, s.Len())
		s.Truncate(m)
		ref = ref[:m]
		if chunkIndex(&s, m) < before {
			crossed = true
		}
		check(step)
	}
	add := func() {
		a := Annotation{Seq: seq}
		seq++
		s.Add(a)
		ref = append(ref, a)
		if len(s.chunks) > 0 && cap(s.chunks[len(s.chunks)-1]) == chunkMaxCap && len(s.chunks[len(s.chunks)-1]) > 0 {
			capped = true
		}
	}

	for step := 0; step < 10000; step++ {
		switch r := rng.Intn(100); {
		case r < 80:
			for k := rng.Intn(20) + 1; k > 0; k-- {
				add()
			}
		case r < 88:
			marks = append(marks, s.Mark())
		case r < 96 && len(marks) > 0:
			m := marks[len(marks)-1]
			marks = marks[:len(marks)-1]
			if m <= len(ref) {
				truncate(step, m)
			}
		case r < 97:
			truncate(step, rng.Intn(len(ref)+1))
		default:
			check(step)
		}
	}
	// Fixed growth sequence past the first capped chunk, then a rollback
	// across several chunks.
	for len(ref) < 8300 {
		add()
	}
	check(-1)
	truncate(-1, 5)
	for len(ref) < 8300 {
		add()
	}
	check(-2)
	if !crossed {
		t.Errorf("seed %d: no truncate crossed a chunk boundary", chunkRandomSeed)
	}
	if !capped {
		t.Errorf("seed %d: no chunk at the %d cap was reached", chunkRandomSeed, chunkMaxCap)
	}
}

// chunkIndex is the index of the chunk holding position n-1 (0 when empty).
func chunkIndex(s *chunkStore, n int) int {
	for i, c := range s.chunks {
		if n <= len(c) {
			return i
		}
		n -= len(c)
	}
	return len(s.chunks)
}
