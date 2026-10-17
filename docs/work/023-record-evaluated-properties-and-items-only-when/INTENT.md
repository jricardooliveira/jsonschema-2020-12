# INTENT #23

- Author: Joao Oliveira
- Date: 2026-10-17
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Stop recording which properties and items each applicator evaluated when nothing can ever use that record. Profiling shows this always-on recording costs about 15% of allocated memory. Recording is needed only when the schema, including schemas it reaches through $ref and $dynamicRef, contains an unevaluatedProperties or unevaluatedItems keyword, or when some report or library Result would show it. Decide this when the schema is loaded, and skip the recording in every other case. Results must stay identical everywhere: the same valid or invalid answer, the same flag, basic, detailed and verbose reports, and the same library Result. All existing tests must pass unchanged, including every unevaluatedProperties and unevaluatedItems test through $ref and $dynamicRef. Every hostile-input limit must keep working exactly as before. That covers nesting depth, ref chain, evaluation-step charging, output units, report bytes and annotation count. Each must trigger at exactly the same inputs, with the same error and exit code 2. Add one regression test that proves this. The compiled Schema must stay immutable and safe for concurrent use, checked with go test -race. Run the project's benchmarks on the same machine before and after the change, and add both sets of numbers to docs/benchmarks.md. The change must show a reduction in allocs/op or B/op, and in ns/op where it targets speed, on the scenarios it targets. No scenario's allocs/op or B/op may increase.

## Operator clarifications
- A program can load a schema once and then ask for different report shapes on each check. The detailed and verbose reports can list which properties each rule looked at. Should the recording switch on just for those checks, or be decided only once at load? — Decide at load, switch on per check when needed — Recommended. At load, the schema is marked as needing the record or not. A check that asks for a report shape that shows it gets the record anyway. Reports stay identical, and flag and basic checks get the saving.
- Should the 'is the record needed?' decision be made for the whole schema at once, or rule by rule? — Whole schema at once — Recommended. If an unevaluatedProperties or unevaluatedItems rule appears anywhere the schema can reach, everything records as today. Otherwise nothing records. Your order schema has no such rule, so it gets the full saving, and the risk of a wrong answer is low.
- $dynamicRef can jump, while a document is being checked, to a schema in the registry that the main schema does not name directly. When the factory decides 'is any unevaluated rule reachable?', what should it count as reachable? — The schema plus everything in the registry it could ever reach — Recommended. If any schema that a $ref or $dynamicRef could land on contains an unevaluated rule, recording stays on. This is cautious, so an answer can never change.
- The project's tests fail if checking or loading a schema uses more memory allocations than an agreed limit (measured count plus about 10%). After this change, flag and basic checks on your order schema should allocate less. Should this story lower those limits to the new counts? — No allocation-limit tests for the order documents exist in the project yet (the only memory check today is the number-amplification test in internal/jsonvalue/number_test.go, which stays unchanged). Story 'Allocation limits in tests' adds every limit once, after all five fixes, from the measured counts plus about 10% (the user's answer). This story adds or changes no allocation-limit test; it records its before and after benchmark numbers in docs/benchmarks.md.
- Timings on your machine vary by about ±5%, while allocation counts are exact. This change mainly saves memory. In the before-and-after numbers in docs/benchmarks.md, what counts as success for speed (ns/op)? — Memory must drop; speed is recorded but not a pass mark — Recommended. The change passes when allocs/op or B/op drop on the targeted scenarios and nothing rises. ns/op is written down for both runs, and a change within the ±5% noise is acceptable.
- Operator note on refining again — Also include this item routed from story 199's release review (tester finding F1): internal/eval/location.go defines nodeLoc and kwLoc types that evaluation never uses (the lazyLoc type in internal/eval/eval.go is what runs), and internal/eval/location_test.go tests only those unused types. In this story, remove the unused types and their tests (if a file cannot be deleted, leave only its package clause and a one-line comment), and add a unit test for lazyLoc proving its keyword location and absolute keyword location equal the old eagerly built strings: through $ref and $dynamicRef, inside if/then/else, with ~0/~1 escaping and percent-encoding, and an empty absolute location before any reference is crossed. Add internal/eval/location.go, internal/eval/location_test.go and the new test file to the claims. Everything else in the plan stays.

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- The inventory is truncated, so it is unconfirmed whether keyword files under internal/keywords/applicator record evaluated items themselves or call eval helpers. If they record directly, those files must also be claimed, and the guard placed there.
- It is unconfirmed how Validate learns the requested report shape. If a Result can be rendered to detailed or verbose after Validate returns without re-evaluating, the record may have to stay on whenever any Result accessor could show it. In that case the saving may apply only where the Result provably never exposes it, and jsonschema.go or options.go may need a small internal change.
- It is unconfirmed whether the 'properties' and 'items' annotations shown in detailed and verbose reports come from the evaluated record or from the separate annotation collection. This determines exactly which check shapes must switch recording on.
- The internal/registry file names are not visible in the truncated inventory. The analysis may need read access to the registry index to enumerate $dynamicAnchor targets.
- Stories 201 and 203 also write internal/eval/eval.go and docs/benchmarks.md. They share files but are not dependencies, and the before and after numbers must be measured on the code as this story finds it.
- Whether location_test.go can be deleted outright or must be reduced to a package clause depends on the factory's file-deletion support.

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
