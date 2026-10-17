# SPEC #33: multipleOf without big integers on the common path

- Date: 2026-10-17
- Type: performance
- Verdict: pass
- Status: proposed (pending Checkpoint 1 sign-off)

## Problem

evalMultipleOf allocates about 6 big.Ints per evaluation, even for ordinary numbers such as prices like 12.99 with multipleOf 0.01. On the 5.5 MB perf document this happens about 100,000 times per validation.

## Why this matters

Round-3 fix 6. It removes an allocation hot spot on number-heavy documents and keeps exact answers (ADR-004: integer arithmetic only, never floats). Allocations are judged exactly. This story is built in parallel with #206, #207 and #211, so all of its guards and records go in new files.

## Acceptance criteria

- internal/jsonvalue exposes a number's inline decimal mantissa (int64) and base-10 exponent through a new accessor in a new file. The accessor allocates nothing and reports 'not small' when the number is held in big form. For every number, the value it returns equals the number's exact value.
- At compile time, multipleOf stores the divisor's small form (int64 mantissa, exponent) when the mantissa fits int64 and the exponent is within ±18, alongside the existing big form. The compiled schema stays immutable.
- When both the instance and the divisor are in the small range (mantissa fits int64, exponent within ±18), divisibility is decided with int64 integer arithmetic only: a remainder check, trailing zeros for the factor 2, and a bounded loop for the factor 5. No floating-point operation appears anywhere in the decision. Every other case, including any intermediate step that would overflow int64, falls back to the existing math/big path unchanged. Zero, negative instances and the absolute-value handling of the divisor give the same answers as today.
- testing.AllocsPerRun reports 0 for evalMultipleOf on small passing numbers, and 0 for the yes/no decision on small failing numbers. Writing the error entry for a failing number may allocate exactly as today. This allocation test is gated under -race the same way the existing allocation tests are.
- A new random differential test compares the int64 path against the math/big path on at least 1,000,000 inputs. It uses the same fixed seed on every run, including -race runs, and prints the seed and the failing instance and divisor when a case fails. The inputs include: at least one third exact multiples (k × divisor, still in the small range); divisors with high powers of 2 and 5 (for example 0.125, 0.0625, 1024, 3125 and 0.00032); negative divisors under today's absolute-value rule; and the int64 mantissa and ±18 exponent boundaries on both the instance and the divisor.
- All existing multipleOf tests and testdata cases are unchanged and pass. No existing test is removed or weakened. go test ./... and go test -race ./... pass.
- CLI stdout, stderr and exit code are byte-identical to the committed v1.1 golden files in testdata/golden/, in flag, basic, detailed and verbose. No golden file is regenerated or edited. Any difference fails the story, and it stops.
- A new root test file allocs_multipleof_test.go holds this story's re-measured allocation limits for the guarded scenarios. Each limit is the measured count plus about 10% headroom, and the measured value is recorded next to it. It reuses the helpers and race gating of allocs_test.go and allocs_race_on/off_test.go read-only. allocs_test.go is not modified.
- A new file docs/benchmarks/r3-210-multipleof.md records before medians (main after #205) and after medians (this story) from go test -bench -count=10 on the same machine. It also records the verdicts: AllocsPerRun = 0, and the Validate5_5MB allocs/op drop caused by this story compared with the required drop (v1.1 count minus 6.25 M). It notes that timings were taken during parallel builds and will be re-measured by the operator at Checkpoint 3. A missed target is recorded and the story is held for the operator's decision.
- The story writes no file claimed by #206, #207 or #211, and does not write allocs_test.go or docs/benchmarks.md. It uses the standard library only, and go.mod is unchanged.

## Non-goals

- Changing number storage or decoding beyond adding a non-allocating accessor
- Speeding up minimum, maximum, exclusive*, enum, const or uniqueItems
- Making the multipleOf error entry allocation-free (error messages and reports stay identical)
- Editing allocs_test.go or docs/benchmarks.md (#208 and #209 own them this round)
- Any float-based shortcut, or preparing for other number representations

## Tasks

- **T1** — Non-allocating inline mantissa/exponent accessor
  - purpose: Expose the small decimal form of a jsonvalue number without allocating.
  - scope: internal/jsonvalue/number_inline.go, internal/jsonvalue/number_inline_test.go
  - done when: The new tests prove that the exact value round-trips against the big form, that big-form and out-of-range numbers report 'not small', and that AllocsPerRun is 0.
  - delivers: AC1
- **T2** — int64 multipleOf decision with compile-time divisor small form
  - purpose: Decide divisibility with integers only on the common path, and fall back to math/big otherwise.
  - scope: internal/keywords/validation/multipleof.go, internal/keywords/validation/multipleof_int64.go, internal/keywords/validation/multipleof_int64_test.go
  - done when: Fixed edge-case tests pass: overflow fallback, zero, negatives, powers of 2 and 5, and the ±18 boundaries. Existing multipleOf tests pass unchanged. A grep shows no float operations in the decision code.
  - depends on: T1
  - delivers: AC2, AC3
- **T3** — Random differential test, 1M inputs, fixed seed
  - purpose: Prove that the int64 path agrees with the math/big path.
  - scope: internal/keywords/validation/multipleof_diff_test.go
  - done when: The test runs at least 1,000,000 cases with the input mix from criterion 5, in both normal and -race runs. On failure it prints the seed and the failing case.
  - depends on: T2
  - delivers: AC5
- **T4** — Zero-allocation test and lowered root allocation limits
  - purpose: Guard the 0-allocation decision and record this story's lowered limits in its own file.
  - scope: internal/keywords/validation/multipleof_allocs_test.go, internal/keywords/validation/multipleof_race_on_test.go, internal/keywords/validation/multipleof_race_off_test.go, allocs_multipleof_test.go
  - done when: AllocsPerRun is 0 for small passing evalMultipleOf calls and for the small failing decision. The new root limits are the measured value plus about 10%, with the measured value noted. allocs_test.go is untouched.
  - depends on: T2
  - delivers: AC4, AC8
- **T5** — Verify, check the golden files and record the benchmarks
  - purpose: Confirm identical behaviour and record before/after numbers and verdicts.
  - scope: docs/benchmarks/r3-210-multipleof.md
  - done when: go test ./... and go test -race ./... pass, the golden comparison passes, the before/after medians at count=10 are recorded, and the target verdicts are stated. If a target is missed, the story is held.
  - depends on: T3, T4
  - delivers: AC6, AC7, AC9, AC10
