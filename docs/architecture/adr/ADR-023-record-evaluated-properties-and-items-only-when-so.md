# ADR #23: Record evaluated properties and items only when something can use the record

- Date: 2026-10-17
- Story: #23
- Status: proposed (pending architecture sign-off)

## Decision
The load-time decision is one immutable boolean on the compiled Schema. It is computed by a new static analysis in internal/compile/evaluated_need.go, which walks subschemas and $ref targets and treats every registry resource holding a $dynamicAnchor that a $dynamicRef in reach could resolve to as reachable. That makes it cautious, so the analysis can only add recording and never remove it wrongly. A visited set makes it terminate on cycles. At the start of each check, eval.go sets a per-call context field 'trackEvaluated' to the Schema's flag OR the needs of the requested shape (detailed or verbose, or a Result view that exposes the record). The recording helpers in evaluated.go and evaluated_items.go return early when that field is false. Keyword files under internal/keywords stay untouched, and no shared state is written after Compile. Limit counters and step charging stay outside the guarded code paths, so limits trigger identically. The crowded files compile.go and eval.go each get only wiring lines. The location.go cleanup is a separate, independent deletion plus a new test file.

## Affected modules
- internal/compile/evaluated_need.go
- internal/compile/evaluated_need_test.go
- internal/compile/compile.go
- internal/eval/eval.go
- internal/eval/evaluated.go
- internal/eval/evaluated_items.go
- internal/eval/evaluated_skip_test.go
- internal/eval/location.go
- internal/eval/location_test.go
- internal/eval/lazyloc_test.go
- jsonschema.go
- evaluated_tracking_limits_test.go
- evaluated_tracking_race_test.go
- docs/benchmarks.md

## Data / API impact
- DB impact: no
- API impact: no
- Migration needed: no

## Constraints for implementation
- Never write to the compiled Schema or anything reachable from it after Compile returns. The switch lives only in the per-call context.
- Reachability through $dynamicRef counts every registry resource it could land on (operator choice: cautious).
- The decision covers the whole schema, not individual rules (operator choice).
- Evaluation-step charging, depth, ref-chain, output-unit, report-byte and annotation-count accounting must not move inside the skipped code.
- Do not edit keyword files under internal/keywords. Guard inside the eval recording helpers.
- Add no allocation-limit test, and leave internal/jsonvalue/number_test.go unchanged.
- No timing assertions in ordinary tests. ns/op is recorded only in docs/benchmarks.md.
- Standard library only, with no go.mod change.

## Risks
- The analysis could miss a path by which an unevaluated keyword is reached, such as a $dynamicRef into a registry resource, an embedded resource with its own $id, or a meta-schema-driven case. That would change an answer, so the analysis errs toward recording, and tests cover each path.
- Detailed or verbose output, or a Result accessor, may read the evaluated record directly. If the check-time need is not detected, reports would differ, so the existing output tests act as the guard.
- Early returns in the recording helpers could change step charging or annotation counting if those are interleaved with recording today.
- Benchmark noise on ns/op is ±5%, so success is judged on allocs/op and B/op.

## Breaking changes
- none

## Sensitivity
- security
