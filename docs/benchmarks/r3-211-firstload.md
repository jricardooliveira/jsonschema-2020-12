# Round 3, story 211: build the built-in meta-schema index once

Machine: Apple M3 Max (16 cores), darwin/arm64, the repo's Go toolchain. Other
stories may have been building in parallel, so **all timings are noisy**. The
operator re-measures them interleaved at Checkpoint 3. B/op and allocs/op are
judged exactly.

- **Before** means main after #205. To get it, the same tree had
  `sharedBuiltin()` in `internal/compile/metacache.go` temporarily call
  `registry.BuildBuiltinIndex()`, which is the second index build that main does.
  That edit was then reverted.
- **After** means this story: `sharedBuiltin()` takes `registry.SharedBuiltinIndex()`.

`buildBuiltin()` still calls `registry.BuildBuiltinIndex()` from scratch in both
states. `BenchmarkFirstLoadOrder` calls it on every iteration, so it still
rebuilds the index on purpose, and the benchmark measures the same thing before
and after.

## First compile of the order schema (fresh process, median of 10)

Run: `go test -run TestFirstLoadTiming -firstload -v .`

| | time | bytes | allocs |
|---|---|---|---|
| before | 1,010,834 ns | 741,672 B | 10,647 |
| after (run 1) | 861,959 ns | 625,224 B | 8,653 |
| after (run 2) | 781,500 ns | 625,176 B | 8,652 |
| change | about -15 % to -23 % | -116,496 B (-15.7 %) | -1,995 |

## In-repo benchmarks (`go test -run '^$' -bench . -benchmem -count=10`, medians)

The packages are `.`, `./internal/compile`, `./internal/jsonvalue` and
`./cmd/jsonschema`. "After" is the second full after run. The first full after
run was slower everywhere, even in `internal/jsonvalue`, which does not use the
registry, so it is treated as machine noise. Rows marked (i) were re-run
interleaved: after, then before, back to back.

| benchmark | before ns/op | after ns/op | B/op before → after | allocs/op before → after |
|---|---|---|---|---|
| CompileOrder | 112,343 | 110,267 | 159,841 → 159,824 (jitter ±0.5 KB in both) | 1,490 → 1,490 |
| ValidateSmall | 24,357 | 25,446 (+4.5 %) | 30,712 → 30,712 | 558 → 558 |
| Validate0_5MB | 23.02 ms | 22.85 ms | 25.78 MB → 25.78 MB | 460,052 → 460,052 |
| Validate5_5MB | 221.1 ms | 221.2 ms | 268.95 MB → 268.95 MB | 4,600,073 → 4,600,074 |
| ValidateErrors | 17.50 ms | 17.28 ms | 21.28 MB → 21.28 MB | 320,029 → 320,029 |
| ValidateErrorsBasic | 17.71 ms | 18.05 ms (+1.9 %) | 23.26 MB → 23.26 MB | 326,057 → 326,056 |
| compile FirstLoadOrder | 413.6 µs | 440.3 µs (+6.5 %) | 340.2 KB → 340.3 KB | 6,020/6,021 → 6,020/6,021 |
| jsonvalue RetainedHeap/0.5MB (\*) | 5.48 ms | 5.78 ms | 6.28 MB → 6.28 MB (retained 4,340,755) | 170,186 → 170,191 |
| jsonvalue RetainedHeap/5.5MB (\*) | 59.99 ms | 60.75 ms | 67.38 MB → 67.38 MB (retained 44,924,6xx) | 1,715,811 → 1,715,811 |
| jsonvalue RetainedHeapExact/0.5MB (\*) | 6.09 ms | 6.73 ms | 7.24 MB → 7.24 MB (retained 5,300,810) | 170,201 → 170,215 |
| jsonvalue RetainedHeapExact/5.5MB (\*) | 65.65 ms | 66.05 ms | 77.04 MB → 77.04 MB | 1,716,686 → 1,716,687 |
| CLI/0.5MB (i) | 31.81 ms | 33.47 ms (+5.2 %) | 39.99 MB → 39.99 MB | 721,731 → 721,731 |
| CLI/5.5MB | 304.4 ms | 308.6 ms | 399.20 MB → 399.20 MB | 7,201,761 → 7,201,760 |
| CLI/Errors | 18.61 ms | 19.24 ms (+3.4 %) | 24.57 MB → 24.57 MB | 327,666 → 327,665 |
| CLI/Verbose0.5MB (i) | 72.75 ms | 74.11 ms (+1.9 %) | 109.57 MB → 109.57 MB | 1,481,841 → 1,481,841 |

(\*) The jsonvalue rows come from the first after run. That package does not
import the registry or the compiler, so its code is identical in both states.
Its spread of up to +10 % shows how noisy timings were on this machine.

## Verdict

- **First compile:** the medians are lower in time, bytes and allocations, which
  meets the success criterion. A process now builds the built-in index once
  (`firstload_count_test.go`). The first-load allocation limit is 9,520 against a
  measured 8,653 (`allocs_firstload_test.go`).
- **B/op and allocs/op:** unchanged in every in-repo benchmark, within ±15 allocs
  of run-to-run jitter. Per operation the code is the same: the change only
  affects what is built once per process.
- **FirstLoadOrder:** measures the same thing as before (6,020 allocs/op, about
  340 KB) because it still rebuilds the index.
- **Timings:** several after medians are higher, by up to +5.2 % on CLI/0.5MB
  interleaved and +4.5 % on ValidateSmall. The same spread also appears in
  `internal/jsonvalue`, which this story cannot affect. Following the story's
  rule, these are recorded and **held for the operator's decision**. They are
  provisional until the interleaved re-measure at Checkpoint 3.
- **Tests:** `go test ./...` (golden CLI test included) and
  `go test -race` on `.`, `./internal/compile` and `./internal/registry` are green.
