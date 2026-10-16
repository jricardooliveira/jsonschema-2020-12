# PLAN #14: Annotation collection: metadata, format, content and unknown keywords, evaluated properties and items

- Date: 2026-10-16
- Status: proposed (pending Checkpoint 2 sign-off)
- Design: see `ADR-014-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `internal/eval/annotations.go`
- `internal/eval/annotations_test.go`
- `internal/eval/eval.go`
- `internal/keywords/registry.go`
- `internal/compile/compile.go`
- `internal/keywords/annotate.go`
- `internal/keywords/annotate_test.go`
- `internal/keywords/metadata/annotations.go`
- `internal/keywords/metadata/annotations_test.go`
- `internal/keywords/formatannotation/format.go`
- `internal/keywords/formatannotation/format_test.go`
- `internal/keywords/content/content.go`
- `internal/keywords/content/content_test.go`
- `internal/keywords/applicator/properties.go`
- `internal/keywords/applicator/patternproperties.go`
- `internal/keywords/applicator/additionalproperties.go`
- `internal/keywords/applicator/prefixitems.go`
- `internal/keywords/applicator/items.go`
- `internal/keywords/applicator/contains.go`
- `annotations.go`
- `jsonschema.go`
- `options.go`
- `annotations_test.go`
- `internal/output/basic.go`
- `cmd/jsonschema/run.go`
- `cmd/jsonschema/annotations_test.go`
- `testdata/annotations/metadata.json`
- `testdata/annotations/format.json`
- `testdata/annotations/content.json`
- `testdata/annotations/unknown.json`
- `testdata/annotations/applicators.json`

Modules the architecture expects to touch:

- `internal/eval/annotations.go`
- `internal/eval/annotations_test.go`
- `internal/eval/eval.go`
- `internal/keywords/registry.go`
- `internal/compile/compile.go`
- `internal/keywords/annotate.go`
- `internal/keywords/annotate_test.go`
- `internal/keywords/metadata/annotations.go`
- `internal/keywords/metadata/annotations_test.go`
- `internal/keywords/formatannotation/format.go`
- `internal/keywords/formatannotation/format_test.go`
- `internal/keywords/content/content.go`
- `internal/keywords/content/content_test.go`
- `internal/keywords/applicator/properties.go`
- `internal/keywords/applicator/patternproperties.go`
- `internal/keywords/applicator/additionalproperties.go`
- `internal/keywords/applicator/prefixitems.go`
- `internal/keywords/applicator/items.go`
- `internal/keywords/applicator/contains.go`
- `annotations.go`
- `jsonschema.go`
- `options.go`
- `annotations_test.go`
- `internal/output/basic.go`
- `cmd/jsonschema/run.go`
- `cmd/jsonschema/annotations_test.go`
- `testdata/annotations/metadata.json`
- `testdata/annotations/format.json`
- `testdata/annotations/content.json`
- `testdata/annotations/unknown.json`
- `testdata/annotations/applicators.json`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — Annotation collection core and evaluation tree in internal/eval
2. **T2** — Shared annotate-only helper with metadata, format and content keywords  _(after T1)_
3. **T3** — Evaluated-location annotations from object and array applicators  _(after T1)_
4. **T4** — Library API: Result.Annotations() and the annotation-limit option  _(after T1)_
5. **T5** — Annotations in basic output and the CLI limit flag  _(after T4)_
6. **T6** — Spec-derived annotation test cases  _(after T2, T3, T4)_

## Risks

- Recording every annotation on large documents costs time and memory. This is accepted by the operator (always collect) and bounded by the limit.
- eval.go and compile.go are crowded files, so changes there must stay minimal and only wire in the code in annotations.go.
- The applicator files from stories 183 and 184 may not exist yet, or may be named differently, when this story starts.

Implementation constraints the coder must honor:

- Only the standard library; go.mod stays without require lines.
- Annotation values only ever come from the schema. Instance values never appear in an annotation, a report or an error.
- Annotations from failing subschemas are discarded and appear nowhere, not even in the internal tree.
- The compiled Schema and everything reachable from it are immutable. Annotation state lives in the per-call evaluation context, and tests run under -race.
- Tests read DefaultMaxAnnotations instead of repeating 10000000.
- main.go is not edited, and no central keyword list or switch is edited: keywords register from init().
- The ordering is deterministic and covered by a test.

- DB impact: no · API impact: yes: adds Result.Annotations(), the Annotation type, Options.MaxAnnotations and a CLI limit flag; passing basic reports gain annotation units · Migration needed: no

Breaking changes:

- A passing 'basic' report now lists annotation units, so it is longer than before (operator-approved).

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: go test -race ./internal/eval/... ./internal/keywords/ ./internal/compile/... passes. The tests show that annotations are dropped under a failing subschema, that the ordering holds, that the tree contains passing parts of a failing evaluation, that the limit error fires, and that an unknown keyword is collected.
- **T2**: Unit tests show that each keyword annotates its value and never fails, that unknown formats are accepted, and that contentSchema is not applied to the instance.
- **T3**: The spec cases in testdata/annotations/applicators.json pass.
- **T4**: Root package tests show an ordered list on a passing result, an empty list on a failing result, and a LimitError when the limit is lowered and exceeded.
- **T5**: CLI contract tests check passing and failing basic output, flag output, the limit flag (exit code 2 when exceeded and for non-positive values), and that no instance values are printed.
- **T6**: go test -race ./... passes, and every case cites a spec section.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
