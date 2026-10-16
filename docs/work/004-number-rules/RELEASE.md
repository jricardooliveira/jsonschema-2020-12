# Release notes — #4: Numeric validation keywords: multipleOf, maximum, exclusiveMaximum, minimum, exclusiveMinimum

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #4
- Status: approved by the operator at Checkpoint 3 (run #139, 2026-10-16 20:04 UTC)

## What changed

Five new numeric validation keywords are now available: multipleOf (exact divisibility), maximum and minimum (inclusive bounds), and exclusiveMaximum and exclusiveMinimum (strict bounds). All comparisons and division use exact big.Int arithmetic — 0.3 is correctly identified as a multiple of 0.1, and huge integers like 12345678901234567890 compare exactly. 1.0 and 1 are treated as equal. Non-number instances (strings, null, objects, arrays) are automatically valid against all numeric keywords.

- internal/keywords/validation/numeric.go: shared helpers for reading numeric keyword values at compile time and checking instance types
- internal/keywords/validation/multipleof.go: exact multipleOf using valuation-based arithmetic to avoid scaling with exponent value
- internal/keywords/validation/maximum.go, minimum.go: inclusive bounds with exact comparison
- internal/keywords/validation/exclusivemaximum.go, exclusiveminimum.go: strict bounds with exact comparison
- internal/keywords/validation/*_test.go (5 files): bounds, edge cases (decimal divisors, big integers, 1.0 vs 1), schema errors, concurrent use, and hostile-scale exponents
- testdata/validation/multipleOf.json, maximum.json, minimum.json, exclusiveMaximum.json, exclusiveMinimum.json: spec-derived positive and negative cases from 2020-12 sections 6.2.1–6.2.5
- internal/keywords/validation/numeric_output_test.go: failure output format (instance and keyword locations, no instance values)
- internal/jsonvalue/number.go: one new read-only Decimal() accessor returning (neg bool, mant *big.Int, exp int64)

## How to verify

- go test ./internal/keywords/validation/ -run MultipleOf: verify 0.3 valid against multipleOf 0.1, 0.35 invalid, big integers exact, divisor 0 or negative rejected at compile
- go test ./internal/keywords/validation/ -run Maximum: verify boundary at 1 is valid, 1.0 at limit 1 valid, big integers differ by 1
- go test ./internal/keywords/validation/ -run Exclusive: verify boundary at 1 is invalid, 1.0 at limit 1 invalid
- go test ./internal/keywords/validation/ -run 'Bounds.*Hostile': verify exponents near ±1e999999999 with MaxNumberExponent cap finish fast with correct answers
- Compile a schema with invalid keyword values (multipleOf: "1", maximum: null) and confirm compile error (exit code 2)
- Validate non-number instances ("5", null, [], {}) against any numeric keyword and confirm all pass
- go test -race ./internal/keywords/validation/: confirm no race conditions in concurrent Validate calls on one compiled schema

## Migration

None. This is a new feature with no data model or schema changes. Existing schemas are unaffected.

## Rollback

Remove internal/keywords/validation/numeric.go, multipleof.go, maximum.go, minimum.go, exclusivemaximum.go, exclusiveminimum.go and all five *_test.go files. Remove testdata/validation/multipleOf.json, maximum.json, minimum.json, exclusiveMaximum.json, exclusiveMinimum.json. In internal/jsonvalue/number.go, remove the Decimal() method added after line 52.

## Known limitations

- Number-size caps and options are owned by story 179; this story adds no cap, flag or library option of its own
- multipleOf uses prime-valuation arithmetic (checking factors of 2 and 5 in the mantissa and comparing exponent differences) instead of literal division, to avoid scaling with extreme exponents like 1e999999999; all cases including decimals and big integers are tested
- These keywords apply only to JSON number instances; all other types (strings, booleans, null, arrays, objects) are automatically valid

## Concerns

- internal/jsonvalue/number.go was edited to add Decimal(), which is outside the normal story file scope. The architecture review and ADR #2 explicitly allow this small read-only accessor for story 181, and number.go is not a crowded hub.
- multipleOf's valuation check is less intuitive than direct division; testdata confirms correctness with decimal divisors, divisors with 2/5 factors, coprime divisors (3, 1.5, 0.3), zero instances, big integers, and a reference check against big.Rat for small values.
