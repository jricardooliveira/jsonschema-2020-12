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

To be filled in by task T6.

## Target verdict (judged by amount saved)

To be filled in by task T6. The required saving is the v1.1 Validate5_5MB B/op
minus 345 MB and the v1.1 allocs/op minus 6.0 M, compared with the code just
before this story (the "Before" table above).

## Speed check (5% rule)

To be filled in by task T6.
