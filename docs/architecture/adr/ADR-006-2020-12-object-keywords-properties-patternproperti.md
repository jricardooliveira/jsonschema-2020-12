# ADR #6: 2020-12 object keywords: properties, patternProperties, additionalProperties, propertyNames, dependentSchemas, maxProperties, minProperties, required, dependentRequired

- Date: 2026-10-16
- Story: #6
- Status: proposed (pending architecture sign-off)

## Decision
There are nine new keyword files. They are split by 2020-12 vocabulary: applicator (properties, patternProperties, additionalProperties, propertyNames, dependentSchemas) and validation (maxProperties, minProperties, required, dependentRequired). Each self-registers through init() into internal/keywords/registry.go. internal/keywords/all/all.go already imports both vocabulary packages, so it is not edited.

The evaluated-property record goes in a new file, internal/eval/evaluated.go. It is a per-call map keyed by instance location and schema object. The crowded internal/eval/eval.go is touched only to add one field or hook to the evaluation context, and only if the skeleton context has no extension point.

additionalProperties finds the sibling properties/patternProperties either at compile time from the same schema object, or at evaluation time from the evaluated record. Compile time is preferred because it does not depend on keyword evaluation order.

Deterministic order requires member iteration in document order, which the jsonvalue object model must preserve.

## Affected modules
- internal/eval/evaluated.go
- internal/eval/evaluated_test.go
- internal/eval/eval.go
- internal/keywords/applicator/properties.go
- internal/keywords/applicator/properties_test.go
- internal/keywords/applicator/patternproperties.go
- internal/keywords/applicator/patternproperties_test.go
- internal/keywords/applicator/additionalproperties.go
- internal/keywords/applicator/additionalproperties_test.go
- internal/keywords/applicator/propertynames.go
- internal/keywords/applicator/propertynames_test.go
- internal/keywords/applicator/dependentschemas.go
- internal/keywords/applicator/dependentschemas_test.go
- internal/keywords/validation/maxproperties.go
- internal/keywords/validation/maxproperties_test.go
- internal/keywords/validation/minproperties.go
- internal/keywords/validation/minproperties_test.go
- internal/keywords/validation/required.go
- internal/keywords/validation/required_test.go
- internal/keywords/validation/dependentrequired.go
- internal/keywords/validation/dependentrequired_test.go
- testdata/applicator/properties.json
- testdata/applicator/patternproperties.json
- testdata/applicator/additionalproperties.json
- testdata/applicator/propertynames.json
- testdata/applicator/dependentschemas.json
- testdata/validation/maxproperties.json
- testdata/validation/minproperties.json
- testdata/validation/required.json
- testdata/validation/dependentrequired.json

## Data / API impact
- DB impact: no
- API impact: no
- Migration needed: no

## Constraints for implementation
- Go standard library only, and go.mod is not edited.
- The compiled Schema stays immutable. Evaluated names live only in the per-call context, verified with go test -race.
- Messages never contain instance values or keys. Keys appear only in the JSON Pointer instance location, with ~0/~1 escaping.
- patternProperties must call the existing ECMA-262-to-RE2 translation step and must not reimplement it. Untranslatable patterns are schema errors (exit code 2).
- Tests read limits and messages from the code where defined, instead of repeating literals.
- Iteration follows document key order. Missing fields follow schema array order.

## Risks
- If the jsonvalue object model (story 179) stores members in a Go map, document key order is lost, and deterministic ordering would require a change to the crowded internal/jsonvalue/decode.go.
- If the evaluation context from the skeleton has no extension point, internal/eval/eval.go (crowded) must be edited. The edit is kept to one wiring line or field.
- Story 191 may define its own annotation mechanism that overlaps the evaluated-names record. 191 should build on internal/eval/evaluated.go rather than duplicate it.

## Breaking changes
- none

## Sensitivity
- pii
