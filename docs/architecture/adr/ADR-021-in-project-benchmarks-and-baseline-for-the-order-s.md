# ADR #21: In-project benchmarks and baseline for the order schema

- Date: 2026-10-17
- Story: #21
- Status: proposed (pending architecture sign-off)

## Decision
All work goes into new files, following the convention that performance guards live in <package>/bench_test.go. The recipe lives in a new package, internal/perfdata, so stories 199–204 reuse the same bytes. That package imports only the standard library, and it never imports the root package, which avoids an import cycle; validity checks that need the library sit in a root test file. internal/perfdata exposes ValidOrder(n int) []byte, ErrorOrder() []byte, and a helper that reads testdata/perf files by a path relative to the module root (located from the caller's package directory). The retained-heap benchmark sits in internal/jsonvalue because it measures that package's decoded form. Benchmark names should be stable (for example BenchmarkCompileOrder, BenchmarkValidateSmall, BenchmarkValidate0_5MB, BenchmarkValidate5_5MB, BenchmarkValidateErrors, BenchmarkValidateErrorsBasic, BenchmarkRetainedHeap/0.5MB, BenchmarkCLI/...), because later stories compare against them. The CLI benchmarks use the existing run entry point in cmd/jsonschema/run.go, read only. The crowded files run.go and main.go are touched only if that entry point cannot take args, streams and return a code.

## Affected modules
- testdata/perf/order_schema.json
- testdata/perf/order_small.json
- internal/perfdata/perfdata.go
- internal/perfdata/perfdata_test.go
- perf_fixtures_test.go
- bench_test.go
- internal/jsonvalue/bench_test.go
- cmd/jsonschema/bench_test.go
- docs/benchmarks.md

## Data / API impact
- DB impact: no
- API impact: no
- Migration needed: no

## Constraints for implementation
- Standard library only. go.mod and go.sum are unchanged, with no other validator anywhere.
- No network. Temporary files come from b.TempDir() or os.CreateTemp under it.
- Large documents are generated once per benchmark, before b.ResetTimer, and never committed.
- Document benchmarks include parsing inside the timed loop and call b.SetBytes(len(doc)).
- Retained heap: runtime.GC, then ReadMemStats, then decode, then runtime.GC, then ReadMemStats. Report the HeapAlloc difference via b.ReportMetric(..., "retained-B") and keep the value alive with runtime.KeepAlive.
- CLI output goes to io.Discard. The benchmark checks the exit code (0 valid, 1 invalid) so a broken run is not timed silently.
- Ordinary tests make no timing assertions. Tests ask perfdata for counts instead of repeating them where practical, but the 492-byte size is the stated contract and may be asserted.
- No validator behaviour changes. If a thin CLI entry point must be added, main.go only delegates, and every existing test passes unchanged.
- Baseline uses -count=10 and records the median and the spread (min–max or percent). Machine, OS, CPU count and Go version go in docs/benchmarks.md.
- The operator's Mac numbers sit in a separate labelled table and are never merged with this story's numbers.

## Risks
- The 5.5 MB valid check in ordinary tests may take a second or more. Keep it to one validation, or skip it under testing.Short().
- Retained-heap readings vary with GC timing. Several GC calls before reading may be needed for a stable number.
- docs/benchmarks.md commands might be picked up by cmd/jsonschema/docs_examples_test.go or docs_flags_test.go if those scan docs/ for CLI examples.
- The baseline is measured on the build machine, not the operator's Mac, so the two tables will differ by design.

## Breaking changes
- none

## Sensitivity
- none
