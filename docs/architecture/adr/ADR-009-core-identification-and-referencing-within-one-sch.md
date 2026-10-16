# ADR #9: Core identification and referencing within one schema document: $id, $anchor, $defs, $ref, $comment

- Date: 2026-10-16
- Story: #9
- Status: proposed (pending architecture sign-off)

## Decision
Resized to 30 files from the previous 36 without dropping any criterion. URI normalization is merged into internal/uri/resolve.go (it is part of comparing resolved URIs). Anchor and $id syntax checks are merged into internal/registry/index.go (they run during the same walk). Root base URI selection is merged into internal/compile/refs.go (it is the first step of load-time resolution). The loop cases are folded into testdata/core/ref.json. Flow: the compiler builds the resource index (internal/registry/index.go) from the parsed document, giving each subschema a base URI and indexing normalized $id values and per-resource anchors. internal/compile/refs.go then resolves every $ref once against that index, so broken references fail at load time regardless of the document. The $ref keyword (internal/keywords/core/ref.go) holds a pointer to its compiled target and evaluates it with its siblings. The evaluation loop guard lives in internal/eval/refguard.go, with per-call state only; eval.go, a crowded file, gets one wiring call. cmd/jsonschema/main.go, also crowded, gets only one flag and passes the schema path; the file-to-URI conversion lives in the library, not in main.go. $id, $anchor, $defs and $comment keep one file each, per the keyword convention. Their tests share core_keywords_test.go plus testdata JSON cases.

## Affected modules
- internal/uri/resolve.go
- internal/uri/resolve_test.go
- internal/uri/pointer.go
- internal/uri/pointer_test.go
- internal/registry/index.go
- internal/registry/index_test.go
- internal/keywords/core/ref.go
- internal/keywords/core/ref_test.go
- internal/keywords/core/id.go
- internal/keywords/core/anchor.go
- internal/keywords/core/defs.go
- internal/keywords/core/comment.go
- internal/keywords/core/core_keywords_test.go
- internal/compile/refs.go
- internal/compile/refs_test.go
- options.go
- baseuri_option_test.go
- internal/eval/refguard.go
- internal/eval/refguard_test.go
- internal/eval/eval.go
- testdata/core/id.json
- testdata/core/anchor.json
- testdata/core/defs.json
- testdata/core/ref.json
- testdata/core/comment.json
- testdata/core/uri_normalization.json
- cmd/jsonschema/main.go
- cmd/jsonschema/refs_cli_test.go
- docs/references.md

## Data / API impact
- DB impact: no
- API impact: yes: new library option and CLI flag for the root base URI; new broken-schema error kinds for references
- Migration needed: no

## Constraints for implementation
- Standard library only (net/url may be used for parsing, but normalization and comparison follow RFC 3986 section 6.2.2 explicitly). No change to go.mod.
- Never fetch over the network and never read a file to resolve a reference. Computing a file:// base URI from a path must not touch the disk beyond the schema file the caller supplied.
- Error messages name schema locations and references, never instance values.
- The loop guard state lives in the per-call evaluation context, never in the compiled schema (race-safe).
- eval.go and main.go are crowded: add only a single wiring call or flag line to each, and put the logic in the new files.
- Keep this story's new code out of the keyword registry file and the all.go blank-import file; the core package must already be listed there by the skeleton.
- Each testdata case cites its 2020-12 core section. Nothing is copied from JSON-Schema-Test-Suite.

## Risks
- Recursion through $ref can recurse deeply; the guard must catch non-descending loops without rejecting legitimate recursion such as trees.
- Normalization that is too lax could make different resources match; the negative tests guard against this.
- Following object targets under unknown keywords must not break the immutability or base URI rules for those targets.

## Breaking changes
- none

## Sensitivity
- security
