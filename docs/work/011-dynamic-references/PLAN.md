# PLAN #11: $dynamicRef and $dynamicAnchor with dynamic scope

- Date: 2026-10-16
- Status: passed gate 2 with nothing for the operator to decide (run #147, 2026-10-16 23:53 UTC)
- Design: see `ADR-011-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `internal/registry/dynamicanchor.go`
- `internal/registry/dynamicanchor_test.go`
- `internal/registry/anchor.go`
- `internal/eval/dynamicscope.go`
- `internal/eval/dynamicscope_test.go`
- `internal/eval/eval.go`
- `internal/keywords/core/dynamicanchor.go`
- `internal/keywords/core/dynamicref.go`
- `internal/keywords/core/dynamicref_test.go`
- `testdata/core/dynamicRef.json`
- `testdata/core/dynamicAnchor.json`
- `internal/keywords/core/dynamicref_spec_test.go`
- `cmd/jsonschema/dynamicref_cli_test.go`
- `cmd/jsonschema/testdata/dynamicref/`
- `internal/eval/dynamicscope_race_test.go`

Modules the architecture expects to touch:

- `internal/registry/dynamicanchor.go`
- `internal/registry/dynamicanchor_test.go`
- `internal/registry/anchor.go`
- `internal/eval/dynamicscope.go`
- `internal/eval/dynamicscope_test.go`
- `internal/eval/dynamicscope_race_test.go`
- `internal/eval/eval.go`
- `internal/keywords/core/dynamicanchor.go`
- `internal/keywords/core/dynamicref.go`
- `internal/keywords/core/dynamicref_test.go`
- `internal/keywords/core/dynamicref_spec_test.go`
- `testdata/core/dynamicRef.json`
- `testdata/core/dynamicAnchor.json`
- `cmd/jsonschema/dynamicref_cli_test.go`
- `cmd/jsonschema/testdata/dynamicref/tree.json`
- `cmd/jsonschema/testdata/dynamicref/strict-tree.json`
- `cmd/jsonschema/testdata/dynamicref/duplicate-anchor.json`
- `cmd/jsonschema/testdata/dynamicref/cycle.json`
- `cmd/jsonschema/testdata/dynamicref/unresolved.json`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — Index $dynamicAnchor names in the resource registry
2. **T2** — Dynamic scope in the per-call evaluation context
3. **T3** — $dynamicAnchor and $dynamicRef keywords  _(after T1, T2)_
4. **T4** — Spec-derived dynamic reference test cases  _(after T3)_
5. **T5** — CLI contract tests for dynamic references  _(after T3)_
6. **T6** — Concurrency check on dynamic scope  _(after T2, T3)_

## Risks

- The point where eval.go enters a schema resource may be spread across $ref handling and embedded-$id handling. Missing one site gives a wrong dynamic scope.
- Reusing story 186's loop key (schema location + instance location) for $dynamicRef needs the dynamically resolved location as the key, not the static one. Otherwise real cycles could slip through or legitimate re-entries could be flagged.
- The location of story 186's anchor index and duplicate check is unknown, so the duplicate-name check may need a small edit in an existing registry file.

Implementation constraints the coder must honor:

- No edit to a central keyword list or switch. Register through init() only.
- The compiled Schema must not be mutated during evaluation. The dynamic scope lives only in the per-call evaluation context.
- $dynamicRef must use story 186's evaluation-time loop detection. Do not add a separate counter or a configurable limit (story 193).
- Touch internal/eval/eval.go only to wire scope push/pop. Put the logic in dynamicscope.go.
- No network or disk access. Targets come only from the compiled document and the caller-supplied registry.
- Reports never include instance values.
- Each test case cites a 2020-12 core section and is written by this project, never copied from the JSON-Schema-Test-Suite.

- DB impact: no · API impact: no · Migration needed: no

Breaking changes:

- (none)

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: Registry tests show '#name' resolving to a $dynamicAnchor, a lookup reporting whether a resource has a matching dynamic anchor, and duplicate names (dynamic+dynamic, dynamic+static) failing compilation.
- **T2**: Unit tests for push/pop and outermost lookup pass. eval.go changes only in the one place that enters and leaves a schema resource.
- **T3**: Keyword unit tests pass for dynamic resolution, bookending, an unresolved reference (schema error), error locations and the loop error.
- **T4**: The shared runner runs both files green under go test -race ./...
- **T5**: The CLI tests pass, with a timeout guarding the cycle case.
- **T6**: go test -race passes with parallel recursive-extension validations giving the expected results.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
