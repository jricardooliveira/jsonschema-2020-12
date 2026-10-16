# Release notes — #3: Error reports: flag and basic output, multiple documents

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #3
- Status: approved by the operator at Checkpoint 3 (run #137, 2026-10-16 19:53 UTC)

## What changed

The validator now renders validation results in the 2020-12 'flag' and 'basic' output shapes, with 'basic' (detailed failure list) as the default. Results show where validation failed and which rule failed, with no instance values exposed, making them safe for logs. The CLI now accepts multiple documents in one invocation, printing one JSON report per document, and the exit code reflects the worst result (2 if any document could not be checked, 1 if any is invalid, 0 if all are valid).

- internal/eval/units.go: New file. Collects failure units in deterministic depth-first order during evaluation. Counts every appended unit against a 10M default cap and returns a LimitError 'output units' when exceeded.
- internal/eval/eval.go: Evaluation now creates a collector and routes keyword failures into it via ctx.Report(). Keywords and applicators can keep or discard child units.
- internal/keywords/registry.go: EvalContext gains InstanceLocation, KeywordLocation, Report(), Units(), Keep() and EvalChild() methods. NewEvalContext() creates a root context.
- internal/compile/compile.go: Tracks the JSON Pointer location of each keyword and node from the schema root, so units can carry it.
- internal/keywords/validation/{type,enum,const}.go: Each keyword now calls ctx.Report() with a value-free error message when validation fails.
- internal/output/registry.go: New single registration point for output shapes. Shapes register from init().
- internal/output/{flag,basic}.go: New shapes. flag renders {valid: bool}. basic renders {valid} or {valid: false, errors: [...]}, with each error as {instanceLocation, keywordLocation, error}.
- output.go: New public API. Result.Output(shape) method renders a result to JSON. DefaultMaxOutputUnits constant exported.
- numberlimits.go: Options gains MaxOutputUnits field (0 means DefaultMaxOutputUnits).
- jsonschema.go: Result now holds units. Validate calls EvaluateUnits and passes the cap from Options.
- cmd/jsonschema/run.go: Complete rewrite. Adds -o/--output flag (default 'basic'). Accepts multiple document paths (or stdin if omitted). Single document prints one output object + newline. Multiple documents print NDJSON {document, output|error} lines. Schema errors reported once on stderr. Exit code is the worst result: 2 if any doc unreadable, 1 if any invalid, 0 if valid.
- cmd/jsonschema/output_test.go, multidoc_test.go: New test files covering CLI contract (both shapes, stdin, multi-doc, mixed valid/invalid/broken, schema errors, unknown -o, all exit codes).
- cmd/jsonschema/run_test.go: Removed the 'two docs' error case (now supported). Updated TestExitValidInvalid type assertion to accept the new basic output format.

## How to verify

- Single valid doc: `jsonschema validate schema.json doc.json` → stdout is `{"valid":true}\n`
- Single invalid doc (default basic): `jsonschema validate schema.json invalid.json` → stdout starts with `{"valid":false,"errors":[`
- Single invalid doc (flag shape): `jsonschema validate -o flag schema.json invalid.json` → stdout is `{"valid":false}\n`
- Multiple docs: `jsonschema validate schema.json doc1.json doc2.json` → stdout has two NDJSON lines, each with {"document": path, "output": {...}} or {"document": path, "error": reason}
- Unknown shape: `jsonschema validate -o fancy schema.json doc.json` → exit code 2, no stdout, error on stderr
- Stdin single: `echo '{"x":1}' | jsonschema validate schema.json` → one output object
- Stdin multi: `echo '{"x":1}' | jsonschema validate schema.json - file.json` → NDJSON with two lines (stdin as "-", then file)
- No instance leak: Plant a recognizable secret string in the document; it does not appear in stdout, stderr or NDJSON lines
- Exit codes: valid doc → 0; invalid doc → 1; broken JSON or missing file → 2; worst result wins with multiple docs

## Migration

Library: Additive. New DefaultMaxOutputUnits constant, new Result.Output(shape) method, new Options.MaxOutputUnits field. Existing code continues to work. CLI: Breaking. Single invalid documents now print the full 2020-12 basic output (`{"valid": false, "errors": [...]}`) instead of the minimal `{"valid": false}`. Scripts expecting the old format must be updated to parse the new shape. Multi-document support is new; existing single-document scripts are unaffected except for the output change.

## Rollback

This change has no database schema migrations and adds no dependencies (go.mod is unchanged). Rollback is simply reverting the commits. The compiled schema structure and validation behaviour are fully backward-compatible at the evaluation level; only the output format and CLI interface have changed.

## Known limitations

- absoluteKeywordLocation is always omitted from output units (story 186 will add it via $id/$ref base URI resolution).
- detailed and verbose output shapes are not available (story 194).
- Annotations are not included in any output shape (story 191 will add annotation collection).
- No CLI flag to raise the output-unit cap beyond the library option (story 193 will add limit-setting flags).
- The regression test for the output-unit cap uses a test-only x-test-passing applicator (story 185 will provide real anyOf/oneOf).

## Concerns

- Breaking change in CLI stdout: Single invalid documents now produce the full 2020-12 basic output instead of the minimal result. This is an accepted breaking change (documented in the architecture), but scripts that parse old output must be updated.
- Per-document error messages are built directly from Go's err.Error() without filtering. File system paths or other system details may appear in stderr (e.g., 'open /path/to/file: permission denied'). The tester approved this as acceptable.
- Output-unit cap memory usage: The 10M default cap, if fully consumed, could accumulate millions of units in memory on very large schemas. This is generous for legitimate use but worth monitoring if extremely large documents are validated.
- Future applicators (story 185) must use the same EvalContext.EvalChild() method to evaluate child schemas, or the output-unit cap could be bypassed. The regression test ensures discarded units are counted; story 185 must maintain this contract.
