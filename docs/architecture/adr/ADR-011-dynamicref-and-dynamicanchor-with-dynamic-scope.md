# ADR #11: $dynamicRef and $dynamicAnchor with dynamic scope

- Date: 2026-10-16
- Story: #11
- Status: proposed (pending architecture sign-off)

## Decision
Dynamic scope is per-evaluation state, so it lives in a new file internal/eval/dynamicscope.go. The crowded internal/eval/eval.go is touched only to push and pop the scope where evaluation enters and leaves a schema resource. Static resolution of the initial target, and the 'has matching dynamic anchor' fact, are computed at compile time and stored immutably. The evaluation step only walks the per-call scope stack, outermost first, and looks up the precomputed per-resource dynamic-anchor table. $dynamicAnchor names share the plain-name namespace with $anchor in the registry, which gives one duplicate check. $dynamicRef reuses story 186's $ref resolution and evaluation-time loop check rather than adding its own, and story 193 later adds the configurable limit on top. The keywords are new files registered through init() into internal/keywords/registry.go, and the core vocabulary is already blank-imported in all.go, so no central list is edited.

## Affected modules
- internal/registry/dynamicanchor.go
- internal/registry/dynamicanchor_test.go
- internal/registry/anchor.go
- internal/eval/dynamicscope.go
- internal/eval/dynamicscope_test.go
- internal/eval/dynamicscope_race_test.go
- internal/eval/eval.go
- internal/keywords/core/dynamicanchor.go
- internal/keywords/core/dynamicref.go
- internal/keywords/core/dynamicref_test.go
- internal/keywords/core/dynamicref_spec_test.go
- testdata/core/dynamicRef.json
- testdata/core/dynamicAnchor.json
- cmd/jsonschema/dynamicref_cli_test.go
- cmd/jsonschema/testdata/dynamicref/tree.json
- cmd/jsonschema/testdata/dynamicref/strict-tree.json
- cmd/jsonschema/testdata/dynamicref/duplicate-anchor.json
- cmd/jsonschema/testdata/dynamicref/cycle.json
- cmd/jsonschema/testdata/dynamicref/unresolved.json

## Data / API impact
- DB impact: no
- API impact: no
- Migration needed: no

## Constraints for implementation
- No edit to a central keyword list or switch. Register through init() only.
- The compiled Schema must not be mutated during evaluation. The dynamic scope lives only in the per-call evaluation context.
- $dynamicRef must use story 186's evaluation-time loop detection. Do not add a separate counter or a configurable limit (story 193).
- Touch internal/eval/eval.go only to wire scope push/pop. Put the logic in dynamicscope.go.
- No network or disk access. Targets come only from the compiled document and the caller-supplied registry.
- Reports never include instance values.
- Each test case cites a 2020-12 core section and is written by this project, never copied from the JSON-Schema-Test-Suite.

## Risks
- The point where eval.go enters a schema resource may be spread across $ref handling and embedded-$id handling. Missing one site gives a wrong dynamic scope.
- Reusing story 186's loop key (schema location + instance location) for $dynamicRef needs the dynamically resolved location as the key, not the static one. Otherwise real cycles could slip through or legitimate re-entries could be flagged.
- The location of story 186's anchor index and duplicate check is unknown, so the duplicate-name check may need a small edit in an existing registry file.

## Breaking changes
- none

## Sensitivity
- security
