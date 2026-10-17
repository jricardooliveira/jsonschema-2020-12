# Release notes — #27: Allocation and memory limit tests for the order scenarios, plus routed test-only fixes

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #27
- Status: proposed (pending Checkpoint 3 sign-off)

**⚠️ the release-agent has concerns.**

## What changed

Test-only story adding allocation and memory limit guards for six validation and compilation scenarios, plus routed test-only fixes for custom meta-schemas, concurrent compilation safety, and Windows portability. No validator behaviour changes; all existing tests pass. Allocation-limit tests use testing.AllocsPerRun and runtime.MemStats, skip under -race, and run on every go test ./..., including the 5.5MB document.

- allocs_test.go: six allocation-limit tests (Compile, Validate small/0.5MB/5.5MB/errors/errors+basic) measuring AllocsPerRun counts and TotalAlloc bytes for large documents, starting from raw bytes, with named constants set to measured value plus ~10% headroom
- allocs_race_on_test.go, allocs_race_off_test.go: build-constrained files declaring raceEnabled constant for race detection in allocation tests
- metaschema_custom_strict_test.go: test of registry-supplied custom meta-schema (unevaluatedProperties: false over 2020-12); verifies unknown keywords at root and nested levels cause compile errors; listed keywords only load and validate
- metaschema_uri_resource_test.go: test of caller resource registered at built-in 2020-12 meta-schema URI; JSON-equal copy accepted, different document gives duplicate-address error
- first_compile_concurrency_test.go: re-exec test spawning child where 32 goroutines concurrently call Compile and Validate on very-first Compile (shared meta-schema set not yet built); parent verifies child exit code and absence of race reports
- cmd/jsonschema/metaschema_custom_strict_cli_test.go: CLI test of strict custom meta-schema; unknown keywords exit code 2, listed keywords validate
- cmd/jsonschema/peakmem_flags_test.go: new unconstrained file holding flag definitions so they exist on all platforms
- cmd/jsonschema/peakmem_other_test.go: new non-unix stub skipping TestCLIPeakMemory with clear message
- cmd/jsonschema/peakmem_test.go: restricted to //go:build unix; getrusage measurement moved behind unix constraint; flags moved to peakmem_flags_test.go
- docs/benchmarks.md: new 'Allocation and memory limits' section listing six allocation limits (2,730 to 7,260,000) and three memory limits (32.2M to 405.9M) measured on go1.27.2 darwin/arm64, with measured values in comments, methods (AllocsPerRun with warm-up and run counts, TotalAlloc delta), and re-measure policy; corrects operator-baseline sentence from 'different Go version' to 'same Go version (1.27.2)'
- internal/compile/metacache_test.go: replaced vacuous assertion with callback counter ensuring callback is called on each MetaChecks invocation; verified that caller resource added to per-call index is not written to shared index

## How to verify

- go test ./... passes with six new allocation tests and 5.5MB scenario (adds a few seconds)
- go test -race ./... shows allocation and memory tests skipped with message 'race detector adds allocations the validator did not cause'
- go test -run Alloc -v . logs alloc count and limit for each of six tests; all pass
- go test -run FirstCompile -v . shows child process executed (output includes TestFirstCompileChild PASS)
- go test -run 'TestCustomStrict|TestCallerResource' -v . verifies custom meta-schema and URI resource tests pass
- GOOS=windows go vet ./cmd/jsonschema and GOOS=linux go vet ./cmd/jsonschema both pass (no undefined syscall.Getrusage)
- docs/benchmarks.md shows new table with six allocation limits and three byte limits matching allocs_test.go constants; baseline sentence names Go 1.27.2

## Migration

none

## Rollback

Remove new test files (allocs_race_on_test.go, allocs_race_off_test.go, allocs_test.go, metaschema_custom_strict_test.go, metaschema_uri_resource_test.go, first_compile_concurrency_test.go, cmd/jsonschema/metaschema_custom_strict_cli_test.go, cmd/jsonschema/peakmem_flags_test.go, cmd/jsonschema/peakmem_other_test.go). Revert docs/benchmarks.md to remove allocation section and restore 'different Go version' wording. Revert cmd/jsonschema/peakmem_test.go to include flags and remove unix constraint. Revert internal/compile/metacache_test.go to original assertions.

## Known limitations

- Allocation and memory limits measured on Apple M3 Max with go1.27.2; counts may differ on other platforms; per operator policy, limits exceeded by >10% are re-measured on purpose and recorded
- Large-document tests add seconds to every go test ./... run and cannot be skipped (operator accepted)
- Strict custom meta-schema tests check for error presence but do not explicitly assert error type or message
- Docs wording on Compile scenario 'includes the built-in meta-schema' does not emphasize that this is the lazy one-time cached build, not a per-call overhead

## Concerns

- Low: strict custom meta-schema tests do not check error type or message, only verify error/no-error (tester noted as low severity)
- Low: docs section on Compile could be clearer that 'includes the built-in meta-schema' refers to lazy one-time cached build, not per-call overhead (tester noted as low severity)
- Tester overall verdict is 'warn' but found no blocking problems; these are minor notes for reviewer attention, not ship blockers
