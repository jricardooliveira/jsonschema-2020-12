# INTENT #22

- Author: Joao Oliveira
- Date: 2026-10-17
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Stop building keyword locations and absolute keyword locations for every keyword evaluated, including keywords that pass. Profiling shows this costs about 30% of allocated memory. Build a location only when a report or a library Result will actually show it, for example by keeping a cheap reference to the schema position and producing the location string on demand. Results must stay identical everywhere: the same valid or invalid answer, the same flag, basic, detailed and verbose reports, and the same library Result. All existing tests must pass unchanged. Verbose output stays correct and keeps its current second evaluation pass. Every hostile-input limit must keep working exactly as before. That covers nesting depth, ref chain, evaluation-step charging, output units, report bytes and annotation count. Each must trigger at exactly the same inputs, with the same error and exit code 2. Add one regression test that proves this. The compiled Schema must stay immutable and safe for concurrent use, checked with go test -race. Run the project's benchmarks on the same machine before and after the change, and add both sets of numbers to docs/benchmarks.md. The change must show a reduction in allocs/op or B/op, and in ns/op where it targets speed, on the validation scenarios. No scenario's allocs/op or B/op may increase.

## Operator clarifications
- How big must the memory saving be for you to accept this change? Profiling says building locations costs about 30% of allocated memory on the validation scenarios. — Any clear reduction, no fixed target — Accepted if allocs/op or B/op go down on the validation scenarios and none goes up. Your overall goal of matching tekuri is still checked later, outside the project.
- The project's tests fail if checking a document uses more memory allocations than an agreed limit. After this change the real counts drop. Should this story lower those limits? — 4
- Timings change by about ±5% from run to run on your Mac. How should the before and after numbers in docs/benchmarks.md be measured? — Run each benchmark 10 times and record the middle value — Smooths out the ±5% noise. It takes a few minutes and uses only the Go tools.
- Earlier you answered '4' when asked whether this story should lower the memory-allocation limits in the tests. I can't tell from that alone what you chose. Which of these do you mean? Your brief sets these limits at 'the measured count after the fixes, plus about 10% headroom'. You also asked that all existing tests pass unchanged, and lowering a limit means editing an existing test's number. — No allocation-limit tests for the order documents exist in the project yet (the only memory check today is the number-amplification test in internal/jsonvalue/number_test.go, which stays unchanged). Story 'Allocation limits in tests' adds every limit once, after all five fixes, from the measured counts plus about 10% (the user's answer). This story adds or changes no allocation-limit test; it records its before and after benchmark numbers in docs/benchmarks.md.
- Building location references once, when the schema is loaded, would make each document check cheaper. But loading the schema might then use a few more allocations. The story says no scenario may use more memory. Does that rule also cover loading the schema? — Yes, loading the schema must not get worse either — Recommended. This is the strictest reading of what you wrote. Locations are then built while a document is being checked, and only when a report needs them. Loading the schema stays the same or gets cheaper.
- Run times vary by about ±5% on your Mac. This change is mainly about memory. What should count as acceptable for speed (ns/op) on the scenarios where documents are checked? — Memory must go down; speed must not get worse beyond the ±5% noise — Recommended. A speed drop that stays within normal run-to-run variation is not a failure. Any speed gain is recorded but not required.
- The new regression test must prove that each safety limit (nesting depth, chain of references, evaluation steps, number of report entries, report size, number of annotations) still triggers at exactly the same point as before. How should the test check that 'same point'? — Right at the edge, for every limit — Recommended. For each limit, one input exactly at the limit still gets a normal valid or invalid answer, and one input just over it gets the same error and exit code 2 as today. Small inputs built by test code, so the test is quick and catches any shift by even one step.

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- The inventory is truncated: internal/output/*, internal/keywords/<vocabulary>/* beyond applicator, the root bench_test.go and options.go contents are not visible. If keyword evaluate functions or the registry's evaluate signature take or build location strings, those files must join the design. In that case, consider splitting into (a) the evaluator core and output and (b) the keyword files.
- It is not confirmed whether the compiled schema already stores per-keyword location strings. If it does, they stay as they are (no compile change), and this story only stops concatenating new strings during evaluation.
- The exact names of the six limits' options and their defaults are assumed to be readable from options.go and cmd/jsonschema/limitflags.go. If any limit is not exposed for a small at-the-edge setting, the regression test has to use its default, which may need larger generated inputs.
- docs/benchmarks.md and the benchmarks are created by story 198 and must exist before the before numbers can be measured.
- Stories 200, 201 and 203 also write internal/eval/eval.go, internal/eval/annotations.go and docs/benchmarks.md. They are not dependencies, but they cannot be built at the same time as this story, and before and after numbers must be measured on whatever code is current when this story starts.

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
