# INTENT #17

- Author: Joao Oliveira
- Date: 2026-10-17
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Add the 2020-12 'detailed' and 'verbose' output formats, as far as practical, as new formats registered through the output-format registration point. Both the library and the CLI output-format option can select them. The nested reports carry the instance location, keyword location and absolute keyword location, and they include annotations where the shapes call for them. They never quote document values. When one document is checked, stdout is exactly one output object. With several documents, each line holds the unchanged output object. Document in the code any part of the shapes that is only partly produced. Write tests for both shapes.

## Operator clarifications
- In the 'detailed' and 'verbose' reports, some annotations repeat parts of the checked document. The 'properties', 'patternProperties', 'additionalProperties' and 'unevaluatedProperties' rules record which property names of the document they matched, and these names can be private, for example a key like "jane.doe@example.com". Other annotations come only from the schema, such as 'title', 'format' or 'default'. Your earlier rule says reports never quote document data. How should annotations that list document property names be handled? — Show property names, never property values — Treat property names as locations, like the instanceLocation path, which already contains them. Closer to the 2020-12 examples, but a sensitive key name can then end up in logs.
- A 'verbose' report includes every rule that was checked, including the ones that passed. For a large or deeply nested document, the report can be much bigger than the document itself. Your untrusted-input rule already limits input size, nesting depth and runaway references. Should the size of the report have its own limit? — Yes, a generous limit you can raise — Recommended. If a report would grow past a large default size, the check stops with a clear error and exit code 2, the same as the other limits. Ordinary documents never reach the limit, and it can be raised with a library option and a command-line flag.
- Operator note on refining again — Add to this story: (a) low finding from story 188: absoluteKeywordLocation must be a valid URI, so percent-encode the JSON Pointer fragment (Resource.AbsoluteLocation in internal/compile/compile.go) for characters not allowed in a URI fragment, in every output shape, with a test using schema keys containing spaces, '%' and non-ASCII; (b) the rendered-report byte cap DefaultMaxReportBytes from story 191 also bounds detailed and verbose output. Keep everything else in the accepted plan.

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- The inventory is truncated after internal/keywords/validation. I could not confirm the exact files in internal/output (registry.go, flag.go, basic.go), whether an internal/uri package exists (the fragment helper is placed in internal/compile to avoid guessing), or where DefaultMaxReportBytes and its option and flag are defined. If the cap is enforced only in a per-shape file or after full rendering, tree.go must apply it itself.
- I don't know if internal/eval/units.go records passing keyword units. If it records only failures, verbose needs that file extended (claimed). If the change instead requires internal/eval/eval.go (crowded), the plan must be revisited.
- If the library render path or the CLI output-format flag validates names against a fixed list instead of the registry, jsonschema.go or cmd/jsonschema/run.go (both crowded) would need a write. They are claimed read-only for now.
- Existing tests (internal/compile/compile_location_test.go, cmd/jsonschema/output_test.go, internal/eval/units_test.go) may assert unencoded absoluteKeywordLocation values or the current unit set. If so, they must be added as write claims.
- Story 192 (unevaluated*, approved, not built) adds unevaluatedProperties annotations. This story handles property-name annotations generically and does not wait for 192, but its tests cannot cover unevaluatedProperties until 192 lands.
- Story 193 (limits, approved) may touch the same limit-error plumbing. This story only reuses the existing cap from 191.

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
