# PLAN #6: 2020-12 object keywords: properties, patternProperties, additionalProperties, propertyNames, dependentSchemas, maxProperties, minProperties, required, dependentRequired

- Date: 2026-10-16
- Status: passed gate 2 with nothing for the operator to decide (run #141, 2026-10-16 20:20 UTC)
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
- `internal/eval/evaluated_items.go`
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

- internal/eval/evaluated_items.go is outside the original reservation. It must gain a property-entry field so that one Mark/Rollback log covers both item and property annotations. Without that, a failed subschema would leak property annotations, or the crowded registry.go would need a second record. The edit is small, and the file is not crowded.
- internal/eval/eval.go is no longer needed, because attachItems already wires the record. The reservation can release it.
- The acceptance criterion and T2/T5 wording ('order the keys appear in the document') are superseded by the operator's ordering rule (sorted keyword names, then sorted member names, then schema order for missing names). Reviewers should check against the ADR rule, not the original story text.
- patternProperties repeats about 6 lines of size check, Translate and Compile from internal/keywords/validation/pattern.go, in internal/keywords/applicator/patternproperties.go. Moving them into ecmaregex would change the translation package, which is a non-goal. Follow-up: share a CompilePattern helper.
- Story 191 may define its own annotation mechanism. It should build on the EvaluatedItems log (evaluated_items.go and evaluated.go) rather than duplicate it.

Implementation constraints the coder must honor:

- Go standard library only, and go.mod is not edited.
- The compiled Schema stays immutable. Compiled data (subschema nodes, *regexp.Regexp, property-name sets, required lists) is built once in Compile and only read afterwards. Evaluated names live only in ctx.Items. This is verified with go test -race.
- Messages never contain instance values or keys. Keys appear only in the JSON Pointer instance location, with ~0/~1 escaping.
- patternProperties, and additionalProperties for its sibling patterns, must run the size check against validation.DefaultMaxPatternSize, then call ecmaregex.Translate and regexp.Compile. They must not reimplement translation. An untranslatable or uncompilable pattern is a compile error (exit code 2).
- Ordering rule (operator decision, record it in the ADR): keywords are evaluated and reported in sorted keyword-name order. Within one keyword, member failures come in sorted member-name order (sort.Strings on a copy of the map keys, never Go map range order), and dependentSchemas/dependentRequired walk their trigger names in sorted order. Missing required and dependentRequired names are reported in schema array order. The decoder (internal/jsonvalue/decode.go) is not touched.
- Document the ordering rule in a doc comment in properties.go.
- additionalProperties takes its siblings from CompileContext.Siblings, never from the evaluated record, because it is evaluated before properties and patternProperties.
- Keywords reach the record only through a local interface type-asserted from ctx.Items. registry.go and eval.go are not edited.
- Tests read limits and messages from the code where they are defined (for example DefaultMaxPatternSize), not from repeated literals.
- The determinism test renders basic output twice for a document with more than 20 keys inserted out of order and compares the bytes. A second test asserts the unit order: keyword order, then sorted member names, then missing names in schema order.

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
