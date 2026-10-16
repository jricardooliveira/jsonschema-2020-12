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
- `numberlimits.go`
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

- numberlimits.go is outside the reservation. The Options struct, and its MaxOutputUnits field, are defined there, so the MaxAnnotations field must be added there too (one line plus a comment). options.go keeps the maxAnnotations() method.
- The plan now has 32 files, which is the cap.
- Recording every annotation on large documents costs time and memory. The operator accepted this ('always collect'), and the limit bounds it.
- eval.go and compile.go are crowded files. Changes there must stay minimal and only wire in the code in annotations.go.
- There is no docs/ file in the plan for the 'names, not values' rule, to stay within the cap. The rule is recorded in the Annotation godoc and in the ADR.

Implementation constraints the coder must honor:

- Only the standard library; go.mod stays without require lines.
- Annotation values come only from the schema. Evaluated property names may appear, because they are already in instanceLocation. Instance VALUES never appear in an annotation, a report or an error.
- Annotations from failing subschemas are discarded and appear nowhere, not even in the internal tree.
- The compiled Schema and everything reachable from it are immutable. Annotation state lives in the per-call evaluation context, and tests run under -race.
- MaxAnnotations copies MaxOutputUnits exactly. Unset or 0 means DefaultMaxAnnotations, a negative value is a compile error, and an explicit CLI value of 0 or less is a usage error. There is never a 'collection off' mode.
- The counter increments on every recorded annotation, including ones later dropped. Going over the limit gives LimitError{Limit: "annotations"}.
- The sort is stable. It compares instance location, then keyword location, both as decoded JSON Pointer tokens (byte order, prefix first), then absolute keyword location, then keyword name, then evaluation sequence. Tests include ties from different dynamic paths and a ~1-escaped token.
- The NDJSON wrapper and the failing-report shape are unchanged.
- Tests read DefaultMaxAnnotations instead of repeating 10000000.
- main.go is not edited, and no central keyword list or switch is edited: keywords register from init().

- DB impact: no · API impact: yes · Migration needed: no

Breaking changes:

- A passing 'basic' report now lists annotation units, so it is longer than before. The operator approved this.

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
