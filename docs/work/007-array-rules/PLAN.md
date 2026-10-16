# PLAN #7: 2020-12 array keywords: prefixItems, items, contains, maxContains, minContains, maxItems, minItems, uniqueItems

- Date: 2026-10-16
- Status: passed gate 2 with nothing for the operator to decide (run #140, 2026-10-16 19:56 UTC)
- Design: see `ADR-007-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `internal/eval/evaluated_items.go`
- `internal/eval/evaluated_items_test.go`
- `internal/eval/evaluated.go`
- `internal/eval/eval.go`
- `internal/keywords/applicator/prefixitems.go`
- `internal/keywords/applicator/prefixitems_test.go`
- `internal/keywords/applicator/items.go`
- `internal/keywords/applicator/items_test.go`
- `testdata/applicator/prefixItems.json`
- `testdata/applicator/items.json`
- `internal/keywords/applicator/contains.go`
- `internal/keywords/applicator/contains_test.go`
- `internal/keywords/validation/mincontains.go`
- `internal/keywords/validation/mincontains_test.go`
- `internal/keywords/validation/maxcontains.go`
- `internal/keywords/validation/maxcontains_test.go`
- `testdata/applicator/contains.json`
- `testdata/validation/minContains.json`
- `testdata/validation/maxContains.json`
- `internal/keywords/validation/maxitems.go`
- `internal/keywords/validation/maxitems_test.go`
- `internal/keywords/validation/minitems.go`
- `internal/keywords/validation/minitems_test.go`
- `testdata/validation/maxItems.json`
- `testdata/validation/minItems.json`
- `internal/keywords/validation/uniqueitems.go`
- `internal/keywords/validation/uniqueitems_test.go`
- `testdata/validation/uniqueItems.json`

Modules the architecture expects to touch:

- `internal/eval/evaluated_items.go`
- `internal/eval/evaluated_items_test.go`
- `internal/eval/evaluated.go`
- `internal/eval/eval.go`
- `internal/keywords/applicator/prefixitems.go`
- `internal/keywords/applicator/prefixitems_test.go`
- `internal/keywords/applicator/items.go`
- `internal/keywords/applicator/items_test.go`
- `internal/keywords/applicator/contains.go`
- `internal/keywords/applicator/contains_test.go`
- `internal/keywords/validation/mincontains.go`
- `internal/keywords/validation/mincontains_test.go`
- `internal/keywords/validation/maxcontains.go`
- `internal/keywords/validation/maxcontains_test.go`
- `internal/keywords/validation/maxitems.go`
- `internal/keywords/validation/maxitems_test.go`
- `internal/keywords/validation/minitems.go`
- `internal/keywords/validation/minitems_test.go`
- `internal/keywords/validation/uniqueitems.go`
- `internal/keywords/validation/uniqueitems_test.go`
- `testdata/applicator/prefixItems.json`
- `testdata/applicator/items.json`
- `testdata/applicator/contains.json`
- `testdata/validation/minContains.json`
- `testdata/validation/maxContains.json`
- `testdata/validation/maxItems.json`
- `testdata/validation/minItems.json`
- `testdata/validation/uniqueItems.json`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — Per-evaluation evaluated-items record
2. **T4** — maxItems and minItems
3. **T5** — uniqueItems with a fingerprint-grouped duplicate check
4. **T2** — prefixItems and items  _(after T1)_
5. **T3** — contains, minContains and maxContains  _(after T1)_

## Risks

- The fingerprint for numbers must normalise exact values (1, 1.0, 1e0) or real duplicates fall into different groups. Tests must cover this.
- Evaluating every element for contains and items increases work on huge arrays. The evaluation-step limits arrive in story 193.

Implementation constraints the coder must honor:

- No instance values in any error message. Messages give positions and counts only.
- No short-circuit: prefixItems, items, contains and uniqueItems examine the whole array.
- No global mutable state. The comparison counter is returned per call.
- Use the exact number model and deep equality from internal/jsonvalue.
- No wall-clock assertions in tests.
- At most one wiring line in internal/eval/eval.go.
- Standard library only. go.mod unchanged.

- DB impact: no · API impact: no · Migration needed: no

Breaking changes:

- (none)

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: evaluated_items_test.go passes under go test -race. The record lives only in the per-call context, and eval.go changes by at most one wiring line.
- **T4**: Spec-derived boundary cases pass, including 2.0 as a limit and non-array instances.
- **T5**: Cases pass for 1 and 1.0, key-order-independent objects, nested arrays and uniqueItems:false. Grouped errors list positions only. The 100,000-element test checks the exact groups and a comparison count far below n*n/100.
- **T2**: The spec-derived positive and negative cases pass, including items:false after prefixItems, non-array instances, multiple failing positions and annotation values.
- **T3**: Cases pass for minContains 0, maxContains bounds, min and max without contains, a single error with matched and required counts and no values, and the indices or true annotation.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
