# Release notes — #9: Core identification and referencing within one schema document: $id, $anchor, $defs, $ref, $comment

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #9
- Status: proposed (pending Checkpoint 3 sign-off)

**⚠️ the release-agent has concerns.**

## What changed

Schemas can now identify themselves and reference their own subschemas: $id sets the base URI (including for embedded resources), $anchor defines plain-name fragments, $ref resolves local and anchor references, $defs holds reusable schemas, and $comment is accepted. All references are validated at load time before any document is checked. Loop detection and depth limits protect against hostile schemas.

- internal/uri: RFC 3986 reference resolution, safe normalization for URI comparison, JSON Pointer parsing with percent-decoding and ~0/~1 unescaping
- internal/registry: Resource index with base URI tracking, normalized $id values indexed by resource, and per-resource anchor mapping; validates $id fragments and anchor syntax
- internal/keywords/core: Five new keywords ($id, $anchor, $defs, $comment, $ref); $ref compiles to a direct pointer to its target; $comment and $defs have no validation effect
- internal/compile/refs: Load-time reference resolution, root base URI selection (override > file path > stdin placeholder), and pure-cycle detection
- internal/eval/refguard: Per-call evaluation guard that detects (schema location, instance location) re-entry loops and enforces evaluation depth limits
- Library API: Options.BaseURI override and Options.SchemaPath for file-based base URI; CLI: --base-uri flag on 'jsonschema validate'
- testdata/core: Six spec-derived JSON test suites for $id, $anchor, $defs, $ref, $comment, and URI normalization, each citing 2020-12 core sections
- docs/references.md: Design decisions on base URI selection, RFC 3986 normalization before comparison, non-schema targets, and loop handling

## How to verify

- go test ./... passes all tests
- go test -race ./internal/eval/... detects no race conditions in the loop guard
- A schema with relative $ref resolves correctly when --base-uri sets the context, and fails to resolve without it
- An unresolved $ref in an unreached branch (e.g. inside a non-matching property) causes load-time exit code 2
- A schema with a pure reference cycle ($ref pointing at itself) causes load-time exit code 2
- A recursive tree schema with properties containing $ref to the root validates correctly on nested documents without hanging or looping

## Migration

No migration required. Schemas without $id, $anchor, or $ref are unaffected. Schemas using these keywords now compile and validate correctly.

## Rollback

Revert to the previous release. Schemas that use $id, $anchor, or $ref will no longer compile until restored to this version or later.

## Known limitations

- External schema registry through caller-supplied mappings (story 187)
- $dynamicRef dynamic-scope resolution (story 188); only $dynamicAnchor syntax is validated
- Meta-schema validation and $schema/$vocabulary enforcement (stories 189, 190)
- Annotation collection through $ref and unevaluated* keywords (stories 191, 192)
- Configurable evaluation limits and CLI flags beyond the fixed recursion backstop (story 193)
- JSON Schema drafts prior to 2020-12

## Concerns

- F7 (medium severity, security and performance): Load-time memory can grow quadratically in pathological cases. When reference targets under unknown keywords are discovered in ancestor-after-descendant order across resolver batches, already-compiled subtrees may be recompiled. A multi-MB hostile schema could trigger this before the depth limit applies. The tester estimated this from code inspection; it has not been measured experimentally. Does not block shipping but should be addressed in story 193 (limits) or in a near-term follow-up to prevent potential denial-of-service.
