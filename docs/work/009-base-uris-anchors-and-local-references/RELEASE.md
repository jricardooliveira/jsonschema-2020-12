# Release notes — #9: Core identification and referencing within one schema document: $id, $anchor, $defs, $ref, $comment

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #9
- Status: approved by the operator at Checkpoint 3 (run #143, 2026-10-16 21:27 UTC)

**⚠️ the release-agent has concerns.**

## What changed

This change implements schema resource identification and reference resolution: $id sets base URIs using RFC 3986 resolution, $anchor defines plain-name fragments, $defs holds reusable subschemas, and $ref resolves references through JSON Pointer fragments, anchor names, and relative paths. References are resolved once at load time (catching broken schemas before any document is checked), verified through safe RFC 3986 normalization, and evaluated at runtime with a per-call loop guard that prevents infinite recursion while allowing legitimate tree descent. All 12 acceptance criteria are delivered with spec-derived test cases, CLI exit-code 2 contracts for broken schemas, and no network or disk access beyond the schema file the caller supplied.

- internal/uri: RFC 3986 reference resolution (relative and absolute), safe URI normalization (lowercase scheme/host, unescape unreserved, dot-segment removal), and JSON Pointer fragment parsing (percent-decode then ~0/~1 unescape)
- internal/registry: Resource index mapping normalized $id to resources and anchors within each resource, duplicate $id/anchor detection, embedded schema base URI computation
- internal/keywords/core: Five new keywords ($id, $anchor, $defs, $ref, $comment) with validation and evaluation; $id/$anchor/$defs/$comment are no-op at evaluation, $ref points to resolved target
- internal/compile: Load-time $ref resolution against resource index, base URI selection (explicit override / file location / stdin placeholder), pure $ref cycle detection, non-schema target following (objects/booleans only)
- internal/eval: Per-call loop guard via callState wrapping EvaluatedItems; tracks (schema location, instance location) pairs and raises LoopError on re-entry without document descent
- cmd/jsonschema/run.go: --base-uri flag passed to CompileWithOptions
- options.go, baseuri_option_test.go: BaseURI and SchemaPath fields; baseURI() helper validates override (absolute, no non-empty fragment) and picks URI
- docs/references.md: Base URI selection rules, resolution process, RFC 3986 normalization choice (with test examples), non-schema target handling, loop detection and depth limits
- testdata/core: Six JSON test suites (id.json, anchor.json, defs.json, ref.json, comment.json, uri_normalization.json) each citing 2020-12 spec section

## How to verify

- RFC 3986 resolution: `go test ./internal/uri/... -run TestResolve -v` shows examples from RFC 3986 5.4 (relative, parent, authority, etc.)
- Safe normalization: `go test ./internal/uri/... -run TestNormalize -v` shows 'https://Example.com/a%7Eb.json' and 'https://example.com/a~b.json' match after normalization, but path case and %2F vs / do not
- JSON Pointer escaping: `go test ./internal/uri/... -run TestParsePointer -v` verifies percent-decode-first (#/$defs/a%25b reaches 'a%b'), then ~1/~0 unescape (#/$defs/a~1b reaches 'a/b')
- Resource index: `go test ./internal/registry/... -v` verifies embedded $id changes base for subschemas, anchors are per-resource, duplicates detected, $id fragments checked
- Load-time resolution: `go test ./internal/compile/... -run TestBrokenRefs -v` shows unresolved references in unreached branches fail at load time (exit 2 before any document is checked), non-schema targets (numbers, strings, arrays, null) rejected
- Loop guard: `go test ./internal/eval/refguard_test.go -run 'TestTreeRecursion|TestNonDescendingLoop|TestEvalDepthLimit' -v` shows tree recursion descends correctly, non-descending loops return LoopError, depth limit returns LimitError
- CLI base URI: `jsonschema validate --base-uri http://x/y/s.json schema-with-relative-refs doc` resolves references relative to the given base; without flag, defaults to stdin placeholder or file path
- CLI broken schema: `jsonschema validate schema-with-{unresolved-ref,bad-percent-encoding,duplicate-anchor,pure-cycle} doc` exits 2 regardless of document content (tests in refs_cli_test.go verify this)

## Migration

None. The library API is backward compatible: existing code using Compile() continues to work (uses stdin placeholder base URI). Two new optional Options fields (BaseURI, SchemaPath) may be set; if unset, the library picks a default. Compiled schemas from before this change are not compatible and must be recompiled (internal representation changed).

## Rollback

Revert the commit. Schemas compiled after this change will no longer load (compiled format unrecognized). The CLI loses --base-uri flag and $ref support entirely. Any schema using $id, $anchor, $defs, or $ref will fail at load time with 'unknown keyword' error (these keywords are now registered but their compile functions would be missing). Callers must recompile schemas with the previous validator.

## Known limitations

- '$dynamicRef' and '$dynamicAnchor' dynamic resolution is not implemented (story 188); only anchor syntax is validated here
- References to schemas outside the loaded document are not resolved; caller must supply external schemas through a registry (story 187)
- Recursion limits are fixed (MaxEvalDepth = 40000 schema nodes; step limit from evaluation context); neither is configurable (story 193)
- Unknown-keyword schema targets (e.g. 'definitions' in 2020-12) are followed if they are objects or booleans; this is documented in docs/references.md as spec-undefined behavior (2020-12 core 9.4.2)

## Concerns

- Low: Error messages name schema locations but use format strings with substring assertions in tests (e.g. 'want %q in err.Error()'). A future refactor to error message wording could pass all tests without catching location format drift. Recommendation before shipping story 187 (external registry): tighten error assertions to exact format or dedicated error types.
- Low: An empty --base-uri flag (--base-uri '') is treated as unset and falls back to stdin placeholder, rather than rejected as an error. The documentation says 'must be absolute', which is correct, but the CLI silently ignores empty input. Recommendation: test and document this edge case explicitly, or reject --base-uri '' with a 'must not be empty' error.
- Low: '$id' appearing inside unknown keywords (e.g. inside 'enum' or 'const') is correctly not indexed, since those are data values not schemas. However, this design choice and its rationale are not explicitly documented in a way a reader of docs/references.md would discover without reading the spec. Recommendation: add a note under 'Non-schema targets' stating that unknown keywords are traversed only to find schema substructures, and that $id/anchor inside data keywords have no effect.
