# SPEC #29: One reusable child evaluation context per depth

- Date: 2026-10-17
- Type: performance
- Verdict: pass
- Status: passed gate 1 with nothing for the operator to decide (run #164, 2026-10-17 21:44 UTC)

## Problem

EvalChild in internal/keywords/registry.go allocates a new EvalContext for every child evaluation. That is about 19% of the bytes Validate allocates.

## Why this matters

This is round-3 story 2. Its target is Validate5_5MB at most 200 MB B/op on top of story 1 (#205), plus measured drops on ValidateErrors and ValidateSmall. Results, output, exit codes, limits, error messages and goroutine safety must not change.

## Acceptance criteria

- EvalChild takes its child EvalContext from the per-call collector, which keeps one context per depth and resets it for each child. A new unit test shows that sibling children at the same depth reuse one context and that every reset clears all per-child state.
- The caller still owns the returned units slice. A new test shows that units returned for one child are not aliased or overwritten when a later sibling at the same depth is evaluated.
- A doc comment next to EvalChild and the per-depth context stack states two rules: no keyword may keep its child context after EvalChild returns, and the caller owns the returned units slice.
- In test binaries only, a reused child context is wiped (poisoned) as soon as EvalChild returns, so any later use fails loudly. A new test proves that a keyword which keeps its child context fails under this check. Production builds behave as without the check.
- Any keyword found to need data from its child context after EvalChild returns keeps a fresh context, as today. Each such keyword is named in the story record and documented next to the code. No applicator, annotate or recording file is edited for this.
- Per-call state (the context stack) is never shared between Validate calls. go test -race ./... is green, including a new concurrent-validation test that exercises reused child contexts.
- All existing tests pass unchanged. None is removed or weakened. allocs_test.go and docs/benchmarks.md are not edited.
- The #205 golden comparison test passes: CLI stdout, stderr and exit code are byte-identical to the committed v1.1 golden files in flag, basic, detailed and verbose. No golden file is regenerated or edited.
- A new root-package file allocs_childctx_test.go holds this story's lowered allocation limits for the guarded scenarios. Each limit is the measured value plus about 10% headroom, with the measured value recorded next to it. The file reuses the helpers and race gating of allocs_test.go and allocs_race_on/off_test.go without editing them.
- A new file docs/benchmarks/r3-206-childctx.md records the following: before medians (main after #205) and after medians from go test -bench -count=10 on the same machine (B/op, allocs/op, ns/op); the absolute values after the story; and a verdict for each target. Targets: Validate5_5MB at most 200 MB B/op combined (the 305 MB stand-alone target is noted as not measured); ValidateErrors drops by at least about 4.8 MB B/op; ValidateSmall drops by at least about 6,200 B/op. Timings are noted as noisy because they were taken during a parallel build, to be re-measured at Checkpoint 3.
- If any target is missed, the numbers are recorded and the story is held for the operator's decision. It is never marked done silently.
- Standard library only. go.mod is unchanged.

## Non-goals

- Measuring the 305 MB stand-alone target on a separate copy of v1.1 code
- Editing allocs_test.go, docs/benchmarks.md or any golden file
- Editing files claimed by #207 (internal/keywords/applicator/*, internal/keywords/annotate.go, internal/keywords/recording.go, internal/eval/evaluated.go), #210 (internal/jsonvalue/number_inline*, internal/keywords/validation/*) or #211 (internal/registry/*, internal/compile/metacache.go)
- Changing verbose output's second evaluation pass
- Any other allocation cut from round 3

## Tasks

- **T1** — Measure the before numbers
  - purpose: On main with #205 integrated, run the in-repo benchmarks (ValidateSmall, Validate0_5MB, Validate5_5MB, ValidateErrors, ValidateErrorsBasic, CompileOrder, CLI) with go test -run '^$' -bench . -benchmem -count=10, and record the medians as 'before'.
  - scope: docs/benchmarks/r3-206-childctx.md
  - done when: Before medians for every scenario recorded in the new doc, with the commit hash and machine.
  - delivers: AC10
- **T2** — Per-depth child context stack
  - purpose: Add the per-call stack of reusable EvalContexts indexed by depth, with a reset that clears all per-child state. Document the invariants next to the code.
  - scope: internal/keywords/childctx.go, internal/keywords/childctx_test.go
  - done when: Unit tests show reuse per depth, a complete reset and no aliasing of returned units.
  - depends on: T1
  - delivers: AC1, AC2, AC3
- **T3** — Switch EvalChild to the reused contexts
  - purpose: Make EvalChild acquire and release the depth's context through the stack, keep the caller's ownership of the units slice, and keep a fresh-context path for any keyword that needs its child context afterwards.
  - scope: internal/keywords/registry.go, internal/eval/eval.go
  - done when: go test ./... and the #205 golden test pass unchanged. Any keyword on the fresh-context path is listed.
  - depends on: T2
  - delivers: AC1, AC2, AC5, AC7, AC8, AC12
- **T4** — Test-only poison check and concurrency test
  - purpose: In test binaries only, wipe a released child context so that a retained use panics. Prove with a test that it catches a retaining keyword, and add a concurrent-validation race test.
  - scope: internal/keywords/childctx_poison.go, internal/keywords/childctx_poison_test.go, childctx_race_test.go
  - done when: The poison test fails a deliberately retaining test keyword. go test -race ./... is green.
  - depends on: T3
  - delivers: AC4, AC6
- **T5** — Lowered allocation limits in a new file
  - purpose: Re-measure the guarded allocation scenarios and add limits (measured value plus about 10%) in allocs_childctx_test.go, with the measured values recorded next to them.
  - scope: allocs_childctx_test.go
  - done when: The new limits pass with and without -race. allocs_test.go is unchanged.
  - depends on: T3
  - delivers: AC9
- **T6** — After numbers and target verdicts
  - purpose: Re-run the benchmarks with -count=10 on the same machine. Record the after medians, the absolute values and the verdict for each target, and hold the story if any target is missed.
  - scope: docs/benchmarks/r3-206-childctx.md
  - done when: The doc shows before and after medians and a pass, miss or hold verdict for each target. go test -race ./... and the golden test are green.
  - depends on: T4, T5
  - delivers: AC7, AC8, AC10, AC11
