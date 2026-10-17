# PLAN #33: multipleOf without big integers on the common path

- Date: 2026-10-17
- Status: proposed (pending Checkpoint 2 sign-off)
- Design: see `ADR-033-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `internal/jsonvalue/number_inline.go`
- `internal/jsonvalue/number_inline_test.go`
- `internal/keywords/validation/multipleof.go`
- `internal/keywords/validation/multipleof_int64.go`
- `internal/keywords/validation/multipleof_int64_test.go`
- `internal/keywords/validation/multipleof_diff_test.go`
- `internal/keywords/validation/multipleof_allocs_test.go`
- `internal/keywords/validation/multipleof_race_on_test.go`
- `internal/keywords/validation/multipleof_race_off_test.go`
- `allocs_multipleof_test.go`
- `docs/benchmarks/r3-210-multipleof.md`

Modules the architecture expects to touch:

- `internal/jsonvalue/number_inline.go`
- `internal/jsonvalue/number_inline_test.go`
- `internal/keywords/validation/multipleof.go`
- `internal/keywords/validation/multipleof_int64.go`
- `internal/keywords/validation/multipleof_int64_test.go`
- `internal/keywords/validation/multipleof_diff_test.go`
- `internal/keywords/validation/multipleof_allocs_test.go`
- `internal/keywords/validation/multipleof_race_on_test.go`
- `internal/keywords/validation/multipleof_race_off_test.go`
- `allocs_multipleof_test.go`
- `docs/benchmarks/r3-210-multipleof.md`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — Non-allocating inline mantissa/exponent accessor
2. **T2** — int64 multipleOf decision with compile-time divisor small form  _(after T1)_
3. **T3** — Random differential test, 1M inputs, fixed seed  _(after T2)_
4. **T4** — Zero-allocation test and lowered root allocation limits  _(after T2)_
5. **T5** — Verify, check the golden files and record the benchmarks  _(after T3, T4)_

## Risks

- Story 203 may hold decimals such as 0.99 as float64 or big.Rat rather than as a decimal mantissa/exponent. If so, a non-allocating accessor needs a change in number.go or the decoder (see uncertainties).
- 1,000,000 cases under -race may take noticeably longer than the 'few seconds' the operator accepted.
- The race detector can add allocations, so the 0-allocation test needs the same race gating as the root allocation tests.
- The Validate5_5MB drop may fall short if the price numbers on the perf document do not reach the small path.

Implementation constraints the coder must honor:

- Integer arithmetic only in the decision: no float32/float64, math.Mod or strconv float parsing (ADR-004).
- Every overflow-prone step (abs of MinInt64, multiplying by powers of 10, exponent difference beyond ±36) falls back to the existing math/big path.
- The existing big path stays functionally unchanged and is the reference in the differential test.
- Error entries, messages, locations and evaluation-step charging stay identical. The golden files are read-only.
- The differential test uses a fixed seed (math/rand/v2 with a constant seed) and at least 1,000,000 cases in every run, including -race runs.
- New test files only; existing multipleOf tests and testdata stay byte-unchanged.
- Do not write allocs_test.go, docs/benchmarks.md, internal/keywords/registry.go, childctx*, internal/keywords/applicator/*, internal/keywords/annotate.go, internal/keywords/recording.go, internal/eval/evaluated.go, internal/registry/* or internal/compile/metacache.go.
- Standard library only; go.mod is unchanged.

- DB impact: no · API impact: no · Migration needed: no

Breaking changes:

- (none)

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: The new tests prove that the exact value round-trips against the big form, that big-form and out-of-range numbers report 'not small', and that AllocsPerRun is 0.
- **T2**: Fixed edge-case tests pass: overflow fallback, zero, negatives, powers of 2 and 5, and the ±18 boundaries. Existing multipleOf tests pass unchanged. A grep shows no float operations in the decision code.
- **T3**: The test runs at least 1,000,000 cases with the input mix from criterion 5, in both normal and -race runs. On failure it prints the seed and the failing case.
- **T4**: AllocsPerRun is 0 for small passing evalMultipleOf calls and for the small failing decision. The new root limits are the measured value plus about 10%, with the measured value noted. allocs_test.go is untouched.
- **T5**: go test ./... and go test -race ./... pass, the golden comparison passes, the before/after medians at count=10 are recorded, and the target verdicts are stated. If a target is missed, the story is held.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
