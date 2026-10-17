# PLAN #23: Record evaluated properties and items only when something can use the record

- Date: 2026-10-17
- Status: proposed (pending Checkpoint 2 sign-off)
- Design: see `ADR-023-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `internal/compile/evaluated_need.go`
- `internal/compile/evaluated_need_test.go`
- `internal/compile/compile.go`
- `internal/eval/eval.go`
- `internal/eval/evaluated.go`
- `internal/eval/evaluated_items.go`
- `internal/eval/evaluated_skip_test.go`
- `jsonschema.go`
- `evaluated_tracking_limits_test.go`
- `evaluated_tracking_race_test.go`
- `internal/eval/location.go`
- `internal/eval/location_test.go`
- `internal/eval/lazyloc_test.go`
- `docs/benchmarks.md`

Modules the architecture expects to touch:

- `internal/compile/evaluated_need.go`
- `internal/compile/evaluated_need_test.go`
- `internal/compile/compile.go`
- `internal/eval/eval.go`
- `internal/eval/evaluated.go`
- `internal/eval/evaluated_items.go`
- `internal/eval/evaluated_skip_test.go`
- `internal/eval/location.go`
- `internal/eval/location_test.go`
- `internal/eval/lazyloc_test.go`
- `jsonschema.go`
- `evaluated_tracking_limits_test.go`
- `evaluated_tracking_race_test.go`
- `docs/benchmarks.md`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — Load-time reachability analysis for unevaluated keywords
2. **T4** — Remove unused location types and test lazyLoc
3. **T2** — Skip evaluated-property/item recording when not needed  _(after T1)_
4. **T3** — Limits regression test and concurrency test  _(after T2)_
5. **T5** — Before and after benchmarks  _(after T2, T3, T4)_

## Risks

- The analysis could miss a path by which an unevaluated keyword is reached, such as a $dynamicRef into a registry resource, an embedded resource with its own $id, or a meta-schema-driven case. That would change an answer, so the analysis errs toward recording, and tests cover each path.
- Detailed or verbose output, or a Result accessor, may read the evaluated record directly. If the check-time need is not detected, reports would differ, so the existing output tests act as the guard.
- Early returns in the recording helpers could change step charging or annotation counting if those are interleaved with recording today.
- Benchmark noise on ns/op is ±5%, so success is judged on allocs/op and B/op.

Implementation constraints the coder must honor:

- Never write to the compiled Schema or anything reachable from it after Compile returns. The switch lives only in the per-call context.
- Reachability through $dynamicRef counts every registry resource it could land on (operator choice: cautious).
- The decision covers the whole schema, not individual rules (operator choice).
- Evaluation-step charging, depth, ref-chain, output-unit, report-byte and annotation-count accounting must not move inside the skipped code.
- Do not edit keyword files under internal/keywords. Guard inside the eval recording helpers.
- Add no allocation-limit test, and leave internal/jsonvalue/number_test.go unchanged.
- No timing assertions in ordinary tests. ns/op is recorded only in docs/benchmarks.md.
- Standard library only, with no go.mod change.

- DB impact: no · API impact: no · Migration needed: no

Breaking changes:

- (none)

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: evaluated_need_test.go covers six cases: no unevaluated keyword (false), a direct keyword (true), a keyword only behind $ref (true), a keyword only in a registry resource reachable by $dynamicRef (true), a cycle of $refs (terminates), and boolean subschemas.
- **T4**: nodeLoc and kwLoc no longer exist in the package. lazyloc_test.go covers $ref, $dynamicRef, if/then/else, ~0/~1 escaping with percent-encoding, and an empty absolute location before any reference is crossed. go vet ./... and go test ./internal/eval pass.
- **T2**: evaluated_skip_test.go shows that recording is off for flag and basic checks on a schema with no unevaluated keyword, and on for detailed and verbose checks and for schemas with unevaluated keywords. go test ./... passes with no existing test changed.
- **T3**: The limits test asserts the same error and exit code 2 at the same boundary inputs for each limit. go test -race ./... passes.
- **T5**: docs/benchmarks.md has a story-200 section with before and after ns/op, B/op, allocs/op and MB/s for every scenario. It shows lower allocs/op or B/op on flag and basic order-document validation and no rise in any scenario.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
