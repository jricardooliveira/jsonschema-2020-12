# ADR #33: multipleOf without big integers on the common path

- Date: 2026-10-17
- Story: #33
- Status: proposed (pending architecture sign-off)

## Decision
There are two layers. (1) internal/jsonvalue gains, in a new file number_inline.go, a method that returns (mantissa int64, exp int32, ok bool) from the number's existing inline representation without allocating. Because the method is in the same package, it reads unexported fields, so number.go needs no edit. (2) The multipleOf keyword keeps its big-form divisor and additionally stores a small form at compile time. A new file, multipleof_int64.go, holds the pure int64 decision. With instance m1·10^e1 and divisor m2·10^e2, and d = e1−e2: when d ≥ 0, reduce m2 by gcd(|m1|, |m2|), then strip up to d factors of 2 (bits.TrailingZeros64) and up to d factors of 5 (a bounded loop); the instance is a multiple only if 1 remains. When d < 0, it needs m2·10^|d| | m1, computed with overflow checks; on overflow the code falls back to the big path. MinInt64 absolute value and any overflow always fall back. evalMultipleOf tries the small path first and otherwise calls the unchanged big path. The error entry is produced exactly as today. For the differential test, both paths must be callable from tests as unexported functions in the same package.

## Affected modules
- internal/jsonvalue/number_inline.go
- internal/jsonvalue/number_inline_test.go
- internal/keywords/validation/multipleof.go
- internal/keywords/validation/multipleof_int64.go
- internal/keywords/validation/multipleof_int64_test.go
- internal/keywords/validation/multipleof_diff_test.go
- internal/keywords/validation/multipleof_allocs_test.go
- internal/keywords/validation/multipleof_race_on_test.go
- internal/keywords/validation/multipleof_race_off_test.go
- allocs_multipleof_test.go
- docs/benchmarks/r3-210-multipleof.md

## Data / API impact
- DB impact: no
- API impact: no
- Migration needed: no

## Constraints for implementation
- Integer arithmetic only in the decision: no float32/float64, math.Mod or strconv float parsing (ADR-004).
- Every overflow-prone step (abs of MinInt64, multiplying by powers of 10, exponent difference beyond ±36) falls back to the existing math/big path.
- The existing big path stays functionally unchanged and is the reference in the differential test.
- Error entries, messages, locations and evaluation-step charging stay identical. The golden files are read-only.
- The differential test uses a fixed seed (math/rand/v2 with a constant seed) and at least 1,000,000 cases in every run, including -race runs.
- New test files only; existing multipleOf tests and testdata stay byte-unchanged.
- Do not write allocs_test.go, docs/benchmarks.md, internal/keywords/registry.go, childctx*, internal/keywords/applicator/*, internal/keywords/annotate.go, internal/keywords/recording.go, internal/eval/evaluated.go, internal/registry/* or internal/compile/metacache.go.
- Standard library only; go.mod is unchanged.

## Risks
- Story 203 may hold decimals such as 0.99 as float64 or big.Rat rather than as a decimal mantissa/exponent. If so, a non-allocating accessor needs a change in number.go or the decoder (see uncertainties).
- 1,000,000 cases under -race may take noticeably longer than the 'few seconds' the operator accepted.
- The race detector can add allocations, so the 0-allocation test needs the same race gating as the root allocation tests.
- The Validate5_5MB drop may fall short if the price numbers on the perf document do not reach the small path.

## Breaking changes
- none

## Sensitivity
- none
