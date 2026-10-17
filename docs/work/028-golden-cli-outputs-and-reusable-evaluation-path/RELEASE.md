# Release notes — #28: Golden CLI files from v1.1, then reuse evaluation path frames

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #28
- Status: proposed (pending Checkpoint 3 sign-off)

**⚠️ the release-agent has concerns.**

## What changed

Per-call path frame stack with embedded lazyLoc and explicit enter/exit calls cuts evaluation allocations by 27–40% (CompileOrder 40%, ValidateSmall 29%, Validate5_5MB 30%, ValidateErrors 33%). Golden files (2,524 test cases from unchanged v1.1) guard every later round-3 story. All targets met; all scenarios faster; go test -race green.

- cmd/jsonschema/golden_cases_test.go: enumerate 631 base cases (perf fixtures, docs/examples, testdata spec-derived documents) in 4 shapes
- cmd/jsonschema/golden_capture_test.go: opt-in CLI capture from unchanged base d7198b7 (production code identical to v1.1 f35d9e4); each case run twice; refuses to overwrite
- cmd/jsonschema/golden_test.go: compare current CLI output byte for byte against golden files; size+SHA-256 for outputs over ~64 KB
- testdata/golden/cli_v1_1.json: 2,512 records with full output
- testdata/golden/cli_v1_1_large.json: 12 records with size+SHA-256 (generated documents and large outputs)
- internal/eval/pathstack.go: per-call []pathFrame indexed by depth; enter(ctx,node) / exit(frame); embedded lazyLoc; fresh per Validate, never pooled
- internal/eval/pathstack_test.go: LIFO enter/exit, frame reuse at depth, per-call isolation, loop detection
- internal/eval/refguard.go: enterNode thin wrapper over callState.enter/exit for compatibility; pathFrame embeds lazyLoc; documented invariant (LIFO, valid only during node evaluation)
- internal/eval/eval.go: evalNode uses explicit st.enter(ctx,n) and defer st.exit(pf) instead of closures; lazyLoc set per keyword
- internal/eval/location.go: lazyLoc type kept for compatibility
- internal/eval/dynamicscope.go: enter/exit pair for explicit LIFO; push wrapper for dynamicscope_test.go
- verbose_nested_frames_test.go: nested $ref, if, allOf; verifies every verbose unit carries own instance, keyword and absolute keyword locations through reused frames
- allocs_test.go: limits lowered with measured values next to each (allocs: 2730→1640, 864→615, 726k→507k, 7.26M→5.06M, 528k→353k, 534.7k→359k; bytes: 39.4M→28.4M, 405.9M→295.9M, 32.2M→23.5M)
- docs/benchmarks.md: story 205 section with before/after medians (-count=10 on M3 Max Go 1.27.2), slowdown judgment (all faster, none ≥5%), target table (all met), golden case count (2,524)

## How to verify

- go test ./... -race: all tests pass; golden test runs ~60 s without -race (CLI builds once, runs 2,524 cases)
- go test -run 'TestGolden$' ./cmd/jsonschema: compares current output byte for byte against golden; fails at once on any difference
- go test -run 'TestPathStack' ./internal/eval: LIFO, reuse, isolation, loop detection all green
- go test -run 'TestVerboseNested' ./: sibling and error locations correct through reused frames
- go test -bench . -benchmem -count=10 ./: CompileOrder 1490 allocs (40% cut), Validate5_5MB 4.6M allocs (30% cut), all ns/op lower
- go test -run '^$' -bench . -benchmem -count=10 ./cmd/jsonschema: CLI allocs/op down 21–32%; no scenario ≥5% slower

## Migration

None. No API change; validation results, output formats, exit codes and goroutine safety identical. Golden files immutable after capture.

## Rollback

Restore internal/eval/refguard.go, eval.go, dynamicscope.go, location.go from before; remove pathstack.go, pathstack_test.go; remove testdata/golden/*; restore allocs_test.go limits; revert docs/benchmarks.md; remove or restore cmd/jsonschema/golden_*.go; restore spec_test.go to include golden directory.

## Known limitations

- Golden files immutable after capture; any output mismatch stops the story immediately (by design, to guard later stories)
- Stack reallocation does not move frames (pointers are stable via frame slice holding *pathFrame), but frame pointers are not held across deeper enters
- Some testdata cases may produce nondeterministic output (e.g., map order); capture stops and lists them for operator decision
- Verbose's second evaluation pass unchanged (not in scope this round); locations render identically through reused frames (verified by test)
- Allocation targets measured on Apple M3 Max, Go 1.27.2; on other hardware compare relative to before-numbers

## Concerns

- Empty file cmd/jsonschema/golden_dbg_test.go should be deleted
- spec_test.go was edited to exclude testdata/golden/ (justified: golden files are CLI records, not spec test cases); this is outside the original scope but was approved as a justified test change
