# Release notes — #21: In-project benchmarks and baseline for the order schema

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #21
- Status: proposed (pending Checkpoint 3 sign-off)

## What changed

Adds comprehensive Go benchmarks to measure the validator's speed and memory before performance work begins. Establishes a baseline with 10-run median and spread measurements. No behavior changes, standard library only, all fixtures and tests work as specified.

- testdata/perf/order_schema.json: order schema fixture from the request
- testdata/perf/order_small.json: 492-byte valid order document (10 items)
- internal/perfdata/: deterministic generator for valid and error orders from a fixed recipe (standard library only, no randomness)
- internal/perfdata/perfdata_test.go: tests that generation is repeatable and matches the 492-byte fixture
- perf_fixtures_test.go: root-level tests that the schema compiles, generated documents validate correctly, and the error document yields exactly 1,000 pattern errors
- bench_test.go: six library benchmarks (compile, small, 0.5 MB, 5.5 MB, error, error+basic), each with b.SetBytes for throughput
- internal/jsonvalue/bench_test.go: retained-heap benchmarks for decoded 0.5 MB and 5.5 MB documents using runtime.GC and runtime.ReadMemStats
- cmd/jsonschema/bench_test.go: four in-process CLI benchmarks calling the existing run entry point with temp files and discarded output
- docs/benchmarks.md: command to run benchmarks, guide to reading the numbers, this story's 10-run baseline with medians and min–max spreads, and the operator's Mac numbers in a separate labelled table
- spec_test.go: added 'perf' to the directory skip list so testdata/perf is not processed as spec test cases

## How to verify

- Run `go test -run '^$' -bench . -benchmem ./...` to see all benchmarks execute and report MB/s, allocs/op and ns/op
- Run `go test -run PerfFixture . -v` to verify the schema compiles, generated documents are valid, and the error document has exactly 1,000 pattern errors
- Run `go test ./internal/perfdata -v` to confirm the recipe matches the 492-byte stored fixture byte for byte and generation is repeatable
- Check docs/benchmarks.md: baseline table lists medians for all benchmarks with min–max ranges across 10 runs, machine details (M3 Max, 16 CPUs, Go 1.27.2), and operator Mac numbers in a separate table

## Migration

None. This story adds measurement only and does not change validator behaviour, schema loading, or library output.

## Rollback

Simple git revert of all new files and the spec_test.go edit. The change introduces only testdata, test code and documentation; no library or CLI code is modified.

## Known limitations

- Benchmark timings vary by about ±5% from run to run; changes smaller than the min–max spread are noise
- Baseline measured on the build machine (Apple M3 Max, 16 CPUs), not the operator's Mac, so absolute timings will differ; only the trend is comparable
- Baseline may include slight inflation from background load running at start of measurement; allocation counts (allocs/op, retained-B) are not affected
- Retained-heap metric depends on GC timing; the benchmark runs multiple GCs before and after to stabilize the reading

## Concerns

- spec_test.go modified outside the declared scope to skip testdata/perf — the change is minimal and necessary (one directory name added to the skip list), and the tester rated it low severity. The perf directory must not be processed as spec test cases.
- Baseline timings slightly inflated by concurrent background work at the start of measurement — documented in docs/benchmarks.md and acceptable for a baseline. The operator must rerun on the same machine in stories 199–204 to compare.
