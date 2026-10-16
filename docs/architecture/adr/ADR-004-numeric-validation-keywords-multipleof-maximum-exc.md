# ADR #4: Numeric validation keywords: multipleOf, maximum, exclusiveMaximum, minimum, exclusiveMinimum

- Date: 2026-10-16
- Story: #4
- Status: proposed (pending architecture sign-off)

## Decision
Five new keyword files in the existing validation vocabulary package. Each self-registers via init() into internal/keywords/registry.go, following the pattern of the type, enum and const keywords from story 179. The vocabulary package is already blank-imported in internal/keywords/all/all.go, so no hub is edited. Exact arithmetic uses the big.Rat held by the internal/jsonvalue number model from story 179. One small in-package helper file (numeric.go) is shared by the five keywords so the conversion logic is not repeated. multipleOf: new(big.Rat).Quo(instance, divisor).IsInt(). Bounds: big.Rat.Cmp. Number-size safety comes entirely from the cap enforced by the story 179 decoder. No crowded file (main.go, eval.go, decode.go) is written.

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

## Data / API impact
- DB impact: no
- API impact: no
- Migration needed: no

## Constraints for implementation
- Standard library only (math/big). go.mod unchanged.
- Do not edit internal/keywords/registry.go, internal/keywords/all/all.go, internal/jsonvalue/decode.go, internal/eval/eval.go or cmd/jsonschema/main.go.
- No number-size cap, option or flag in this story. Rely on story 179.
- Never use float64 for comparisons or division.
- Error messages must not include the instance value. Including the schema's limit value is allowed, since it is not document data.
- Compiled keyword state is immutable and safe for concurrent use.
- Tests ask the code or the test data for values rather than repeating limits, and each test file cites the 2020-12 validation spec section.

## Risks
- The exact accessor exposed by the story 179 number model (for example a Rat() method or exported field) is not yet known. If it returns a shared *big.Rat, the keywords must not mutate it and should use fresh result values.
- big.Rat division on numbers near the story 179 size cap may still be slow. Its cost is bounded by that cap.

## Breaking changes
- none

## Sensitivity
- security
