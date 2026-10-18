package validation

import (
	"math"
	"math/bits"
	"math/rand/v2"
	"strconv"
	"testing"
)

// Fixed seed: every run, including -race, checks the same cases (ADR-004).
const (
	diffSeed1 = 0x5eed0210
	diffSeed2 = 0xd1ff
	diffCases = 1_000_000
)

var (
	diffMants = []int64{1, 2, 3, 5, 7, 10, 25, 32, 125, 625, 1024, 3125, 32000, 99,
		math.MaxInt64, math.MaxInt64 - 1, math.MaxInt64 / 2, 1 << 62, 1<<62 + 1, math.MinInt64 + 1}
	diffExps = []int{-18, -17, -1, 0, 1, 17, 18}
	// Divisors with high powers of 2 and 5: 0.125, 0.0625, 1024, 3125, 0.00032.
	diffDivisors = [][2]int64{{125, -3}, {625, -4}, {1024, 0}, {3125, 0}, {32, -5}, {1, -2}, {1, -18}, {1, 18}, {5, -1}, {1, 0}}
)

func numStr(m int64, e int) string { return strconv.FormatInt(m, 10) + "e" + strconv.Itoa(e) }

// TestMultipleSmallMatchesBigDifferential compares multipleSmall against
// isMultiple (math/big) on diffCases inputs.
func TestMultipleSmallMatchesBigDifferential(t *testing.T) {
	rng := rand.New(rand.NewPCG(diffSeed1, diffSeed2))
	randMant := func() int64 {
		switch rng.IntN(4) {
		case 0:
			return diffMants[rng.IntN(len(diffMants))]
		case 1:
			return rng.Int64N(1000) + 1
		case 2:
			return rng.Int64N(math.MaxInt64) + 1
		}
		return rng.Int64N(1<<20) + 1
	}
	randExp := func() int {
		if rng.IntN(2) == 0 {
			return diffExps[rng.IntN(len(diffExps))]
		}
		return rng.IntN(37) - 18
	}
	var decided, multiples int
	for i := 0; i < diffCases; i++ {
		var dm int64
		var de int
		if rng.IntN(3) == 0 {
			d := diffDivisors[rng.IntN(len(diffDivisors))]
			dm, de = d[0], int(d[1])
		} else {
			dm, de = randMant(), randExp()
		}
		var im int64
		var ie int
		switch i % 3 {
		case 0: // exact multiple k × divisor, kept in int64 range
			hi, lo := bits.Mul64(uint64(rng.Int64N(1<<20)), uint64(dm))
			if hi != 0 || lo > math.MaxInt64 {
				lo = uint64(dm)
			}
			im, ie = int64(lo), de
		default:
			im, ie = randMant(), randExp()
		}
		if rng.IntN(2) == 0 {
			im = -im
		}
		if rng.IntN(50) == 0 {
			im = 0
		}
		inst, div := mustNum(t, numStr(im, ie)), mustNum(t, numStr(dm, de))
		m1, e1, ok1 := inst.Inline()
		m2, e2, ok2 := div.Inline()
		if !ok1 || !ok2 {
			continue // outside the small range: big path only
		}
		got, ok := multipleSmall(m1, e1, m2, e2)
		if !ok {
			continue // declared fallback
		}
		decided++
		_, ma, ea := inst.Decimal()
		want := ma.Sign() == 0 || isMultiple(ma, ea, compileForTest(t, div))
		if got != want {
			t.Fatalf("seed=(%#x,%#x) case %d: instance %s multipleOf %s: small=%v big=%v",
				uint64(diffSeed1), uint64(diffSeed2), i, numStr(im, ie), numStr(dm, de), got, want)
		}
		if got {
			multiples++
		}
		// Negative divisor: the small path must refuse (big path handles it).
		if i%1000 == 0 && m1 != 0 {
			if _, ok := multipleSmall(m1, e1, -m2, e2); ok {
				t.Fatalf("seed=(%#x,%#x) case %d: negative divisor decided", uint64(diffSeed1), uint64(diffSeed2), i)
			}
		}
	}
	if decided < diffCases/2 || multiples < diffCases/4 {
		t.Fatalf("weak coverage: decided=%d multiples=%d of %d", decided, multiples, diffCases)
	}
}
