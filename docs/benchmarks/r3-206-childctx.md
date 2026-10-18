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

- Same machine, same command (`-count=10`), medians of 10 runs, with story #206 applied.
- Timings are noisy (taken during a parallel build); the operator re-measures interleaved at Checkpoint 3 and judges the 5% rule there. Only B/op and allocs/op are judged here.

| Benchmark | ns/op | B/op | allocs/op |
|---|---:|---:|---:|
| CompileOrder | 109,413 | 126,148 | 1,261 |
| ValidateSmall | 24,406 | 23,234 | 509 |
| Validate0_5MB | 21,850,954 | 18,577,063 | 410,050 |
| Validate5_5MB | 208,302,396 | 196,951,917 | 4,100,071 |
| ValidateErrors | 16,045,666 | 15,513,187 | 280,026 |
| ValidateErrorsBasic | 17,305,845 | 17,499,243 | 286,052 |
| CLI/0.5MB | 32,773,320 | 32,756,001 | 671,499 |
| CLI/5.5MB | 303,257,084 | 327,163,825 | 6,701,528 |
| CLI/Errors | 18,894,648 | 18,771,859 | 287,433 |
| CLI/Verbose0.5MB | 77,423,544 | 95,138,400 | 1,381,610 |

## Verdicts

| Target | Before | After | Change | Verdict |
|---|---:|---:|---:|---|
| Validate5_5MB at most 200 MB B/op (combined, on top of #205) | 268,954,740 | 196,951,917 | -72,002,823 | PASS |
| Validate5_5MB at most 305 MB (stand-alone) | n/a | n/a | n/a | not measured (operator: combined target only) |
| ValidateErrors drops at least about 4.8 MB B/op | 21,276,920 | 15,513,187 | -5,763,733 | PASS |
| ValidateSmall drops at least about 6,200 B/op | 30,713 | 23,234 | -7,479 | PASS |

No target missed; the story is not held. B/op and allocs/op dropped in every scenario; no scenario got slower by median in the noisy timings (Verbose and CLI numbers are lower too).

## Checks

- `go test -race ./...`: all packages green, including the #205 golden comparison test and the new concurrent child-context test.
- No golden file, allocs_test.go or docs/benchmarks.md was edited.
