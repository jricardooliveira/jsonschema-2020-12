# PLAN #3: Error reports: flag and basic output, multiple documents

- Date: 2026-10-16
- Status: approved by the operator at Checkpoint 2 (run #137, 2026-10-16 19:02 UTC)
- Design: see `ADR-003-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `internal/eval/units.go`
- `internal/eval/units_test.go`
- `internal/eval/eval.go`
- `internal/keywords/registry.go`
- `internal/compile/compile.go`
- `internal/keywords/validation/type.go`
- `internal/keywords/validation/enum.go`
- `internal/keywords/validation/const.go`
- `internal/output/registry.go`
- `internal/output/registry_test.go`
- `internal/output/flag.go`
- `internal/output/flag_test.go`
- `internal/output/basic.go`
- `internal/output/basic_test.go`
- `jsonschema.go`
- `output.go`
- `output_test.go`
- `numberlimits.go`
- `cmd/jsonschema/run.go`
- `cmd/jsonschema/output_test.go`
- `cmd/jsonschema/multidoc_test.go`
- `cmd/jsonschema/run_test.go`

Modules the architecture expects to touch:

- `internal/eval/units.go`
- `internal/eval/units_test.go`
- `internal/eval/eval.go`
- `internal/keywords/registry.go`
- `internal/compile/compile.go`
- `internal/keywords/validation/type.go`
- `internal/keywords/validation/enum.go`
- `internal/keywords/validation/const.go`
- `internal/output/registry.go`
- `internal/output/registry_test.go`
- `internal/output/flag.go`
- `internal/output/flag_test.go`
- `internal/output/basic.go`
- `internal/output/basic_test.go`
- `jsonschema.go`
- `output.go`
- `output_test.go`
- `numberlimits.go`
- `cmd/jsonschema/run.go`
- `cmd/jsonschema/output_test.go`
- `cmd/jsonschema/multidoc_test.go`
- `cmd/jsonschema/run_test.go`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — Collect output units during evaluation, with a cap that counts every appended unit
2. **T2** — Output-format registry with 'flag' and 'basic'  _(after T1)_
3. **T3** — Library API: render a Result in a chosen shape, with an output-unit cap option  _(after T1, T2)_
4. **T4** — CLI: -o/--output, several documents, NDJSON lines and worst-result exit code  _(after T3)_
5. **T5** — CLI contract tests  _(after T4)_

## Risks

- anyOf/oneOf do not exist yet (story 185), so the regression test must use a test-only applicator registered from the test file. Story 185 must use the same EvalContext child-evaluation helper, or the cap could be bypassed.
- eval.go, compile.go, registry.go, jsonschema.go and run.go are crowded files, so this story serialises with other stories that write them.
- Changing invalid-document stdout from the minimal result to basic may break exact-output assertions in run_test.go.

Implementation constraints the coder must honor:

- Standard library only; go.mod is not edited.
- No instance value may appear in any unit, error sentence, stderr reason or NDJSON line.
- The cap counts every appended unit, including units appended by children of passing applicators that are later discarded. It never uses the retained slice length.
- Per-evaluation state (collector, counter) lives in the per-call context; the compiled Schema is never mutated.
- EvaluateFunc's signature stays unchanged; new capabilities are methods on EvalContext.
- Tests read DefaultMaxOutputUnits and lower it through the option instead of repeating the number.
- Each feature's tests go in new test files; existing tests are edited only where the new stdout breaks them.

- DB impact: no · API impact: yes: new Result.Output, DefaultMaxOutputUnits and Options.MaxOutputUnits; CLI gains -o/--output and multiple documents. · Migration needed: no

Breaking changes:

- CLI stdout for an invalid document changes from the minimal result to the 2020-12 'basic' object (the default).

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: go test -race ./internal/eval/... ./internal/keywords/... passes. units_test.go registers a test-only passing applicator whose discarded child failures exceed a lowered cap, and asserts that evaluation stops with the LimitError 'output units'. It also asserts the deterministic order and that no instance value appears in any unit.
- **T2**: go test ./internal/output/... passes. Tests cover the exact JSON for valid and invalid flag/basic, omission of absoluteKeywordLocation when unknown, the unknown-shape error and duplicate registration.
- **T3**: go test -race . passes. output_test.go checks flag and basic output from the library, absence of instance values, the LimitError from Validate when the cap is lowered, and that the test reads the cap from DefaultMaxOutputUnits instead of repeating the number.
- **T4**: The CLI builds with go build ./cmd/jsonschema, and the tests in T5 pass against it.
- **T5**: go test -race ./cmd/jsonschema/... passes. Every exit code (0, 1, 2) is asserted with one and with several documents.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
