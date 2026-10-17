# PLAN #22: Build keyword locations only when a report or Result shows them

- Date: 2026-10-17
- Status: passed gate 2 with nothing for the operator to decide (run #157, 2026-10-17 10:54 UTC)
- Design: see `ADR-022-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `docs/benchmarks.md`
- `internal/eval/location.go`
- `internal/eval/location_test.go`
- `internal/eval/eval.go`
- `internal/eval/units.go`
- `internal/eval/annotations.go`
- `cmd/jsonschema/limits_edge_regression_test.go`

Modules the architecture expects to touch:

- `internal/eval/location.go`
- `internal/eval/location_test.go`
- `internal/eval/eval.go`
- `internal/eval/units.go`
- `internal/eval/annotations.go`
- `cmd/jsonschema/limits_edge_regression_test.go`
- `docs/benchmarks.md`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — Record the before benchmark numbers
2. **T2** — Lazy location reference in the evaluator
3. **T4** — Limits-at-the-edge regression test
4. **T3** — Wire lazy locations into evaluation, units and annotations  _(after T2)_
5. **T5** — Race and immutability check  _(after T3)_
6. **T6** — Record the after benchmark numbers and verify there is no regression  _(after T1, T3)_

## Risks

- Location strings may also be built inside keyword evaluate functions (internal/keywords/**) or in the keyword registry's evaluate signature, not only in internal/eval. In that case more files would need changing.
- A lazy reference that captures mutable per-call state could render the wrong path if it is evaluated after that state changes, for example dynamic scope in verbose's second pass.
- The report-bytes limit depends on rendered string sizes. Rendering at a different moment must not change when the limit trips.

Implementation constraints the coder must honor:

- Rendered locations must match today's strings exactly, including JSON Pointer escaping (~0/~1), percent-encoding in absolute locations, $ref/$dynamicRef path segments and embedded-resource base URIs.
- Do not add fields or caches to the compiled Schema or any compile-time structure. Compile must not allocate more.
- Evaluation-step charging, output-unit counting, report-byte accounting and annotation counting must occur at the same points with the same amounts.
- Verbose keeps its second evaluation pass.
- No existing test file is edited, and no allocation-limit test is added or changed.
- Standard library only, with no change to go.mod.
- Benchmarks: 10 runs per scenario on the same machine, record the median (for example -count=10), using Go tools only.

- DB impact: no · API impact: no · Migration needed: no

Breaking changes:

- (none)

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: docs/benchmarks.md has a story-199 'before' table of median ns/op, B/op, allocs/op and MB/s for every scenario, with the command used.
- **T2**: Unit tests in location_test.go show the rendered strings equal today's strings for nested keywords, ~0/~1 escaping, $ref and $dynamicRef hops, and embedded $id resources.
- **T4**: The test passes both before and after T3, which proves that no limit's trigger point shifted.
- **T3**: go test ./... passes with no existing test edited, including the output, verbose, annotation, unevaluated* and CLI contract tests.
- **T5**: go test -race ./... passes, and location_test.go includes a parallel render case.
- **T6**: docs/benchmarks.md shows before and after medians side by side for story 199, with the reduction stated per validation scenario. No allocation-limit test file was added or changed.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
