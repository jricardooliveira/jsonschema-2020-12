# ADR #4: Numeric validation keywords: multipleOf, maximum, exclusiveMaximum, minimum, exclusiveMinimum

- Date: 2026-10-16
- Story: #4
- Status: passed gate 2 with nothing for the operator to decide (run #139, 2026-10-16 19:57 UTC)

## Decision
This revision only replaces the big.Rat arithmetic, as the boundary review asked. Every file, name, constraint and test from the reviewed design stays the same. I checked the existing code: internal/jsonvalue/number.go defines *Number as {neg, mant big.Int, exp int64, ndig}. The value is (-1)^neg × mant × 10^exp, mant has no trailing zeros, and zero has a single form. Cmp(a, b) compares the sign first and then the adjusted exponent (ndig+exp). It aligns mantissas only when the adjusted exponents are equal, so its cost is bounded by the digit cap. Rat() exists but must not be used here. Keywords in internal/keywords/validation register from init() with VocabularyURI, the same way as const.go and type.go. Compile errors are plain errors that name the keyword (e.g. fmt.Errorf("multipleOf: must be a number")). Failures go through ctx.Report(msg). The root TestSpecCases in spec_test.go globs testdata/*/*.json, so the new testdata files are found with no edit to the runner.

numeric.go holds one in-package helper: numberValue(keyword string, v any) (*jsonvalue.Number, error). It type-asserts *jsonvalue.Number and returns a schema error that names the keyword when the value is not a number. The evaluate functions type-assert the instance to *jsonvalue.Number and return valid at once for any other type, including the string "5". The helper never calls Rat().

Bounds: maximum is Cmp(inst, lim) <= 0, minimum is >= 0, exclusiveMaximum is < 0, exclusiveMinimum is > 0. 1.0 and 1 normalise to the same form, so 1.0 at limit 1 compares as equal.

multipleOf works on (mant, exp) through a new read-only accessor in number.go. I chose the modular check but did not take the textbook form mant_a×10^(exp_a−e) mod mant_b×10^(exp_b−e) literally. That form multiplies by 10^|exp_a−exp_b|, which grows with the exponent value (1e999999999 against 0.1 would build a billion-digit integer). The check uses an equivalent prime-valuation form instead:

Let a = ma·10^ea and b = mb·10^eb, and d = ea−eb. a/b is an integer iff all three hold:
- (i) r_b divides ma, where r_b is mb with all factors of 2 and 5 removed. This is a big.Int Mod on numbers of at most digit-cap size.
- (ii) v2(ma) + d ≥ v2(mb).
- (iii) v5(ma) + d ≥ v5(mb).

v2 and v5 are counted by repeated division of the mantissa, at most about 3.4×ndig steps each. Conditions (ii) and (iii) compare the int64 d with a small number, written overflow-safe (compare ea against eb plus a small bound, guarding int64 overflow). An instance equal to zero is always a multiple. The divisor is checked at compile time to be greater than zero (sign > 0 via Cmp with a zero Number, or the accessor's sign). This puts the multipleOf arithmetic in validation/multipleof.go.

The only edit outside new files is one small accessor in internal/jsonvalue/number.go. number.go is not a crowded file, and decode.go is not touched. No crowded hub (registry.go, all.go, eval.go, compile.go, main.go, run.go, jsonschema.go) is touched. Skipped: a shared decimal-arithmetic package; add one if a later keyword needs divisibility.

## Affected modules
- internal/keywords/validation/numeric.go
- internal/keywords/validation/multipleof.go
- internal/keywords/validation/multipleof_test.go
- internal/keywords/validation/maximum.go
- internal/keywords/validation/maximum_test.go
- internal/keywords/validation/minimum.go
- internal/keywords/validation/minimum_test.go
- internal/keywords/validation/exclusivemaximum.go
- internal/keywords/validation/exclusivemaximum_test.go
- internal/keywords/validation/exclusiveminimum.go
- internal/keywords/validation/exclusiveminimum_test.go
- testdata/validation/multipleOf.json
- testdata/validation/maximum.json
- testdata/validation/minimum.json
- testdata/validation/exclusiveMaximum.json
- testdata/validation/exclusiveMinimum.json
- internal/jsonvalue/number.go

## Data / API impact
- DB impact: no
- API impact: no
- Migration needed: no

## Constraints for implementation
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

## Risks
- internal/jsonvalue/number.go is outside the reservation. It needs one small read-only accessor (Decimal), which ADR #2 explicitly gives to story 181. The factory should widen the reservation for this file. It is not a crowded file.
- No code path in these five keywords scales with the exponent value. Cmp aligns mantissas only when the adjusted exponents are equal, which is bounded by the digit cap. multipleOf uses only mantissa-sized big.Int Mod/Quo and int64 comparisons of exponents. Rat() is never called.
- The valuation form of multipleOf is less obvious than division, so a wrong edge case could give a wrong answer. Mitigation: the testdata covers both signs of d, divisors with 2/5 factors (0.5, 2.5, 0.125), coprime divisors (3, 1.5, 0.3), zero instances, big integers and the 1.0/2.0 cases. The test file also checks it against big.Rat Quo/IsInt for small values only, where Rat is safe.
- int64 overflow when comparing ea−eb at extreme exponents with absurdly raised caps. The comparison must be written overflow-safe, and a test covers it with exponents near the cap.

## Breaking changes
- none

## Sensitivity
- security
