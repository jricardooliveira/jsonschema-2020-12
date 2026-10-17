# SPEC #27: Allocation and memory limit tests for the order scenarios, plus routed test-only fixes

- Date: 2026-10-17
- Type: feature
- Verdict: pass
- Status: passed gate 1 with nothing for the operator to decide (run #162, 2026-10-17 13:42 UTC)

## Problem

The performance fixes from stories 199 to 203 are measured only by benchmarks, and nothing in an ordinary go test run fails when a later change brings allocations back. Benchmarks are run by hand. Allocation counts are stable, so ordinary tests can guard them. Earlier release reviews also left test-only gaps. Story 202 lacks tests for strict custom meta-schemas, for a caller resource at the built-in meta-schema URI and for a reliable very-first-Compile concurrency test, and one of its assertions is vacuous. The story 203 peak-memory test does not compile on Windows. docs/benchmarks.md also has a wrong sentence about the operator's Go version.

## Why this matters

The operator chose 'benchmarks plus allocation limits in tests' so the speed gains are not slowly lost. Each limit is the measured count after the fixes plus about 10% headroom. Total-memory limits on the large documents cover the operator's main complaint, which is memory. The routed items close review findings without changing any validator behaviour.

## Acceptance criteria

- A new root-package test file allocs_test.go has one allocation-limit test for each of these six scenarios: (1) compiling the order schema, starting from the raw schema bytes and including loading of the built-in meta-schema; (2) validating the small document from testdata/perf; (3) validating the ~0.5 MB valid document; (4) validating the ~5.5 MB valid document; (5) validating the 1,000-error document; (6) validating the 1,000-error document and rendering the basic report. Each document scenario starts from the raw document bytes, so decoding the JSON is included. The large documents come from the fixed recipe in internal/perfdata. Counts are measured with testing.AllocsPerRun, and a test fails when the count exceeds its limit.
- Each allocation-count limit is a named constant set to the count the current code uses, measured on this build machine, plus about 10% headroom. A comment next to each constant records the measured value.
- For the ~0.5 MB valid, ~5.5 MB valid and 1,000-error documents, the tests also enforce a total-bytes-allocated limit. This is the runtime.MemStats TotalAlloc delta across the operation, which counts every piece requested, even pieces freed straight away. Each limit is the measured value plus about 10% headroom and is recorded the same way. The small-document and compile scenarios check counts only.
- When the test binary is built with -race, every allocation and memory-limit test skips and says the race detector adds allocations the validator did not cause. Race detection uses a pair of build-constrained test files (//go:build race and //go:build !race). A plain go test ./... enforces every limit.
- The large-document checks always run. There is no -short or environment-variable skip.
- docs/benchmarks.md gains a section on the allocation and memory limits. For each scenario it lists the measured count, the measured total bytes where one applies, the limit and its headroom, and how the value was measured: the API calls, raw bytes as the starting point, AllocsPerRun or MemStats TotalAlloc, and the run count. It also states the Go version, OS and machine used for the measurement. It states the policy that one limit applies on all platforms and that a limit exceeded on a newer Go is re-measured on purpose, with the reason recorded. The same edit corrects the operator-baseline sentence that says the operator's Mac uses 'a different machine and a different Go version' to say it used Go 1.27.2, the same version as the build machine.
- A new test supplies a custom meta-schema through the registry that uses unevaluatedProperties: false over an allOf $ref to the 2020-12 meta-schema. A schema using that meta-schema with an unknown keyword at the root, or nested under properties, is an invalid-schema load error. Through the CLI the same case exits with code 2. A schema using only listed keywords loads and validates.
- A new test registers a caller resource at the built-in 2020-12 meta-schema URI. A JSON-equal copy of the embedded meta-schema is accepted. A different document gives today's duplicate-address error.
- A new test proves the very first Compile is safe under concurrency. It re-executes its own test binary in a fresh process, selected by an environment variable plus a -test.run filter. In the child, 32 goroutines released together call Compile and then Validate on the order schema before anything else has built the shared meta-schema set. The parent fails on a non-zero child exit or on a race report when run under go test -race.
- The vacuous assertion that story 202's tester reported as F4 is replaced with an assertion that fails when the behaviour it names is broken.
- cmd/jsonschema/peakmem_test.go is restricted to //go:build unix. A new //go:build !unix stub file makes TestCLIPeakMemory skip with a clear message. Any test flags the peak-memory test uses are defined in a file with no build constraint, so they exist on every platform. GOOS=windows go vet ./... and GOOS=linux go vet ./... both pass.
- No non-test Go file changes, and validator results are unchanged. All existing tests pass unchanged except the replaced F4 assertion. go test ./... and go test -race ./... pass, the new tests use only the standard library, and no test uses the network.

## Non-goals

- Changing any validator, CLI or output behaviour, or any non-test Go file
- Per-platform or per-Go-version limits (the operator chose one limit for all)
- Peak or retained heap limits in tests (the operator chose total allocated bytes)
- Changing the verbose report's second evaluation pass
- Comparing against any other validator inside the project
- Any timing-based assertion

## Tasks

- **T1** — Race-mode detection and allocation-count limit tests
  - purpose: Guard the six scenarios with AllocsPerRun count limits measured on this build machine plus about 10% headroom, starting from raw bytes, and skip under -race.
  - scope: allocs_test.go, allocs_race_on_test.go, allocs_race_off_test.go
  - done when: go test -run Alloc -v . shows six passing tests, each logging its count against its limit. go test -race -run Alloc -v . shows them skipped with the race message.
  - delivers: AC1, AC2, AC4, AC5, AC12
- **T2** — Total-memory limits for the three large documents
  - purpose: Add MemStats TotalAlloc-delta limits, measured plus about 10%, for the 0.5 MB, 5.5 MB and 1,000-error documents.
  - scope: allocs_test.go
  - done when: go test -run Alloc -v . logs the bytes and the limit for the three large scenarios and passes.
  - depends on: T1
  - delivers: AC3, AC4, AC5
- **T3** — Record the limits in docs/benchmarks.md and fix the Go-version sentence
  - purpose: Document each measured value, limit, method, Go version, OS, machine and the re-measure policy, and correct the operator-baseline sentence.
  - scope: docs/benchmarks.md
  - done when: The docs section lists six count limits and three byte limits that match the constants in allocs_test.go. The baseline sentence names Go 1.27.2 as the same version as the build machine.
  - depends on: T1, T2
  - delivers: AC6
- **T4** — Story 202 follow-up tests: strict custom meta-schema, built-in URI resource, F4
  - purpose: Add the missing tests for registry custom meta-schemas and for a caller resource at the built-in URI, and replace the vacuous F4 assertion.
  - scope: metaschema_custom_strict_test.go, cmd/jsonschema/metaschema_custom_strict_cli_test.go, metaschema_uri_resource_test.go, compile_cache_limits_test.go, internal/compile/metacache_test.go
  - done when: The new tests pass on the unchanged code. The F4 assertion fails when the behaviour it names is deliberately broken locally, and the code is then reverted.
  - delivers: AC7, AC8, AC10, AC12
- **T5** — Reliable very-first-Compile concurrency test via re-exec
  - purpose: Prove that the shared meta-schema set is built safely when 32 goroutines race on the very first Compile in a fresh process.
  - scope: first_compile_concurrency_test.go
  - done when: go test -race -run FirstCompile -v . passes and shows that the child process ran. The parent fails if the child exits non-zero or prints 'WARNING: DATA RACE'.
  - delivers: AC9, AC12
- **T6** — Make the CLI peak-memory test portable
  - purpose: Put the Getrusage measurement behind //go:build unix, add a non-unix skip stub, and keep the flag definitions on every platform.
  - scope: cmd/jsonschema/peakmem_test.go, cmd/jsonschema/peakmem_other_test.go, cmd/jsonschema/peakmem_flags_test.go
  - done when: GOOS=windows go vet ./... and GOOS=linux go vet ./... both pass, and go test ./cmd/jsonschema passes on darwin.
  - delivers: AC11, AC12
