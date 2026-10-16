# PLAN #4: Numeric validation keywords: multipleOf, maximum, exclusiveMaximum, minimum, exclusiveMinimum

- Date: 2026-10-16
- Status: proposed (pending Checkpoint 2 sign-off)
- Design: see `ADR-004-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `internal/keywords/validation/numeric.go`
- `internal/keywords/validation/multipleof.go`
- `internal/keywords/validation/multipleof_test.go`
- `testdata/validation/multipleOf.json`
- `internal/keywords/validation/maximum.go`
- `internal/keywords/validation/maximum_test.go`
- `testdata/validation/maximum.json`
- `internal/keywords/validation/minimum.go`
- `internal/keywords/validation/minimum_test.go`
- `testdata/validation/minimum.json`
- `internal/keywords/validation/exclusivemaximum.go`
- `internal/keywords/validation/exclusivemaximum_test.go`
- `testdata/validation/exclusiveMaximum.json`
- `internal/keywords/validation/exclusiveminimum.go`
- `internal/keywords/validation/exclusiveminimum_test.go`
- `testdata/validation/exclusiveMinimum.json`

Modules the architecture expects to touch:

- `internal/keywords/validation/numeric.go`
- `internal/keywords/validation/multipleof.go`
- `internal/keywords/validation/multipleof_test.go`
- `internal/keywords/validation/maximum.go`
- `internal/keywords/validation/maximum_test.go`
- `internal/keywords/validation/minimum.go`
- `internal/keywords/validation/minimum_test.go`
- `internal/keywords/validation/exclusivemaximum.go`
- `internal/keywords/validation/exclusivemaximum_test.go`
- `internal/keywords/validation/exclusiveminimum.go`
- `internal/keywords/validation/exclusiveminimum_test.go`
- `testdata/validation/multipleOf.json`
- `testdata/validation/maximum.json`
- `testdata/validation/minimum.json`
- `testdata/validation/exclusiveMaximum.json`
- `testdata/validation/exclusiveMinimum.json`
- `internal/jsonvalue/number.go`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — Shared exact-number helpers for the validation vocabulary
2. **T2** — multipleOf keyword  _(after T1)_
3. **T3** — maximum and minimum keywords  _(after T1)_
4. **T4** — exclusiveMaximum and exclusiveMinimum keywords  _(after T1)_
5. **T5** — Verify no cap, no flag and the race-free full suite  _(after T2, T3, T4)_

## Risks

- internal/jsonvalue/number.go is outside the reservation. It needs one small read-only accessor (Decimal), which ADR #2 explicitly gives to story 181. The factory should widen the reservation for this file. It is not a crowded file.
- No code path in these five keywords scales with the exponent value. Cmp aligns mantissas only when the adjusted exponents are equal, which is bounded by the digit cap. multipleOf uses only mantissa-sized big.Int Mod/Quo and int64 comparisons of exponents. Rat() is never called.
- The valuation form of multipleOf is less obvious than division, so a wrong edge case could give a wrong answer. Mitigation: the testdata covers both signs of d, divisors with 2/5 factors (0.5, 2.5, 0.125), coprime divisors (3, 1.5, 0.3), zero instances, big integers and the 1.0/2.0 cases. The test file also checks it against big.Rat Quo/IsInt for small values only, where Rat is safe.
- int64 overflow when comparing ea−eb at extreme exponents with absurdly raised caps. The comparison must be written overflow-safe, and a test covers it with exponents near the cap.

Implementation constraints the coder must honor:

- Standard library only (math/big). go.mod unchanged.
- Do not edit internal/keywords/registry.go, internal/keywords/all/all.go, internal/jsonvalue/decode.go, internal/eval/eval.go or cmd/jsonschema/main.go.
- No number-size cap, option or flag in this story. Rely on story 179.
- Never use float64 for comparisons or division.
- Never call Number.Rat() and never construct a big.Rat in these keywords or in numeric.go. Never call pow10 or big.Int.Exp with an exponent taken from a number.
- Bounds use jsonvalue.Cmp only.
- multipleOf uses the (mant, exp) valuation check: r_b | ma, v2(ma)+d ≥ v2(mb), v5(ma)+d ≥ v5(mb), with d = ea−eb compared overflow-safe in int64. No step may scale with the exponent value; work must depend only on the mantissa digit counts.
- numeric.go only type-asserts *jsonvalue.Number. A non-number keyword value returns a compile error naming the keyword, which is a schema error with exit 2. A multipleOf divisor ≤ 0 is also a compile error.
- The number.go accessor is read-only and returns a copy of mant. Add no other accessor.
- Error messages must not include the instance value. Including the schema's limit value is allowed, since it is not document data.
- Compiled keyword state is immutable and safe for concurrent use.
- Tests ask the code or the test data for values rather than repeating limits, and each test file cites the 2020-12 validation spec section.
- multipleof_test.go includes a hostile-scale case. With raised caps through jsonschema.Options (MaxNumberExponent), instances such as 1e999999999 and 1e-999999999 against multipleOf 0.1 / 3, and against maximum 1, finish fast with the correct exact answer. It also includes a -race concurrent Validate on one compiled schema.

- DB impact: no · API impact: no · Migration needed: no

Breaking changes:

- (none)

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: Package builds. The helpers are used by the keyword files in T2 and T3 and are covered by their tests.
- **T2**: go test ./internal/keywords/validation/ -run MultipleOf passes. Cases include 0.3/0.1 valid, 0.35/0.1 invalid, big integers, 2.0/1 valid, non-number instances ignored, and divisor 0 or -1 as a schema error.
- **T3**: go test for Maximum and Minimum passes. Cases include the boundary as valid, 1.0 vs 1, big integers differing by 1, long decimals differing in the last digit, and non-numbers ignored.
- **T4**: go test for ExclusiveMaximum and ExclusiveMinimum passes. Cases include the boundary as invalid, 1.0 at limit 1 as invalid, big-number and decimal edges, and non-numbers ignored.
- **T5**: go test -race ./... passes. A git diff shows no change to internal/jsonvalue/decode.go, cmd/jsonschema/main.go, go.mod or the keyword registry.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
