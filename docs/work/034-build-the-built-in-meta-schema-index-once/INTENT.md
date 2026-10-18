# INTENT #34

- Author: Joao Oliveira
- Date: 2026-10-17
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Stop building the built-in meta-schema index twice. Today compile/metacache.go calls registry.BuildBuiltinIndex() even though registry.AddBuiltins in registry.go has already built and shared the index. The second build costs about 165 µs and 118 KB per process. Use one shared builder for both. Add a counter test that shows exactly one built-in index build per process. The FirstLoadOrder benchmark rebuilds on purpose, so add a separate first-compile measure. Validation behaviour, output, exit codes, limits, load errors, error messages and goroutine safety must not change at all. The shared index must stay immutable and safe to share across goroutines. Use the standard library only. Do not remove or weaken any existing test. go test -race must pass. CLI stdout, stderr and exit code must stay byte-identical to the committed v1.1 golden files in flag, basic, detailed and verbose, for the perf fixtures, docs/examples and the testdata cases. Never regenerate or edit a golden file. If output differs, the story has failed and must stop. Record before and after numbers, including the new first-compile measure, from the in-repo benchmarks on the same machine in docs/benchmarks.md. Re-measure the allocation limits in allocs_test.go and lower them on purpose, as docs/benchmarks.md describes. If any number gets worse, record it and hold the story for the operator's decision.

## Operator clarifications
- This story, unlike the other six, has no number target. The new first-compile measure should save about 165 µs and 118 KB, since the index is no longer built twice. What should count as success? — Clear saving, no fixed number — Recommended. The first-compile medians must be lower than before in both time and memory, and the before and after numbers are recorded. If there is no saving, the story is held for your decision. This avoids a hard target that may not hold on the build machine.
- To prove the index is built only once, the test has to count the builds. That means a small counter in the product code that only tests read. Is that acceptable? — Yes, a tiny internal counter — Recommended. One counter inside the registry package. It cannot be seen from outside the library, it does not change any result, and it is safe when several checks run at the same time.
- The saving only happens once per program run, at the very first schema load. A normal speed test repeats the same step many times in one run, so after the first time the index already exists and nothing is saved. How should the new first-load measure get a clean start each time? — Start a fresh test program for each run — Recommended. A test starts its own program again, about 10 times, and times only the first schema load in each. This measures exactly what users get. It uses only the standard Go toolchain and records the result in docs/benchmarks.md without asserting a time.
- Should the first schema load also get an allocation limit in allocs_test.go, so a later change cannot quietly bring the second index build back? — Yes, a limit from the new count plus about 10% — Recommended. Same rule as the other limits. Together with the counter test, it catches the double build returning.
- Operator note on refining again — Operator note (parallel build): stories #206, #207, #210 and #211 will be built AT THE SAME TIME after #205 is integrated, so none of them may write a file another of them writes. Change only this:
1. Do NOT write allocs_test.go or docs/benchmarks.md (shared files; #208 and #209 stay serial and keep writing them). Instead:
   a. Put this story's re-measured, lowered allocation limits in a NEW root-package test file allocs_<topic>_test.go (for example allocs_childctx_test.go, allocs_objectkeys_test.go, allocs_multipleof_test.go, allocs_firstload_test.go), reusing the existing helpers and race gating of allocs_test.go / allocs_race_on/off_test.go read-only. Each new limit = this story's measured value + about 10% headroom, measured value recorded next to it. The existing limits in allocs_test.go stay untouched (they remain valid, looser upper bounds); that is the deliberate lowering for this story.
   b. Record this story's before/after medians (go test -count=10, before = main after #205, after = with this story) and the target verdicts in a NEW file docs/benchmarks/r3-<story number>-<topic>.md.
2. Do not write any file claimed by the other three stories (#206: internal/keywords/registry.go and childctx*; #207: internal/keywords/applicator/*, internal/keywords/annotate.go, internal/keywords/recording.go, internal/eval/evaluated.go; #210: internal/jsonvalue/number_inline*, internal/keywords/validation/*; #211: internal/registry/*, internal/compile/metacache.go). If you need one at the gate, the answer will be no.
3. Benchmark timings taken while other stories build in parallel are noisy: record them, but the operator re-measures interleaved at Checkpoint 3 and judges the 5% rule on those numbers. B/op and allocs/op are exact and judged as before.
Everything else in the accepted plan stays unchanged.
- The new first-load measure starts a fresh test program about 10 times, which adds a few seconds and only makes sense on a quiet machine. When should the timing part run? — Only when asked, like the peak-memory test — Recommended. Skipped in a normal 'go test' run and switched on with a test flag, in the same way as the existing opt-in TestCLIPeakMemory. Normal test runs stay fast and never depend on timings. The numbers are recorded in docs/benchmarks/r3-211-firstload.md.
- The 'one index build per process' counter test and the new first-load allocation limit only make sense in a fresh program, because other tests and the FirstLoadOrder benchmark rebuild the index on purpose. Should these two checks run in every normal test run? — Yes, every run, each in one fresh test program — Recommended. Each check starts one fresh program (well under a second) and checks the build count and the allocation count. Allocation counts are exact, so this is reliable and catches a second build coming back on any machine. The race-detector setting of the existing allocation tests is followed.

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- The inventory is truncated, and internal/registry is not listed. The exact file holding BuildBuiltinIndex and AddBuiltins is assumed to be internal/registry/registry.go, as the request states. If the builder lives in another file inside internal/registry, that file replaces registry.go in the write claim (still inside #211's claimed area).
- AddBuiltins may already hold a once-guarded shared index. In that case T1 only adds the counter, and T2 just points metacache at the existing accessor.
- Where FirstLoadOrder lives (root bench_test.go or internal/compile/bench_test.go) and how it forces a rebuild is unknown. If it depends on metacache calling BuildBuiltinIndex, keeping its 'rebuild on purpose' behaviour may need a change to a benchmark file. That would be a read-to-write change that must be confirmed so that no existing test is weakened.
- It is unknown whether any root test file compiles a schema at package init. That would make the fresh-process first-load allocation measure include a prior load, and must be checked before fixing the limit.
- The 'before' numbers must come from main after #205. Parallel builds of #206, #207 and #210 make timings noisy, so the time verdict is provisional until the operator re-measures at Checkpoint 3.

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
