# PLAN #31: Annotations collected in growing chunks, flattened lazily under sync.Once

- Date: 2026-10-17
- Status: proposed (pending Checkpoint 2 sign-off)
- Design: see `ADR-031-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `internal/eval/annotation_chunks.go`
- `internal/eval/annotation_chunks_test.go`
- `internal/eval/annotation_chunks_random_test.go`
- `internal/eval/annotation_sink.go`
- `internal/eval/annotations.go`
- `annotations.go`
- `jsonschema.go`
- `annotations_once_race_test.go`
- `cmd/jsonschema/golden_dbg_test.go`
- `allocs_test.go`
- `docs/benchmarks.md`

Modules the architecture expects to touch:

- `internal/eval/annotation_chunks.go`
- `internal/eval/annotation_chunks_test.go`
- `internal/eval/annotation_chunks_random_test.go`
- `internal/eval/annotation_sink.go`
- `internal/eval/annotations.go`
- `annotations.go`
- `jsonschema.go`
- `annotations_once_race_test.go`
- `allocs_test.go`
- `docs/benchmarks.md`
- `cmd/jsonschema/golden_dbg_test.go`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — Chunked annotation store with mark/truncate and reuse
2. **T4** — Remove stray golden_dbg_test.go
3. **T2** — Wire the chunk store into the annotation sink  _(after T1)_
4. **T3** — Lazy flatten and sort under sync.Once, release chunks  _(after T2)_
5. **T5** — Measure, lower allocation limits, record numbers  _(after T3, T4)_

## Risks

- Rollback across a chunk boundary could leave stale entries visible if the cursor arithmetic is off; the random test targets this case.
- ValidateSmall: more than 16 annotations means several small blocks, so allocs/op could rise. The operator rule makes that a hold, not a raised limit.
- Releasing chunks outside the Once would race with concurrent readers; the race test covers this.
- If 206's per-depth child contexts each own a sink, every depth gets its own chunk store. Byte totals could differ from the prototype, so this must be measured.

Implementation constraints the coder must honor:

- Chunk capacities grow 16, 32, … and cap at 4,096; fixed-size chunks are forbidden.
- Fresh chunks per check; no sync.Pool or cross-call reuse.
- Flatten, sort and chunk release all happen inside the same sync.Once.
- The sort order must be exactly the current order: reuse annotation_order.go, do not reimplement it.
- Do not edit internal/eval/eval.go, internal/keywords/registry.go, testdata/golden/*, allocs_<topic>_test.go or docs/benchmarks/r3-*.md.
- The random test uses math/rand with a fixed seed printed on failure. It goes in its own new _test.go file.
- Standard library only; go.mod unchanged.
- Delete only cmd/jsonschema/golden_dbg_test.go, using git rm.

- DB impact: no · API impact: no · Migration needed: no

Breaking changes:

- (none)

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: go test -race ./internal/eval -run 'Chunk' passes. The random test reports reaching a cross-boundary truncate and a capped 4,096 chunk. Without -race it runs under 1 s.
- **T4**: git status shows exactly one deleted file. go test ./cmd/jsonschema passes, including the golden test.
- **T2**: Existing internal/eval annotation, order and sink tests pass unchanged under -race.
- **T3**: go test -race -run 'Annotations|Once' . passes. The existing annotations_test.go and annotations_ondemand_test.go tests pass unchanged.
- **T5**: docs/benchmarks.md shows the before and after medians, the computations with sources and the verdict per target. allocs_test.go limits are lowered, with measured values. go test -race ./... is green.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
