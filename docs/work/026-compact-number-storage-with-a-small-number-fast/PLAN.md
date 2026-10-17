# PLAN #26: Compact exact number storage in decoded documents (fix 5)

- Date: 2026-10-17
- Status: passed gate 2 with nothing for the operator to decide (run #161, 2026-10-17 13:04 UTC)
- Design: see `ADR-026-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `internal/jsonvalue/number.go`
- `internal/jsonvalue/number_canonical.go`
- `internal/jsonvalue/decode.go`
- `internal/jsonvalue/number_fastpath_test.go`
- `internal/jsonvalue/equal.go`
- `internal/jsonvalue/equal_fastpath_test.go`
- `internal/keywords/validation/multipleof.go`
- `internal/keywords/validation/type_numberpath_test.go`
- `internal/keywords/validation/multipleof_numberpath_test.go`
- `internal/keywords/validation/maximum_numberpath_test.go`
- `internal/keywords/validation/exclusivemaximum_numberpath_test.go`
- `internal/keywords/validation/minimum_numberpath_test.go`
- `internal/keywords/validation/exclusiveminimum_numberpath_test.go`
- `internal/keywords/validation/enum_numberpath_test.go`
- `internal/keywords/validation/const_numberpath_test.go`
- `internal/keywords/validation/uniqueitems_numberpath_test.go`
- `number_limits_regression_test.go`
- `number_race_test.go`
- `docs/benchmarks.md`

Modules the architecture expects to touch:

- `internal/jsonvalue/number.go`
- `internal/jsonvalue/number_canonical.go`
- `internal/jsonvalue/decode.go`
- `internal/jsonvalue/equal.go`
- `internal/jsonvalue/number_fastpath_test.go`
- `internal/jsonvalue/equal_fastpath_test.go`
- `internal/keywords/validation/multipleof.go`
- `internal/keywords/validation/type_numberpath_test.go`
- `internal/keywords/validation/multipleof_numberpath_test.go`
- `internal/keywords/validation/maximum_numberpath_test.go`
- `internal/keywords/validation/exclusivemaximum_numberpath_test.go`
- `internal/keywords/validation/minimum_numberpath_test.go`
- `internal/keywords/validation/exclusiveminimum_numberpath_test.go`
- `internal/keywords/validation/enum_numberpath_test.go`
- `internal/keywords/validation/const_numberpath_test.go`
- `internal/keywords/validation/uniqueitems_numberpath_test.go`
- `number_limits_regression_test.go`
- `number_race_test.go`
- `docs/benchmarks.md`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — Compact Number representation
2. **T2** — Exact mixed-representation equality and hashing  _(after T1)_
3. **T3** — Per-rule fast/big/boundary tests and the multipleOf helper  _(after T1, T2)_
4. **T4** — Limits regression and race test  _(after T1, T2)_
5. **T5** — Before/after benchmarks and CLI peak memory  _(after T3, T4)_

## Risks

- Non-canonical representations (for example 10 as mant=10 vs mant=1,exp=1) would break equality or uniqueItems hashing; canonicalisation must be total.
- An int64 overflow in mantissa alignment for Cmp or multipleOf would silently corrupt answers; boundary tests at int64 max must cover this.
- Number-text preservation may need the original text, which limits the per-number memory saving.
- docs/benchmarks.md is crowded (6 stories), and stories 199 to 202 also append before/after numbers there, so their before baselines interleave.
- Retained-heap measurement noise could hide a small gain.

Implementation constraints the coder must honor:

- Never use float64 to decide an answer; do not add a float64 field.
- Never compute 10^n for a large n. Exponent alignment stays within the existing digit cap.
- Digit and exponent caps and their LimitErrors are identical, and are read from internal/jsonvalue/limits.go rather than repeated as numbers in tests.
- Do not edit internal/jsonvalue/number_test.go, equal_test.go or any existing test file.
- Canonical() and String() output are byte-identical to today for every input.
- Do not touch internal/eval/eval.go, internal/compile/compile.go, jsonschema.go, options.go, cmd/jsonschema/run.go or any keyword other than multipleof.go.
- No allocation-limit test is added or changed.
- The 'before' benchmark numbers come from running the current benchmarks, including retained heap, on the old code on the same machine.
- Standard library only; go.mod is unchanged.

- DB impact: no · API impact: no · Migration needed: no

Breaking changes:

- (none)

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: go test ./internal/jsonvalue/... passes with number_test.go unchanged. The new fast/big/boundary unit tests pass.
- **T2**: New equality and hash tests over mixed operands pass. The existing equal_test.go and the allocation-free-equality tests pass unchanged.
- **T3**: All nine new test files pass, and every existing keyword test passes unchanged.
- **T4**: go test ./... and go test -race ./... pass.
- **T5**: docs/benchmarks.md holds both sets of numbers and the CLI peak memory. The document scenarios show lower allocs/op or B/op and lower retained heap, and no scenario regresses.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
