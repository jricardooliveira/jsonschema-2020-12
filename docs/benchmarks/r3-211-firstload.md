# Round 3, story 211: build the built-in meta-schema index once

Machine: Apple M3 Max, darwin/arm64, Go toolchain of the repo. Timings were taken
while other stories may build in parallel: **noisy**, the operator re-measures
interleaved at Checkpoint 3. B/op and allocs/op are exact.

"Before" = same tree with `internal/compile/metacache.go` temporarily calling
`registry.BuildBuiltinIndex()` (second build, as on main after #205). "After" =
`registry.SharedBuiltinIndex()` (this story). The temporary edit was reverted.

## First compile of the order schema (fresh process, median of 10)

Run: `go test -run TestFirstLoadTiming -firstload -v .`

| | time | bytes | allocs |
|---|---|---|---|
| before | 997,833 ns | 741,768 B | 10,649 |
| after | 862,250 ns | 625,176 B | 8,652 |
| change | -13.6% | -15.7% | -1,997 |

## Benchmarks (go test -bench Order -benchmem -count=10, medians approx.)

| benchmark | before | after |
|---|---|---|
| internal/compile FirstLoadOrder | ~445,000 ns, 340.2 KB, 6,020 allocs | ~277,700 ns, 222.3 KB, 4,025 allocs |
| CompileOrder (root) | not re-measured before (path unchanged: cached meta-schema) | ~110,000 ns, 159.7 KB, 1,490 allocs |

Note: FirstLoadOrder calls `buildBuiltin()`, which now takes the shared index
instead of rebuilding it, so its "after" no longer includes an index build. The
benchmark file is unchanged; its numbers drop for that reason.

## Verdict

First-compile median is lower in time, bytes and allocations: success criterion met.
Time numbers are provisional (noisy). `go test ./...` and `go test -race ./...`
are green, including the golden CLI test. Fresh-process tests: one build
(`firstload_count_test.go`), first-load allocation limit 9,520 against measured 8,653
(`allocs_firstload_test.go`).
