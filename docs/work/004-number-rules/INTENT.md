# INTENT #4

- Author: Joao Oliveira
- Date: 2026-10-16
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Add the 2020-12 numeric validation keywords multipleOf, maximum, exclusiveMaximum, minimum and exclusiveMinimum. Each keyword lives in its own file and registers through the keyword registry. All comparisons use the exact number model. For example, 0.3 is a multiple of 0.1, comparisons on huge integers are exact, and 1.0 behaves as the integer 1. The keywords apply only to number instances and ignore every other type. Write spec-derived positive and negative tests, including decimal and big-number cases.

## Operator clarifications
- A number can be written very short but stand for something enormous, for example 1e999999999 (a 1 followed by almost a billion zeros). Checking 'multipleOf' or 'maximum' exactly on such a number could make the tool use huge amounts of memory or run for a very long time. Your earlier answer says crafted files must never hang the tool. What should happen with numbers like this? — Limit on number size, clear error when exceeded — Recommended. Numbers whose exponent or digit count goes past a generous built-in limit stop the check with a clear error (exit code 2), never a guessed valid or invalid. Ordinary numbers, including very large integers and long decimals, are never affected. The limit can be raised with the same library option and command-line flag style as the other limits.
- Operator note on refining again — Do not build a number-size cap in this story: story 179 (Exact JSON values) already owns the number-size cap, its library option, its CLI flag and its limit error, inside the JSON decoder. This story reuses that exact number model and adds no cap, no option and no flag of its own. Remove the dependency on story 193 (Limits); this story depends only on 178 and 179. Keep the five numeric keywords, exact math/big arithmetic and the spec-derived tests (decimal multipleOf, big numbers, exclusive bounds).

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- The inventory is empty, so the exact names in story 179's number model (the type name and how it exposes a big.Rat) and the location of the shared table-driven test runner are assumed, not verified. Story 179's jsonvalue files and runner must be read before coding. If the runner lives in a file each keyword test must edit, that file becomes a hub and should be reported.
- The vocabulary package folder is assumed to be internal/keywords/validation/, created by story 179 for type, enum and const. If story 179 used a different vocabulary folder name, every path above moves accordingly.
- If story 179's number model exposes no exact big.Rat accessor, a small accessor would be needed in a new file under internal/jsonvalue/ (not decode.go). That would add one write claim.
- The output from story 180 is assumed to render keyword failures generically from the evaluate result, so no output file changes. If story 180 is not yet built, the failure-reporting criterion is verified only at the library Result level.

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
