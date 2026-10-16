# SPEC #9: Core identification and referencing within one schema document: $id, $anchor, $defs, $ref, $comment

- Date: 2026-10-16
- Type: feature
- Verdict: pass
- Status: passed gate 1 with nothing for the operator to decide (run #143, 2026-10-16 20:56 UTC)

## Problem

The validator cannot yet identify schema resources or follow references. Without '$id' base URI resolution, '$anchor', '$defs' and '$ref', real 2020-12 schemas that reuse subschemas inside one document cannot be checked. Unresolvable or looping references must also be caught as errors and never reported as 'invalid'.

## Why this matters

The brief's mandatory scope (3) requires $id and base URI resolution per RFC 3986 including embedded resources, $anchor, local and relative $ref with ~0/~1 escaping and percent-decoding, $defs and $comment. Later stories (registry 187, $dynamicRef 188, meta-schemas 189, unevaluated* 192) build on this resource index and reference resolution.

## Acceptance criteria

- '$id' sets the base URI of its schema resource by RFC 3986 reference resolution against the enclosing base URI. This includes embedded schema resources with their own '$id', whose subschemas resolve against the embedded resource's base. An '$id' ending in an empty fragment ('#') is accepted and the empty fragment is ignored. An '$id' with a non-empty fragment is a broken schema (exit code 2), and the error names the schema location.
- '$anchor' defines a plain-name fragment in its schema resource. '$anchor' and '$dynamicAnchor' values must match ^[A-Za-z_][-A-Za-z0-9._]*$, otherwise the schema is broken (exit code 2). A repeated '$id' or a repeated anchor name within one schema resource is a broken schema (exit code 2), and the error names the repeated '$id' or anchor. This story checks '$dynamicAnchor' syntax only; dynamic resolution belongs to story 188.
- '$defs' holds reusable subschemas. They are compiled with the base URI in effect at their location, can be reached by '$ref', and have no effect on validation unless referenced.
- '$ref' resolves local JSON Pointer fragments. Pointer tokens are percent-decoded first and then ~1 and ~0 are unescaped, in that order, so '#/$defs/a~1b', '#/$defs/a~0b' and '#/$defs/a%25b' reach the correct keys. A reference with invalid percent-encoding (for example '%zz') is a broken schema (exit code 2), and the error names the schema location.
- '$ref' resolves plain-name anchor fragments and relative references against the current base URI. Before comparison, both the reference and every declared '$id' get RFC 3986 safe normalization: lowercase scheme and host, unescaping of unreserved characters, uppercase percent-escape hex digits, and removal of '.' and '..' segments. For example, 'https://Example.com/a%7Eb.json' matches 'https://example.com/a~b.json', while addresses differing in anything else (for example path case or a different path) do not match. Positive and negative tests cover this choice.
- A '$ref' whose target is an object or a boolean that is not a recognised subschema location (for example '#/definitions/x', where 'definitions' is an unknown keyword in 2020-12) is followed. The target is compiled as a subschema with the base URI in effect at that location. A target that is a number, string, array or null is an unresolvable reference (exit code 2). This choice is documented with the note that 2020-12 core 9.4.2 leaves it undefined.
- '$ref' works alongside sibling keywords in the same schema object: the referenced schema and every sibling keyword are all applied, and the document is valid only when all of them pass.
- '$comment' is accepted in any schema object and has no effect on validation results.
- Every '$ref' in a schema is resolved once when the schema is loaded, before any document is checked. A reference that cannot be resolved is a broken-schema error (exit code 2, never 'invalid') naming the reference and its schema location, even when it sits in a branch that a given document never reaches.
- A root schema without '$id' gets its base URI from its file location (for example file:///home/me/schema.json). The caller can override it through a library option and a CLI flag. A schema piped in without a file gets a fixed, documented placeholder base URI. Working out the base URI never reads anything from disk.
- Legitimate recursive schemas whose references descend into the document (trees, linked lists) validate correctly. At load time, a pure reference cycle (schema objects containing only '$ref' that point at each other) is a broken-schema error (exit code 2). During evaluation, re-entering the same schema location at the same instance location on the current evaluation path is a loop error (exit code 2, never a hang and never valid or invalid), with the recursion limit as the backstop.
- The validator never fetches a reference target from the network or reads it from disk. A reference to another resource that is not inside the loaded document is an unresolved-reference error (exit code 2). Spec-derived positive and negative tests cite their 2020-12 core sections (8.2.1 $id, 8.2.2 $anchor, 8.2.3.1 $ref, 8.2.4 $defs, 8.3 $comment, 9.1 base URI, 9.2 embedded resources, 9.4 JSON Pointer fragments). docs/references.md documents the default base URI, the stdin placeholder, the normalization rule, non-schema targets and loop handling.

## Non-goals

- Resolving references to other documents through a caller-supplied registry (story 187).
- '$dynamicRef' and '$dynamicAnchor' dynamic-scope resolution (story 188); only '$dynamicAnchor' syntax is checked here.
- Meta-schema validation and '$schema' / '$vocabulary' handling (stories 189, 190).
- Annotation collection through '$ref' and unevaluated* interaction (stories 191, 192).
- Configurable limits and their CLI flags beyond using the existing recursion backstop (story 193).
- Any preparation for older drafts (for example 'definitions' as a keyword or '$ref' overriding siblings).

## Tasks

- **T1** — URI resolution, normalization and JSON Pointer
  - purpose: Provide RFC 3986 reference resolution, safe normalization for comparing addresses, percent-encoding validation, and JSON Pointer fragment parsing with percent-decoding followed by ~1/~0 unescaping, in internal/uri.
  - scope: internal/uri/resolve.go, internal/uri/resolve_test.go, internal/uri/pointer.go, internal/uri/pointer_test.go
  - done when: go test ./internal/uri/... passes, with resolution, normalization (positive and negative) and pointer escaping cases.
  - delivers: AC1, AC4, AC5
- **T2** — Resource index with $id, $anchor and duplicate detection
  - purpose: Walk a loaded schema document, compute each subschema's base URI, index schema resources by normalized $id and anchors by resource, check the $id fragment and anchor syntax rules, and report duplicates.
  - scope: internal/registry/index.go, internal/registry/index_test.go
  - done when: go test ./internal/registry/... passes, covering embedded resources, empty or non-empty $id fragments, anchor syntax and duplicate $id or anchor errors.
  - depends on: T1
  - delivers: AC1, AC2, AC3
- **T3** — Core keyword files
  - purpose: Register $id, $anchor, $defs, $comment and $ref through the keyword registry, each in its own file. $ref evaluates its resolved target alongside sibling keywords; $comment and $defs have no validation effect.
  - scope: internal/keywords/core/id.go, internal/keywords/core/anchor.go, internal/keywords/core/defs.go, internal/keywords/core/comment.go, internal/keywords/core/ref.go, internal/keywords/core/ref_test.go, internal/keywords/core/core_keywords_test.go
  - done when: go test ./internal/keywords/core/... passes, including $ref with siblings and $comment having no effect.
  - depends on: T2
  - delivers: AC3, AC7, AC8
- **T4** — Load-time reference resolution and root base URI
  - purpose: At compile time, choose the root base URI (option override, otherwise file location, otherwise stdin placeholder) and resolve every $ref once against the index. Follow object or boolean non-schema targets, reject other values, reject invalid percent-encoding and pure $ref cycles, and report unresolved references as broken-schema errors.
  - scope: internal/compile/refs.go, internal/compile/refs_test.go
  - done when: go test ./internal/compile/... passes, with unreached broken refs, non-schema targets, pure cycles and base URI selection cases.
  - depends on: T2
  - delivers: AC6, AC9, AC10, AC11, AC12
- **T5** — Evaluation-time loop guard
  - purpose: Track (schema location, instance location) pairs on the current evaluation path in the per-call evaluation context and raise a loop error on re-entry. Wire it into eval.go with a single call so that legitimate recursion that descends into the document keeps working.
  - scope: internal/eval/refguard.go, internal/eval/refguard_test.go, internal/eval/eval.go
  - done when: go test -race ./internal/eval/... passes. A tree schema validates and a non-descending loop returns a loop error without hanging.
  - depends on: T3
  - delivers: AC11
- **T6** — Base URI option, CLI flag, spec-derived test cases and docs
  - purpose: Expose the base URI override as a library option and a CLI flag. Have the CLI pass the schema file's location (or nothing for stdin). Add the spec-derived testdata cases and CLI contract tests for exit code 2 on unresolved, broken and looping references. Document the decisions in docs/references.md.
  - scope: options.go, baseuri_option_test.go, cmd/jsonschema/main.go, cmd/jsonschema/refs_cli_test.go, testdata/core/id.json, testdata/core/anchor.json, testdata/core/defs.json, testdata/core/ref.json, testdata/core/comment.json, testdata/core/uri_normalization.json, docs/references.md
  - done when: go test ./... passes with no network. CLI tests show exit 2 for an unresolved reference in an unreached branch, an invalid percent-encoding, a duplicate anchor and a $ref loop, and show that the base URI flag changes how a relative reference resolves.
  - depends on: T4, T5
  - delivers: AC1, AC2, AC3, AC4, AC5, AC6, AC7, AC8, AC9, AC10, AC11, AC12
