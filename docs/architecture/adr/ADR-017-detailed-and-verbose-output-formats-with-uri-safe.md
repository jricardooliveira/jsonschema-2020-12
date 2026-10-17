# ADR #17: Detailed and verbose output formats, with URI-safe absoluteKeywordLocation

- Date: 2026-10-17
- Story: #17
- Status: proposed (pending architecture sign-off)

## Decision
The new shapes are new files in internal/output that register through the existing output registry, the same way flag.go and basic.go do. The library's render path and the CLI output-format option resolve shapes by name, so neither jsonschema.go nor cmd/jsonschema/run.go nor main.go should need editing (read-only claims). A shared builder, internal/output/tree.go, turns the per-call evaluation units and annotations into the 2020-12 unit tree and enforces the existing DefaultMaxReportBytes cap while building. The only edit to a crowded file is a one-line call in internal/compile/compile.go (Resource.AbsoluteLocation) to a new helper in internal/compile/fragment.go. eval.go is not touched. If verbose needs passing units that are not recorded today, the change goes into internal/eval/units.go, which is not crowded.

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
- API impact: yes: two new output-format values; absoluteKeywordLocation fragment encoding changes in all shapes
- Migration needed: no

## Constraints for implementation
- Standard library only; go.mod unchanged
- Each shape is its own file registered from init(); no central list or switch edit
- Never emit instance values; property names only where an annotation lists matched names
- Enforce DefaultMaxReportBytes during building, not after building the whole tree in memory
- Do not edit internal/eval/eval.go; touch internal/compile/compile.go only for the one-line call
- Tests read limits and defaults from code constants
- Compiled Schema stays immutable; per-call state only (go test -race)

## Risks
- Encoding the fragment changes absoluteKeywordLocation for schemas with special characters in keys. Existing tests that assert the unencoded form will need updating.
- Verbose may need passing units that the evaluator does not currently record, which would grow the per-call memory cost.
- Verbose output for large documents can hit the report cap. That is intended, but users may need the raise flag.

## Breaking changes
- absoluteKeywordLocation for schema keys with spaces, '%' or non-ASCII characters is now percent-encoded in flag/basic output too

## Sensitivity
- security
- pii
