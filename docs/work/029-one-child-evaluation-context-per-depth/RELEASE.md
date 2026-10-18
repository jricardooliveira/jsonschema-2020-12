# Release notes — #29: One reusable child evaluation context per depth

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #29
- Status: approved by the operator at Checkpoint 3 (run #164, 2026-10-17 22:09 UTC)

**⚠️ the release-agent has concerns.**

## What changed

Internal optimization: EvalChild now reuses one child evaluation context per depth from a per-call stack instead of allocating a new context for every child. This saves about 19% of the bytes Validate allocates. Validation results, output, exit codes, error messages and goroutine safety are unchanged. All three targets passed; all tests pass including golden comparison and race test. Two low-severity notes for reviewers (see concerns).

- internal/keywords/registry.go: EvalChild refactored to acquire child contexts from a per-call stack (stk *childStack field) indexed by depth; resets each context before evaluation, hands units to caller without copying
- internal/keywords/childctx.go: new childStack struct with at(depth) for stack access; context.reset(parent, instLoc, tree) clears all per-child state while keeping capacity and shared per-call fields
- internal/keywords/childctx_poison.go: new poison() function zeroing a released context in test binaries only (testing.Testing()), so retained use panics
- internal/keywords/childctx_test.go: new unit tests proving per-depth reuse, complete reset of per-child state, and no aliasing of returned units slices
- internal/keywords/childctx_poison_test.go: new tests proving the poison check catches a retained context and returned units survive poisoning
- childctx_race_test.go: new concurrent validation test (8 goroutines × 200 iterations) run with -race, verifying per-call context isolation
- allocs_childctx_test.go: new allocation-limit tests with measured values +~10% headroom (ValidateSmall 560 allocs / 25,600 B/op, Validate5_5MB 216.7 MB B/op, ValidateErrors 17.1 MB B/op)
- docs/benchmarks/r3-206-childctx.md: new benchmark record with before/after medians (go test -count=10), targets verdict, commit 079149b and Apple M3 Max machine

## How to verify

- go test ./... passes all existing tests unchanged, including golden comparison test
- go test -race ./... passes, including new concurrent validation test
- go test -run '^$' -bench . -benchmem -count=10 . ./cmd/jsonschema medians: Validate5_5MB 196.95 MB B/op (target ≤200 MB PASS), ValidateErrors 15.51 MB (−5.76 MB, target ≥−4.8 MB PASS), ValidateSmall 23,234 B/op (−7,479, target ≥−6,200 PASS)
- CLI output byte-identical to v1.1 golden files in all formats (flag, basic, detailed, verbose)
- TestPoisonCatchesRetainedChildContext in childctx_poison_test.go: poison check panics on a deliberately retained child context

## Migration

none

## Rollback

Revert internal/keywords/registry.go EvalChild to allocate fresh contexts (remove stk.at() and reset() calls); delete internal/keywords/childctx.go, internal/keywords/childctx_poison.go, internal/keywords/childctx_test.go, internal/keywords/childctx_poison_test.go, childctx_race_test.go, allocs_childctx_test.go and docs/benchmarks/r3-206-childctx.md. All existing tests pass unchanged.

## Known limitations

- The poison check (zeroing a released context) runs only in test binaries when testing.Testing() is true; production builds do not poison, so a retained context silently sees sibling state. Mitigation: poison test proves the check works; no current keyword retains.
- Any future keyword that needs data from its child context after EvalChild returns must be explicitly placed on the fresh-context path (stk==nil branch) and documented. Currently no keyword requires this.

## Concerns

- No test exercises reuse and non-aliasing through EvalChild itself at the call site; childctx unit tests prove the stack property, but a higher-level test over multiple EvalChild calls on siblings would strengthen confidence. Low severity: units detached by reference not copied, so aliasing would cause data corruption or be caught by the poison test.
- The per-call context stack lives on its own field (stk *childStack) rather than embedded in the collector. Low severity: design choice for clarity; indirection is minimal.
