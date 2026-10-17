# PLAN #34: Build the built-in meta-schema index once per process

- Date: 2026-10-17
- Status: proposed (pending Checkpoint 2 sign-off)
- Design: see `ADR-034-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `internal/registry/registry.go`
- `internal/registry/buildcount.go`
- `internal/compile/metacache.go`
- `firstload_helper_test.go`
- `firstload_count_test.go`
- `allocs_firstload_test.go`
- `firstload_timing_test.go`
- `docs/benchmarks/r3-211-firstload.md`

Modules the architecture expects to touch:

- `internal/registry/registry.go`
- `internal/registry/buildcount.go`
- `internal/compile/metacache.go`
- `firstload_helper_test.go`
- `firstload_count_test.go`
- `allocs_firstload_test.go`
- `firstload_timing_test.go`
- `docs/benchmarks/r3-211-firstload.md`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — One shared built-in index builder with a build counter
2. **T2** — Meta-schema cache reuses the shared index  _(after T1)_
3. **T3** — Fresh-process helper and one-build counter test  _(after T2)_
4. **T4** — First-load allocation limit  _(after T3)_
5. **T5** — Opt-in first-compile timing measure  _(after T3)_
6. **T6** — Measure before/after and record the verdict  _(after T2, T4, T5)_

## Risks

- Package-level variables or init code in root test files (or example_test.go) may compile a schema before the child's measurement. The counter would still read 1, but the first-load allocation count would not be a true first load. The child must measure at the very start of its own test, and this must be verified.
- FirstLoadOrder may reset the meta-schema cache through a test hook that relies on metacache's own BuildBuiltinIndex call. Removing that call could silently stop the benchmark from rebuilding the index and change what it measures.
- If AddBuiltins and metacache currently use slightly different index contents or options, sharing one index could change load errors. The golden test and the existing metacache/registry tests must catch any difference.
- Timing medians taken while #206, #207 and #210 build in parallel are noisy, so the time verdict waits for the operator's interleaved re-measure at Checkpoint 3.
- Under -race, allocation counts differ, so the first-load limit must follow the existing race gating, or it becomes flaky.

Implementation constraints the coder must honor:

- Standard library only (sync, sync/atomic, os, os/exec, flag, runtime, testing). go.mod unchanged.
- Do not write allocs_test.go, docs/benchmarks.md, testdata/golden/*, internal/keywords/registry.go, childctx*, internal/keywords/applicator/*, internal/keywords/annotate.go, internal/keywords/recording.go, internal/eval/evaluated.go, internal/jsonvalue/number_inline* or internal/keywords/validation/*.
- Do not remove or weaken any existing test. FirstLoadOrder keeps rebuilding on purpose.
- The shared index must stay immutable after it is built and safe to share across goroutines. go test -race ./... must be green.
- No timing or peak-memory assertion in default runs. The timing measure is opt-in via a test flag.
- The child-process checks must run only the intended test (an anchored -test.run plus an env marker) so that no other test compiles a schema first.
- Allocation limit = measured value + about 10%, with the measured value recorded next to it. Follow the existing race gating.
- Any output differing from the golden files means the story stops. Golden files are never edited.

- DB impact: no · API impact: no · Migration needed: no

Breaking changes:

- (none)

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: go test -race ./internal/registry/... green; the counter increments once per actual build; existing registry tests unchanged and passing.
- **T2**: go test -race ./internal/compile/... ./... green, including metacache_test.go, first_compile_concurrency_test.go and compile_cache_limits_test.go unchanged; the golden test passes.
- **T3**: go test ./ runs the counter test in a child process and passes. Temporarily restoring the second build makes it fail (checked locally, not committed). The test follows the race gating.
- **T4**: go test ./ and go test -race ./ green. The measured value and limit are recorded in the file. allocs_test.go is unchanged (git diff empty).
- **T5**: Skipped by default in go test ./. With the flag set it prints medians over about 10 runs. No assertion depends on timing.
- **T6**: The doc holds before/after medians, including first-compile time and B. go test ./... and go test -race ./... are green, and the golden test matches. If the first-compile time or memory is not lower, or any number gets worse, this is recorded and the story is held for the operator.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
