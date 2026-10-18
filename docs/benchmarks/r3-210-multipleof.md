# Round 3, story 210: multipleOf without big integers on the common path

Machine: Apple M3 Max, Go as installed on the build machine, `go test -run XXX -bench Validate -benchmem -count=10 .`
(root package). Medians of 10 runs.

- **Before** = the code at base commit 9ef52f6 (main after #205). Measured on this tree with the small path in
  `evalMultipleOf` switched off for the run (`&& false` on the small-path condition, since reverted), so the
  big-integer path decides every case, as before this story.
- **After** = this story (small path active).

## Medians

| Benchmark | Before ns/op | After ns/op | Before B/op | After B/op | Before allocs/op | After allocs/op |
|---|---|---|---|---|---|---|
| ValidateSmall | 23,200 | 22,770 | 23,234 | 22,673 | 509 | 469 |
| Validate0_5MB | 20.13 ms | 19.35 ms | 18,577,200 | 17,937,450 | 410,050 | 370,049 |
| Validate5_5MB | 208.6 ms | 196.75 ms | 196,951,950 | 191,031,950 | 4,100,071 | 3,700,070 |
| ValidateErrors | 16.00 ms | 16.12 ms | 15,513,030 | 15,513,000 | 280,026 | 280,026 |
| ValidateErrorsBasic | 15.96 ms | 16.65 ms | 17,499,250 | 17,499,250 | 286,053 | 286,052 |

The error scenarios have no multipleOf evaluations that reach the small path, so their counts are unchanged.

## Timing caveat

Timings were taken during parallel builds (and the "after" run overlapped a `go test -race` run), so they are noisy.
The operator re-measures interleaved at Checkpoint 3 and judges the 5% rule on those numbers. B/op and allocs/op are
exact. The largest apparent slowdown is ValidateErrorsBasic (+4.3%), within the stated ±5% run-to-run noise, with
identical allocations.

## Verdicts

- **AllocsPerRun of evalMultipleOf on small numbers = 0: MET.** `TestMultipleOfSmallZeroAllocs`
  (internal/keywords/validation) reports 0 for passing instances and 0 for the failing yes/no decision; gated under
  `-race` like the root allocation tests.
- **Validate5_5MB allocs/op drop (operator decision, recorded as given):**
  - v1.1 source of the number: release f35d9e4 (round 2 final), `Validate5_5MB` = 6,600,096 allocs/op
    (allocs_test.go history "measured: 6600096" and the round-2 final table in docs/benchmarks.md).
  - Required drop = 6,600,096 - 6,250,000 = **350,096** allocs/op.
  - Measured drop caused by this story = 4,100,070 - 3,700,070 = **400,000** allocs/op (median-to-median, same machine).
  - 400,000 >= 350,096, so the criterion is **MET**. The absolute count after this story (3,700,070) is also below
    6.25 M.
- **Behaviour:** `go test` and `go test -race` for the root package, cmd/jsonschema (includes the golden
  comparison against the committed v1.1 files), internal/keywords/validation and internal/jsonvalue pass.
  No golden file was changed.
- **Root allocation limits:** allocs_multipleof_test.go holds the lowered limits (measured plus about 10%);
  allocs_test.go is unchanged.
