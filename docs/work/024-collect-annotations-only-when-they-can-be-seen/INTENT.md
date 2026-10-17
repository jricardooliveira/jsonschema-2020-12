# INTENT #24

- Author: Joao Oliveira
- Date: 2026-10-17
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Stop collecting and sorting annotations when nothing can ever show them. Also stop splitting JSON Pointers on every comparison when annotations are sorted. The annotations in the library Result must stay identical for every caller: the library may not skip annotations unless a caller asks for them. Skip collection only where no Result accessor and no report could ever show the annotations. Otherwise, build them on demand with identical content and order. Sort annotations without re-splitting pointers, for example by comparing precomputed segments, and keep exactly the same order as today. Results must stay identical everywhere: the same valid or invalid answer, the same flag, basic, detailed and verbose reports, and the same library Result. All existing tests must pass unchanged. Every hostile-input limit must keep working exactly as before. That covers nesting depth, ref chain, evaluation-step charging, output units, report bytes and annotation count. Each must trigger at exactly the same inputs, with the same error and exit code 2. Add one regression test that proves this. The compiled Schema must stay immutable and safe for concurrent use, checked with go test -race. Run the project's benchmarks on the same machine before and after the change, and add both sets of numbers to docs/benchmarks.md. The change must show a reduction in allocs/op or B/op, and in ns/op where it targets speed, on the scenarios it targets. No scenario's allocs/op or B/op may increase.

## Operator clarifications
- Today, programmers get annotations in the library Result without asking for them. You said the library may only skip them when a caller asks. Should this story add a new library option so a caller can say "I don't need annotations"? — No new library option and no API change. Every caller keeps getting today's annotations in the Result, with the same content and order. Skip collecting them only where no Result accessor and no report could ever show them; if they are built on demand instead, the annotation-count limit must still fail Validate at exactly the same inputs with the same limit error and exit code 2, and the Result stays immutable and safe to read from several goroutines (go test -race).
- Timings on your Mac vary by about 5% from run to run, while memory allocation counts are reliable. How big must the improvement be for this story to count as done? — Fewer allocations required; speed may not get worse — Easier to meet. Speed only has to stay within the normal 5% wobble. Allocation and memory rules stay the same.
- The project has tests that fail if checking a document or loading a schema uses more memory allocations than a set limit. After this story, the code should use fewer allocations than those limits allow. Should this story also lower those limits to match the new code? — No allocation-limit tests for the order documents exist in the project yet (the only memory check today is the number-amplification test in internal/jsonvalue/number_test.go, which stays unchanged). Story 'Allocation limits in tests' adds every limit once, after all five fixes, from the measured counts plus about 10% (the user's answer). This story adds or changes no allocation-limit test; it records its before and after benchmark numbers in docs/benchmarks.md.

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- The inventory is truncated, so it is not confirmed whether the Result type and its annotation accessors live in jsonschema.go, annotations.go or another root file. The write claim on jsonschema.go is precautionary.
- It is not confirmed whether the Result exposes annotations on an invalid result, or whether detailed and verbose show annotations from failing subschemas. Either one shrinks the set of subtrees that can safely be skipped. The coder must check internal/output and the existing tests before choosing the skip rule.
- The names of the story 198 benchmarks and their bench_test.go locations are not known yet, because story 198 is approved but not released.
- It is not known how many annotations the order schema produces in this codebase, for example whether properties and items results count as annotations. This decides how large the reduction in AC7 can be.
- Stories 199, 200 and 203 also change internal/eval/eval.go or the benchmark numbers. Their order relative to this story changes the 'before' baseline in docs/benchmarks.md.

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
