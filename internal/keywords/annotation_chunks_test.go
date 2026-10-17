package keywords

import "testing"

func chunkAnn(i int) Annotation { return Annotation{Seq: i} }

func TestChunkStoreGrowthCapacities(t *testing.T) {
	var s chunkStore
	for i := 0; i < 16+32+64+128+256+512+1024+2048+4096+5; i++ {
		s.Add(chunkAnn(i))
	}
	want := []int{16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 4096}
	if len(s.chunks) != len(want) {
		t.Fatalf("chunks = %d, want %d", len(s.chunks), len(want))
	}
	for i, w := range want {
		if cap(s.chunks[i]) != w {
			t.Errorf("chunk %d cap = %d, want %d", i, cap(s.chunks[i]), w)
		}
	}
}

func TestChunkStoreTruncateKeepsAndReusesChunks(t *testing.T) {
	var s chunkStore
	for i := 0; i < 100; i++ {
		s.Add(chunkAnn(i))
	}
	nchunks := len(s.chunks)
	first := &s.chunks[2][0]
	s.Truncate(10) // crosses chunk boundaries
	if s.Len() != 10 || len(s.chunks) != nchunks {
		t.Fatalf("len %d chunks %d", s.Len(), len(s.chunks))
	}
	for i := 10; i < 100; i++ {
		s.Add(chunkAnn(i))
	}
	if len(s.chunks) != nchunks || first != &s.chunks[2][0] {
		t.Fatal("emptied chunks were reallocated")
	}
	for i, a := range s.Flatten() {
		if a.Seq != i {
			t.Fatalf("entry %d has Seq %d", i, a.Seq)
		}
	}
}

func TestChunkStoreEmptyAndNoopTruncate(t *testing.T) {
	var s chunkStore
	s.Truncate(0)
	if got := s.Flatten(); len(got) != 0 || s.Len() != 0 {
		t.Fatal("empty store not empty")
	}
	s.Add(chunkAnn(1))
	s.Truncate(s.Mark())
	s.Truncate(0)
	if s.Len() != 0 || len(s.Flatten()) != 0 {
		t.Fatal("truncate to 0 failed")
	}
}
