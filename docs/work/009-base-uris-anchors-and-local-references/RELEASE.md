# Release notes — #9: Core identification and referencing within one schema document: $id, $anchor, $defs, $ref, $comment

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #9
- Status: proposed (pending Checkpoint 3 sign-off)

**⚠️ the release-agent has concerns.**

## What changed

Enables core $id, $anchor, $defs, $ref, and $comment keywords for within-document schema identification and reference resolution. $id sets base URIs using RFC 3986 resolution; $ref resolves local JSON Pointer fragments and plain-name anchors with full normalization and escaping; load-time validation catches broken schemas (unresolved refs, bad syntax, pure cycles); evaluation-time loop detection stops non-descending recursion. The change is tested at library and CLI levels, with spec-derived cases. One medium-severity limitation remains: a schema with branching $ref can take exponential time in document depth because there is no general evaluation-step limit; this is deferred to story 193 and does not block.

- internal/uri/resolve.go, pointer.go: RFC 3986 reference resolution, safe URI normalization (lowercase scheme/host, unescape unreserved, remove . and .. segments), and JSON Pointer parsing with percent-decoding followed by ~0/~1 unescaping
- internal/registry/index.go: resource index building walks the schema, computes base URIs for each subschema, indexes normalized $id values and per-resource $anchor names, validates $id fragments and anchor syntax, reports duplicates
- internal/keywords/core/{id,anchor,defs,comment,ref}.go: five keyword files registering themselves. $id, $anchor, $defs, $comment have no effect at evaluation time. $ref holds a compiled target pointer and evaluates it alongside sibling keywords
- internal/compile/refs.go: load-time reference resolution. Chooses root base URI (override, file location, or stdin placeholder). Resolves every $ref once against the resource index. Follows object/boolean targets under any keyword; rejects other values. Reports unresolved refs, bad percent-encoding, and pure $ref cycles as broken schema (exit code 2)
- internal/eval/refguard.go: per-call evaluation context tracks (schema location, instance location) pairs. Detects re-entry at the same schema and instance location as a loop error. Enforces MaxEvalDepth = 4 × decoder max to prevent stack overflow on deep recursion
- internal/eval/eval.go: one wiring call to enter/exit nodes for the loop guard. keywordLocation now follows the evaluation path through $ref targets
- jsonschema.go, numberlimits.go, options.go: new Options fields BaseURI (override) and SchemaPath (file location, string work only). Root base URI selection via RootBaseURI in compile package
- cmd/jsonschema/run.go: one --base-uri flag; passes schema path to Options
- testdata/core/{id,anchor,defs,ref,comment,uri_normalization}.json: spec-derived positive and negative test cases citing 2020-12 core sections
- docs/references.md: documents base URI selection, resolution, normalization (RFC 3986 6.2.2), non-schema targets, loop handling, and output locations

## How to verify

- Run `go test ./internal/uri/...`: tests RFC 3986 resolution against RFC 3986 section 5.4.1/5.4.2 examples, safe normalization matching rules (scheme/host lowercase, unreserved unescaping, dot segment removal), and JSON Pointer escaping (percent-decode before ~1/~0)
- Run `go test ./internal/registry/...`: tests resource indexing with embedded $id, empty/non-empty $id fragment rules, $anchor syntax validation, duplicate $id and anchor detection
- Run `go test ./internal/keywords/core/...`: tests $ref evaluation with siblings (both must pass), $comment and $defs having no effect
- Run `go test ./internal/compile/...`: tests reference resolution of reachable and unreachable targets, non-schema targets (object/boolean followed, others rejected), invalid percent-encoding, pure reference cycles, base URI selection
- Run `go test -race ./internal/eval/...`: tests non-descending loops caught at evaluation time, tree recursion (descending) allowed, depth limit as a backstop
- Run the CLI: `jsonschema validate --base-uri http://x/y/s.json schema.json doc.json` with a relative $ref in the schema; use --base-uri to change resolution; omit it to get the file-path default; verify exit code 2 for unresolved refs and reference loops
- Inspect docs/references.md: base URI selection rules, normalization algorithm, non-schema target handling, loop error conditions

## Migration

None. This story adds new keywords; existing schemas without $id, $anchor, $defs, $ref, or $comment are unaffected. Schemas using these keywords are now supported where they were previously rejected as unknown.

## Rollback

Remove the 30 new files and revert the three modified files (jsonschema.go, numberlimits.go, cmd/jsonschema/run.go) to their prior state. No database or persistent state is affected.

## Known limitations

- $ref recursion with branching (multiple $ref to the same schema at different paths in the document) can take exponential time in document depth, because there is no per-evaluation-step limit yet. A deeply nested document with a schema like {allOf: [{$ref: #}, {$ref: #}]} can cause exponential re-evaluation. This is deferred to story 193 (configurable evaluation limits) and does not block this release.
- References to schemas outside the loaded document are unresolved errors. Cross-document resolution via a registry is story 187.
- $dynamicRef and dynamic-scope $anchor resolution are not implemented. Only $dynamicAnchor syntax is validated. Story 188 adds dynamic resolution.
- No network fetching or autonomous disk reading. References must be satisfied within the loaded document.
- A root schema has no default $id; the validator assigns a base URI from file location (file:///...), --base-uri override, or a fixed stdin placeholder (https://stdin.invalid/schema.json).

## Concerns

- Exponential-time evaluation: a schema with branching $ref (e.g. {allOf: [{$ref: #}, {$ref: #}]}) on a deeply nested document can cause exponential re-evaluation because each branch re-enters the same subschema. This is not a correctness issue but a performance issue that would manifest in denial-of-service scenarios. The tester noted it as medium severity and deferred to story 193. A future story should add a per-evaluation-step limit to catch runaway evaluation deterministically.
