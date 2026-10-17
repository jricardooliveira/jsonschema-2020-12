# Round 3, story 2 (#206): one child evaluation context per depth

## Before (main after #205)

- Commit: `079149b` (worktree base, with #205 integrated)
- Machine: Apple M3 Max (16 cores), darwin/arm64, Go 1.27.2
- Command: `go test -run '^$' -bench . -benchmem -count=10 . ./cmd/jsonschema`
- Values are medians of 10 runs (B/op and allocs/op are stable to within 0.01%).
- Timings are noisy: taken during a parallel build, to be re-measured at Checkpoint 3.

| Benchmark | ns/op | B/op | allocs/op |
|---|---:|---:|---:|
| CompileOrder | 112,867 | 159,752 | 1,490 |
| ValidateSmall | 25,220 | 30,713 | 558 |
| Validate0_5MB | 22,557,882 | 25,779,300 | 460,052 |
| Validate5_5MB | 219,008,830 | 268,954,740 | 4,600,074 |
| ValidateErrors | 16,491,452 | 21,276,920 | 320,030 |
| ValidateErrorsBasic | 18,494,238 | 23,263,500 | 326,056 |
| CLI/0.5MB | 34,924,714 | 39,990,350 | 721,731 |
| CLI/5.5MB | 316,631,120 | 399,199,900 | 7,201,760 |
| CLI/Errors | 21,069,442 | 24,568,600 | 327,663 |
| CLI/Verbose0.5MB | 73,835,954 | 109,574,500 | 1,481,841 |

## Targets (judged on the combined code, on top of #205)

- Validate5_5MB: at most 200 MB B/op (305 MB stand-alone target: not measured).
- ValidateErrors: drop of at least about 4.8 MB B/op.
- ValidateSmall: drop of at least about 6,200 B/op.

## After

To be recorded by task T6.
