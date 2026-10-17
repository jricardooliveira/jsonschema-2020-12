# ADR #24: Collect and sort annotations only when they can be shown, without re-splitting JSON Pointers

- Date: 2026-10-17
- Story: #24
- Status: proposed (pending architecture sign-off)

## Decision
New logic goes into new files in internal/eval: annotation_order.go for segment-based sorting, annotation_record.go for compact records and on-demand building, and annotation_sink.go for discard-mode evaluation. Two crowded files, internal/eval/eval.go and jsonschema.go, are touched only to wire these in: one switch to the sink in eval.go, and a lazy field and accessor guard on Result where the Result type is defined. The annotation-count limit stays where it is charged today, at record time, so it fires during Validate at the same inputs. Building on demand reads only data owned by the Result. It never reads pooled or per-call evaluation state, which keeps the Result immutable and race-free. Discard mode affects only annotation recording. Step charging, depth, ref-chain and output-unit accounting run unchanged.

## Affected modules
- internal/eval/annotations.go
- internal/eval/annotation_order.go
- internal/eval/annotation_order_test.go
- internal/eval/annotation_record.go
- internal/eval/annotation_sink.go
- internal/eval/annotation_sink_test.go
- internal/eval/eval.go
- internal/eval/units.go
- annotations.go
- jsonschema.go
- annotations_ondemand_test.go
- cmd/jsonschema/annotations_limits_regression_test.go
- docs/benchmarks.md

## Data / API impact
- DB impact: no
- API impact: no
- Migration needed: no

## Constraints for implementation
- No new library option, CLI flag or exported identifier.
- Annotation content and order must be byte-identical in the Result and in every output shape.
- The annotation-count limit must be charged when an annotation is recorded, not when it is built.
- Discard mode must not change evaluation-step charging or any other limit counter.
- Building on demand must not touch pooled or reused per-call evaluation state, and the Result must be safe for concurrent reads.
- Add new code in new files. Change eval.go and jsonschema.go only to wire the new code in.
- Do not add or change any allocs_test.go or allocation-limit test, and leave internal/jsonvalue/number_test.go untouched.
- Use the standard library only, with no timing assertions in tests.

## Risks
- Skipping collection in a subtree whose annotations a shape does show would change a report. Verbose and detailed may expose annotations from failing subschemas, so the skip rule must be checked against internal/output before any subtree is discarded.
- Lazy building could hold references into per-call state that is later reused, which would cause data races or wrong content.
- If the order schema produces few annotations, the measurable gain on the order scenarios may be small. Then sorting and deferred building are the main contributors, and AC7 depends on them.
- eval.go is also changed by stories 199 and 200, so merges and rebases on it are likely.

## Breaking changes
- none

## Sensitivity
- security
