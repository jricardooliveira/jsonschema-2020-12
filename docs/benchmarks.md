# Benchmarks

The project measures its own speed and memory with standard-library Go
benchmarks. No other validator is used or compared here.

## Command

All benchmarks, with memory figures:

    go test -run '^$' -bench . -benchmem ./...

For a baseline, repeat each benchmark 10 times:

    go test -run '^$' -bench . -benchmem -count=10 ./...

Run one package at a time (for example `./cmd/jsonschema`) when you want quiet,
comparable timings, and do not run other work on the machine meanwhile.

The inputs are the order schema and the 492-byte document in `testdata/perf/`.
The large documents are generated in code by `internal/perfdata` from a fixed
recipe, so they are identical on every run and are not committed:

| Document | Items | Size |
|---|---|---|
| small valid | 10 | 492 B |
| medium valid | 10,000 | about 0.5 MB |
| large valid | 100,000 | about 5.5 MB |
| error | 10,000 (1,000 with a bad `sku`) | about 0.33 MB |

## Benchmarks

| Name | Package | What it times |
|---|---|---|
| `BenchmarkCompileOrder` | root | Compile the order schema |
| `BenchmarkValidateSmall` | root | Parse and validate the small document |
| `BenchmarkValidate0_5MB` | root | Parse and validate the 0.5 MB document |
| `BenchmarkValidate5_5MB` | root | Parse and validate the 5.5 MB document |
| `BenchmarkValidateErrors` | root | Parse and validate the error document |
| `BenchmarkValidateErrorsBasic` | root | The same, plus rendering the basic report |
| `BenchmarkRetainedHeap/0.5MB`, `/5.5MB` | internal/jsonvalue | Heap kept alive by the decoded document |
| `BenchmarkCLI/0.5MB`, `/5.5MB`, `/Errors`, `/Verbose0.5MB` | cmd/jsonschema | One in-process `run` call on a file, output discarded |

## How to read the numbers

- **ns/op**: time for one operation (one compile, or one parse and validate).
  Lower is better. Timings move by about 5% from run to run.
- **MB/s**: document size divided by the time. Higher is better. Shown only for
  the document benchmarks (they call `b.SetBytes`).
- **B/op**: bytes allocated on the heap per operation, including garbage that is
  freed again. A proxy for memory pressure.
- **allocs/op**: number of heap allocations per operation. This is the most
  stable figure; it barely changes between runs.
- **retained-B**: heap still alive after decoding one document, measured with
  `runtime.GC` and `runtime.ReadMemStats` before and after, with
  `runtime.KeepAlive` on the value. It is the in-project stand-in for peak
  memory, and it shows the effect of how numbers and values are stored.

To compare a later round with this baseline, run the same command on the same
machine and compare medians. A change smaller than the spread is noise.

## Baseline of this story (measured before any fix)

Measured on the build machine, not the operator's Mac. Each benchmark was run
10 times (`-count=10`). The table gives the median and the min–max range of the
10 runs. B/op and allocs/op are medians; they varied by well under 1%.

| | |
|---|---|
| Machine | Apple M3 Max, 16 logical CPUs (`-16`) |
| OS | macOS (Darwin 25.6.0), arm64 |
| Go | go1.27.2 darwin/arm64 |
| Runs | 10 per benchmark |

Each package was run on its own, one after another. A separate full-tree run
of the same command was going on in the background at the start, so a few
timings may be a little high; allocation counts and retained heap are not
affected.

### Library (package root)

| Benchmark | ns/op median (min–max) | MB/s | B/op | allocs/op |
|---|---|---|---|---|
| CompileOrder | 1,046,911 (1,019,577–1,069,349) | – | 1,040,111 | 28,410 |
| ValidateSmall | 52,816 (52,486–53,787) | 9.31 | 75,195 | 1,582 |
| Validate0_5MB | 49,025,188 (48,057,646–50,260,439) | 10.54 | 73,940,970 | 1,423,839 |
| Validate5_5MB | 453,248,167 (451,129,153–460,111,333) | 12.06 | 755,891,300 | 14,183,144 |
| ValidateErrors | 32,862,684 (32,595,169–33,318,962) | 10.14 | 53,849,900 | 844,097 |
| ValidateErrorsBasic | 33,828,981 (33,360,468–34,037,904) | 9.85 | 55,836,000 | 850,123 |

### Decoded document (package internal/jsonvalue)

| Benchmark | ns/op median (min–max) | retained-B | B/op | allocs/op |
|---|---|---|---|---|
| RetainedHeap/0.5MB | 6,125,789 (6,046,646–6,269,659) | 4,340,750 | 6,282,100 | 170,200 |
| RetainedHeap/5.5MB | 60,273,987 (59,654,511–60,468,647) | 44,924,650 | 67,382,585 | 1,715,812 |

### CLI, in process (package cmd/jsonschema)

| Benchmark | ns/op median (min–max) | MB/s | B/op | allocs/op |
|---|---|---|---|---|
| CLI/0.5MB | 59,998,600 (58,938,088–60,481,050) | 8.6 | 91,260,000 | 1,502,415 |
| CLI/5.5MB | 538,371,781 (534,553,646–550,519,438) | 10.1 | 909,393,000 | 14,711,733 |
| CLI/Errors | 38,781,395 (37,436,460–39,063,515) | 8.6 | 63,067,100 | 878,657 |
| CLI/Verbose0.5MB | 125,118,900 (115,829,713–127,075,616) | 4.1 | 195,700,000 | 2,956,264 |

The CLI figures include reading the file and writing the report to
`io.Discard`, which the library benchmarks do not.

## Story 199 (lazy keyword locations): before

The "before" set for story 199, measured on the code as it stood before any
change of this story. Same machine and Go version as the baseline above
(Apple M3 Max, 16 logical CPUs, macOS arm64, go1.27.2). Medians of 10 runs
(for 10 values, the mean of the 5th and 6th sorted). The "after" set is added
next to this one when the change is done.

Command (one package at a time, `-count=10`):

    go test -run '^$' -bench . -benchmem -count=10 .
    go test -run '^$' -bench . -benchmem -count=10 ./internal/jsonvalue
    go test -run '^$' -bench . -benchmem -count=10 ./cmd/jsonschema

Note: a full-tree run of the same command was going on in the background
during part of the measurement, so timings may be a little high or noisy
(`CLI/Errors` shows two clusters, about 35 ms and 37.7 ms). Allocation counts
and retained heap are not affected.

### Library (package root)

| Benchmark | ns/op | MB/s | B/op | allocs/op |
|---|---|---|---|---|
| CompileOrder | 1,041,904 | – | 1,039,896 | 28,410 |
| ValidateSmall | 53,168 | 9.25 | 75,195 | 1,582 |
| Validate0_5MB | 49,923,275 | 10.35 | 73,940,908 | 1,423,839 |
| Validate5_5MB | 454,568,674 | 12.03 | 755,891,277 | 14,183,144 |
| ValidateErrors | 32,893,218 | 10.12 | 53,849,932 | 844,097 |
| ValidateErrorsBasic | 34,122,732 | 9.76 | 55,835,666 | 850,122 |

### Decoded document (package internal/jsonvalue)

| Benchmark | ns/op | retained-B | B/op | allocs/op |
|---|---|---|---|---|
| RetainedHeap/0.5MB | 6,120,433 | 4,340,755 | 6,282,191 | 170,201 |
| RetainedHeap/5.5MB | 60,079,201 | 44,924,652 | 67,382,582 | 1,715,811 |

### CLI, in process (package cmd/jsonschema)

| Benchmark | ns/op | MB/s | B/op | allocs/op |
|---|---|---|---|---|
| CLI/0.5MB | 59,672,058 | 8.66 | 91,260,846 | 1,502,415 |
| CLI/5.5MB | 531,462,573 | 10.29 | 909,393,436 | 14,711,730 |
| CLI/Errors | 36,637,432 | 9.09 | 63,067,414 | 878,656 |
| CLI/Verbose0.5MB | 125,242,953 | 4.13 | 195,722,282 | 2,956,262 |

## Story 199 (lazy keyword locations): after, side by side

The "after" set, measured on the finished change with the same machine, Go
version, commands and method as the "before" set above (10 runs per
benchmark, one package at a time, median; for 10 values, the mean of the 5th
and 6th sorted). Change is after against before; negative is better.

### Library (package root)

| Benchmark | ns/op before → after | B/op before → after | allocs/op before → after |
|---|---|---|---|
| CompileOrder | 1,041,904 → 894,530 (−14.2%) | 1,039,896 → 900,985 (−13.4%) | 28,410 → 25,922 (−8.8%) |
| ValidateSmall | 53,168 → 35,766 (−32.7%) | 75,195 → 56,517 (−24.8%) | 1,582 → 1,132 (−28.4%) |
| Validate0_5MB | 49,923,275 → 33,672,266 (−32.6%) | 73,940,908 → 55,375,300 (−25.1%) | 1,423,839 → 973,830 (−31.6%) |
| Validate5_5MB | 454,568,674 → 314,837,599 (−30.7%) | 755,891,277 → 570,285,550 (−24.6%) | 14,183,144 → 9,683,136 (−31.7%) |
| ValidateErrors | 32,893,218 → 21,618,687 (−34.3%) | 53,849,932 → 40,656,891 (−24.5%) | 844,097 → 510,077 (−39.6%) |
| ValidateErrorsBasic | 34,122,732 → 22,778,565 (−33.2%) | 55,835,666 → 42,642,749 (−23.6%) | 850,122 → 516,101 (−39.3%) |

Throughput (MB/s) after: ValidateSmall 13.76, Validate0_5MB about 15.3,
Validate5_5MB 17.36, ValidateErrors about 15.4, ValidateErrorsBasic about 14.6.

### Decoded document (package internal/jsonvalue)

This package is not touched by the story; the figures only confirm no change.

| Benchmark | ns/op before → after | retained-B before → after | B/op before → after | allocs/op before → after |
|---|---|---|---|---|
| RetainedHeap/0.5MB | 6,120,433 → 5,944,167 (noise) | 4,340,755 → 4,340,750 | 6,282,191 → 6,281,822 | 170,201 → about 170,190 |
| RetainedHeap/5.5MB | 60,079,201 → 60,236,852 (noise) | 44,924,652 → 44,924,640 | 67,382,582 → 67,382,570 | 1,715,811 → 1,715,811 |

### CLI, in process (package cmd/jsonschema)

| Benchmark | ns/op before → after | B/op before → after | allocs/op before → after |
|---|---|---|---|
| CLI/0.5MB | 59,672,058 → 41,480,889 (−30.5%) | 91,260,846 → 66,151,492 (−27.5%) | 1,502,415 → 1,049,917 (−30.1%) |
| CLI/5.5MB | 531,462,573 → 377,559,097 (−29.0%) | 909,393,436 → 659,646,050 (−27.5%) | 14,711,730 → 10,209,229 (−30.6%) |
| CLI/Errors | 36,637,432 → 25,539,853 (−30.3%) | 63,067,414 → 44,679,631 (−29.2%) | 878,656 → 542,143 (−38.3%) |
| CLI/Verbose0.5MB | 125,242,953 → 88,516,996 (−29.3%) | 195,722,282 → 145,015,790 (−25.9%) | 2,956,262 → 2,053,762 (−30.5%) |

### Acceptance check

- Every validation scenario (library and CLI) has lower allocs/op and lower
  B/op: allocs/op by 28% to 40%, B/op by 24% to 29%.
- No scenario is higher on allocs/op or B/op. Compiling the order schema is
  lower too (this story does not touch the compiler; the drop comes from the
  code base as measured now and is not claimed as a result of this story).
- ns/op is lower on every validation scenario (about 29% to 34%), so it is
  not worse beyond the ±5% noise. The speed gain is recorded, not required.
- The decoded-document figures are unchanged, as expected.
- No allocation-limit test was added or changed.

## Operator baseline (the operator's Mac, from the brief)

These are the operator's own measurements, copied from the brief. They are a
different machine and a different Go version, so do not compare them directly
with the table above. Hardware: one Mac, 16 cores, Go 1.27.2. Timings vary
about ±5%; allocation counts are reliable. The brief gives no spread.

| Scenario | Time | Memory / allocations |
|---|---|---|
| Library: small document (about 0.5 KB), per validation | 56 µs | 1,582 allocs |
| Library: 500 KB valid document | 48 ms | 74 MB allocated |
| Library: 330 KB document with 1,000 errors | 33 ms | 54 MB allocated |
| Library: compile the schema | 1.1 ms | about 28k allocs |
| Library: throughput | about 10 MB/s | – |
| Library: parse the 500 KB document alone with `encoding/json` | 7.7 ms | – |
| CLI: 0.5 MB document | 70 ms | 39 MB peak |
| CLI: 5.5 MB document | 580 ms | 330 MB peak (about 60× the file) |
| CLI: 0.33 MB document with 1,000 errors | 50 ms | 32 MB peak |
| CLI: 0.5 MB document, verbose output | 140 ms | 65 MB peak |

The operator's CLI figures are of the real program (peak process memory); the
CLI benchmarks above run in process and report allocated bytes, so only the
trend, not the absolute number, is comparable.
