# INTENT #29

- Author: Joao Oliveira
- Date: 2026-10-17
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Cut evaluation allocations by reusing child evaluation contexts. Today EvalChild in internal/keywords/registry.go allocates a new EvalContext for every child evaluation, which accounts for about 19% of the bytes allocated. Change the per-call collector to keep one context per depth and reset it for each child. The caller still owns the returned units slice. Document that no keyword may keep its child context after EvalChild returns. Validation behaviour, output, exit codes, limits, error messages and goroutine safety must not change at all. Use the standard library only and add no dependency. Do not remove or weaken any existing test. go test -race must pass. CLI stdout, stderr and exit code must stay byte-identical to the committed v1.1 golden files in flag, basic, detailed and verbose, for the perf fixtures, docs/examples and the testdata cases. Never regenerate or edit a golden file. If output differs, the story has failed and must stop. Targets, measured on an Apple M3 Max with Go 1.27.2 (on other hardware, compare relative to the before numbers): Validate5_5MB at most 305 MB B/op for this change on its own, and at most 200 MB together with the path-frame change; ValidateErrors at most 24.5 MB B/op; ValidateSmall at most 35,500 B/op. Record before and after numbers from the in-repo benchmarks on the same machine in docs/benchmarks.md. Re-measure the allocation limits in allocs_test.go and lower them on purpose, as docs/benchmarks.md describes. If a target is missed, record the numbers and hold the story for the operator's decision.

## Operator clarifications
- This change has two memory targets for the 5.5 MB document: at most 305 MB on its own, and at most 200 MB together with the earlier change that reuses path frames (story 1). Story 1 is built first, so this story starts from code that already includes it. Measuring this change on its own would mean also building it on a separate copy of the old v1.1 code. Which target should decide whether this story passes? — Combined target only (200 MB), on top of story 1 — Recommended. The 'before' numbers are taken from the code with story 1 already in, and the story passes if the result is at most 200 MB. No separate copy of the old code is needed. The 305 MB target is noted as not measured.
- Suppose story 1 (path frames) is held for your decision because it missed one of its targets. What should happen to this story? — Wait until you decide on story 1 — Recommended. This story starts only from the code you have accepted, so its before and after numbers stay clear.
- The 200 MB target for the 5.5 MB document is judged on top of story 1. The other two targets (ValidateErrors at most 24.5 MB and ValidateSmall at most 35,500 B/op) do not say whether they apply to this change alone or to the combined code. How should they be judged? — Judged on the combined code like the 200 MB target, but by the drop this story causes, as for stories 3 and 6: measured before and after this story on the same machine, ValidateErrors must drop by at least the v1.1 gap to its target (v1.1 about 29.3 MB, so at least about 4.8 MB less) and ValidateSmall by at least about 6,200 B/op (v1.1 41,682 B/op to 35,500). Absolute values after the story are recorded too. A smaller drop is recorded and the story is held for my decision.
- Each depth will now reuse one child context. If some keyword quietly kept its child context after EvalChild returned, a later sibling could overwrite it, and a report could be wrong without any test noticing. How should the project guard against that? — Also a test-only check that clears reused contexts — In tests only, a context is wiped as soon as EvalChild returns, so any keyword that still uses it fails loudly. This catches more mistakes, but it adds a little test-build-only code.
- Suppose the builder finds a keyword that really needs data from its child context after EvalChild returns (for example, for the verbose report). What should happen? — Leave that keyword on a fresh context — That one keyword keeps allocating a new context as it does today. This is safe, but part of the memory saving is lost.
- Operator note on refining again — Operator note (parallel build): stories #206, #207, #210 and #211 will be built AT THE SAME TIME after #205 is integrated, so none of them may write a file another of them writes. Change only this:
1. Do NOT write allocs_test.go or docs/benchmarks.md (shared files; #208 and #209 stay serial and keep writing them). Instead:
   a. Put this story's re-measured, lowered allocation limits in a NEW root-package test file allocs_<topic>_test.go (for example allocs_childctx_test.go, allocs_objectkeys_test.go, allocs_multipleof_test.go, allocs_firstload_test.go), reusing the existing helpers and race gating of allocs_test.go / allocs_race_on/off_test.go read-only. Each new limit = this story's measured value + about 10% headroom, measured value recorded next to it. The existing limits in allocs_test.go stay untouched (they remain valid, looser upper bounds); that is the deliberate lowering for this story.
   b. Record this story's before/after medians (go test -count=10, before = main after #205, after = with this story) and the target verdicts in a NEW file docs/benchmarks/r3-<story number>-<topic>.md.
2. Do not write any file claimed by the other three stories (#206: internal/keywords/registry.go and childctx*; #207: internal/keywords/applicator/*, internal/keywords/annotate.go, internal/keywords/recording.go, internal/eval/evaluated.go; #210: internal/jsonvalue/number_inline*, internal/keywords/validation/*; #211: internal/registry/*, internal/compile/metacache.go). If you need one at the gate, the answer will be no.
3. Benchmark timings taken while other stories build in parallel are noisy: record them, but the operator re-measures interleaved at Checkpoint 3 and judges the 5% rule on those numbers. B/op and allocs/op are exact and judged as before.
Everything else in the accepted plan stays unchanged.

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- The inventory is truncated: it is not confirmed whether the per-call collector struct lives in internal/keywords/registry.go or internal/eval/eval.go. internal/eval/eval.go is claimed for at most a one-field wiring edit. If the collector lives entirely in registry.go, eval.go stays untouched.
- It is not yet known which keywords, if any, read their child context after EvalChild returns (likely candidates are applicators building verbose units in internal/keywords/applicator/*). Those files belong to #207 and are not written here. Any fresh-context exception must be selected from registry.go. If that is impossible without editing an applicator file, the story stops and asks the operator.
- Using testing.Testing() for the test-only poison check imports the testing package into a production package. If that is unacceptable, the alternative is a build-tagged file pair, but default go test runs would then not poison.
- The golden comparison test and its file names come from #205, which is still in progress. This story must start from #205's accepted code. If #205 is held, this story waits, per the operator's clarification.
- Benchmark timings taken during the parallel build of #206, #207, #210 and #211 are noisy. Only B/op and allocs/op are judged here; ns/op is re-measured by the operator at Checkpoint 3.
- The ValidateErrors and ValidateSmall drop thresholds (about 4.8 MB and about 6,200 B/op) were derived from v1.1 numbers on the M3 Max. On other hardware, only the relative drop is meaningful.

## Proposed outcome

See `SPEC.md` in this directory — the spec-agent's acceptance criteria are
the machine-readable statement of the outcome, derived from the problem above.

## Affected users and systems

See `PLAN.md` → *Files that change*, and the project's `project-spec.json`
for the systems in scope.

## Constraints

The project's `PROJECT_RULES.md` and `project-spec.json` (stack, conventions,
NFRs and the `forbidden` list) are injected into every agent and bind this work.

## Open questions

Recorded in `SPEC.md` → *Open questions*. A non-empty list parks the run at
Checkpoint 1 for the operator rather than failing it.
