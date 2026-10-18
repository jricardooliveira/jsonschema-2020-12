# SPEC #34: Build the built-in meta-schema index once per process

- Date: 2026-10-17
- Type: performance
- Verdict: pass
- Status: passed gate 1 with nothing for the operator to decide (run #165, 2026-10-17 21:46 UTC)

## Problem

internal/compile/metacache.go calls registry.BuildBuiltinIndex() even though registry.AddBuiltins (internal/registry/registry.go) has already built and shared the built-in meta-schema index. Every process builds the index twice, which costs about 165 µs and 118 KB at the first schema load. The FirstLoadOrder benchmark rebuilds on purpose, so the saving cannot be seen in any existing measure.

## Why this matters

Round 3, fix 7: schema loading should cost less, with results that stay exactly the same. A counter test and a first-load allocation limit stop the double build from coming back later.

## Acceptance criteria

- One shared builder in internal/registry produces the built-in meta-schema index. registry.AddBuiltins and internal/compile/metacache.go both get the index from it, so a process builds the index exactly once. The shared index stays immutable after it is built and is safe to share across goroutines (sync.Once / sync.OnceValue).
- A single build counter lives inside internal/registry. It is an atomic counter incremented by every built-in index build. Only tests read it (internal package, so it is not part of the public API). It changes no result, error or output, and it is safe under concurrent use.
- A counter test in a new root-package test file runs in every normal go test run. It starts one fresh test program (a re-exec of the test binary that runs only the child), loads a schema through the public API on the path that uses both AddBuiltins and the meta-schema cache, and asserts that the build counter is exactly 1. It honours the race gating used by allocs_race_on_test.go / allocs_race_off_test.go.
- A new root-package file allocs_firstload_test.go adds an allocation limit for the first schema load. The test runs in every normal run, in one fresh test program, and counts allocations of the first Compile of the order schema in testdata/perf. The limit is the measured count after this story plus about 10%, with the measured value recorded next to it. It reuses the existing allocation helpers and race gating read-only. allocs_test.go is not modified.
- A first-compile timing measure is skipped in a normal go test run and enabled only by a dedicated test flag, in the same way as TestCLIPeakMemory -peakmem. It starts a fresh test program about 10 times, measures only the first schema load in each (time, bytes and allocations), and reports the medians. It asserts no timing.
- A new file docs/benchmarks/r3-211-firstload.md records before medians (main after #205) and after medians (this story), measured with go test -count=10 on the same machine. It covers the in-repo benchmarks and the new first-compile measure, and states the verdict. Success means the first-compile median is lower in both time and memory. If there is no saving, or any number gets worse, the story is recorded and held for the operator's decision. Timings taken during parallel builds are marked as noisy and re-measured by the operator at Checkpoint 3. B/op and allocs/op are judged exactly.
- Validation results, flag/basic/detailed/verbose output, library Result, load errors, error messages, limits, exit codes and goroutine safety are unchanged. All existing tests pass unmodified, go test -race ./... is green, and the golden test matches the committed v1.1 golden files byte for byte. No golden file, allocs_test.go, docs/benchmarks.md or file claimed by #206, #207 or #210 is written. Only the standard library is used and go.mod is unchanged.

## Non-goals

- Changing what the FirstLoadOrder benchmark measures or how it rebuilds on purpose
- Any further compile-time optimisation beyond removing the second built-in index build
- A fixed time or memory target for the first-compile measure
- Editing allocs_test.go or docs/benchmarks.md (stories #208 and #209 keep those)
- Exposing the build counter in the public API

## Tasks

- **T1** — One shared built-in index builder with a build counter
  - purpose: Make internal/registry the single owner of the built-in index: one sync.Once-guarded shared builder that AddBuiltins uses, and an atomic build counter that only tests read.
  - scope: internal/registry/registry.go, internal/registry/buildcount.go
  - done when: go test -race ./internal/registry/... green; the counter increments once per actual build; existing registry tests unchanged and passing.
  - delivers: AC1, AC2
- **T2** — Meta-schema cache reuses the shared index
  - purpose: Replace the registry.BuildBuiltinIndex() call in metacache.go with the shared builder, so the meta-schema cache no longer rebuilds the index. Load errors and behaviour stay the same.
  - scope: internal/compile/metacache.go
  - done when: go test -race ./internal/compile/... ./... green, including metacache_test.go, first_compile_concurrency_test.go and compile_cache_limits_test.go unchanged; the golden test passes.
  - depends on: T1
  - delivers: AC1, AC7
- **T3** — Fresh-process helper and one-build counter test
  - purpose: Add a small root-package test helper that re-executes the test binary to run one child check in a fresh process. Use it for a counter test, run in every normal run, asserting exactly one built-in index build per process.
  - scope: firstload_helper_test.go, firstload_count_test.go
  - done when: go test ./ runs the counter test in a child process and passes. Temporarily restoring the second build makes it fail (checked locally, not committed). The test follows the race gating.
  - depends on: T2
  - delivers: AC3
- **T4** — First-load allocation limit
  - purpose: Measure the allocations of the first Compile of the order schema in a fresh process. Set the limit to the measured value plus about 10%, with the measured value recorded next to it.
  - scope: allocs_firstload_test.go
  - done when: go test ./ and go test -race ./ green. The measured value and limit are recorded in the file. allocs_test.go is unchanged (git diff empty).
  - depends on: T3
  - delivers: AC4
- **T5** — Opt-in first-compile timing measure
  - purpose: Add a flag-enabled test that starts about 10 fresh test programs, times only the first schema load in each, and reports median time, bytes and allocations without asserting them.
  - scope: firstload_timing_test.go
  - done when: Skipped by default in go test ./. With the flag set it prints medians over about 10 runs. No assertion depends on timing.
  - depends on: T3
  - delivers: AC5
- **T6** — Measure before/after and record the verdict
  - purpose: Run the in-repo benchmarks with -count=10 and the new first-compile measure, before (main after #205) and after, on the same machine. Record the medians and the verdict, and run the full verification.
  - scope: docs/benchmarks/r3-211-firstload.md
  - done when: The doc holds before/after medians, including first-compile time and B. go test ./... and go test -race ./... are green, and the golden test matches. If the first-compile time or memory is not lower, or any number gets worse, this is recorded and the story is held for the operator.
  - depends on: T2, T4, T5
  - delivers: AC6, AC7
