# INTENT #30

- Author: Joao Oliveira
- Date: 2026-10-17
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Remove wasted copies in the object keywords properties, additionalProperties and patternProperties. Today each object's member names are copied to the heap and sorted twice, the applied slice grows from a capacity of 1, the names are copied again for the annotation, and the keyword location is rendered for the evaluated record even when recording is switched off. Make these changes: sort into a small stack buffer that keeps the same byte order as today, using the heap only above 8 members; pre-size the applied slice; annotate without a copy; and render the location only when the record is kept, behind a Recording() check. unevaluatedProperties and unevaluatedItems must keep recording. Annotation values and order must stay unchanged. Validation behaviour, output, exit codes, limits, error messages and goroutine safety must not change at all. Use the standard library only. Do not remove or weaken any existing test. go test -race must pass. CLI stdout, stderr and exit code must stay byte-identical to the committed v1.1 golden files in flag, basic, detailed and verbose, for the perf fixtures, docs/examples and the testdata cases. Never regenerate or edit a golden file. If output differs, the story has failed and must stop. Targets, measured on an Apple M3 Max with Go 1.27.2 (on other hardware, compare relative to the before numbers): Validate5_5MB at most 345 MB B/op and at most 6.0 M allocs/op for this change on its own. Record before and after numbers from the in-repo benchmarks on the same machine in docs/benchmarks.md. Re-measure the allocation limits in allocs_test.go and lower them on purpose, as docs/benchmarks.md describes. If a target is missed, record the numbers and hold the story for the operator's decision.

## Operator clarifications
- The targets for this story (at most 345 MB and 6.0 M allocations on Validate5_5MB) are 'for this change on its own'. Stories 1 and 2 may already be merged, and together they are expected to bring the number to about 200 MB, so this target would be met before this story even starts. How should this story's target be judged? — Judge by the amount saved — Recommended. Take the saving the target implies over v1.1 (v1.1 number minus 345 MB, and minus 6.0 M allocations). The story must save at least that much compared with the code just before it, on the same machine. Gives a real check without extra branches. If it saves less, the numbers are recorded and the story is held for you.
- Story 1 said 'no scenario more than 5% slower', judged by comparing medians of 10 runs. This story's request does not mention speed. Should the same speed rule apply here? — Yes, same rule as story 1 — Recommended. Run each benchmark 10 times before and after and compare the medians; any scenario more than 5% slower is reported and the story is held for your decision. Not asserted in a test.
- The change sorts property names in a small fixed buffer for objects with up to 8 properties, and uses the old way for larger objects. The existing tests and golden files may not include objects right at that limit. Should this story add its own test to prove the results stay the same on both sides of the limit? — Yes, add a boundary test — Recommended. A new test checks objects with 0, 1, 8, 9 and many properties, including names that sort differently by byte order. It confirms that validation results, annotation values and their order match today's code in all four report shapes. It also checks that unevaluatedProperties still sees every evaluated property. This is one small extra test file and no existing test changes.
- Operator note on refining again — Operator note (parallel build): stories #206, #207, #210 and #211 will be built AT THE SAME TIME after #205 is integrated, so none of them may write a file another of them writes. Change only this:
1. Do NOT write allocs_test.go or docs/benchmarks.md (shared files; #208 and #209 stay serial and keep writing them). Instead:
   a. Put this story's re-measured, lowered allocation limits in a NEW root-package test file allocs_<topic>_test.go (for example allocs_childctx_test.go, allocs_objectkeys_test.go, allocs_multipleof_test.go, allocs_firstload_test.go), reusing the existing helpers and race gating of allocs_test.go / allocs_race_on/off_test.go read-only. Each new limit = this story's measured value + about 10% headroom, measured value recorded next to it. The existing limits in allocs_test.go stay untouched (they remain valid, looser upper bounds); that is the deliberate lowering for this story.
   b. Record this story's before/after medians (go test -count=10, before = main after #205, after = with this story) and the target verdicts in a NEW file docs/benchmarks/r3-<story number>-<topic>.md.
2. Do not write any file claimed by the other three stories (#206: internal/keywords/registry.go and childctx*; #207: internal/keywords/applicator/*, internal/keywords/annotate.go, internal/keywords/recording.go, internal/eval/evaluated.go; #210: internal/jsonvalue/number_inline*, internal/keywords/validation/*; #211: internal/registry/*, internal/compile/metacache.go). If you need one at the gate, the answer will be no.
3. Benchmark timings taken while other stories build in parallel are noisy: record them, but the operator re-measures interleaved at Checkpoint 3 and judges the 5% rule on those numbers. B/op and allocs/op are exact and judged as before.
Everything else in the accepted plan stays unchanged.

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- The inventory is truncated. I could not confirm whether internal/keywords/recording.go already exists or whether Recording() is already defined elsewhere, such as on EvalContext in registry.go. If it is defined in registry.go, this story only calls it. Adding or changing it there is forbidden (#206), so it would need to be raised at the gate.
- The exact name of #205's golden comparison test and its golden file paths are not known yet. This story only runs and reads them.
- The v1.1 Validate5_5MB B/op and allocs/op baselines needed for the saving-based verdict are read from docs/benchmarks.md. If v1.1 numbers on this machine are missing there, the verdict uses the recorded v1.1 figures relative to before, and this is stated in the doc.
- Whether the copy-free annotation can avoid every copy depends on how annotate.go stores values. If a retained copy is needed for safety, it is kept and the reason is recorded, rather than risking changed annotation values.
- Escape analysis may prevent the [8]string buffer from staying on the stack when it is passed through interface or closure boundaries. The resulting saving may then be smaller than the prototype's.
- Existing applicator tests (for example properties_test.go and additionalproperties_test.go) are expected to keep compiling because internal signatures are unchanged. If a signature change proves unavoidable, those test files must be listed and must not be weakened.

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
