# PLAN #21: In-project benchmarks and baseline for the order schema

- Date: 2026-10-17
- Status: proposed (pending Checkpoint 2 sign-off)
- Design: see `ADR-021-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `testdata/perf/order_schema.json`
- `testdata/perf/order_small.json`
- `internal/perfdata/perfdata.go`
- `internal/perfdata/perfdata_test.go`
- `perf_fixtures_test.go`
- `bench_test.go`
- `internal/jsonvalue/bench_test.go`
- `cmd/jsonschema/bench_test.go`
- `docs/benchmarks.md`

Modules the architecture expects to touch:

- `testdata/perf/order_schema.json`
- `testdata/perf/order_small.json`
- `internal/perfdata/perfdata.go`
- `internal/perfdata/perfdata_test.go`
- `perf_fixtures_test.go`
- `bench_test.go`
- `internal/jsonvalue/bench_test.go`
- `cmd/jsonschema/bench_test.go`
- `docs/benchmarks.md`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — Perf fixtures and deterministic generator
2. **T2** — Fixture correctness checks through the library  _(after T1)_
3. **T3** — Library benchmarks  _(after T1)_
4. **T4** — Retained-heap metric of decoded documents  _(after T1)_
5. **T5** — In-process CLI benchmarks  _(after T1)_
6. **T6** — Measure baseline and write docs/benchmarks.md  _(after T2, T3, T4, T5)_

## Risks

- The 5.5 MB valid check in ordinary tests may take a second or more. Keep it to one validation, or skip it under testing.Short().
- Retained-heap readings vary with GC timing. Several GC calls before reading may be needed for a stable number.
- docs/benchmarks.md commands might be picked up by cmd/jsonschema/docs_examples_test.go or docs_flags_test.go if those scan docs/ for CLI examples.
- The baseline is measured on the build machine, not the operator's Mac, so the two tables will differ by design.

Implementation constraints the coder must honor:

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

- DB impact: no · API impact: no · Migration needed: no

Breaking changes:

- (none)

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: go test ./internal/perfdata passes: the recipe for 10 items equals the stored file byte for byte and is 492 bytes, repeated generation gives identical bytes, and the error document has 1,000 items with sku "bad".
- **T2**: go test -run PerfFixture . passes.
- **T3**: go test -run '^$' -bench . -benchmem . prints all six benchmarks with MB/s on the document benchmarks.
- **T4**: go test -run '^$' -bench Retained -benchmem ./internal/jsonvalue prints a retained-bytes metric for both sizes.
- **T5**: go test -run '^$' -bench . -benchmem ./cmd/jsonschema prints four CLI benchmarks with the expected exit codes, and every existing cmd/jsonschema test passes.
- **T6**: docs/benchmarks.md contains the command, the reading guide, this story's baseline table (median and spread, 10 runs, machine and Go version) and the operator baseline table. go test ./... and go test -race ./... pass, and go.mod is unchanged.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
