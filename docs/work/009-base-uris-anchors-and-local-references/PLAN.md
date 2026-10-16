# PLAN #9: Core identification and referencing within one schema document: $id, $anchor, $defs, $ref, $comment

- Date: 2026-10-16
- Status: proposed (pending Checkpoint 2 sign-off)
- Design: see `ADR-009-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `internal/uri/resolve.go`
- `internal/uri/resolve_test.go`
- `internal/uri/pointer.go`
- `internal/uri/pointer_test.go`
- `internal/registry/index.go`
- `internal/registry/index_test.go`
- `internal/keywords/core/id.go`
- `internal/keywords/core/anchor.go`
- `internal/keywords/core/defs.go`
- `internal/keywords/core/comment.go`
- `internal/keywords/core/ref.go`
- `internal/keywords/core/ref_test.go`
- `internal/keywords/core/core_keywords_test.go`
- `internal/compile/refs.go`
- `internal/compile/refs_test.go`
- `internal/eval/refguard.go`
- `internal/eval/refguard_test.go`
- `internal/eval/eval.go`
- `options.go`
- `baseuri_option_test.go`
- `cmd/jsonschema/main.go`
- `cmd/jsonschema/refs_cli_test.go`
- `testdata/core/id.json`
- `testdata/core/anchor.json`
- `testdata/core/defs.json`
- `testdata/core/ref.json`
- `testdata/core/comment.json`
- `testdata/core/uri_normalization.json`
- `docs/references.md`

Modules the architecture expects to touch:

- `internal/uri/resolve.go`
- `internal/uri/resolve_test.go`
- `internal/uri/pointer.go`
- `internal/uri/pointer_test.go`
- `internal/registry/index.go`
- `internal/registry/index_test.go`
- `internal/keywords/core/ref.go`
- `internal/keywords/core/ref_test.go`
- `internal/keywords/core/id.go`
- `internal/keywords/core/anchor.go`
- `internal/keywords/core/defs.go`
- `internal/keywords/core/comment.go`
- `internal/keywords/core/core_keywords_test.go`
- `internal/compile/refs.go`
- `internal/compile/refs_test.go`
- `options.go`
- `baseuri_option_test.go`
- `internal/eval/refguard.go`
- `internal/eval/refguard_test.go`
- `internal/eval/eval.go`
- `testdata/core/id.json`
- `testdata/core/anchor.json`
- `testdata/core/defs.json`
- `testdata/core/ref.json`
- `testdata/core/comment.json`
- `testdata/core/uri_normalization.json`
- `cmd/jsonschema/main.go`
- `cmd/jsonschema/refs_cli_test.go`
- `docs/references.md`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — URI resolution, normalization and JSON Pointer
2. **T2** — Resource index with $id, $anchor and duplicate detection  _(after T1)_
3. **T3** — Core keyword files  _(after T2)_
4. **T4** — Load-time reference resolution and root base URI  _(after T2)_
5. **T5** — Evaluation-time loop guard  _(after T3)_
6. **T6** — Base URI option, CLI flag, spec-derived test cases and docs  _(after T4, T5)_

## Risks

- Recursion through $ref can recurse deeply; the guard must catch non-descending loops without rejecting legitimate recursion such as trees.
- Normalization that is too lax could make different resources match; the negative tests guard against this.
- Following object targets under unknown keywords must not break the immutability or base URI rules for those targets.

Implementation constraints the coder must honor:

- Standard library only (net/url may be used for parsing, but normalization and comparison follow RFC 3986 section 6.2.2 explicitly). No change to go.mod.
- Never fetch over the network and never read a file to resolve a reference. Computing a file:// base URI from a path must not touch the disk beyond the schema file the caller supplied.
- Error messages name schema locations and references, never instance values.
- The loop guard state lives in the per-call evaluation context, never in the compiled schema (race-safe).
- eval.go and main.go are crowded: add only a single wiring call or flag line to each, and put the logic in the new files.
- Keep this story's new code out of the keyword registry file and the all.go blank-import file; the core package must already be listed there by the skeleton.
- Each testdata case cites its 2020-12 core section. Nothing is copied from JSON-Schema-Test-Suite.

- DB impact: no · API impact: yes: new library option and CLI flag for the root base URI; new broken-schema error kinds for references · Migration needed: no

Breaking changes:

- (none)

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: go test ./internal/uri/... passes, with resolution, normalization (positive and negative) and pointer escaping cases.
- **T2**: go test ./internal/registry/... passes, covering embedded resources, empty or non-empty $id fragments, anchor syntax and duplicate $id or anchor errors.
- **T3**: go test ./internal/keywords/core/... passes, including $ref with siblings and $comment having no effect.
- **T4**: go test ./internal/compile/... passes, with unreached broken refs, non-schema targets, pure cycles and base URI selection cases.
- **T5**: go test -race ./internal/eval/... passes. A tree schema validates and a non-descending loop returns a loop error without hanging.
- **T6**: go test ./... passes with no network. CLI tests show exit 2 for an unresolved reference in an unreached branch, an invalid percent-encoding, a duplicate anchor and a $ref loop, and show that the base URI flag changes how a relative reference resolves.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
