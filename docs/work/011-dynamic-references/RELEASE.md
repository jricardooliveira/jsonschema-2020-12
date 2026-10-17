# Release notes — #11: $dynamicRef and $dynamicAnchor with dynamic scope

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #11
- Status: proposed (pending Checkpoint 3 sign-off)

## What changed

Schemas using $dynamicRef and $dynamicAnchor with dynamic scope can now be validated correctly, including recursive extension and official meta-schemas. $dynamicRef resolves its initial target like $ref, but when that target carries a matching $dynamicAnchor, it uses the outermost matching $dynamicAnchor in the dynamic scope instead. The dynamic scope is per-call and safe for concurrent use.

- internal/registry/dynamicanchor.go, dynamicanchor_test.go: $dynamicAnchor indexing in the registry, sharing the plain-name namespace with $anchor; duplicate names (dynamic+dynamic, dynamic+static) are schema errors
- internal/eval/dynamicscope.go, dynamicscope_test.go, dynamicscope_race_test.go: per-call dynamic scope stack with push/pop and outermost-match lookup; verified thread-safe under -race
- internal/keywords/core/dynamicanchor.go: $dynamicAnchor keyword (stores anchor name for dynamic lookup)
- internal/keywords/core/dynamicref.go, dynamicref_test.go, dynamicref_spec_test.go: $dynamicRef keyword with static resolution, bookending, and dynamic scope resolution; uses story 186's loop check
- internal/compile/compile.go: Resource struct (URI, Pointer, Dynamic map) to track compiled $dynamicAnchor subschemas per resource
- internal/compile/refs.go: extended compilation to fill Resource.Dynamic tables after compile returns
- internal/eval/eval.go: added viaRef tracking and absoluteKeywordLocation reporting when evaluation crosses a reference
- internal/eval/refguard.go: integrated dynamic scope push/pop at node entry/exit
- internal/keywords/registry.go, internal/output/basic.go: added AbsoluteKeywordLocation field to error units and basic output
- testdata/core/dynamicRef.json, dynamicAnchor.json: spec-derived test cases covering outermost resolution, bookending, embedded resources, registry documents, duplicate names, unresolved targets, and recursive extension
- cmd/jsonschema/dynamicref_cli_test.go, cmd/jsonschema/testdata/dynamicref/: CLI contract tests for exit codes (2 on schema error/cycle, 0/1 on valid/invalid), output location names, and recursive extension

## How to verify

- Compile a recursive tree schema with `$dynamicAnchor: "node"` and `items: {$dynamicRef: "#node"}`, then validate nested tree objects; they pass if type constraints are met
- Extend the tree schema with `$ref: <tree-uri>` and add `required: ["data"]` at the root; validate: objects at every nesting level require 'data'. Validate the plain tree separately; same data values pass without 'data'
- Create a schema with `$dynamicRef: "#T"` pointing to a number type, validate a string value, and check basic output: keywordLocation contains `/$dynamicRef`, no instance value is quoted, and absoluteKeywordLocation shows the resolved target's URI
- Compile schemas with duplicate `$dynamicAnchor: "name"` in one resource (or dynamic+static anchor name): expect schema error exit code 2
- Compile unresolved `$dynamicRef: "#missing"`: expect schema error exit code 2
- Create a cycle with `$dynamicRef` that never descends and validate: expect loop error (not valid/invalid), exit code 2, no hang (test includes 10-second timeout)

## Migration

none

## Rollback

Remove internal/keywords/core/dynamicanchor.go and dynamicref.go, internal/registry/dynamicanchor.go, and internal/eval/dynamicscope.go. Revert dynamic scope push/pop and viaRef tracking from eval.go and refguard.go. Revert Resource struct and AbsoluteLocation method from compile.go. No data migration needed; this is a new feature with no mutation of existing keywords.

## Known limitations

- $dynamicRef recursion is limited only by story 186's evaluation-time loop detection; story 193 adds a configurable limit
- unevaluatedProperties and unevaluatedItems do not yet work through $dynamicRef (story 192)
- Meta-schema validation and embedded meta-schemas (story 189) not yet implemented
- $recursiveRef and $recursiveAnchor from 2019-09 are not supported (2020-12 only)

## Concerns

- Fragment encoding (low hardening note): absoluteKeywordLocation fragments are built from JSON Pointers without additional percent-encoding. While RFC 3986 allows unescaped '/' and '~' in fragments, property names containing literal '%' would not be percent-encoded (e.g., '#/foo%bar' instead of '#/foo%25bar'). This is unlikely in practice and does not affect correctness; mentioned for future hardening review.
