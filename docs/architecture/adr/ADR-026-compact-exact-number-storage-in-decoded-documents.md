# ADR #26: Compact exact number storage in decoded documents (fix 5)

- Date: 2026-10-17
- Story: #26
- Status: passed gate 2 with nothing for the operator to decide (run #161, 2026-10-17 13:04 UTC)

## Decision
The change is confined to jsonvalue.Number's private fields. Number becomes {mant int64, exp int64, big *big.Int (nil on the fast path), original text as today, lazily computed and counted rat}. The representation is canonical. Trailing zeros of the mantissa are folded into the exponent, a big mantissa that fits int64 is demoted to the fast form, and zero has a single form. Equality and hashing can then compare (mant, exp) directly for fast/fast pairs and use big alignment only within the digit cap. Cmp on fast/fast aligns exponents only when the shift stays within int64 decimal digits; otherwise it falls back to the existing capped big alignment, never a large 10^n. IsInteger reduces to exp >= 0 on the canonical form. multipleOf keeps the exact decimal/modular method, and a fast/fast operand pair may take an int64 modular shortcut that is still exact. Everything else (keywords, eval, output, root API) stays untouched because the Go shape of decoded documents is unchanged.

## Affected modules
- internal/jsonvalue/number.go
- internal/jsonvalue/number_canonical.go
- internal/jsonvalue/decode.go
- internal/jsonvalue/equal.go
- internal/jsonvalue/number_fastpath_test.go
- internal/jsonvalue/equal_fastpath_test.go
- internal/keywords/validation/multipleof.go
- internal/keywords/validation/type_numberpath_test.go
- internal/keywords/validation/multipleof_numberpath_test.go
- internal/keywords/validation/maximum_numberpath_test.go
- internal/keywords/validation/exclusivemaximum_numberpath_test.go
- internal/keywords/validation/minimum_numberpath_test.go
- internal/keywords/validation/exclusiveminimum_numberpath_test.go
- internal/keywords/validation/enum_numberpath_test.go
- internal/keywords/validation/const_numberpath_test.go
- internal/keywords/validation/uniqueitems_numberpath_test.go
- number_limits_regression_test.go
- number_race_test.go
- docs/benchmarks.md

## Data / API impact
- DB impact: no
- API impact: no
- Migration needed: no

## Constraints for implementation
- Never use float64 to decide an answer; do not add a float64 field.
- Never compute 10^n for a large n. Exponent alignment stays within the existing digit cap.
- Digit and exponent caps and their LimitErrors are identical, and are read from internal/jsonvalue/limits.go rather than repeated as numbers in tests.
- Do not edit internal/jsonvalue/number_test.go, equal_test.go or any existing test file.
- Canonical() and String() output are byte-identical to today for every input.
- Do not touch internal/eval/eval.go, internal/compile/compile.go, jsonschema.go, options.go, cmd/jsonschema/run.go or any keyword other than multipleof.go.
- No allocation-limit test is added or changed.
- The 'before' benchmark numbers come from running the current benchmarks, including retained heap, on the old code on the same machine.
- Standard library only; go.mod is unchanged.

## Risks
- Non-canonical representations (for example 10 as mant=10 vs mant=1,exp=1) would break equality or uniqueItems hashing; canonicalisation must be total.
- An int64 overflow in mantissa alignment for Cmp or multipleOf would silently corrupt answers; boundary tests at int64 max must cover this.
- Number-text preservation may need the original text, which limits the per-number memory saving.
- docs/benchmarks.md is crowded (6 stories), and stories 199 to 202 also append before/after numbers there, so their before baselines interleave.
- Retained-heap measurement noise could hide a small gain.

## Breaking changes
- none

## Sensitivity
- none
