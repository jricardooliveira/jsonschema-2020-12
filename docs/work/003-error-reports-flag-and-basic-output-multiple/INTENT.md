# INTENT #3

- Author: Joao Oliveira
- Date: 2026-10-16
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Make the validator produce the standard 2020-12 output formats 'flag' and 'basic' as JSON, in both the library and the CLI. The CLI picks the shape with an output-format option, and 'basic' is the default. Every reported error gives the instance location, the keyword location, the absolute keyword location where one is known, and which rule failed. A report never quotes the document's values, so it is safe to save in logs. Give output formats one registration point so later shapes can be added as new files. The CLI accepts several document paths in one run. When exactly one document is checked, stdout is exactly one 2020-12 output object for the chosen shape and nothing else. When several documents are checked, the CLI prints one JSON report per line. Each line carries a label for its document, and the label does not change the standard output object inside the line. The worst result sets the exit code: 2 if any document could not be checked, otherwise 1 if any document is invalid, otherwise 0. Add CLI contract tests for both shapes, stdin input, multiple documents and exit codes.

## Operator clarifications
- When several documents are checked, each output line needs a label saying which document it belongs to. What should one line look like? — A wrapper with the file path and the report — Example: {"document": "orders/a.json", "output": {"valid": false, "errors": [...]}}. The standard report sits unchanged inside "output". A document read from standard input is labelled "-".
- Sometimes a document cannot be checked at all, for example a broken JSON file or a size limit that is exceeded. What should the tool print for that document? — Explanation on the error stream, nothing on stdout — With one document, stdout stays empty and the reason goes to the separate error stream, with exit code 2. With several documents, that document's line carries an "error" field with the reason instead of "output", and the other documents are still checked.
- The 2020-12 'basic' shape may also list the notes collected while checking a valid document (for example titles, descriptions, format names and which property names were checked). Should this story include them? — Not in this story: here a valid document gives just {"valid": true} in basic, because annotation collection does not exist yet. The later story 'Annotation collection' adds annotations to the output shapes as 2020-12 describes. Annotations may name schema keywords, locations and instance property names (JSON Pointer locations), but never instance values.
- What should the switch that picks the report shape be called on the command line? — --output with a short form -o — Example: jsonschema validate -o flag schema.json doc.json. A common, easy-to-remember name. Unknown values such as -o fancy give exit code 2.
- Sometimes the schema itself cannot be used, for example it is broken JSON, breaks the 2020-12 rules, or points to a schema that cannot be found. Then no document can be checked. What should the tool print when several documents were given? — One explanation on the error stream, nothing on stdout — Recommended. The schema problem is reported once on the error stream with exit code 2, and no per-document lines are printed. It is the same whether one or several documents were given, so scripts handle it one way.
- Each reported error also carries a short sentence for people saying which rule failed. Since reports must never quote the document's values, what may that sentence mention? — The rule and the schema's own limit, never the document's value — Recommended. Example: "must be at most 3" or "missing required property 'email'". The limit and the property names come from the schema, so nothing from the document is shown.
- When a document fails deep inside the schema, the 'basic' list can show only the rules that actually failed, or also each surrounding rule that failed because of them. For example, an order whose price is too high fails 'maximum' at /price, and because of that also fails 'properties' around it. Which should the list show? — Both the failing rule and each surrounding rule, in a fixed order — Recommended. This follows the 2020-12 examples, where wrapping rules such as 'properties' or 'allOf' also appear. The order is always the same for the same input, so tests and evaluators can compare reports exactly.
- Operator note on refining again — Relaunch after run #136 failed review. Keep the approved design of run #136 unchanged (2020-12 flag/basic output, multi-document NDJSON lines, -o/--output in cmd/jsonschema/run.go, failure-unit cap DefaultMaxOutputUnits with LimitError 'output units'). Concrete fix the tester required: the output-unit cap must count EVERY appended unit, including units appended by children of passing applicators that are later discarded, not the retained slice length; add a regression test where many discarded child failures under passing anyOf/oneOf branches exceed a lowered cap and evaluation stops with the LimitError. cmd/jsonschema/run.go is in scope (the CLI flags live there).

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- anyOf/oneOf (story 185) are not built yet, so the required regression test uses a test-only passing applicator registered in internal/eval/units_test.go. Once story 185 exists, a test with real anyOf/oneOf would strengthen this.
- It is unclear whether the 'output units' LimitError should reuse internal/jsonvalue.LimitError or be a root-level error type. The design reuses the existing LimitError if its fields allow a limit name; otherwise a new error type goes in internal/eval/units.go. decode.go is read only.
- absoluteKeywordLocation can only be filled once $id/$ref base URIs exist (stories 186/187). Until then it is omitted, which this story's criteria allow.
- cmd/jsonschema/run_test.go may assert the old minimal stdout for invalid documents. It is claimed for write in case those assertions must be updated.
- options.go is listed as crowded but does not exist in the inventory; Options currently lives in numberlimits.go, so the new option field goes there. If another story creates options.go first and moves Options, this claim moves with it.
- The CLI flag that raises the output-unit cap is left to story 193 (limit flags). Only the library option is added here.

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
