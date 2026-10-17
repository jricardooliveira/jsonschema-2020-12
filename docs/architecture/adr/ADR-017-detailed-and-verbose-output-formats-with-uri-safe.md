# ADR #17: Detailed and verbose output formats, with URI-safe absoluteKeywordLocation

- Date: 2026-10-17
- Story: #17
- Status: proposed (pending architecture sign-off)

## Decision
The design is unchanged except for one point the reviewer settled: the report byte cap. This story uses only the existing library option Options.MaxReportBytes (numberlimits.go) and the existing constant DefaultMaxReportBytes (internal/output/registry.go, re-exported in output.go). It adds no CLI flag for the cap and does not edit cmd/jsonschema/run.go. Story 193 owns that flag and its 'raise the cap through the CLI' test. Acceptance criterion 6 is met here by library-level tests: the cap is enforced for detailed and verbose, and raising Options.MaxReportBytes lets a large report through. The CLI tests in this story only check that going over the default cap gives exit 2 with an empty stdout. The percent-encoding breaking change is accepted.

The rest is as reviewed:
- The new shapes are new files in internal/output. They register through the existing output registry with RegisterReport and render through RenderReport, which already applies the cap through the Budget writer. This is the same way basic.go works.
- The library's render path (Result.Output) and the CLI output-format option look shapes up by name, so jsonschema.go, cmd/jsonschema/run.go and main.go are only read, not edited.
- A shared builder, internal/output/tree.go, turns the per-call evaluation units and annotations into the 2020-12 unit tree. It writes each unit through the Budget as it goes, so the cap is enforced while building.
- The only edit to a crowded file is a one-line call in Resource.AbsoluteLocation (internal/compile/compile.go) to a new helper in internal/compile/fragment.go.
- eval.go is not touched. If verbose needs passing units that are not recorded today, the change goes into internal/eval/units.go, which is not crowded.

## Affected modules
- internal/compile/compile.go
- internal/compile/fragment.go
- internal/compile/fragment_test.go
- internal/output/tree.go
- internal/output/nested_limit_test.go
- internal/output/detailed.go
- internal/output/detailed_test.go
- internal/output/verbose.go
- internal/output/verbose_test.go
- internal/eval/units.go
- internal/eval/units_verbose_test.go
- output_nested_test.go
- cmd/jsonschema/output_nested_test.go
- cmd/jsonschema/abslocation_escape_test.go

## Data / API impact
- DB impact: no
- API impact: yes
- Migration needed: no

## Constraints for implementation
- Standard library only; go.mod unchanged
- Each shape is its own file, registered from init() with output.RegisterReport; no edit to a central list or switch
- Never emit instance values; property names appear only where an annotation lists matched names
- Enforce the cap while building, by writing through the existing output.Budget, not after building the whole tree in memory
- Do not edit internal/eval/eval.go; touch internal/compile/compile.go only for the one-line call
- Do not edit cmd/jsonschema/run.go and do not add a report-bytes CLI flag (story 193 owns it)
- Cap tests use DefaultMaxReportBytes and Options.MaxReportBytes from the code instead of repeating the number; the 'raise the cap' test is library-level (output_nested_test.go, internal/output/nested_limit_test.go)
- The compiled Schema stays immutable; all state is per call (go test -race)

## Risks
- Encoding the fragment changes absoluteKeywordLocation for schemas with special characters in their keys. Existing tests that assert the unencoded form will need updating.
- Verbose may need passing units that the evaluator does not record today, which would raise the memory cost of each call.
- Verbose output for large documents can hit the report cap. That is intended. Until story 193 adds the CLI flag, CLI users cannot raise the cap; only library callers can, through Options.MaxReportBytes.

## Breaking changes
- absoluteKeywordLocation for schema keys that contain spaces, '%' or non-ASCII characters is now percent-encoded, in flag and basic output too (accepted by the reviewer)

## Sensitivity
- security
- pii
