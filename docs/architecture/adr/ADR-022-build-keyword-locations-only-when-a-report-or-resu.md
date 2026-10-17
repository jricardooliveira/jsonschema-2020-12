# ADR #22: Build keyword locations only when a report or Result shows them

- Date: 2026-10-17
- Story: #22
- Status: proposed (pending architecture sign-off)

## Decision
New code goes in internal/eval/location.go: a small, allocation-light location reference (parent pointer plus segment, and the base URI of the current resource). Strings are rendered only at the points where an error, output unit, annotation or Result entry is created. The crowded file internal/eval/eval.go is touched only to pass the reference instead of the string and to call render at those creation points. internal/eval/units.go and internal/eval/annotations.go render on creation. The compiled schema is not changed, so Compile allocations cannot rise. Per-call memoisation, if any, lives in the evaluation context. Verbose keeps its second pass, which also uses the lazy reference.

## Affected modules
- internal/eval/location.go
- internal/eval/location_test.go
- internal/eval/eval.go
- internal/eval/units.go
- internal/eval/annotations.go
- cmd/jsonschema/limits_edge_regression_test.go
- docs/benchmarks.md

## Data / API impact
- DB impact: no
- API impact: no
- Migration needed: no

## Constraints for implementation
- Rendered locations must match today's strings exactly, including JSON Pointer escaping (~0/~1), percent-encoding in absolute locations, $ref/$dynamicRef path segments and embedded-resource base URIs.
- Do not add fields or caches to the compiled Schema or any compile-time structure. Compile must not allocate more.
- Evaluation-step charging, output-unit counting, report-byte accounting and annotation counting must occur at the same points with the same amounts.
- Verbose keeps its second evaluation pass.
- No existing test file is edited, and no allocation-limit test is added or changed.
- Standard library only, with no change to go.mod.
- Benchmarks: 10 runs per scenario on the same machine, record the median (for example -count=10), using Go tools only.

## Risks
- Location strings may also be built inside keyword evaluate functions (internal/keywords/**) or in the keyword registry's evaluate signature, not only in internal/eval. In that case more files would need changing.
- A lazy reference that captures mutable per-call state could render the wrong path if it is evaluated after that state changes, for example dynamic scope in verbose's second pass.
- The report-bytes limit depends on rendered string sizes. Rendering at a different moment must not change when the limit trips.

## Breaking changes
- none

## Sensitivity
- security
