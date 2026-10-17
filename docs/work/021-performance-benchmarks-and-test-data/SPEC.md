# SPEC #21: In-project benchmarks and baseline for the order schema

- Date: 2026-10-17
- Type: feature
- Verdict: pass
- Status: passed gate 1 with nothing for the operator to decide (run #156, 2026-10-17 10:33 UTC)

## Problem

Later stories (199–204) must show a measured gain in speed and memory, and none may regress any scenario. The project has no benchmarks, no stored performance fixtures and no recorded baseline, so before and after cannot be compared inside the project.

## Why this matters

The operator's success criterion for this round is roughly matching the outside comparison in speed and memory. The project itself must guard the gains with standard-library Go benchmarks that cover the same scenarios as the operator's own measurements: compile, small, 0.5 MB and 5.5 MB valid documents, the 1,000-error document, error plus basic report, and the CLI runs including the error document and verbose output. The baseline must be recorded before any fix.

## Acceptance criteria

- testdata/perf/order_schema.json holds the order schema from the request, unchanged in meaning: same keywords, values, $defs and $ref. The library compiles it without error.
- testdata/perf/order_small.json is the valid document of 10 items, exactly 492 bytes. Its bytes are exactly what the recipe produces, with separators ", " and ": " and no trailing newline. Item i (from 0) is {"sku": "ABC-i", "qty": i, "price": i.99}.
- Package internal/perfdata uses only the standard library and generates, deterministically and without randomness, the valid document for any item count (used with 10, 10,000 and 100,000 items) and the error document. The error document is {"id": 1, "items": [...]} with 10,000 items {"sku": "ABC-i", "qty": i}, where sku is "bad" when i is divisible by 10. Repeated calls return identical bytes. No large document is committed to the repository.
- Ordinary tests, not benchmarks, check the inputs. The recipe for 10 items equals the stored small document byte for byte and is 492 bytes long. The generated 10,000-item and 100,000-item valid documents are valid against the order schema. The error document is invalid, with exactly 1,000 failing 'pattern' errors in the basic report. These tests run in well under a second apart from the 5.5 MB check and use no timing assertions.
- Library benchmarks in bench_test.go at the module root cover these scenarios: compiling the order schema; validating the small, 0.5 MB and 5.5 MB valid documents; validating the 1,000-error document; and validating the 1,000-error document plus rendering the basic report. Each document benchmark includes parsing the JSON in its timed loop and calls b.SetBytes with the document size, so MB/s is reported. Large documents are generated once, outside the timer.
- A benchmark reports the retained heap of a decoded 0.5 MB and 5.5 MB valid document, in the validator's own decoded form (internal/jsonvalue), as a custom metric with b.ReportMetric. The measurement uses runtime.ReadMemStats after runtime.GC before and after decoding, and keeps the value alive with runtime.KeepAlive.
- In-process CLI benchmarks in cmd/jsonschema/bench_test.go call the CLI's run entry point directly, with no subprocess, for these runs: the 0.5 MB valid document; the 5.5 MB valid document; the 1,000-error document; and the 0.5 MB document with the verbose output format. Each document is written to a temporary file once before the timer starts and passed as a file path. The report output is discarded with io.Discard, and each run's exit code is checked to be the expected 0 or 1.
- If the CLI does not already have an entry point that takes arguments, input and output streams and returns the exit code, a thin one is added. main.go only calls it, so exit codes and output stay identical, and all existing CLI tests pass unchanged.
- docs/benchmarks.md gives the exact command (go test -run '^$' -bench . -benchmem ./..., with -count=10 for baselines) and explains how to read ns/op, MB/s, B/op, allocs/op and the retained-heap metric. It records this story's baseline, measured before any fix: each benchmark run 10 times, with the middle value and the spread, plus the machine, OS, CPU count, Go version and run count. It has a separate, clearly labelled table with the operator's own Mac baseline numbers from the brief, and the two baselines are never mixed.
- No other validator appears in go.mod, go.sum or any test. go.mod is unchanged and gains no require line. No test or benchmark uses the network. Validator behaviour is unchanged, and go test ./... and go test -race ./... pass.

## Non-goals

- Any performance fix (stories 199–203)
- Allocation-limit tests (story 204)
- Comparison with any other validator inside the project
- Changing the verbose second evaluation pass
- Timing-based assertions in ordinary tests
- Committing generated large documents
- Using benchstat or any non-stdlib tool

## Tasks

- **T1** — Perf fixtures and deterministic generator
  - purpose: Store the order schema and the 492-byte small document. Add internal/perfdata, which builds the valid and error documents from the fixed recipe and also reads the stored fixtures from testdata/perf.
  - scope: testdata/perf/order_schema.json, testdata/perf/order_small.json, internal/perfdata/perfdata.go, internal/perfdata/perfdata_test.go
  - done when: go test ./internal/perfdata passes: the recipe for 10 items equals the stored file byte for byte and is 492 bytes, repeated generation gives identical bytes, and the error document has 1,000 items with sku "bad".
  - delivers: AC1, AC2, AC3, AC4
- **T2** — Fixture correctness checks through the library
  - purpose: Prove the benchmark inputs measure what they claim: the schema compiles, the generated valid documents pass, and the error document gives exactly 1,000 pattern errors in the basic report.
  - scope: perf_fixtures_test.go
  - done when: go test -run PerfFixture . passes.
  - depends on: T1
  - delivers: AC1, AC4
- **T3** — Library benchmarks
  - purpose: Add testing.B benchmarks for compile, the small, 0.5 MB and 5.5 MB valid documents, the error document, and the error document plus basic report. Each includes parsing and calls b.SetBytes.
  - scope: bench_test.go
  - done when: go test -run '^$' -bench . -benchmem . prints all six benchmarks with MB/s on the document benchmarks.
  - depends on: T1
  - delivers: AC5
- **T4** — Retained-heap metric of decoded documents
  - purpose: Report the retained heap of decoded 0.5 MB and 5.5 MB documents, in the validator's own form, through b.ReportMetric after runtime.GC and runtime.ReadMemStats.
  - scope: internal/jsonvalue/bench_test.go
  - done when: go test -run '^$' -bench Retained -benchmem ./internal/jsonvalue prints a retained-bytes metric for both sizes.
  - depends on: T1
  - delivers: AC6
- **T5** — In-process CLI benchmarks
  - purpose: Benchmark the CLI run entry point with temporary file paths and discarded output for the 0.5 MB, 5.5 MB, 1,000-error and 0.5 MB verbose runs. Add a thin entry point only if none exists.
  - scope: cmd/jsonschema/bench_test.go
  - done when: go test -run '^$' -bench . -benchmem ./cmd/jsonschema prints four CLI benchmarks with the expected exit codes, and every existing cmd/jsonschema test passes.
  - depends on: T1
  - delivers: AC7, AC8
- **T6** — Measure baseline and write docs/benchmarks.md
  - purpose: Run every benchmark 10 times before any fix. Record the middle value and the spread with the machine details. Add the operator's Mac numbers in a separate labelled table, plus the command and how to read the numbers.
  - scope: docs/benchmarks.md
  - done when: docs/benchmarks.md contains the command, the reading guide, this story's baseline table (median and spread, 10 runs, machine and Go version) and the operator baseline table. go test ./... and go test -race ./... pass, and go.mod is unchanged.
  - depends on: T2, T3, T4, T5
  - delivers: AC9, AC10
