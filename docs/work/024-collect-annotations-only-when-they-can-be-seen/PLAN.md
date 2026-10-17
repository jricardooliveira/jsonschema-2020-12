# PLAN #24: Collect and sort annotations only when they can be shown, without re-splitting JSON Pointers

- Date: 2026-10-17
- Status: proposed (pending Checkpoint 2 sign-off)
- Design: see `ADR-024-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `internal/eval/annotation_order.go`
- `internal/eval/annotation_order_test.go`
- `internal/eval/annotations.go`
- `internal/eval/annotation_record.go`
- `annotations.go`
- `jsonschema.go`
- `internal/eval/annotation_sink.go`
- `internal/eval/eval.go`
- `internal/eval/annotation_sink_test.go`
- `cmd/jsonschema/annotations_limits_regression_test.go`
- `annotations_ondemand_test.go`
- `docs/benchmarks.md`

Modules the architecture expects to touch:

- `internal/eval/annotations.go`
- `internal/eval/annotation_order.go`
- `internal/eval/annotation_order_test.go`
- `internal/eval/annotation_record.go`
- `internal/eval/annotation_sink.go`
- `internal/eval/annotation_sink_test.go`
- `internal/eval/eval.go`
- `internal/eval/units.go`
- `annotations.go`
- `jsonschema.go`
- `annotations_ondemand_test.go`
- `cmd/jsonschema/annotations_limits_regression_test.go`
- `docs/benchmarks.md`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — Sort annotations by precomputed pointer segments
2. **T2** — Record annotations cheaply and build them on demand  _(after T1)_
3. **T3** — Skip collection where no output can show annotations  _(after T2)_
4. **T4** — Regression test: hostile-input limits unchanged  _(after T3)_
5. **T5** — On-demand annotations: identical content and race-safety test  _(after T3)_
6. **T6** — Before and after benchmarks recorded  _(after T5)_

## Risks

- Skipping collection in a subtree whose annotations a shape does show would change a report. Verbose and detailed may expose annotations from failing subschemas, so the skip rule must be checked against internal/output before any subtree is discarded.
- Lazy building could hold references into per-call state that is later reused, which would cause data races or wrong content.
- If the order schema produces few annotations, the measurable gain on the order scenarios may be small. Then sorting and deferred building are the main contributors, and AC7 depends on them.
- eval.go is also changed by stories 199 and 200, so merges and rebases on it are likely.

Implementation constraints the coder must honor:

- No new library option, CLI flag or exported identifier.
- Annotation content and order must be byte-identical in the Result and in every output shape.
- The annotation-count limit must be charged when an annotation is recorded, not when it is built.
- Discard mode must not change evaluation-step charging or any other limit counter.
- Building on demand must not touch pooled or reused per-call evaluation state, and the Result must be safe for concurrent reads.
- Add new code in new files. Change eval.go and jsonschema.go only to wire the new code in.
- Do not add or change any allocs_test.go or allocation-limit test, and leave internal/jsonvalue/number_test.go untouched.
- Use the standard library only, with no timing assertions in tests.

- DB impact: no · API impact: no · Migration needed: no

Breaking changes:

- (none)

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: annotation_order_test.go shows that the old comparator and the new one give the same order on the listed edge cases, and existing annotation tests pass unchanged.
- **T2**: Existing tests for annotations in the root package, internal/eval, internal/keywords and the CLI pass unchanged, and the limit tests pass unchanged.
- **T3**: annotation_sink_test.go shows that discarded subtrees record nothing and charge the same evaluation steps, and all existing tests pass unchanged.
- **T4**: The test passes before and after T2 and T3, and it reads each limit value from the code instead of repeating the number.
- **T5**: `go test -race ./...` passes, including the new test.
- **T6**: docs/benchmarks.md lists the before and after ns/op, B/op and allocs/op for every scenario. The targeted scenarios show a reduction, no scenario's allocs/op or B/op increases, and no allocs test changed.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
