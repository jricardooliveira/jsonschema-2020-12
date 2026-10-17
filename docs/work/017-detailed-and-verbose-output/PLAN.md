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

- Encoding the fragment changes absoluteKeywordLocation for schemas with special characters in keys. Existing tests that assert the unencoded form will need updating.
- Verbose may need passing units that the evaluator does not currently record, which would grow the per-call memory cost.
- Verbose output for large documents can hit the report cap. That is intended, but users may need the raise flag.

Implementation constraints the coder must honor:

- Standard library only; go.mod unchanged
- Each shape is its own file registered from init(); no central list or switch edit
- Never emit instance values; property names only where an annotation lists matched names
- Enforce DefaultMaxReportBytes during building, not after building the whole tree in memory
- Do not edit internal/eval/eval.go; touch internal/compile/compile.go only for the one-line call
- Tests read limits and defaults from code constants
- Compiled Schema stays immutable; per-call state only (go test -race)

- DB impact: no · API impact: yes: two new output-format values; absoluteKeywordLocation fragment encoding changes in all shapes · Migration needed: no

Breaking changes:

- absoluteKeywordLocation for schema keys with spaces, '%' or non-ASCII characters is now percent-encoded in flag/basic output too

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
