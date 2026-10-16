# Release notes — #3: Error reports: flag and basic output, multiple documents

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #3
- Status: proposed (pending Checkpoint 3 sign-off)

**⚠️ the release-agent has concerns.**

## What changed

The validator now produces 2020-12 'flag' and 'basic' output shapes as JSON for both the library and CLI. Single documents produce one standard output object; multiple documents produce one JSON report per line, each labelled with its file path. Error reports include instance and keyword locations and rule names, never instance values. The CLI accepts -o/--output to choose the shape (default: basic) and multiple document paths, with exit code 0 for valid, 1 for invalid, and 2 for errors.

- internal/eval/units.go: output unit collection with per-call counter that caps EVERY appended unit, including discarded children
- internal/eval/eval.go: keywords now report failures through ctx.Report() with value-free error sentences; Evaluate renamed to EvaluateUnits internally
- internal/keywords/registry.go: EvalContext gained InstanceLocation, KeywordLocation, Report(), Keep(), EvalChild() and Units() methods; EvaluateFunc signature unchanged
- internal/compile/compile.go: each compiled node and item now carries its JSON Pointer location from schema root
- internal/keywords/validation/{type,enum,const}.go: keywords report failures with rule sentences instead of silent false returns
- internal/output/: new package with single registry (registry.go), self-registering flag (flag.go) and basic (basic.go) shapes; flag is {valid}, basic is {valid:true} or {valid:false, errors:[...]}
- jsonschema.go: Result gained units field; Schema gained output-unit cap; Validate returns units via EvaluateUnits()
- numberlimits.go: Options gained MaxOutputUnits field (0 means DefaultMaxOutputUnits)
- output.go: public API Result.Output(shape string) renders to bytes; DefaultMaxOutputUnits exported
- cmd/jsonschema/run.go: added -o/--output flag (default: basic), removed single-document-only restriction, implemented multi-document NDJSON output, worst-result exit codes, per-document error lines, and one-time schema error reporting on stderr
- cmd/jsonschema/run_test.go: removed 'two docs' usage error test (now allowed); widened assertion from map[string]bool to map[string]any to accept basic's errors array

## How to verify

- Single document, default basic shape: `jsonschema validate schema.json doc.json` outputs exactly one JSON object {"valid":...} or {"valid":false,"errors":[...]} and newline
- Single document, flag shape: `jsonschema validate -o flag schema.json doc.json` outputs exactly {"valid":true|false} and newline
- Multiple documents: `jsonschema validate schema.json doc1.json doc2.json` outputs NDJSON, each line is {"document":"...","output":{...}} or {"document":"...","error":"..."}, in argument order
- Stdin, explicit: `echo '1' | jsonschema validate schema.json -` treats - as stdin
- Stdin, implicit: `echo '1' | jsonschema validate schema.json` with no document argument reads stdin
- Multiple with stdin twice: `jsonschema validate schema.json doc.json - - < input.json` does not hang; second - gives {"document":"-","error":"stdin already consumed"}
- Unknown shape: `jsonschema validate -o fancy schema.json doc.json` exits code 2, prints nothing on stdout
- Schema error with multiple docs: `jsonschema validate broken-schema.json doc1.json doc2.json` outputs one error line on stderr, exit code 2, stdout empty
- Confirm no leaks: `jsonschema validate schema.json doc-with-secret.json 2>&1 | grep -i secret` returns empty (secret value never in stdout or stderr)
- Output-unit cap exceeded: `jq 'limit(1;range(100000))|.x'` as instance against {"properties":{...}} with cap lowered to <100 gives exit code 2 and LimitError message on stderr
- Exit codes: 0 if all valid, 1 if any invalid, 2 if any cannot be read or schema is broken

## Migration

The CLI's default output format for invalid documents changed from minimal {"valid":false} to the 2020-12 'basic' shape {"valid":false,"errors":[...]}. Scripts and test harnesses that parse the old minimal format must be updated. The error units give instance and keyword locations as JSON Pointers, with each failure from both the keyword itself and all wrapping applicators that failed because of it. If you had worked around this by parsing, migrate to the standard structure. The library API is backward compatible: Compile returns the same type, and new callers of Result.Output() get the new shapes; existing code that only calls Validate() sees no change except that Result now holds units (which were previously inaccessible).

## Rollback

Revert cmd/jsonschema/run.go to single-document-only mode: accept only zero or one document path (not a list), use the old minimal {"valid":...} output, and remove the -o/--output flag. Revert internal/eval/eval.go and internal/output/ to drop the units collection and shape rendering. Revert jsonschema.go to return Result{valid:ok} without units. The compiled schema compilation (internal/compile) can keep the location tracking; it does not break anything that does not use it.

## Known limitations

- absoluteKeywordLocation is omitted in all error units. It will be added when $id/$ref and base URI resolution arrive (stories 186, 187).
- Detailed and verbose output shapes are not included; story 194 will add them.
- Annotations (titles, descriptions, format names, evaluated property names) are not included in basic shape output; story 191 will add them.
- The CLI has no flag to raise the output-unit cap; story 193 will add limit flags.
- The regression test for the output-unit cap uses a test-only passing applicator because anyOf/oneOf do not exist yet (story 185). Once story 185 lands, replace it with a test using real anyOf.
- The CLI does not validate that the schema itself uses only known vocabularies (that check comes with the meta-schema enforcement story).

## Concerns

- The build gate does not report that tests were executed with go test -race, so there is no proof that the concurrent Validate/Output access is race-free. The code passes local -race testing, but verify the factory's build process also runs -race.
- There is no CLI contract test that verifies a LimitError from the output-unit cap results in exit code 2 (and stderr output). The library test in output_test.go covers this, but the CLI path is untested.
- Keyword locations in error units are not rooted by a base URI (story 186/187). This is intentional and documented, but it means locations are always relative to the schema root. Consumers must not assume an absolute keyword location until those stories arrive.
