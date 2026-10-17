# Release notes — #23: Record evaluated properties and items only when something can use the record

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #23
- Status: approved by the operator at Checkpoint 3 (run #159, 2026-10-17 12:37 UTC)

## What changed

Schema validation now skips recording evaluated properties and items when no unevaluatedProperties or unevaluatedItems keyword can reach the evaluator. This internal optimization saves memory (15–28% on validating documents against the order schema) with no change to any result: same valid/invalid answers, same flag/basic/detailed/verbose reports, same library Result. The decision is made once at Compile and applies to every subsequent check on that schema.

- internal/compile: added NeedsEvaluated static analysis that scans the schema at Compile time for reachability of unevaluatedProperties and unevaluatedItems through $ref and every registry resource a $dynamicRef could land on
- internal/compile/compile.go: added NoUnevaluated boolean field to root Node, set by the analysis
- internal/eval: replaced EvaluateLazy call in Schema.Validate with EvaluateValidate, which sets a per-check skip flag in the evaluation context when NoUnevaluated is true and the check needs no report shape that shows the evaluated record
- internal/eval/evaluated.go, evaluated_items.go: guarded RecordProperties and all recording helpers with early return when skip is true; step charging and all limit counters stay outside the guard
- internal/eval: removed dead nodeLoc and kwLoc types; replaced their tests with unit tests for the live lazyLoc implementation checking locations through $ref, $dynamicRef, if/then/else, ~0/~1 escaping and empty absolute locations
- docs/benchmarks.md: added story-200 benchmark section with before/after measurements showing 17–28% drop in B/op and 1–2% in allocs/op on flag/basic validation, and no scenario rising

## How to verify

- Run `go test -race ./...` to confirm all tests pass and race detector finds no data races on schemas that skip and keep recording
- Compile a schema with no unevaluatedProperties or unevaluatedItems and one with them: both validate the same documents to the same results in all report shapes (flag, basic, detailed, verbose)
- Run `go test -run '^$' -bench . -benchmem . ./internal/jsonvalue ./cmd/jsonschema` and observe lower B/op and allocs/op on Validate*, ValidateErrors* and CLI benchmarks compared to the baseline (documented in docs/benchmarks.md)
- Verify that all limit tests pass on schemas with and without unevaluatedProperties/Items keywords, using the same boundary inputs: nesting depth, ref chain, evaluation steps, output units, report bytes, and annotation count

## Migration

None. This is a transparent internal optimization with no API or behavioral changes.

## Rollback

Revert to the previous commit. The change only affects internal evaluation state and introduces no persistent schema changes.

## Known limitations

- The load-time reachability decision is conservative: if a schema has any external $ref or $dynamicRef, all registry resources are scanned (the cautious rule). A schema that could theoretically reach no unevaluated keywords may be marked as needing recording if the registry contains one with such keywords, but this errs toward correctness and avoids missed answers.
- The decision is whole-schema, not per-rule: if unevaluatedProperties or unevaluatedItems appears anywhere the schema can reach, recording stays on for the entire check. A finer-grained decision is a future optimization.
- Verbose output still evaluates the document twice (once with details, once with the full tree), so it records in both passes. This was not changed by this story and is noted as a future improvement.

## Concerns

- The skip switch is correctly fail-safe: its zero value is false (recording on), so only EvaluateValidate sets it. Tree and lazy evaluation paths always record, preserving all outputs.
- The load-time analysis recurses depth-first over the already-bounded schema and allocates nothing (no heap stack), so CompileOrder allocs/op stays at the baseline.
- The old location.go tests for dead nodeLoc/kwLoc types were replaced by lazyloc_test.go, which tests the live lazyLoc implementation through the real evaluator. This is a legitimate test upgrade per AC6 (remove unused types and test lazyLoc); the new test covers all required cases.
- Benchmark improvements are primarily in memory (B/op and allocs/op), which was the goal. Speed improvements (ns/op) are within the ±5% noise band and are recorded for reference only.
