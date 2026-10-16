# ADR #3: Error reports: flag and basic output, multiple documents

- Date: 2026-10-16
- Story: #3
- Status: approved by the operator at Checkpoint 2 (run #137, 2026-10-16 19:02 UTC)

## Decision
Keep the approved design of run #136. The engine collects failure units in a new file, internal/eval/units.go. Each unit holds instanceLocation, keywordLocation, an optional absoluteKeywordLocation and a value-free error sentence. A per-call collector in the evaluation context holds them, so the compiled schema stays immutable. Applicators evaluate children through an EvalContext helper that returns the child's units. The parent keeps them, in a fixed order with the child units before the wrapping applicator's unit, or discards them when the applicator passes. The collector increments its counter on EVERY append, kept or later discarded. When the counter exceeds the cap, the collector returns a LimitError naming 'output units', which propagates as an error. internal/eval/eval.go changes only to create the collector and route keyword failures into it. The existing Evaluate stays as a thin wrapper so current callers and tests keep working. internal/keywords/registry.go gains only the reporting and child-evaluation methods on EvalContext, and the EvaluateFunc signature stays unchanged so dispatch_test.go and other keyword stories are not broken. internal/compile/compile.go records each node's keyword location (a JSON Pointer from the schema root) so units can carry it. internal/output holds the single shape registry (Register/Lookup) and one file per shape (flag.go, basic.go), each registering from init(). The root package adds output.go with Result.Output(shape) and DefaultMaxOutputUnits. jsonschema.go gains only the units field on Result and passes the cap into evaluation, and numberlimits.go gains the MaxOutputUnits option field. cmd/jsonschema/run.go holds the CLI logic: the -o/--output flag, positional document list ('-' for stdin), single-object stdout for one document, NDJSON {document, output|error} lines for several documents, schema errors reported once on stderr, and the worst-result exit code. main.go is not edited.

## Affected modules
- internal/eval/units.go
- internal/eval/units_test.go
- internal/eval/eval.go
- internal/keywords/registry.go
- internal/compile/compile.go
- internal/keywords/validation/type.go
- internal/keywords/validation/enum.go
- internal/keywords/validation/const.go
- internal/output/registry.go
- internal/output/registry_test.go
- internal/output/flag.go
- internal/output/flag_test.go
- internal/output/basic.go
- internal/output/basic_test.go
- jsonschema.go
- output.go
- output_test.go
- numberlimits.go
- cmd/jsonschema/run.go
- cmd/jsonschema/output_test.go
- cmd/jsonschema/multidoc_test.go
- cmd/jsonschema/run_test.go

## Data / API impact
- DB impact: no
- API impact: yes: new Result.Output, DefaultMaxOutputUnits and Options.MaxOutputUnits; CLI gains -o/--output and multiple documents.
- Migration needed: no

## Constraints for implementation
- Standard library only; go.mod is not edited.
- No instance value may appear in any unit, error sentence, stderr reason or NDJSON line.
- The cap counts every appended unit, including units appended by children of passing applicators that are later discarded. It never uses the retained slice length.
- Per-evaluation state (collector, counter) lives in the per-call context; the compiled Schema is never mutated.
- EvaluateFunc's signature stays unchanged; new capabilities are methods on EvalContext.
- Tests read DefaultMaxOutputUnits and lower it through the option instead of repeating the number.
- Each feature's tests go in new test files; existing tests are edited only where the new stdout breaks them.

## Risks
- anyOf/oneOf do not exist yet (story 185), so the regression test must use a test-only applicator registered from the test file. Story 185 must use the same EvalContext child-evaluation helper, or the cap could be bypassed.
- eval.go, compile.go, registry.go, jsonschema.go and run.go are crowded files, so this story serialises with other stories that write them.
- Changing invalid-document stdout from the minimal result to basic may break exact-output assertions in run_test.go.

## Breaking changes
- CLI stdout for an invalid document changes from the minimal result to the 2020-12 'basic' object (the default).

## Sensitivity
- security
- pii
