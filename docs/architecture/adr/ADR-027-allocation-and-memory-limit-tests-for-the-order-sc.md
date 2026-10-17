# ADR #27: Allocation and memory limit tests for the order scenarios, plus routed test-only fixes

- Date: 2026-10-17
- Story: #27
- Status: passed gate 2 with nothing for the operator to decide (run #162, 2026-10-17 13:42 UTC)

## Decision
Test-only story. The allocation guards follow the convention that allocation-limit tests live in <package>/allocs_test.go. They live in the root package because they go through the public API: Compile from raw schema bytes, decoding from raw document bytes, Validate, and basic rendering. They reuse the order fixtures in testdata/perf and the generators in internal/perfdata, the same ones bench_test.go uses. Race detection is done with two tiny build-constrained test files that declare a raceEnabled constant. The 202 follow-ups and the first-Compile re-exec test are new root files. The CLI exit-code half of item (a) is a new file in cmd/jsonschema. The peak-memory fix splits peakmem_test.go into a unix-only measurement, a non-unix stub and an unconstrained flags file. docs/benchmarks.md is the only crowded file written, and only because the story requires the limits recorded there.

## Affected modules
- allocs_test.go
- allocs_race_on_test.go
- allocs_race_off_test.go
- docs/benchmarks.md
- metaschema_custom_strict_test.go
- cmd/jsonschema/metaschema_custom_strict_cli_test.go
- metaschema_uri_resource_test.go
- first_compile_concurrency_test.go
- compile_cache_limits_test.go
- internal/compile/metacache_test.go
- cmd/jsonschema/peakmem_test.go
- cmd/jsonschema/peakmem_other_test.go
- cmd/jsonschema/peakmem_flags_test.go

## Data / API impact
- DB impact: no
- API impact: no
- Migration needed: no

## Constraints for implementation
- No non-test .go file may change. If a new test exposes a defect, stop and report it rather than fixing production code in this story.
- Standard library only, no network, and nothing added to go.mod or go.sum.
- Start every measurement from raw bytes. Schema compile includes loading the built-in meta-schema, and document scenarios include decoding.
- Keep the run counts small for the 5.5 MB scenario so the extra test time stays at a few seconds. Never skip it under -short.
- Measure memory as the TotalAlloc delta from runtime.ReadMemStats around the operation, after a warm-up run, averaged or taken over a fixed run count that is recorded in the docs.
- Limits are named constants equal to the measured value times about 1.1, with the measured value in a comment, and the docs repeat the same numbers.
- The re-exec child must be selected by an environment variable plus -test.run on os.Args[0], and must not touch Compile anywhere before the 32 goroutines are released.
- The peak-memory flags must be defined in a file without build constraints. Only the syscall.Getrusage use moves behind //go:build unix.
- No timing-based assertions anywhere.

## Risks
- Allocation counts differ between darwin and linux and between Go versions. The limits are measured on the build machine, so they may fail on the operator's Mac if its counts are over 10% higher. Per the operator's answer, that is handled by re-measuring on purpose.
- TotalAlloc wobbles more than counts do, which could make the byte limits flaky. The 10% headroom is the operator's choice.
- The 5.5 MB scenario adds seconds to every go test ./... run (accepted by the operator).
- A new 202 follow-up test may reveal a real defect that this story is not allowed to fix.

## Breaking changes
- none

## Sensitivity
- none
