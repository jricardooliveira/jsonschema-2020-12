# Release notes — #34: Build the built-in meta-schema index once per process

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #34
- Status: proposed (pending Checkpoint 3 sign-off)

## What changed

The built-in meta-schema index is now built exactly once per process instead of twice. Both registry.AddBuiltins and the meta-schema cache get the index from a new shared builder guarded by sync.Once. Validation behaviour, output, exit codes, limits, error messages and goroutine safety are unchanged. The first-compile cost drops from ~1.01 ms and 741 KB to ~0.86 ms and 625 KB (about −15 % to −23 % time, −116 KB, −1,995 allocations).

- internal/registry/registry.go: Added SharedBuiltinIndex() accessor, returns built-in index built once per process via sync.Once. BuildBuiltinIndex() increments atomic counter on every call.
- internal/registry/buildcount.go: New file with atomic counter builtinBuilds and BuiltinBuilds() reader, visible only inside the internal package for tests.
- internal/compile/metacache.go: sharedBuiltin() now calls registry.SharedBuiltinIndex() instead of registry.BuildBuiltinIndex(), eliminating the second index build. buildBuiltin() kept intact so FirstLoadOrder benchmark still rebuilds on purpose.
- firstload_helper_test.go: New file with freshProcess() helper that re-execs the test binary in a child process with env marker and anchored -test.run, so checks measure at true program start.
- firstload_count_test.go: New file with TestFirstLoadBuildsBuiltinIndexOnce, runs in fresh process in every normal go test run, asserts exactly 1 build via registry.BuiltinBuilds().
- allocs_firstload_test.go: New file with TestAllocsFirstLoadOrder, measures first Compile allocations in fresh process (measured 8,653, limit 9,520), runs every normal test in race mode if enabled.
- firstload_timing_test.go: New file with TestFirstLoadTiming, opt-in via -firstload flag, runs ~10 fresh processes and logs medians of time/bytes/allocations without asserting them.
- docs/benchmarks/r3-211-firstload.md: New file recording before/after medians from in-repo benchmarks (go test -count=10), the new first-compile measure, and the verdict.

## How to verify

- Run go test ./ to confirm the counter test passes in a fresh process: registry.BuiltinBuilds() must be 1 after the first public Compile.
- Run go test ./ to confirm the allocation limit test passes: first-load allocations must stay ≤ 9,520.
- Run go test ./... and go test -race ./... on root, ./internal/compile and ./internal/registry to confirm all existing tests pass unchanged.
- Run the CLI golden test (cmd/jsonschema golden_test.go) to confirm stdout/stderr/exit code stay byte-identical to v1.1 in flag/basic/detailed/verbose.
- Run go test -run TestFirstLoadTiming -firstload -v . to see the first-compile medians (time, bytes, allocations) without any timing assertion.
- Temporarily edit registry.go to call BuildBuiltinIndex() twice in SharedBuiltinIndex() and confirm the counter test fails (BuiltinBuilds() becomes 2).

## Migration

none

## Rollback

Remove allocs_firstload_test.go, firstload_count_test.go, firstload_helper_test.go, firstload_timing_test.go, and docs/benchmarks/r3-211-firstload.md. Revert internal/registry/registry.go to remove SharedBuiltinIndex() and the counter increment. Remove internal/registry/buildcount.go. Revert internal/compile/metacache.go to have sharedBuiltin() call buildBuiltin() directly instead of SharedBuiltinIndex().

## Known limitations

- BenchmarkFirstLoadOrder still rebuilds the index on purpose (by calling buildBuiltin() directly from the benchmark, which still calls registry.BuildBuiltinIndex() each iteration). It measures the same thing as before, so the benchmark results stay comparable.
- The improvement (−165 µs, −118 KB per process) happens only at the first schema load in each process. Subsequent compilations in the same process use the cached index and see no change.
- The new first-load timing measure is opt-in (−firstload flag) and reports medians without asserting them, because benchmark timings taken during parallel builds (#206, #207, #210) are noisy. The operator re-measures interleaved at Checkpoint 3 to judge the time.
- Some in-repo benchmark medians are higher by +1.9 % to +6.5 %, but internal/jsonvalue (unaffected by this story) shows similar +10 % jitter, indicating machine noise rather than a story regression. These are recorded and provisional.

## Concerns

- Timing verdicts are held for the operator's decision: the before and after time medians are 1,010,834 ns vs 781,500–861,959 ns (showing savings), but several other benchmarks show up to +6.5 % jitter. All numbers are recorded in docs/benchmarks/r3-211-firstload.md and the operator re-measures them interleaved at Checkpoint 3 on a quiet machine.
- The shared built-in index is immutable after SharedBuiltinIndex() returns: callers must never write to it or its resources, and may safely share it across goroutines. This invariant is documented in the code and tested by go test -race.
- The allocation limit of 9,520 depends on the measured count of 8,653 plus ~10% headroom. If a later change adds allocations to the first Compile, the test will fail and guard against silent regression.
- The counter test and allocation limit test both run in fresh processes (one entry point each per run), which adds ~10 ms per test but is necessary to measure at true program start before any other test code executes.
