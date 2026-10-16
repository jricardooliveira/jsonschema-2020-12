# PLAN #8: Applicator keywords allOf, anyOf, oneOf, not and if/then/else, with keyword locations rooted at the schema root

- Date: 2026-10-16
- Status: approved by the operator at Checkpoint 2 (run #142, 2026-10-16 20:40 UTC)
- Design: see `ADR-008-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `internal/keywords/registry.go`
- `internal/compile/compile.go`
- `internal/compile/compile_location_test.go`
- `internal/eval/eval.go`
- `internal/eval/eval_subschema_test.go`
- `internal/output/units_test.go`
- `internal/keywords/applicator/subschemas.go`
- `internal/keywords/applicator/allof.go`
- `internal/keywords/applicator/allof_test.go`
- `internal/keywords/applicator/not.go`
- `internal/keywords/applicator/not_test.go`
- `internal/keywords/applicator/anyof.go`
- `internal/keywords/applicator/anyof_test.go`
- `internal/keywords/applicator/oneof.go`
- `internal/keywords/applicator/oneof_test.go`
- `internal/keywords/applicator/if.go`
- `internal/keywords/applicator/if_test.go`
- `internal/keywords/applicator/then.go`
- `internal/keywords/applicator/then_test.go`
- `internal/keywords/applicator/else.go`
- `internal/keywords/applicator/else_test.go`
- `testdata/applicator/allOf.json`
- `testdata/applicator/anyOf.json`
- `testdata/applicator/oneOf.json`
- `testdata/applicator/not.json`
- `testdata/applicator/if-then-else.json`
- `testdata/applicator/boolean-subschemas.json`
- `testdata/applicator/nested.json`
- `spec_applicator_test.go`

Modules the architecture expects to touch:

- `internal/keywords/registry.go`
- `internal/compile/compile.go`
- `internal/compile/compile_location_test.go`
- `internal/eval/eval.go`
- `internal/eval/eval_subschema_test.go`
- `internal/output/units_test.go`
- `internal/keywords/applicator/subschemas.go`
- `internal/keywords/applicator/allof.go`
- `internal/keywords/applicator/allof_test.go`
- `internal/keywords/applicator/anyof.go`
- `internal/keywords/applicator/anyof_test.go`
- `internal/keywords/applicator/oneof.go`
- `internal/keywords/applicator/oneof_test.go`
- `internal/keywords/applicator/not.go`
- `internal/keywords/applicator/not_test.go`
- `internal/keywords/applicator/if.go`
- `internal/keywords/applicator/if_test.go`
- `internal/keywords/applicator/then.go`
- `internal/keywords/applicator/then_test.go`
- `internal/keywords/applicator/else.go`
- `internal/keywords/applicator/else_test.go`
- `testdata/applicator/allOf.json`
- `testdata/applicator/anyOf.json`
- `testdata/applicator/oneOf.json`
- `testdata/applicator/not.json`
- `testdata/applicator/if-then-else.json`
- `testdata/applicator/boolean-subschemas.json`
- `testdata/applicator/nested.json`
- `spec_applicator_test.go`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — Root keyword locations: compile and evaluate subschemas at their schema location
2. **T2** — allOf and not  _(after T1)_
3. **T3** — anyOf and oneOf  _(after T2)_
4. **T4** — if/then/else  _(after T2)_
5. **T5** — Spec-derived applicator cases  _(after T3, T4)_

## Risks

- Story 180 is still in progress: the output-unit type, its units_test.go path and how basic renders nested child failures may change before this story builds.
- Adding location to CompileContext changes the signature seen by the type/enum/const keywords and by approved stories 181-184, which all write registry.go and compile.go.
- 'if' needing sibling access may require exposing the parent schema object in CompileContext, if it does not already.

Implementation constraints the coder must honor:

- Reports never include instance values; only instanceLocation, keywordLocation and error text.
- No central keyword list or switch; registration only via init() in each keyword's file.
- No new dependency; go.mod unchanged.
- anyOf and oneOf evaluate every choice (operator decision for unevaluated*).
- Compiled nodes and their locations are immutable after Compile returns; per-call state stays in the evaluation context.
- Tests ask the code for messages and locations where defined, rather than repeating literals unnecessarily.
- Each JSON case cites a 2020-12 spec section (the speccases loader rejects missing citations).

- DB impact: no · API impact: no · Migration needed: no

Breaking changes:

- Keyword locations of nested units change from child-relative to schema-root-relative; story 180's TestUnitOrderAndNoValues assertion is updated accordingly.

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: go test -race ./internal/... passes. A child const inside allOf/0 reports keywordLocation /allOf/0/const. TestUnitOrderAndNoValues asserts the rooted location.
- **T2**: allof_test.go and not_test.go pass, covering valid, invalid, boolean subschemas, report entries without values, and schema errors.
- **T3**: anyof_test.go and oneof_test.go pass, including the every-choice-evaluated test and the 'choices 0 and 2 both matched' message.
- **T4**: if/then/else tests pass: then/else ignored without if, if failure alone is valid, and a failing branch reports the branch entry plus its failures and nothing for if.
- **T5**: go test ./... runs every testdata/applicator case, each with a citation; all pass.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
