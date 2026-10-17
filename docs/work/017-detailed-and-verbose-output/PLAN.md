# PLAN #17: Detailed and verbose output formats, with URI-safe absoluteKeywordLocation

- Date: 2026-10-17
- Status: proposed (pending Checkpoint 2 sign-off)
- Design: see `ADR-017-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `internal/compile/fragment.go`
- `internal/compile/fragment_test.go`
- `internal/compile/compile.go`
- `cmd/jsonschema/abslocation_escape_test.go`
- `internal/output/tree.go`
- `internal/output/nested_limit_test.go`
- `internal/output/detailed.go`
- `internal/output/detailed_test.go`
- `internal/output/verbose.go`
- `internal/output/verbose_test.go`
- `internal/eval/units.go`
- `internal/eval/units_verbose_test.go`
- `output_nested_test.go`
- `cmd/jsonschema/output_nested_test.go`

Modules the architecture expects to touch:

- `internal/compile/compile.go`
- `internal/compile/fragment.go`
- `internal/compile/fragment_test.go`
- `internal/output/registry.go`
- `internal/output/tree.go`
- `internal/output/nested_limit_test.go`
- `internal/output/detailed.go`
- `internal/output/detailed_test.go`
- `internal/output/verbose.go`
- `internal/output/verbose_test.go`
- `jsonschema.go`
- `output.go`
- `output_nested_test.go`
- `output_unchanged_test.go`
- `cmd/jsonschema/output_nested_test.go`
- `cmd/jsonschema/abslocation_escape_test.go`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — Percent-encode the absoluteKeywordLocation fragment
2. **T2** — Shared nested-report builder bounded by the report byte cap  _(after T1)_
3. **T3** — Detailed shape  _(after T2)_
4. **T4** — Verbose shape, including passing units  _(after T2)_
5. **T5** — Library and CLI contract tests for detailed and verbose  _(after T3, T4)_

## Risks

- internal/output/registry.go, jsonschema.go and output.go were read-only in the reservation and are now write claims. The Tree field on Report and RegisterTreeReport/NeedsTree are the smallest seam that lets a shape receive the tree, and Result has to keep the root node and instance so it can rebuild the tree. jsonschema.go is crowded; the change is three fields and one line.
- A Result now keeps the decoded instance alive for as long as the caller holds it. Validate does no extra work and records nothing extra, but memory is held longer when callers keep many Results.
- Detailed and verbose evaluate the document a second time inside Output. That doubles the CPU cost only for those shapes. Flag and basic are unaffected.
- Verbose does not list passing leaf assertions that have no subschema and no annotation, because the reviewer ruled out recording passing units. This is documented in the shape files.
- Encoding the fragment changes absoluteKeywordLocation for schemas with special characters in their keys. Existing tests that assert the unencoded form will need updating.
- Verbose output for large documents can hit the report cap. That is intended. Until story 193 adds the CLI flag, only library callers can raise the cap, through Options.MaxReportBytes.
- output_unchanged_test.go relies on hashes taken before the change. If the coder takes them after changing flag or basic, the test proves nothing. The new fragment encoding does not affect it because the test schema's keys need no encoding.

Implementation constraints the coder must honor:

- Standard library only; go.mod unchanged
- Each shape is its own file, registered from init() with output.RegisterTreeReport; no edit to a central list or switch
- Schema.Validate keeps calling eval.EvaluateFull: it builds no tree and records no passing units. Only Result.Output, for a shape where NeedsTree is true, calls eval.EvaluateTree, with the same MaxOutputUnits and MaxAnnotations
- Passing tree nodes never count against MaxOutputUnits; do not add Report calls or passing units anywhere
- Do not edit internal/eval/eval.go, internal/eval/units.go or internal/keywords/registry.go
- Touch internal/compile/compile.go only for the one-line fragment call, and jsonschema.go only for the Result fields and the line in Validate that fills them
- Do not edit cmd/jsonschema/run.go and do not add a report-bytes CLI flag (story 193 owns it). A nil Report.Tree must render a minimal valid root unit, so run.go's output.Render name probe accepts the new shapes
- Enforce the cap while rendering, by writing each unit through the existing output.Budget, not after building the whole report in memory
- Never emit instance values; property names appear only where an annotation lists matched names
- Write output_unchanged_test.go first, at the base commit, so its flag/basic SHA-256 values and CLI exit codes come from the code before this story
- Cap tests use DefaultMaxReportBytes and Options.MaxReportBytes from the code instead of repeating the number; the 'raise the cap' test is library-level
- The compiled Schema and Result stay immutable; the tree is per Output call and never cached (go test -race)
- Comments in detailed.go and verbose.go record the parts produced only partly: no unit for a passing leaf assertion that has no subschema and no annotation, and absoluteKeywordLocation only where it is known

- DB impact: no · API impact: yes · Migration needed: no

Breaking changes:

- absoluteKeywordLocation for schema keys that contain spaces, '%' or non-ASCII characters is now percent-encoded, in flag and basic output too (accepted by the reviewer)

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: Unit tests cover keys with a space, '%' and non-ASCII characters, and the result parses with net/url. A CLI basic-shape test shows encoded absoluteKeywordLocation. Existing location tests still pass.
- **T2**: Tests show that exceeding the cap gives the limit error (exit-2 class), that raising the option lets the report through, and that the cap value is read from the code.
- **T3**: Table tests on valid and invalid instances check nesting, locations, annotations and that secret values are absent.
- **T4**: Tests show passing keywords present in verbose and absent from detailed, with no document values. go test -race passes.
- **T5**: CLI tests parse stdout as a single JSON object for one document and per line for several documents, and assert exit codes 0, 1 and 2.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
