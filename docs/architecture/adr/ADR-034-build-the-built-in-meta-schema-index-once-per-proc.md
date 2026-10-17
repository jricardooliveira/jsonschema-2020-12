# ADR #34: Build the built-in meta-schema index once per process

- Date: 2026-10-17
- Story: #34
- Status: proposed (pending architecture sign-off)

## Decision
The change is small and lives entirely in #211's claimed area (internal/registry/*, internal/compile/metacache.go).

Shared builder and counter:
- internal/registry gets one shared builder (sync.Once / sync.OnceValue over the existing build function). The function that actually builds the index increments an atomic.Int64 build counter, defined in a new file buildcount.go. Only an internal accessor reads it, and it lives in an internal package, so it is invisible outside the module.
- AddBuiltins keeps using the shared index.
- metacache.go switches from registry.BuildBuiltinIndex() to the shared accessor.
- The built index is never mutated after construction. That invariant is documented next to the builder.
- BuildBuiltinIndex stays available for the FirstLoadOrder benchmark, which rebuilds on purpose.

Tests (all new root-package _test files):
- A helper re-execs os.Args[0] with -test.run=^TestName$ and a marker env var, so the child runs exactly one check in a fresh process and reports its result on stdout.
- The counter test (every run) asserts the build count is 1 after a public Compile.
- allocs_firstload_test.go (every run) measures first-Compile allocations in the child with runtime.MemStats Mallocs and compares them to a limit of measured + ~10%. It reuses the race gating from allocs_race_on/off_test.go read-only.
- firstload_timing_test.go is opt-in via its own test flag. It runs about 10 children and reports medians.

Shared files avoided:
- allocs_test.go and docs/benchmarks.md are not touched. Results go to docs/benchmarks/r3-211-firstload.md.
- No crowded file is written.

## Affected modules
- internal/registry/registry.go
- internal/registry/buildcount.go
- internal/compile/metacache.go
- firstload_helper_test.go
- firstload_count_test.go
- allocs_firstload_test.go
- firstload_timing_test.go
- docs/benchmarks/r3-211-firstload.md

## Data / API impact
- DB impact: no
- API impact: no
- Migration needed: no

## Constraints for implementation
- Standard library only (sync, sync/atomic, os, os/exec, flag, runtime, testing). go.mod unchanged.
- Do not write allocs_test.go, docs/benchmarks.md, testdata/golden/*, internal/keywords/registry.go, childctx*, internal/keywords/applicator/*, internal/keywords/annotate.go, internal/keywords/recording.go, internal/eval/evaluated.go, internal/jsonvalue/number_inline* or internal/keywords/validation/*.
- Do not remove or weaken any existing test. FirstLoadOrder keeps rebuilding on purpose.
- The shared index must stay immutable after it is built and safe to share across goroutines. go test -race ./... must be green.
- No timing or peak-memory assertion in default runs. The timing measure is opt-in via a test flag.
- The child-process checks must run only the intended test (an anchored -test.run plus an env marker) so that no other test compiles a schema first.
- Allocation limit = measured value + about 10%, with the measured value recorded next to it. Follow the existing race gating.
- Any output differing from the golden files means the story stops. Golden files are never edited.

## Risks
- Package-level variables or init code in root test files (or example_test.go) may compile a schema before the child's measurement. The counter would still read 1, but the first-load allocation count would not be a true first load. The child must measure at the very start of its own test, and this must be verified.
- FirstLoadOrder may reset the meta-schema cache through a test hook that relies on metacache's own BuildBuiltinIndex call. Removing that call could silently stop the benchmark from rebuilding the index and change what it measures.
- If AddBuiltins and metacache currently use slightly different index contents or options, sharing one index could change load errors. The golden test and the existing metacache/registry tests must catch any difference.
- Timing medians taken while #206, #207 and #210 build in parallel are noisy, so the time verdict waits for the operator's interleaved re-measure at Checkpoint 3.
- Under -race, allocation counts differ, so the first-load limit must follow the existing race gating, or it becomes flaky.

## Breaking changes
- none

## Sensitivity
- none
