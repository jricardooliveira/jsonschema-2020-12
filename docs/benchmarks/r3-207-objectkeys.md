# Story 207: object keywords without wasted copies

Machine: Apple M3 Max (arm64, 16 cores). Command:
`go test -run NONE -bench Validate -benchmem -count=10 .`

Timings are medians of 10 runs, read by hand from the sorted runs. They were
taken while other stories may build in parallel, so they are noisy. The operator
re-measures interleaved at Checkpoint 3 and judges the 5% rule there. Speed is
never asserted in a test. B/op and allocs/op are stable across runs.

## Before (main after #205, no production change from this story)

| Benchmark | ns/op (median) | B/op | allocs/op |
|---|---|---|---|
| ValidateSmall | 24,370 | 23,234 | 509 |
| Validate0_5MB | 21,500,000 | 18,577,000 | 410,050 |
| Validate5_5MB | 205,100,000 | 196,951,400 | 4,100,070 |
| ValidateErrors | 15,520,000 | 15,513,000 | 280,026 |
| ValidateErrorsBasic | 16,500,000 | 17,499,200 | 286,053 |

## After (with this story)

Same machine and command, medians of 10 runs.

| Benchmark | ns/op (median) | B/op | allocs/op |
|---|---|---|---|
| ValidateSmall | 22,637 | 19,855 | 435 |
| Validate0_5MB | 19,473,000 | 15,375,190 | 340,043 |
| Validate5_5MB | 189,729,000 | 164,949,700 | 3,400,064 |
| ValidateErrors | 14,890,000 | 14,103,800 | 241,023 |
| ValidateErrorsBasic | 15,330,000 | 16,090,100 | 247,050 |

## Target verdict (judged by amount saved)

The operator decided: the criterion is MET. "v1.1" means release f35d9e4
(round 2 final), whose Validate5_5MB is 368,958,349 B/op and 6,600,096
allocs/op (round-2 final table in docs/benchmarks.md; allocs_test.go before
round 3: "measured: 6600096", "measured: 368958280").

| Quantity | Required saving | Achieved (vs Before) | Verdict |
|---|---|---|---|
| B/op | 368,958,349 - 345,000,000 = 23,958,349 | 196,951,400 - 164,949,700 = 32,001,700 | met |
| allocs/op | 6,600,096 - 6,000,000 = 600,096 | 4,100,070 - 3,400,064 = 700,006 | met |

Absolute values (164.95 MB, 3.40 M allocs/op) are also under 345 MB and 6.0 M.
Final verdict: PASS.

## Speed check (5% rule)

Medians of 10 runs, before vs after:

| Benchmark | Before ns/op | After ns/op | Change |
|---|---|---|---|
| ValidateSmall | 24,370 | 22,637 | -7.1% |
| Validate0_5MB | 21,500,000 | 19,473,000 | -9.4% |
| Validate5_5MB | 205,100,000 | 189,729,000 | -7.5% |
| ValidateErrors | 15,520,000 | 14,890,000 | -4.1% |
| ValidateErrorsBasic | 16,500,000 | 15,330,000 | -7.1% |

No scenario is more than 5% slower; none is flagged. These timings were taken
while other stories may build in parallel, so they are noisy. The operator
re-measures interleaved at Checkpoint 3 and judges the 5% rule there. Speed is
not asserted in any test.
