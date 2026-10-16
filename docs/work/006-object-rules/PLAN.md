# PLAN #6: 2020-12 object keywords: properties, patternProperties, additionalProperties, propertyNames, dependentSchemas, maxProperties, minProperties, required, dependentRequired

- Date: 2026-10-16
- Status: proposed (pending Checkpoint 2 sign-off)
- Design: see `ADR-006-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `internal/eval/evaluated.go`
- `internal/eval/evaluated_test.go`
- `internal/eval/eval.go`
- `internal/keywords/applicator/properties.go`
- `internal/keywords/applicator/properties_test.go`
- `internal/keywords/applicator/patternproperties.go`
- `internal/keywords/applicator/patternproperties_test.go`
- `internal/keywords/applicator/additionalproperties.go`
- `internal/keywords/applicator/additionalproperties_test.go`
- `testdata/applicator/properties.json`
- `testdata/applicator/patternproperties.json`
- `testdata/applicator/additionalproperties.json`
- `internal/keywords/applicator/propertynames.go`
- `internal/keywords/applicator/propertynames_test.go`
- `internal/keywords/applicator/dependentschemas.go`
- `internal/keywords/applicator/dependentschemas_test.go`
- `testdata/applicator/propertynames.json`
- `testdata/applicator/dependentschemas.json`
- `internal/keywords/validation/maxproperties.go`
- `internal/keywords/validation/maxproperties_test.go`
- `internal/keywords/validation/minproperties.go`
- `internal/keywords/validation/minproperties_test.go`
- `internal/keywords/validation/required.go`
- `internal/keywords/validation/required_test.go`
- `internal/keywords/validation/dependentrequired.go`
- `internal/keywords/validation/dependentrequired_test.go`
- `testdata/validation/maxproperties.json`
- `testdata/validation/minproperties.json`
- `testdata/validation/required.json`
- `testdata/validation/dependentrequired.json`

Modules the architecture expects to touch:

- `internal/eval/evaluated.go`
- `internal/eval/evaluated_test.go`
- `internal/eval/eval.go`
- `internal/keywords/applicator/properties.go`
- `internal/keywords/applicator/properties_test.go`
- `internal/keywords/applicator/patternproperties.go`
- `internal/keywords/applicator/patternproperties_test.go`
- `internal/keywords/applicator/additionalproperties.go`
- `internal/keywords/applicator/additionalproperties_test.go`
- `internal/keywords/applicator/propertynames.go`
- `internal/keywords/applicator/propertynames_test.go`
- `internal/keywords/applicator/dependentschemas.go`
- `internal/keywords/applicator/dependentschemas_test.go`
- `internal/keywords/validation/maxproperties.go`
- `internal/keywords/validation/maxproperties_test.go`
- `internal/keywords/validation/minproperties.go`
- `internal/keywords/validation/minproperties_test.go`
- `internal/keywords/validation/required.go`
- `internal/keywords/validation/required_test.go`
- `internal/keywords/validation/dependentrequired.go`
- `internal/keywords/validation/dependentrequired_test.go`
- `testdata/applicator/properties.json`
- `testdata/applicator/patternproperties.json`
- `testdata/applicator/additionalproperties.json`
- `testdata/applicator/propertynames.json`
- `testdata/applicator/dependentschemas.json`
- `testdata/validation/maxproperties.json`
- `testdata/validation/minproperties.json`
- `testdata/validation/required.json`
- `testdata/validation/dependentrequired.json`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — Per-evaluation record of evaluated property names
2. **T3** — propertyNames and dependentSchemas
3. **T4** — maxProperties, minProperties, required, dependentRequired
4. **T2** — properties, patternProperties, additionalProperties  _(after T1)_
5. **T5** — All failures and deterministic order  _(after T2, T4)_

## Risks

- If the jsonvalue object model (story 179) stores members in a Go map, document key order is lost, and deterministic ordering would require a change to the crowded internal/jsonvalue/decode.go.
- If the evaluation context from the skeleton has no extension point, internal/eval/eval.go (crowded) must be edited. The edit is kept to one wiring line or field.
- Story 191 may define its own annotation mechanism that overlaps the evaluated-names record. 191 should build on internal/eval/evaluated.go rather than duplicate it.

Implementation constraints the coder must honor:

- Go standard library only, and go.mod is not edited.
- The compiled Schema stays immutable. Evaluated names live only in the per-call context, verified with go test -race.
- Messages never contain instance values or keys. Keys appear only in the JSON Pointer instance location, with ~0/~1 escaping.
- patternProperties must call the existing ECMA-262-to-RE2 translation step and must not reimplement it. Untranslatable patterns are schema errors (exit code 2).
- Tests read limits and messages from the code where defined, instead of repeating literals.
- Iteration follows document key order. Missing fields follow schema array order.

- DB impact: no · API impact: no · Migration needed: no

Breaking changes:

- (none)

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: evaluated_test.go shows names recorded and read back per instance location and schema object, with no shared state between two concurrent evaluations (go test -race).
- **T3**: Tests show a rejected key appearing only in the instance location and never in the message, and dependentSchemas applying only when its trigger is present.
- **T4**: Tests cover boundary counts, several missing fields reported in schema order, a message such as 'missing required property: email', and non-object instances passing.
- **T2**: Keyword tests and testdata cases pass, including non-object instances, boolean subschemas, an untranslatable pattern giving a schema error, and additionalProperties ignoring names matched in the same object only.
- **T5**: A test with two wrong fields and one missing field shows three errors in the required order in basic output, and compares the rendered output from two runs byte for byte.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
