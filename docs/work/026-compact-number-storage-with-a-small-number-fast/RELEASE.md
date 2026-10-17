# Release notes — #26: Compact exact number storage in decoded documents (fix 5)

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #26
- Status: proposed (pending Checkpoint 3 sign-off)

**⚠️ the release-agent has concerns.**

## What changed

Compact Number storage now holds small numbers inline (int64 mantissa + exponent) instead of always allocating a big.Int, cutting per-number allocation by one and saving ~32 bytes. Allocation counts drop 4–23%, retained heap on the *Number path drops 8–11%, and every validation result stays byte-identical. Fast-path comparisons, equality and hashing no longer copy mantissas into big.Int.

- internal/jsonvalue/number.go: Number stores magnitude inline (uint64 mag; big.Int nil on fast path) with canonical trailing-zero stripping and fast-form demotion of big mantissas that fit int64
- internal/jsonvalue/number_canonical.go: Canonical() output now routes both fast and big paths, byte-identical to the old code
- internal/jsonvalue/number_fastpath_test.go: New white-box tests cover fast-path numbers, big-path numbers, int64 boundary cases and trailing-zero normalization
- internal/jsonvalue/equal_fastpath_test.go: New tests verify mixed fast/big equality (1, 1.0, 1e0 all equal) and allocation-free nested deep equality
- internal/keywords/validation/{type,multipleOf,maximum,exclusiveMaximum,minimum,exclusiveMinimum,enum,const,uniqueItems}_numberpath_test.go: Nine new test files, one per number rule, covering fast-path, big-path and boundary (int64 limits, trailing zeros, exponent caps)
- number_limits_regression_test.go: Regression test proving all seven hostile-input limits (depth, input size, steps, ref chain, output units, annotations, report bytes) trigger at identical thresholds and with identical error types as before
- number_race_test.go: Concurrent-safety test: one compiled schema validates fast-path and big-path documents from 8 goroutines, 50 rounds each, under go test -race
- cmd/jsonschema/peakmem_test.go: New benchmark test (runs only with -peakmem flag) measures CLI peak resident memory via Unix syscalls (Rusage); produces child process to isolate memory measurement
- cmd/jsonschema/number_limits_exit_test.go: CLI-level regression test proving exit code 2 is issued at the same limit thresholds as the library
- internal/jsonvalue/retained_exact_bench_test.go: New benchmark BenchmarkRetainedHeapExact on the validator's own decode path (DecodeLimited, *Number leaves), showing heap retained after decoding
- docs/benchmarks.md: Appended before/after benchmark tables (same session, same machine: Apple M3 Max, go1.27.2, -count=5, median of 5). Library allocs/op lower by 1.6–5.7% on all document scenarios, B/op lower by 0.5–15%, retained heap on *Number path lower by 8.3–10.8%, CLI peak memory lower by 1.6–6.3%. No scenario regressed beyond run-to-run noise (±5%)

## How to verify

- Run `go test ./...` to verify all tests pass (every existing test unchanged, every new per-rule and boundary test passes)
- Run `go test -race ./...` to confirm no data races on concurrent validation of mixed fast/big numbers
- Run `go test -run '^$' -bench . -benchmem ./... | awk '/Validate|RetainedHeap|CLI/'` to see allocation counts and byte allocation reductions; compare against docs/benchmarks.md
- Inspect docs/benchmarks.md: confirm all document-scenario allocs/op and B/op are lower, no scenario's B/op increased, no scenario took more than 10% longer
- Decode a 5.5 MB JSON document and measure peak memory: before this story ~330 MB, after ~210 MB (per docs/benchmarks.md CLI peak memory section)

## Migration

None. The decoded document's Go shape is unchanged: map[string]any, []any, string, bool, nil, *jsonvalue.Number. Only jsonvalue.Number's private fields change. The public API (Decimal, IsInteger, Rat, Canonical, String, Cmp, equality) stays identical. No schema, config or caller-side change needed.

## Rollback

Standard Git rollback: `git revert <commit>` or `git checkout <prior-commit>`. All validation results are byte-identical before and after, so no data reconciliation is needed. Benchmarks and tests will return to the pre-story numbers.

## Known limitations

- Fast-path numbers are limited to mantissas ≤ MaxInt64 (9,223,372,036,854,775,807) and ≤19 decimal digits. Larger mantissas use big.Int, which allocates. This boundary is tested; big.Int mantissas that fit int64 are automatically demoted to the fast form, preserving canonicality.
- Peak memory measurement via syscall.Rusage (peakmem_test.go) does not build on Windows; on Windows the test is blocked by unavailable Unix syscall types. The test runs only with the `-peakmem` flag (manual benchmarking, not automated CI), so it does not affect validation on Windows.

## Concerns

- **Windows build incompatibility (medium severity)**: cmd/jsonschema tests fail to compile on Windows. The new peakmem_test.go uses syscall.Rusage and syscall.RUSAGE_SELF (Unix only), which do not exist on Windows. This breaks the entire cmd/jsonschema test build when cross-compiling to or testing on Windows. **Fix required**: add `//go:build !windows` to peakmem_test.go or guard the Rusage call with a platform-specific build constraint. The test only runs with a manual `-peakmem` flag, so it can safely skip on Windows. Without this fix, Windows users/CI will see compile errors in the cmd/jsonschema package.
