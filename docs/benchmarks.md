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
