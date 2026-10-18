# PLAN #29: One reusable child evaluation context per depth

- Date: 2026-10-17
- Status: passed gate 2 with nothing for the operator to decide (run #164, 2026-10-17 21:44 UTC)
- Design: see `ADR-029-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `docs/benchmarks/r3-206-childctx.md`
- `internal/keywords/childctx.go`
- `internal/keywords/childctx_test.go`
- `internal/keywords/registry.go`
- `internal/eval/eval.go`
- `internal/keywords/childctx_poison.go`
- `internal/keywords/childctx_poison_test.go`
- `childctx_race_test.go`
- `allocs_childctx_test.go`

Modules the architecture expects to touch:

- `internal/keywords/registry.go`
- `internal/keywords/childctx.go`
- `internal/keywords/childctx_test.go`
- `internal/keywords/childctx_poison.go`
- `internal/keywords/childctx_poison_test.go`
- `internal/eval/eval.go`
- `childctx_race_test.go`
- `allocs_childctx_test.go`
- `docs/benchmarks/r3-206-childctx.md`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — Measure the before numbers
2. **T2** — Per-depth child context stack  _(after T1)_
3. **T3** — Switch EvalChild to the reused contexts  _(after T2)_
4. **T4** — Test-only poison check and concurrency test  _(after T3)_
5. **T5** — Lowered allocation limits in a new file  _(after T3)_
6. **T6** — After numbers and target verdicts  _(after T4, T5)_

## Risks

- A keyword that keeps its child context (for example for verbose units) would produce a wrong report that tests may not notice. The poison check is the mitigation.
- If units are detached by copying, the B/op saving shrinks. Ownership must be handed over without an extra copy where possible.
- If testing.Testing() poisoning adds cost inside benchmarks, B/op and allocs/op are unaffected (it only zeroes a struct), but ns/op may shift slightly.
- The combined 200 MB target may be missed. In that case the story is held for the operator.

Implementation constraints the coder must honor:

- Results, output, exit codes, limits, error messages and goroutine safety must not change.
- No golden file is edited. Any golden diff means the story fails and stops.
- Do not write allocs_test.go, docs/benchmarks.md, or any file claimed by #207, #210 or #211.
- Reset must clear every per-child field, so no state leaks between siblings.
- The returned units slice must not alias the reused context's buffer after the next reset.
- The poison check must not allocate and must not run in production binaries.
- Standard library only. go.mod is unchanged.

- DB impact: no · API impact: no · Migration needed: no

Breaking changes:

- (none)

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: Before medians for every scenario recorded in the new doc, with the commit hash and machine.
- **T2**: Unit tests show reuse per depth, a complete reset and no aliasing of returned units.
- **T3**: go test ./... and the #205 golden test pass unchanged. Any keyword on the fresh-context path is listed.
- **T4**: The poison test fails a deliberately retaining test keyword. go test -race ./... is green.
- **T5**: The new limits pass with and without -race. allocs_test.go is unchanged.
- **T6**: The doc shows before and after medians and a pass, miss or hold verdict for each target. go test -race ./... and the golden test are green.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
