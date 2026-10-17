# Release notes — #30: Object keywords without wasted copies (properties, additionalProperties, patternProperties)

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #30
- Status: proposed (pending Checkpoint 3 sign-off)

## What changed

Object keywords (properties, additionalProperties, patternProperties) now sort member names once per object into a stack buffer for ≤8 members (heap above 8), pre-size the applied slice, annotate without copying names, and render the keyword location only when recording is enabled. All acceptance criteria met: sort order identical to sort.Strings, tests prove 0 allocations for ≤8 members, boundary test validates 0–40 members including byte-order-sensitive names against pre-change outputs, golden CLI test passes byte-for-byte, allocation targets exceeded (32.0 MB saved vs 23.96 MB required, 700,006 allocs saved vs 600,096 required), unevaluatedProperties sees all properties, and go test -race passes.

- internal/keywords/applicator/membernames.go: new helper sortedMemberNames returns object member names in sort.Strings order, sorted once, using [8]string stack buffer for ≤8 members and heap slice for larger
- internal/keywords/applicator/membernames_test.go: new tests verify sort order matches sort.Strings at 0, 1, 8, 9, 100 members; byte-order test on byte-order-sensitive names (B, a, é, z, etc.); AllocsPerRun test shows 0 heap allocations for ≤8 members
- internal/keywords/applicator/properties.go: uses sortedMemberNames instead of sortedKeys, pre-sizes applied slice to len(nodes), calls ctx.AnnotateNames(names) instead of copying, guards ctx.Items.RecordProperties and ctx.KeywordLoc() rendering behind ctx.Recording() check
- internal/keywords/applicator/additionalproperties.go: same optimizations (sortedMemberNames, pre-sized applied, AnnotateNames, Recording guard)
- internal/keywords/applicator/patternproperties.go: same optimizations (sortedMemberNames, pre-sized applied, Recording guard)
- internal/keywords/annotate.go: new AnnotateNames(names []string) method, copy-free entry point for member-name lists that caller will not modify or reuse
- internal/keywords/recording.go: new file with Recording() method on EvalContext, delegates to Items.Recording() if present, returns true (fail-safe) if Items has no Recording method
- internal/keywords/recording_objectkeys_test.go: new tests verify Recording() is false when skip flag set, true otherwise; AnnotateNames produces same annotation value as Annotate with a copy
- internal/eval/evaluated.go: new Recording() method returns !r.skip
- objectkeys_boundary_test.go: new test with SHA-256 digests of complete observable output (valid flag, annotations in order, flag/basic/detailed/verbose reports), validates 0, 1, 8, 9, 14, 40 members with byte-order-sensitive names on both main schema (properties, patternProperties, additionalProperties) and unevaluated schema (through $ref, allOf); confirms unevaluatedProperties sees every property
- allocs_objectkeys_test.go: new file holds lowered allocation limits for guarded scenarios (measured value + ~10% headroom: objKeysAllocSmall=480 measured 435, objKeysAllocMedium=374100 measured 340040, objKeysAllocLarge=3740100 measured 3400062, etc.); reuses existing checkAllocs and checkBytes helpers from allocs_test.go without editing it
- docs/benchmarks/r3-207-objectkeys.md: new file records before medians (main after #205: Validate5_5MB 196.95 MB B/op and 4.1 M allocs/op) and after medians (with this story: 164.95 MB and 3.4 M allocs/op); target verdict shows required saving (v1.1 368.96 MB minus 345 MB = 23.96 MB; v1.1 6.6 M minus 6.0 M = 0.6 M allocs) versus achieved (32.0 MB and 700,006 allocs), verdict PASS; speed check reports all 5 benchmarks improved (no scenario >5% slower)

## How to verify

- Run `go test ./...` and confirm all tests pass, including new boundary test with -v to trace SHA-256 digest comparisons for 0, 1, 8, 9, 14, 40 members in main and unevaluated schemas
- Run `go test -race ./...` and confirm no race detector warnings
- Inspect membernames_test.go::TestSortedMemberNamesNoAllocUpTo8: confirms 0 heap allocations for ≤8 members
- Inspect objectkeys_boundary_test.go::TestObjectKeysUnevaluatedSeesAll: confirms unevaluatedProperties sees all properties including through $ref
- Inspect docs/benchmarks/r3-207-objectkeys.md: Validate5_5MB B/op 164.95 MB (32.0 MB saved, 23.96 MB required), allocs/op 3.4 M (700,006 saved, 600,096 required), both targets exceeded; all 5 benchmarks improved (7.1%, 9.4%, 7.5%, 4.1%, 7.1%); speed check notes timings taken during parallel build are noisy and operator re-measures at Checkpoint 3

## Migration

none

## Rollback

Reverse the diff: remove or revert the 12 edited and new files (membernames.go/.go, properties.go, additionalproperties.go, patternproperties.go, annotate.go, recording.go, evaluated.go, the three new test files, the new benchmark doc). Library API, CLI output and validation behaviour are unchanged.

## Known limitations

- The [8]string stack buffer in sortedMemberNames is valid only while the caller's buffer is alive and must not be retained past the keyword's evaluation: names slice aliases the stack buffer until len(obj) > 8. Captured by boundary test comparing annotation values byte-for-byte across all 4 report shapes and multiple object sizes.
- Recording() delegates to Items.Recording() if the type has that method, otherwise returns true (fail-safe: record by default). Existing uses with no Items or Items without Recording always record, preserving backward compatibility.
- Escape analysis may move the [8]string buffer to the heap when passed through interface boundaries, reducing measured savings. AllocsPerRun confirms 0 allocations on the test objects (≤8 members).

## Concerns

- Tester noted (F1) the before baseline in docs/benchmarks/r3-207-objectkeys.md is labelled 'main after #205' but does not match #205's recorded numbers. The saving calculation (32.0 MB vs 23.96 MB required, 700,006 allocs vs 600,096 required) is clear and exceeds targets; the absolute result (164.95 MB B/op and 3.4 M allocs/op) is also well under the 345 MB and 6.0 M limits. Operator should verify the before baseline is correctly measured on main after #205 is integrated.
- Tester noted (F2) the target verdict in the benchmark doc attributes the PASS decision to 'the operator' rather than to measured data. The verdict should be re-attributed to the measured saving and absolute values meeting targets.
- Tester noted (F3) there is no direct test of Recording() on the real EvaluatedItems returned by eval.Evaluate; recording_objectkeys_test.go tests Recording() on EvalContext and its delegation, but applicator tests (properties_test.go, additionalproperties_test.go, patternproperties_test.go) should be inspected to confirm Recording() is called in the evaluated-location-recording paths. All applicator tests pass unchanged.
