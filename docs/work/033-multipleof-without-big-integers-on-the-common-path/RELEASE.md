# Release notes — #33: multipleOf without big integers on the common path

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #33
- Status: approved by the operator at Checkpoint 3 (run #167, 2026-10-17 22:39 UTC)

**⚠️ the release-agent has concerns.**

## What changed

multipleOf now decides divisibility on small numbers (mantissa ≤ int64, exponent ±18) with int64 arithmetic only, never floats (ADR-004), and falls back to the existing math/big path for big mantissas or huge exponents. The decision allocates nothing. All existing multipleOf tests and CLI golden files pass unchanged, Validate5_5MB allocations drop by 400,000 (target 350,000: met), AllocsPerRun = 0 on small numbers (verified), and a fixed-seed 1M-case differential test proves the small path matches the big path exactly.

- internal/jsonvalue/number_inline.go: new Inline() accessor returns (mantissa int64, exponent int32, ok bool) without allocating; ok=false when mantissa held in big form or |exp|>18
- internal/jsonvalue/number_inline_test.go: tests for exact value round-trip, non-small detection, and AllocsPerRun=0
- internal/keywords/validation/multipleof.go: stores divisor's small form (mantissa int64, exponent int32, small bool) at compile time; evalMultipleOf tries multipleSmall first when both forms are small, falls back to unchanged isMultiple (math/big)
- internal/keywords/validation/multipleof_int64.go: multipleSmall() decides divisibility with int64 only (d≥0 branch: gcd reduction, trailing-zero stripping for factors 2 and 5; d<0 branch: overflow-checked multiply, remainder check); returns (multiple, decided bool)
- internal/keywords/validation/multipleof_int64_test.go: edge cases (zero, negatives, powers of 2/5, int64 and ±18 boundaries) verified against isMultiple
- internal/keywords/validation/multipleof_diff_test.go: TestMultipleSmallMatchesBigDifferential runs 1M cases with fixed seed (0x5eed0210, 0xd1ff) comparing multipleSmall vs isMultiple; input mix: ≥1/3 exact multiples (k×divisor in int64 range), high powers of 2/5 (0.125, 0.0625, 1024, 3125, 0.00032), negative divisors, int64 and ±18 exponent boundaries; prints seed and failing case on mismatch; coverage floor enforced
- internal/keywords/validation/multipleof_allocs_test.go: TestMultipleOfSmallZeroAllocs verifies AllocsPerRun=0 for yes/no decision on small numbers, skipped under -race
- internal/keywords/validation/multipleof_race_on_test.go, multipleof_race_off_test.go: race gating constant raceEnabled for allocation tests
- allocs_multipleof_test.go (root): lowered per-story allocation limits guarding Validate0_5MB (407,000 measured 370,045), Validate5_5MB (4,070,000 measured 3,700,065), ValidateErrors (308,000 measured 280,003), ValidateErrorsBasic (315,000 measured 286,036); reuses helpers from allocs_test.go read-only; allocs_test.go unchanged
- docs/benchmarks/r3-210-multipleof.md: benchmark table (before/after ns/op, B/op, allocs/op on 10 runs), AllocsPerRun=0 verdict met, Validate5_5MB drop (400,000 measured ≥ 350,000 required: met), golden comparison pass, timing caveat (parallel builds, operator re-measures at Checkpoint 3)

## How to verify

- go test ./... and go test -race ./... pass, all existing tests and golden files match v1.1
- go test -run TestMultipleSmallMatchesBigDifferential internal/keywords/validation -v to confirm 1M differential test runs, prints fixed seed, and completes without failure
- go test -run TestInlineAllocsZero internal/jsonvalue confirms Inline() allocates zero
- go test -run TestMultipleOfSmallZeroAllocs internal/keywords/validation (skipped if -race detected) confirms yes/no decision allocates zero
- Spot-check multipleof.go line ~78–85 for multipleSmall call only when both m.Inline() and div.small report ok=true
- go test ./cmd/jsonschema confirms golden comparison passes for all four output formats (-o flag, basic, detailed, verbose)
- Optional: run go test -run XXX -bench Validate -benchmem -count=10 . on the same machine and compare medians to docs/benchmarks/r3-210-multipleof.md; operator re-measures at Checkpoint 3

## Migration

none; internal optimization, no public API or data model changes, no persisted data affected

## Rollback

Revert multipleof.go: remove the d.sm, d.se, d.small fields from multipleOfDivisor, remove the Inline() call in compileMultipleOf, and remove the small-path branch in evalMultipleOf (lines calling multipleSmall). The code falls back entirely to the unchanged math/big path. Optionally remove allocs_multipleof_test.go and restore allocs_test.go looser limits. Validation results stay identical in both directions.

## Known limitations

- The small-path (int64 arithmetic) is taken only when both instance and divisor can be read as (mantissa int64, exponent ±18 via Inline()). Numbers with big-form mantissas or exponents outside ±18 automatically fall through to the unchanged math/big path, so their performance and allocation are as before.
- The zero-allocation guarantee (AllocsPerRun=0) applies only to the yes/no divisibility decision itself; writing an error entry (context.Report) for a failing number may allocate exactly as today, so error messages and reports are unchanged.
- Under -race, the race detector may add allocations beyond what the code causes; the AllopsPerRun assertion is skipped under -race, consistent with existing allocation tests.

## Concerns

- Benchmark documentation clarity (tester flagged as 'low' severity): (1) the v1.1 baseline reference in docs/benchmarks/r3-210-multipleof.md should cite its source more explicitly (e.g. 'allocs_test.go git history or docs/benchmarks.md from round 2 final'); (2) the 'Before' column header should note it represents 'the small path disabled' (not a true pre-release binary), while 'After' is 'this story's small path active'; (3) ValidateSmall's allocation drop (509→469 allocs/op, ~7% improvement) is not guarded in the new allocs_multipleof_test.go file, though the original allocs_test.go remains the looser upper bound. The operator will re-measure all numbers at Checkpoint 3 interleaved and will judge against those, not these in-repo medians, so the impact is documentation context only.
- Benchmark medians in docs/benchmarks/r3-210-multipleof.md were measured during parallel builds (stories #206, #207, #211 running concurrently), introducing timing noise (~±5%); the operator's Checkpoint 3 re-measure will be interleaved and will judge the 5% rule on those numbers instead. B/op and allocs/op are exact.
- The differential test (1M cases, fixed seed) runs under -race despite the added wall-clock cost; if CI time budgets are tight, the operator can decide to -race-skip the differential test separately at Checkpoint 3 if coverage from other test runs is sufficient.
