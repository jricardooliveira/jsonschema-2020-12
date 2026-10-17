# PLAN #15: unevaluatedProperties and unevaluatedItems (2020-12 unevaluated vocabulary)

- Date: 2026-10-17
- Status: passed gate 2 with nothing for the operator to decide (run #150, 2026-10-17 00:40 UTC)
- Design: see `ADR-015-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `internal/keywords/registry.go`
- `internal/eval/eval.go`
- `internal/eval/order_test.go`
- `internal/keywords/unevaluated/unevaluatedproperties.go`
- `internal/keywords/unevaluated/unevaluatedproperties_test.go`
- `internal/keywords/unevaluated/unevaluateditems.go`
- `internal/keywords/unevaluated/unevaluateditems_test.go`
- `testdata/unevaluated/unevaluatedProperties.json`
- `testdata/unevaluated/unevaluatedItems.json`

Modules the architecture expects to touch:

- `internal/keywords/registry.go`
- `internal/eval/eval.go`
- `internal/eval/order_test.go`
- `internal/keywords/unevaluated/unevaluatedproperties.go`
- `internal/keywords/unevaluated/unevaluatedproperties_test.go`
- `internal/keywords/unevaluated/unevaluateditems.go`
- `internal/keywords/unevaluated/unevaluateditems_test.go`
- `testdata/unevaluated/unevaluatedProperties.json`
- `testdata/unevaluated/unevaluatedItems.json`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — Evaluate after siblings
2. **T2** — unevaluatedProperties keyword  _(after T1)_
3. **T3** — unevaluatedItems keyword  _(after T1)_
4. **T4** — Spec-derived cases  _(after T2, T3)_

## Risks

- If story 191 does not track evaluated items/properties per instance location, or drops annotations at the wrong granularity, this story cannot be correct without changing 191's files.
- anyOf/oneOf short-circuiting in story 185 would hide annotations from later passing branches.
- $dynamicRef extension cases depend on story 188 propagating annotations from the resolved target.

Implementation constraints the coder must honor:

- No central list or switch of keyword names; ordering comes from the registration flag.
- Compiled Schema stays immutable; all evaluated-location state lives in the per-call evaluation context (tests run with -race).
- Errors never include instance values, only instance location, keyword location and absolute keyword location.
- Only annotations from successful subschemas count.
- Standard library only; go.mod unchanged.
- Tests cite 2020-12 core sections; nothing from the official JSON-Schema-Test-Suite or Bowtie is copied or referenced.

- DB impact: no · API impact: no · Migration needed: no

Breaking changes:

- (none)

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: internal/eval/order_test.go shows that a flagged test keyword sees sibling annotations whatever the key order. The existing registry and eval tests still pass.
- **T2**: go test -race ./internal/keywords/unevaluated/ passes, including the per-property error locations and the no-value check.
- **T3**: go test -race ./internal/keywords/unevaluated/ passes, including the per-item error locations and the no-value check.
- **T4**: go test ./... runs TestSpecCases over both files, and all cases pass.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
