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
| `BenchmarkRetainedHeap/0.5MB`, `/5.5MB` | internal/jsonvalue | Heap kept alive by the decoded document (`Decode`, `json.Number` leaves) |
| `BenchmarkRetainedHeapExact/0.5MB`, `/5.5MB` | internal/jsonvalue | The same on the validator's decode path (`DecodeLimited`, `*Number` leaves) |
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

## Story 201 (annotations collected and sorted only when shown): before and after

Both sets were measured on the same machine as above (Apple M3 Max, 16
logical CPUs, macOS arm64, go1.27.2), one package at a time, `-count=5`,
median of 5. "Before" is the code at the start of this story (commit 386bf22,
which already includes stories 199 and 200); "after" is the finished story.
Change is after against before; negative is better.

Command:

    go test -run '^$' -bench . -benchmem -count=5 . ./internal/jsonvalue ./cmd/jsonschema

### Library (package root)

| Benchmark | ns/op before → after | B/op before → after | allocs/op before → after |
|---|---|---|---|
| CompileOrder | 899,140 → 673,243 (−25.1%) | 900,842 → 696,303 (−22.7%) | 25,922 → 14,072 (−45.7%) |
| ValidateSmall | 35,704 → 30,333 (−15.0%) | 56,516 → 51,208 (−9.4%) | 1,132 → 846 (−25.3%) |
| Validate0_5MB | 33,949,334 → 28,212,720 (−16.9%) | 55,375,356 → 50,724,540 (−8.4%) | 973,830 → 710,113 (−27.1%) |
| Validate5_5MB | 314,784,073 → 271,604,364 (−13.7%) | 570,285,988 → 524,157,112 (−8.1%) | 9,683,136 → 7,100,144 (−26.7%) |
| ValidateErrors | 22,111,058 → 22,293,748 (+0.8%, noise) | 40,656,864 → 40,657,289 (+0.001%, noise) | 510,077 → 510,080 (noise) |
| ValidateErrorsBasic | 22,337,740 → 22,179,905 (−0.7%) | 42,642,596 → 42,642,832 (+0.0006%, noise) | 516,100 → 516,102 (noise) |

The two error scenarios have no valid-result annotations to collect, so they
are unchanged: the differences are a few hundred bytes and 2 to 3 allocations
out of 40 MB and 510,000, inside the run-to-run spread (before-runs of
ValidateErrors alone ranged 510,076 to 510,079 allocations).

### Decoded document (package internal/jsonvalue)

Not touched by this story; shown only to confirm no change.

| Benchmark | ns/op before → after | retained-B before → after | B/op before → after | allocs/op before → after |
|---|---|---|---|---|
| RetainedHeap/0.5MB | 5,999,266 → 5,930,574 (noise) | 4,340,753 → 4,340,754 | 6,282,123 → 6,281,994 | 170,200 → 170,197 |
| RetainedHeap/5.5MB | 60,080,042 → 59,823,958 (noise) | 44,924,650 → 44,924,646 | 67,382,574 → 67,382,582 | 1,715,811 → 1,715,811 |

### CLI, in process (package cmd/jsonschema)

| Benchmark | ns/op before → after | B/op before → after | allocs/op before → after |
|---|---|---|---|
| CLI/0.5MB | 41,726,897 → 40,222,769 (−3.6%) | 66,151,160 → 65,472,149 (−1.0%) | 1,049,917 → 984,372 (−6.2%) |
| CLI/5.5MB | 376,142,181 → 370,964,347 (−1.4%) | 659,647,045 → 654,931,621 (−0.7%) | 10,209,231 → 9,714,408 (−4.8%) |
| CLI/Errors | 25,310,125 → 24,038,087 (−5.0%) | 44,679,512 → 44,477,816 (−0.5%) | 542,144 → 530,295 (−2.2%) |
| CLI/Verbose0.5MB | 86,826,378 → 87,357,247 (+0.6%, noise) | 145,015,786 → 144,733,062 (−0.2%) | 2,053,761 → 1,934,518 (−5.8%) |

The CLI gain is smaller than the library gain because the CLI renders the
report, which needs the annotations, so they are built and sorted on demand
there (once, in place, with the segments of each pointer computed once).

### Acceptance check

- Valid-document scenarios (library and CLI) have lower allocs/op and lower
  B/op; allocs/op by 2% to 27%, B/op by 0.2% to 9%.
- No scenario's allocs/op or B/op rose beyond a few allocations or bytes of
  noise (the two error scenarios in the library table).
- ns/op is not worse beyond the ±5% spread on any scenario.
- The compile figures are lower as well (the meta-schema check at Compile no
  longer sorts annotations up front); this is recorded as measured.
- No allocation-limit test was added or changed.

## Story 200 (evaluated properties and items recorded only when needed): before and after

Both sets were measured on the same machine as above (Apple M3 Max, 16
logical CPUs, macOS arm64, go1.27.2), in the same tree and the same session,
`-count=5`, median of 5.

- **Before**: the finished code with the load-time decision forced off
  (`n.NoUnevaluated = false`, so recording stays on everywhere, as it did
  before this story). This was done by a temporary one-line edit of
  `jsonschema.go`, reverted before the "after" run. Its figures equal the
  story 201 "after" figures to within noise, as expected.
- **After**: the finished code, unchanged.

Command:

    go test -run '^$' -bench . -benchmem -count=5 . ./internal/jsonvalue ./cmd/jsonschema

The order schema has no unevaluated* keyword, so flag and basic checks skip
the record. Change is after against before; negative is better.

### Library (package root)

| Benchmark | ns/op before → after | MB/s before → after | B/op before → after | allocs/op before → after |
|---|---|---|---|---|
| CompileOrder | 667,268 → 682,721 (noise) | – | 696,779 → 696,654 (noise) | 14,072 → 14,072 (0) |
| ValidateSmall | 30,154 → 28,701 (−4.8%) | 16.32 → 17.14 | 51,208 → 42,641 (−16.7%) | 846 → 829 (−2.0%) |
| Validate0_5MB | 29,076,265 → 26,535,065 (−8.7%) | 17.80 → 19.47 | 50,724,601 → 36,742,878 (−27.6%) | 710,113 → 700,087 (−1.4%) |
| Validate5_5MB | 269,257,406 → 253,485,354 (−5.9%) | 20.30 → 21.57 | 524,157,710 → 377,118,288 (−28.1%) | 7,100,145 → 7,000,109 (−1.4%) |
| ValidateErrors | 22,006,129 → 19,721,135 (−10.4%) | 15.13 → 16.88 | 40,656,893 → 29,849,153 (−26.6%) | 510,078 → 500,049 (−2.0%) |
| ValidateErrorsBasic | 22,715,679 → 20,308,417 (−10.6%) | 14.66 → 16.39 | 42,642,644 → 31,835,532 (−25.3%) | 516,101 → 506,077 (−1.9%) |

### Decoded document (package internal/jsonvalue)

Not touched by the story; shown only to confirm no change.

| Benchmark | ns/op before → after | retained-B before → after | B/op before → after | allocs/op before → after |
|---|---|---|---|---|
| RetainedHeap/0.5MB | 5,944,405 → 6,046,076 (noise) | 4,340,752 → 4,340,751 | 6,282,021 → 6,282,028 | 170,198 → 170,198 |
| RetainedHeap/5.5MB | 59,948,123 → 60,161,680 (noise) | 44,924,641 → 44,924,641 | 67,382,573 → 67,382,574 | 1,715,811 → 1,715,811 |

### CLI, in process (package cmd/jsonschema)

| Benchmark | ns/op before → after | MB/s before → after | B/op before → after | allocs/op before → after |
|---|---|---|---|---|
| CLI/0.5MB | 39,534,288 → 38,480,686 (−2.7%) | 13.07 → 13.43 | 65,469,570 → 51,485,678 (−21.4%) | 984,372 → 974,347 (−1.0%) |
| CLI/5.5MB | 372,067,181 → 352,835,722 (−5.2%) | 14.69 → 15.49 | 654,932,610 → 507,894,690 (−22.5%) | 9,714,409 → 9,614,379 (−1.0%) |
| CLI/Errors | 25,324,144 → 23,913,738 (−5.6%) | 13.15 → 13.92 | 44,477,457 → 33,673,367 (−24.3%) | 530,296 → 520,272 (−1.9%) |
| CLI/Verbose0.5MB | 86,439,052 → 84,931,387 (−1.7%) | 5.98 → 6.08 | 144,733,299 → 130,751,052 (−9.7%) | 1,934,520 → 1,924,494 (−0.5%) |

Verbose is lower too: its first pass (Validate) skips the record, while its
second pass (the tree) records, as designed.

### Acceptance check

- Flag and basic validation of the order documents (library and CLI) have
  lower B/op (about 17% to 28% in the library, 21% to 24% in the CLI for
  flag/basic-style runs) and lower allocs/op (1% to 2%).
- No scenario's allocs/op or B/op rose. The decoded-document figures are
  unchanged.
- `CompileOrder`: a first version of the load-time reachability scan used a
  heap stack and cost 3 allocations (14,072 → 14,075). The scan now recurses
  over the already depth-limited schema value and allocates nothing, so
  `CompileOrder` is back at 14,072 allocs/op (runs: 14,071 to 14,072) and
  its B/op is inside run-to-run spread. The `CompileOrder` row and the root
  and CLI rows were re-measured after that fix (same command, `-count=5`,
  median); the validation rows matched the figures above to within a few
  allocations and bytes, so they are kept.
- ns/op is lower on every validation scenario (about 2% to 11%); the speed
  gain is recorded, not required. Changes within ±5% are noise.
- No allocation-limit test was added or changed.

## Story 202 (meta-schema built once per process): before

Measured before any change of this story, on the same machine as above
(Apple M3 Max, 16 logical CPUs, macOS arm64, go1.27.2), `-count=5`, median of 5.
The other scenarios keep the story 201 "after" figures above. The "after" set
is added next to this one when the change is done.

Commands:

    go test -run '^$' -bench CompileOrder -benchmem -count=5 .
    go test -run '^$' -bench FirstLoad -benchmem -count=5 ./internal/compile

| Benchmark | Package | ns/op | B/op | allocs/op |
|---|---|---|---|---|
| CompileOrder (repeated compile) | root | 689,619 | 696,549 | 14,072 |
| FirstLoadOrder (one-shot) | internal/compile | 422,503 | 339,838 | 6,019 |

`BenchmarkFirstLoadOrder` rebuilds the built-in meta-schema index and compiles
the 2020-12 meta-schema from scratch on every iteration, then compiles the
order schema against it. It does not run the meta-schema check evaluation, so
it is smaller than `CompileOrder`. Before the change it runs the build steps
`Compile` runs on every call; once the shared build function exists it calls
that function instead, so the "after" figure is comparable.

## Story 202 (meta-schema built once per process): after

Same machine, Go version and method as the "before" set (`-count=5`, median
of 5). "Before" is the story 202 "before" table above for the compile rows and
the story 200 "after" tables for the other scenarios (nothing else changed in
between). Command:

    go test -run '^$' -bench . -benchmem -count=5 . ./internal/jsonvalue ./cmd/jsonschema ./internal/compile

| Benchmark | ns/op before → after | B/op before → after | allocs/op before → after |
|---|---|---|---|
| CompileOrder (repeated compile) | 689,619 → 125,328 (−81.8%) | 696,549 → 209,407 (−69.9%) | 14,072 → 2,489 (−82.3%) |
| FirstLoadOrder (one-shot) | 422,503 → 425,224 (noise) | 339,838 → 339,988 (noise) | 6,019 → 6,020 (noise) |
| ValidateSmall | 28,701 → 29,123 (noise) | 42,641 → 42,643 | 829 → 829 |
| Validate0_5MB | 26,535,065 → 26,792,761 (noise) | 36,742,878 → 36,742,889 | 700,087 → 700,087 |
| Validate5_5MB | 253,485,354 → 254,783,750 (noise) | 377,118,288 → 377,118,316 | 7,000,109 → 7,000,109 |
| ValidateErrors | 19,721,135 → 20,477,864 (noise) | 29,849,153 → 29,849,586 | 500,049 → 500,051 |
| ValidateErrorsBasic | 20,308,417 → 21,140,034 (noise) | 31,835,532 → 31,834,696 | 506,077 → 506,076 |
| RetainedHeap/0.5MB | 6,046,076 → 5,976,054 (noise) | 6,282,028 → 6,282,026 | 170,198 → 170,198 |
| RetainedHeap/5.5MB | 60,161,680 → 61,137,884 (noise) | 67,382,574 → 67,382,576 | 1,715,811 → 1,715,811 |
| CLI/0.5MB | 38,480,686 → 39,090,659 (noise) | 51,485,678 → 51,001,696 (−0.9%) | 974,347 → 962,762 (−1.2%) |
| CLI/5.5MB | 352,835,722 → 351,065,708 | 507,894,690 → 507,413,272 | 9,614,379 → 9,602,792 (−0.1%) |
| CLI/Errors | 23,913,738 → 21,834,825 (−8.7%) | 33,673,367 → 33,187,324 (−1.4%) | 520,272 → 508,685 (−2.2%) |
| CLI/Verbose0.5MB | 84,931,387 → 82,405,022 (−3.0%) | 130,751,052 → 130,269,925 (−0.4%) | 1,924,494 → 1,912,909 (−0.6%) |

The CompileOrder "before" is the figure measured at the start of this story.

`FirstLoadOrder` is unchanged by design: it calls the build function itself on
every iteration, so it measures the one-time cost a one-shot CLI run still
pays once (building the built-in set, then compiling the order schema). The
saving shows in repeated compiles, where each Compile after the first reuses
the shared set.

### Acceptance check

- The compile benchmark is lower in all three numbers: ns/op −82%, B/op −70%,
  allocs/op −82%.
- No scenario's allocs/op or B/op rose beyond a few allocations or bytes of
  run-to-run noise (`ValidateErrors` +2 allocs of 500,000). The CLI rows are
  lower because each CLI run compiles the schema once, and the meta-schema
  check no longer collects annotations.
- No allocation-limit test was added or changed by this story.

## Story 203 (compact exact number storage): before and after

Same machine, Go version and method as above (Apple M3 Max, 16 logical CPUs,
macOS arm64, go1.27.2), `-count=5`, median of 5. Both sets were measured in
the same session, one straight after the other, with the same benchmark
files. Change is after against before; negative is better.

- **Before**: the code just before this story. `internal/jsonvalue/number.go`
  and `number_canonical.go` were set back to their content at the commit
  before the story (`big.Int` mantissa in every `Number`), with the story's
  new white-box test `number_fastpath_test.go` emptied so the package built.
  Every other file was the finished one. All tests passed on that tree,
  including the story's new per-rule, limit and race tests.
- **After**: the finished code, with the three files restored exactly
  (`git diff` shows no change to them).

Command (both sets):

    go test -run '^$' -bench . -benchmem -count=5 . ./internal/jsonvalue ./cmd/jsonschema

### Library (package root)

| Benchmark | ns/op before → after | B/op before → after | allocs/op before → after |
|---|---|---|---|
| CompileOrder | 131,433 → 124,464 (−5.3%) | 209,160 → 209,249 (noise, see note) | 2,489 → 2,482 (−0.3%) |
| ValidateSmall | 29,817 → 28,242 (−5.3%) | 42,643 → 41,682 (−2.3%) | 829 → 785 (−5.3%) |
| Validate0_5MB | 27,195,606 → 25,649,156 (−5.7%) | 36,742,738 → 35,783,591 (−2.6%) | 700,087 → 660,084 (−5.7%) |
| Validate5_5MB | 260,193,136 → 253,744,833 (−2.5%) | 377,118,280 → 368,958,400 (−2.2%) | 7,000,108 → 6,600,106 (−5.7%) |
| ValidateErrors | 19,973,389 → 19,456,277 (−2.6%) | 29,849,143 → 29,283,397 (−1.9%) | 500,049 → 480,050 (−4.0%) |
| ValidateErrorsBasic | 20,702,509 → 20,304,891 (−1.9%) | 31,835,034 → 31,270,190 (−1.8%) | 506,076 → 486,079 (−4.0%) |

Note on CompileOrder B/op: +89 bytes of 209,000 (0.04%). Its five before-runs
ranged 208,703 to 209,495 and its after-runs 208,582 to 209,382, so this is
run-to-run spread; allocs/op fell by 7 (the schema's own numbers no longer
allocate a `big.Int` word).

### Decoded document (package internal/jsonvalue)

`RetainedHeap` decodes with `Decode`, which keeps `json.Number` leaves and
never builds a `*Number`, so it cannot show this story's change; it is listed
only to confirm nothing else moved. `RetainedHeapExact` (new, in
`internal/jsonvalue/retained_exact_bench_test.go`) measures the same thing on
the validator's own decode path, `DecodeLimited`, which builds `*Number`.

| Benchmark | ns/op before → after | retained-B before → after | B/op before → after | allocs/op before → after |
|---|---|---|---|---|
| RetainedHeapExact/0.5MB | 8,215,172 → 6,780,282 (−17.5%) | 5,940,707 → 5,300,807 (−10.8%) | 8,524,450 → 7,243,177 (−15.0%) | 220,254 → 170,223 (−22.7%) |
| RetainedHeapExact/5.5MB | 83,973,500 → 68,860,998 (−18.0%) | 59,484,716 → 54,524,704 (−8.3%) | 88,931,526 → 77,171,933 (−13.2%) | 2,224,999 → 1,718,765 (−22.8%) |
| RetainedHeap/0.5MB | 5,954,634 → 5,938,769 (noise) | 4,340,754 → 4,340,754 (no change) | 6,282,024 → 6,282,193 (noise) | 170,198 → 170,201 (noise) |
| RetainedHeap/5.5MB | 62,157,023 → 62,326,363 (noise) | 44,924,643 → 44,924,641 (no change) | 67,382,582 → 67,382,586 (noise) | 1,715,812 → 1,715,811 |

On the `*Number` path a fast-path number is now one allocation (the
`Number`, 64-byte size class instead of 80); before, `big.Int.SetString` also
allocated the word slice and its own scratch. The 0.5 MB document holds
20,001 numbers and lost 50,031 allocations and 639,900 retained bytes, about
2.5 allocations and 32 bytes per number.

Where the validation saving comes from: decoding, and fast-path `Cmp` and
equality, which no longer copy mantissas into `big.Int`. `multipleOf` is not
changed by this story: `multipleof.go` still calls `Decimal()`, which builds
a fresh `big.Int` on every evaluation, fast path included.

### CLI, in process (package cmd/jsonschema)

| Benchmark | ns/op before → after | B/op before → after | allocs/op before → after |
|---|---|---|---|
| CLI/0.5MB | 37,394,113 → 36,830,701 (−1.5%) | 51,001,946 → 50,043,284 (−1.9%) | 962,763 → 922,755 (−4.2%) |
| CLI/5.5MB | 355,032,097 → 355,866,111 (+0.2%, noise) | 507,412,722 → 499,252,978 (−1.6%) | 9,602,795 → 9,202,786 (−4.2%) |
| CLI/Errors | 22,801,719 → 23,303,234 (+2.2%, noise) | 33,187,774 → 32,622,777 (−1.7%) | 508,686 → 488,678 (−3.9%) |
| CLI/Verbose0.5MB | 83,873,962 → 85,369,327 (+1.8%, noise) | 130,269,298 → 129,630,004 (−0.5%) | 1,912,908 → 1,882,895 (−1.6%) |

### CLI peak memory (information only)

Peak resident memory (getrusage `maxrss`) of one CLI run per document, each
in a fresh process, measured with the same before and after trees as above,
5 runs each, median. "floor" is the same process on a 22-byte document, so
the start-up cost can be read off. Command:

    go test -count=5 -run 'TestCLIPeakMemory$' -v ./cmd/jsonschema -args -peakmem

The test (`cmd/jsonschema/peakmem_test.go`) writes the generated documents to
a temporary directory and re-runs its own test binary as the child, which
calls `run` once; it asserts nothing and is skipped without `-peakmem`.

| Document | before | after | change |
|---|---|---|---|
| floor (22 B) | 8.63 MB | 8.81 MB | noise |
| 0.5 MB valid | 33.88 MB | 32.82 MB | −3.1% |
| 5.5 MB valid | 223.28 MB | 209.21 MB | −6.3% |
| 0.33 MB, 1,000 errors | 26.43 MB | 26.00 MB | −1.6% |

Runs varied by about ±3% (5.5 MB before: 214.7 to 223.9 MB; after: 205.1 to
213.6 MB), so only the 5.5 MB drop is clearly outside the spread.

### Acceptance check

- allocs/op is lower on every document scenario (library 4.0% to 5.7%, CLI
  1.6% to 4.2%, decoded `*Number` document 22.7% to 22.8%); B/op is lower on
  every document scenario (0.5% to 15.0%).
- Retained heap on the `*Number` decode path is lower: −10.8% (0.5 MB) and
  −8.3% (5.5 MB). The `json.Number` path is unchanged, as expected.
- No scenario's allocs/op or B/op rose beyond run-to-run noise (CompileOrder
  B/op +89 bytes, inside its spread; its allocs/op fell).
- No scenario is more than 10% slower (largest: CLI/Errors +2.2%).
- CLI peak memory is recorded above for information.
- No allocation-limit test was added or changed.

## Story 205 (path frames reused): before and after

Same machine, Go version and method as above (Apple M3 Max, 16 logical CPUs,
macOS arm64, go1.27.2). Both sets were measured in one session, one straight
after the other, `-count=10`, one package at a time, median of 10 (mean of the
5th and 6th sorted; a median ending in .5 is rounded up). Change is after
against before; negative is better.

- **Before**: the evaluation code at the story's base d7198b7.
  `internal/eval/eval.go` and `internal/eval/refguard.go` were set back to
  their base logic (a heap `pathFrame`, a leave closure and a heap `lazyLoc`
  per node; only comments differed from d7198b7) and `pathstack.go` was
  emptied. `dynamicscope.go` kept its compatibility `push`, which allocates
  exactly as the base `push` did. Every benchmark file was the finished one.
- **After**: the finished code; `git diff HEAD` showed no change to the
  production files once they were restored.

Commands (each set):

    go test -run '^$' -bench . -benchmem -count=10 .
    go test -run '^$' -bench . -benchmem -count=10 ./cmd/jsonschema

### Library (package root)

| Benchmark | ns/op before → after | B/op before → after | allocs/op before → after |
|---|---|---|---|
| CompileOrder | 121,821 → 107,082 (−12.1%) | 209,509 → 159,776 (−23.7%) | 2,482 → 1,490 (−40.0%) |
| ValidateSmall | 27,040 → 24,361 (−9.9%) | 41,682 → 30,712 (−26.3%) | 785 → 558 (−28.9%) |
| Validate0_5MB | 24,076,961 → 21,539,069 (−10.5%) | 35,783,469 → 25,779,348 (−28.0%) | 660,085 → 460,052 (−30.3%) |
| Validate5_5MB | 244,068,325 → 215,975,330 (−11.5%) | 368,958,349 → 268,954,649 (−27.1%) | 6,600,105 → 4,600,073 (−30.3%) |
| ValidateErrors | 18,305,033 → 16,628,118 (−9.2%) | 29,283,472 → 21,276,804 (−27.3%) | 480,051 → 320,029 (−33.3%) |
| ValidateErrorsBasic | 19,169,953 → 17,227,037 (−10.1%) | 31,270,044 → 23,263,360 (−25.6%) | 486,078 → 326,057 (−32.9%) |

### Decoded document (package internal/jsonvalue)

`internal/jsonvalue` does not import `internal/eval`, so its code is the same
in both trees; it was measured once in the same session (`-count=10`, median):

| Benchmark | ns/op | retained-B | B/op | allocs/op |
|---|---|---|---|---|
| RetainedHeap/0.5MB | 5,555,501 | 4,340,754 | 6,281,435 | 170,185 |
| RetainedHeap/5.5MB | 60,428,592 | 44,924,648 | 67,382,572 | 1,715,811 |
| RetainedHeapExact/0.5MB | 6,104,132 | 5,300,792 | 7,242,459 | 170,206 |
| RetainedHeapExact/5.5MB | 66,414,690 | 54,524,696 | 77,105,323 | 1,717,726 |

These match the story 203 "after" figures within run-to-run spread.

### CLI, in process (package cmd/jsonschema)

| Benchmark | ns/op before → after | B/op before → after | allocs/op before → after |
|---|---|---|---|
| CLI/0.5MB | 34,846,863 → 33,022,870 (−5.2%) | 50,043,681 → 39,990,447 (−20.1%) | 922,754 → 721,731 (−21.8%) |
| CLI/5.5MB | 339,660,785 → 310,899,568 (−8.5%) | 499,252,608 → 399,200,032 (−20.0%) | 9,202,785 → 7,201,762 (−21.7%) |
| CLI/Errors | 21,097,672 → 18,578,935 (−11.9%) | 32,622,852 → 24,568,684 (−24.7%) | 488,679 → 327,666 (−32.9%) |
| CLI/Verbose0.5MB | 79,214,468 → 73,675,186 (−7.0%) | 129,630,119 → 109,574,878 (−15.5%) | 1,882,894 → 1,481,841 (−21.3%) |

### Slowdown judgement (reported, not asserted)

No scenario is slower: every ns/op median is lower (−5.2% to −12.1%), so none
is more than 5% slower.

### Targets

| Target | Measured | Status |
|---|---|---|
| Validate5_5MB B/op at most 280 MB | 268,954,649 B (268.95 MB) | met |
| Validate5_5MB allocs/op at most 4.8 M | 4,600,073 | met |
| ValidateErrors B/op at most 22.5 MB | 21,276,804 B (21.28 MB) | met |
| CompileOrder allocs/op at most 1,600 | 1,490 | met |
| No scenario more than 5% slower | all faster | met |

All targets are met on this machine; nothing is held for the operator.

### Golden files

`testdata/golden/cli_v1_1.json` holds 2,512 entries and
`cli_v1_1_large.json` holds 12, so 2,524 entries in total: 631 base cases in
each of the 4 output shapes. The golden files were not touched by this task.
`TestGolden` (`go test -count=1 -v -run 'TestGolden$' ./cmd/jsonschema`, no
`-race`) took 8.1 s on this machine, including the CLI build. It fails at once
when a recorded case's args or stdin size and SHA-256 differ from the current
case list or the generated stdin.

## Story 208 (annotations in growing chunks, flattened under sync.Once)

Machine and Go as above (Apple M3 Max, 16 logical CPUs, macOS arm64,
go1.27.2), root package. Both sets were measured in one session, one straight
after the other, with the same benchmark files:

    go test -run NONE -bench . -benchmem -count=10 .

Median of 10 (mean of the 5th and 6th sorted; a median ending in .5 is
rounded up). Change is after against before; negative is better.

- **Before**: the code just before this story (base 1525f40, which holds
  stories 205 to 207, 210 and 211). The production files of this story
  (`internal/keywords/annotate.go`, `internal/keywords/registry.go`,
  `internal/eval/annotations.go`, `jsonschema.go`) were set back to their
  base logic: one growing annotation slice, flattened as is at Validate and
  sorted in the old `eval.Lazy`. Only comments and one unused accessor
  differed from 1525f40; the new files (`annotation_chunks.go`) were present
  but unused.
- **After**: the finished code, with those files restored. Annotations are
  kept in growing chunks (16 … 4,096) and handed to the Result unflattened;
  the first `Annotations()` or `Output()` call flattens, sorts and releases
  the chunks inside one `sync.Once` (`eval.Pending`).

| Benchmark | ns/op before → after | B/op before → after | allocs/op before → after |
|---|---|---|---|
| CompileOrder | 101,424 → 101,389 (noise) | 127,826 → 127,878 (noise) | 1,203 → 1,203 |
| ValidateSmall | 20,795 → 20,502 (−1.4%) | 19,295 → 17,942 (−7.0%) | 395 → 392 (−0.8%) |
| Validate0_5MB | 17,380,310 → 17,316,667 (−0.4%) | 14,735,945 → 12,243,601 (−16.9%) | 300,044 → 300,037 |
| Validate5_5MB | 179,733,098 → 173,864,549 (−3.3%) | 159,030,455 → 124,672,670 (−21.6%) | 3,000,063 → 3,000,071 (+8) |
| ValidateErrors | 14,138,448 → 13,829,770 (−2.2%) | 14,103,979 → 11,610,884 (−17.7%) | 241,023 → 241,013 |
| ValidateErrorsBasic | 15,054,734 → 14,690,494 (−2.4%) | 16,089,749 → 13,597,460 (−15.5%) | 247,050 → 247,040 |

Validate5_5MB needs 8 more allocations of 3 million: past the growing chunks
it allocates one chunk per 4,096 annotations (plus the `Pending` holder),
where the old slice regrew by doubling; the bytes fall by 34 MB. The
benchmarks never call `Annotations()` or `Output()`, so no flat copy is
made. No allocation limit has to rise for it (see below).

### Computation against v1.1 (f35d9e4, round 2 final, operator's values)

The targets are relative to the code just before this story: the required
drop is the v1.1 value minus the target (operator's note).

| Target | v1.1 | Required | Before | After | Change | Verdict |
|---|---|---|---|---|---|---|
| Validate5_5MB B/op | 368,958,349 | drop at least 26,958,349 (368,958,349 − 342,000,000) | 159,030,455 | 124,672,670 | drop 34,357,785 | met |
| ValidateErrors B/op | 29,282,834 | drop at least 1,882,834 (29,282,834 − 27,400,000) | 14,103,979 | 11,610,884 | drop 2,493,095 | met |
| ValidateSmall B/op | 41,682 | rise at most 318 (42,000 − 41,682) | 19,295 | 17,942 | drop 1,353 | met |
| ValidateSmall allocs/op | 785 | must not rise | 395 | 392 | drop 3 | met |

No scenario is slower by median. All targets are met on this machine.

### Allocation limits lowered (allocs_test.go)

Measured on the finished code with `go test -count=1 -v -run TestAllocs .`;
limit = measured + about 10%, rounded up. "Old" is the story 204 limit. No
limit rose. Where measured + 10% rounds to the old limit or just above it,
the old limit is kept (it is already within about 10% of the measured value).

| Scenario | allocs: old limit → new (measured) | bytes: old limit → new (measured) |
|---|---|---|
| Compile order | 1,640 → 1,325 (1,203) | – |
| Validate small | 615 → 432 (392) | – |
| Validate 0.5 MB | 507,000 → 331,000 (300,035) | 28,400,000 → 13,468,000 (12,243,154) |
| Validate 5.5 MB | 5,061,000 → 3,301,000 (3,000,069) | 295,900,000 → 137,141,000 (124,673,165) |
| Validate errors | 353,000 → 266,000 (240,999) | 23,500,000 → 12,770,000 (11,611,584; +10% would be 12,772,743) |
| Validate errors + basic | 359,000 → 272,000 (247,028) | – |

The `allocs_<topic>_test.go` files and `docs/benchmarks/r3-*.md` were not
touched.

## Story 209 (report rendered without copies): before

Measured before any change of this story, on the code after stories 205 to 208
(base bb263ff). Machine: Apple M3 Max, 16 logical CPUs, macOS (Darwin 25.6.0)
arm64, go1.27.2. Median of 10 (mean of the 5th and 6th sorted; a median ending
in .5 is rounded up). Commands, one package at a time:

    go test -run '^$' -bench . -benchmem -count=10 .
    go test -run '^$' -bench . -benchmem -count=10 ./cmd/jsonschema

Timing caveat: a second `-count=10` run of both packages was still going in the
background while the root package ran, so ns/op of the root rows (and possibly
the start of the CLI rows) may be high or noisy. B/op and allocs/op are not
affected. The "after" run must be made on an idle machine.

### Library (package root)

| Benchmark | ns/op | B/op | allocs/op |
|---|---|---|---|
| CompileOrder | 105,856 | 127,832 | 1,203 |
| ValidateSmall | 21,281 | 17,942 | 392 |
| Validate0_5MB | 17,953,528 | 12,243,428 | 300,036 |
| Validate5_5MB | 176,738,646 | 124,672,629 | 3,000,071 |
| ValidateErrors | 14,536,018 | 11,610,886 | 241,013 |
| ValidateErrorsBasic | 15,507,701 | 13,597,467 | 247,040 |

### CLI, in process (package cmd/jsonschema)

| Benchmark | ns/op | B/op | allocs/op |
|---|---|---|---|
| CLI/0.5MB | 29,423,930 | 27,307,104 | 561,428 |
| CLI/5.5MB | 273,590,120 | 263,691,661 | 5,601,469 |
| CLI/Errors | 23,695,346 | 14,870,477 | 248,362 |
| CLI/Verbose0.5MB | 65,119,383 | 84,481,021 | 1,181,534 |

### Peak RSS, 5.5 MB document, default output (basic)

`TestCLIPeakMemory` (`go test -count=5 -run 'TestCLIPeakMemory$' -v
./cmd/jsonschema -args -peakmem`), maxrss of one fresh child process per run,
in bytes. The child calls `run` with the default shape (basic).

| Run | 5.5 MB maxrss (bytes) |
|---|---|
| 1 | 222,314,496 |
| 2 | 199,753,728 |
| 3 | 222,380,032 |
| 4 | 222,330,880 |
| 5 | 222,150,656 |
| Median | 222,314,496 (212.02 MiB) |

An earlier single `-count=1` run gave 221,904,896 (211.63 MiB). Floor (22 B)
8.37 to 8.60 MB, 0.5 MB 31.4 to 33.9 MB, errors 21.8 to 22.7 MB in the same runs.

### Peak RSS with `/usr/bin/time -l` on the built CLI, 5.5 MB, `-o basic`

The opt-in `TestCLIPeakMemoryTimeL` (`cmd/jsonschema/peakmem_time_test.go`,
macOS only, skipped without `-peakmem`) builds the CLI with `go build` into a
temporary directory, writes the order schema and the generated 5.5 MB document
(5,466,732 bytes) there, and runs
`/usr/bin/time -l <cli> validate -o basic schema.json 5.5MB.json` five times,
logging "maximum resident set size" (bytes). It asserts nothing. Command:

    go test -count=1 -run TestCLIPeakMemoryTimeL -v ./cmd/jsonschema -args -peakmem

| Run | maximum resident set size (bytes) |
|---|---|
| 1 | 220,217,344 |
| 2 | 220,397,568 |
| 3 | 220,512,256 |
| 4 | 220,102,656 |
| 5 | 220,119,040 |
| Median | 220,217,344 (210.02 MiB) |

The RSS target (at least 30 MiB = 31,457,280 bytes lower) therefore needs an
"after" `/usr/bin/time -l` median of at most 188,760,064 bytes
(220,217,344 − 31,457,280), measured with the same command on the same machine.

### Early reading against the targets (not a verdict)

The "before" CLI/5.5MB is already 263,691,661 B/op and 5,601,469 allocs/op,
below the absolute targets (450,000,000 and 7,300,000). The required drops
(499,252,608 − 450,000,000 = 49,252,608 B/op; 9,202,785 − 7,300,000 =
1,902,785 allocs/op) are measured against this before value, so this story
alone would need to reach at most 214,439,053 B/op and 3,698,684 allocs/op
(263,691,661 − 49,252,608; 5,601,469 − 1,902,785). That is a drop of 18.7% and
34.0%; whether rendering changes can do this is decided by the "after" run.

## Allocation and memory limits (story 204)

The limits in the table below are the story 204 values; story 208 lowered
them (see the table above and the constants in `allocs_test.go`).

Ordinary tests in `allocs_test.go` (root package) fail when a scenario uses
more allocations, or more total bytes, than its limit. They run with a plain
`go test ./...`, including the 5.5 MB document; there is no `-short` or
environment skip. Under `-race` they skip, because the race detector adds
allocations the validator did not cause (`allocs_race_on_test.go` and
`allocs_race_off_test.go` set `raceEnabled`).

Measured on: Apple M3 Max, 16 logical CPUs, macOS (Darwin 25.6.0) arm64,
go1.27.2 darwin/arm64.

Method:

- Every scenario starts from raw bytes: the schema from
  `testdata/perf/order_schema.json`, the small document from
  `testdata/perf/order_small.json`, the large ones from `internal/perfdata`.
- Compile: `Compile(schemaBytes)` (includes the built-in meta-schema). One
  warm-up call, then `testing.AllocsPerRun` with 20 runs.
- Validate: `Schema.Validate(docBytes)` (decoding included), plus
  `Result.Output("basic")` for the last scenario. One warm-up call, then
  `testing.AllocsPerRun` with 50 runs (small), 3 runs (0.5 MB, errors, errors
  plus basic) or 2 runs (5.5 MB).
- Bytes: the `runtime.MemStats.TotalAlloc` delta (every byte requested, even
  if freed at once) around 3 runs after a warm-up, divided by 3.
- Limit = measured value plus about 10%, rounded up.

| Scenario | Measured allocs | Allocs limit | Measured bytes | Bytes limit |
|---|---|---|---|---|
| Compile order schema | 1,490 | 1,640 | – | – |
| Validate small | 558 | 615 | – | – |
| Validate 0.5 MB | 460,045 | 507,000 | 25,777,949 | 28,400,000 |
| Validate 5.5 MB | 4,600,065 | 5,061,000 | 268,954,661 | 295,900,000 |
| Validate errors (1,000) | 320,005 | 353,000 | 21,277,448 | 23,500,000 |
| Validate errors + basic report | 326,032 | 359,000 | – | – |

Re-measured and lowered on purpose by story 205 (path frames reused); the
previous limits were 2,730 / 864 / 726,000 / 7,260,000 / 528,000 / 534,700
allocs and 39.4 M / 405.9 M / 32.2 M bytes.

Policy: one limit applies on every platform. If a limit is exceeded on a newer
Go or another OS, the limit is re-measured on purpose and the reason recorded
here and next to the constant; it is not raised silently.

## Operator baseline (the operator's Mac, from the brief)

These are the operator's own measurements, copied from the brief. They are a
different machine but the same Go version (Go 1.27.2, the version of the build
machine), so do not compare the timings directly with the table above. Hardware: one Mac, 16 cores, Go 1.27.2. Timings vary
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
