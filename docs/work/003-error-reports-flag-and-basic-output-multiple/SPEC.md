# SPEC #3: Error reports: flag and basic output, multiple documents

- Date: 2026-10-16
- Type: feature
- Verdict: pass
- Status: approved by the operator at Checkpoint 1 (run #137, 2026-10-16 19:01 UTC)

## Problem

The validator prints only a minimal valid or invalid result. Scripts, evaluators and people cannot see where a document failed or which rule failed. The CLI also checks only one document per run.

## Why this matters

The brief makes the 2020-12 'flag' and 'basic' output shapes mandatory, with 'basic' as the default. Reports must be safe for logs. The CLI must check several documents in one run, with the worst result setting the exit code. Later shapes ('detailed', 'verbose') and later keywords must be able to add output without editing shared files.

## Acceptance criteria

- The library can render a validation Result as a 2020-12 'flag' object ({"valid": true|false} and nothing else) and as a 2020-12 'basic' object. The shape is selected by name through ONE output-format registration point (internal/output/registry.go). Each shape lives in its own file and registers itself, so a later shape is added as a new file only. An unknown shape name is an error.
- In 'basic', a valid document gives exactly {"valid": true}. No annotations are included in this story. An invalid document gives {"valid": false, "errors": [...]}. Each error unit has 'instanceLocation' and 'keywordLocation' as JSON Pointers, 'absoluteKeywordLocation' only where an absolute URI is known, and 'error', a short sentence naming the failed rule.
- The 'basic' errors list contains each failing keyword and each surrounding applicator unit that failed because of it. The order is fixed and deterministic, so the same schema and document always give byte-identical output.
- No output unit, error sentence or CLI line ever contains an instance value. Error sentences name only the rule, schema-side limits and schema-side property names (for example "missing required property 'email'"). This is covered by a test that plants a recognisable value in a failing document and asserts it is absent from both shapes.
- Evaluation counts EVERY output unit it appends, including units appended under passing applicators and later discarded. It does not count the length of the retained list. When the count exceeds DefaultMaxOutputUnits (or the caller's raised option value), evaluation stops with a LimitError naming 'output units'. That is an error, never a valid or invalid result. A regression test lowers the cap and shows that many discarded child failures under passing anyOf/oneOf-style branches trigger the LimitError.
- The CLI accepts '-o' and '--output' to choose the shape ('flag' or 'basic'). The default is 'basic'. An unknown value such as '-o fancy' exits with code 2 and prints nothing on stdout.
- The CLI accepts one or more document paths. '-' or no path reads standard input. When exactly one document is checked, stdout is exactly one 2020-12 output object for the chosen shape, followed by a newline and nothing else.
- When several documents are checked, stdout has one JSON line per document, in argument order. Each line has the shape {"document": <path, or "-" for stdin>, "output": <unchanged standard output object>}. A document that cannot be checked (broken JSON, size limit, unreadable file) instead gets {"document": ..., "error": <reason without document content>} on its line, and the other documents are still checked. With one document, that case gives empty stdout, the reason on stderr and exit code 2.
- When the schema itself cannot be used (broken JSON, compile error, unresolved reference), the problem is reported once on stderr with exit code 2. Nothing is printed on stdout, however many documents were given.
- The exit code is the worst result: 2 if any document could not be checked, otherwise 1 if any document is invalid, otherwise 0.
- CLI contract tests cover both shapes, stdin input, multiple documents (including a mix of valid, invalid and broken documents), schema errors with several documents, the unknown -o value, and every exit code. All tests use no network and pass under go test -race ./....

## Non-goals

- 'detailed' and 'verbose' shapes (story 194).
- Annotations in any output shape (story 'Annotation collection', 191).
- CLI flags for the other limits (story 193). This story adds only the output-unit cap, its library option and its default.
- Absolute keyword locations through $id/$ref resolution (stories 186/187). Until then 'absoluteKeywordLocation' is omitted where no absolute base URI is known.
- New validation or applicator keywords. anyOf/oneOf themselves arrive in story 185.
- Documentation of the CLI contract (story 197).

## Tasks

- **T1** — Collect output units during evaluation, with a cap that counts every appended unit
  - purpose: Give keywords a way to report a failure with a value-free message. Record instance and keyword locations for each failure, and let applicators evaluate children whose units are kept or discarded in a fixed order. Count every appended unit against the output-unit cap, and stop with a LimitError 'output units' when it is exceeded.
  - scope: internal/eval/units.go, internal/eval/units_test.go, internal/eval/eval.go, internal/keywords/registry.go, internal/compile/compile.go, internal/keywords/validation/type.go, internal/keywords/validation/enum.go, internal/keywords/validation/const.go
  - done when: go test -race ./internal/eval/... ./internal/keywords/... passes. units_test.go registers a test-only passing applicator whose discarded child failures exceed a lowered cap, and asserts that evaluation stops with the LimitError 'output units'. It also asserts the deterministic order and that no instance value appears in any unit.
  - delivers: AC3, AC4, AC5
- **T2** — Output-format registry with 'flag' and 'basic'
  - purpose: Create the single output-format registration point and add the 'flag' and 'basic' shapes as self-registering files that render collected units to 2020-12 JSON.
  - scope: internal/output/registry.go, internal/output/registry_test.go, internal/output/flag.go, internal/output/flag_test.go, internal/output/basic.go, internal/output/basic_test.go
  - done when: go test ./internal/output/... passes. Tests cover the exact JSON for valid and invalid flag/basic, omission of absoluteKeywordLocation when unknown, the unknown-shape error and duplicate registration.
  - depends on: T1
  - delivers: AC1, AC2
- **T3** — Library API: render a Result in a chosen shape, with an output-unit cap option
  - purpose: Expose output rendering on Result. Add DefaultMaxOutputUnits and a raisable Options field. Make exceeding the cap an error returned from Validate.
  - scope: jsonschema.go, output.go, output_test.go, numberlimits.go
  - done when: go test -race . passes. output_test.go checks flag and basic output from the library, absence of instance values, the LimitError from Validate when the cap is lowered, and that the test reads the cap from DefaultMaxOutputUnits instead of repeating the number.
  - depends on: T1, T2
  - delivers: AC1, AC2, AC4, AC5
- **T4** — CLI: -o/--output, several documents, NDJSON lines and worst-result exit code
  - purpose: Wire the output option, multi-document input, labelled lines, per-document errors, one-time schema errors and worst-result exit codes into the thin CLI.
  - scope: cmd/jsonschema/run.go
  - done when: The CLI builds with go build ./cmd/jsonschema, and the tests in T5 pass against it.
  - depends on: T3
  - delivers: AC6, AC7, AC8, AC9, AC10
- **T5** — CLI contract tests
  - purpose: Prove the CLI contract for both shapes, stdin, multiple documents, schema errors, unknown -o values and exit codes. Update existing CLI tests whose expected stdout changes.
  - scope: cmd/jsonschema/output_test.go, cmd/jsonschema/multidoc_test.go, cmd/jsonschema/run_test.go
  - done when: go test -race ./cmd/jsonschema/... passes. Every exit code (0, 1, 2) is asserted with one and with several documents.
  - depends on: T4
  - delivers: AC6, AC7, AC8, AC9, AC10, AC11
