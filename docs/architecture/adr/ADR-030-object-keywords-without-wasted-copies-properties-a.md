# ADR #30: Object keywords without wasted copies (properties, additionalProperties, patternProperties)

- Date: 2026-10-17
- Story: #30
- Status: passed gate 2 with nothing for the operator to decide (run #166, 2026-10-17 22:13 UTC)

## Decision
The change is confined to the applicator vocabulary package and two small engine seams.
- A new helper (applicator/membernames.go) returns sorted member names into a caller-provided [8]string stack array and falls back to a heap slice above 8. It sorts with sort.Strings semantics (byte order), once per object, and the three keywords share it.
- Recording() is added as a method on EvalContext in internal/keywords/recording.go. Go allows methods in any file of the package, so registry.go (#206) is not edited. Recording() delegates to a read-only accessor on the per-call evaluated tracker in internal/eval/evaluated.go.
- annotate.go gains a copy-free variant. The existing API and its tests remain.
- The keyword location for the evaluated record is rendered lazily, only inside the Recording() branch.
- New tests and limits go in new files. Before/after numbers go in docs/benchmarks/r3-207-objectkeys.md.

## Affected modules
- internal/keywords/applicator/membernames.go
- internal/keywords/applicator/membernames_test.go
- internal/keywords/applicator/properties.go
- internal/keywords/applicator/additionalproperties.go
- internal/keywords/applicator/patternproperties.go
- internal/keywords/recording.go
- internal/keywords/recording_objectkeys_test.go
- internal/keywords/annotate.go
- internal/eval/evaluated.go
- objectkeys_boundary_test.go
- allocs_objectkeys_test.go
- docs/benchmarks/r3-207-objectkeys.md

## Data / API impact
- DB impact: no
- API impact: no
- Migration needed: no

## Constraints for implementation
- Sort order must equal sort.Strings (byte order) exactly. Stack buffer for 8 or fewer members, heap only above 8.
- Never write internal/keywords/registry.go, childctx*, internal/keywords/validation/*, internal/jsonvalue/number_inline*, internal/registry/*, internal/compile/metacache.go, allocs_test.go or docs/benchmarks.md.
- Never edit testdata/golden/*. If golden output differs, stop as failed.
- Copy-free annotation must not alias a buffer that is reused or mutated later. The stack buffer must not escape into a retained annotation. If escape analysis forces a heap move, record it rather than change semantics.
- unevaluatedProperties and unevaluatedItems must keep receiving every evaluated record whenever Recording() is true.
- Standard library only. go.mod unchanged. No existing test changed, removed or weakened.
- Timings are reported only, never asserted.

## Risks
- Aliasing: annotating without a copy could expose the stack buffer or a slice later mutated, changing annotation values. Mitigation: annotate from the stable slice only, and the boundary test compares annotation values in all four shapes.
- Escape analysis may move the [8]string buffer to the heap when it is passed through interfaces, reducing the saving. Verify with -gcflags=-m and AllocsPerRun.
- Recording() could be false in a path where unevaluated* actually needs the record (through $ref or $dynamicRef). The unevaluated* tests and the new recording test cover this.
- Benchmark timings are noisy during the parallel build. B/op and allocs/op are exact.
- The saving target may be missed. The story then records the numbers and is held for the operator.

## Breaking changes
- none

## Sensitivity
- none
