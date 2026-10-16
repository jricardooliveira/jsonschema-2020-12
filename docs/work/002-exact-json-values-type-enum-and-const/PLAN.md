# PLAN #2: Exact JSON value model, deep equality, and the type, enum and const keywords

- Date: 2026-10-16
- Status: proposed (pending Checkpoint 2 sign-off)
- Design: see `ADR-002-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `internal/jsonvalue/value.go`
- `internal/jsonvalue/number.go`
- `internal/jsonvalue/decode.go`
- `internal/jsonvalue/limits.go`
- `internal/jsonvalue/number_test.go`
- `internal/jsonvalue/decode_test.go`
- `internal/jsonvalue/equal.go`
- `internal/jsonvalue/equal_test.go`
- `compiler.go`
- `numberlimits.go`
- `internal/keywords/registry.go`
- `internal/eval/eval.go`
- `cmd/jsonschema/main.go`
- `numberlimits_test.go`
- `cmd/jsonschema/parse_errors_test.go`
- `internal/keywords/validation/type.go`
- `internal/keywords/validation/type_test.go`
- `testdata/validation/type.json`
- `internal/keywords/validation/enum.go`
- `internal/keywords/validation/const.go`
- `internal/keywords/validation/enum_test.go`
- `internal/keywords/validation/const_test.go`
- `testdata/validation/enum.json`
- `testdata/validation/const.json`
- `internal/speccases/runner.go`
- `internal/speccases/runner_test.go`

Modules the architecture expects to touch:

- `internal/jsonvalue/value.go`
- `internal/jsonvalue/number.go`
- `internal/jsonvalue/decode.go`
- `internal/jsonvalue/limits.go`
- `internal/jsonvalue/equal.go`
- `internal/jsonvalue/number_test.go`
- `internal/jsonvalue/decode_test.go`
- `internal/jsonvalue/equal_test.go`
- `compiler.go`
- `numberlimits.go`
- `numberlimits_test.go`
- `internal/keywords/registry.go`
- `internal/eval/eval.go`
- `cmd/jsonschema/main.go`
- `cmd/jsonschema/parse_errors_test.go`
- `internal/keywords/validation/type.go`
- `internal/keywords/validation/enum.go`
- `internal/keywords/validation/const.go`
- `internal/keywords/validation/type_test.go`
- `internal/keywords/validation/enum_test.go`
- `internal/keywords/validation/const_test.go`
- `testdata/validation/type.json`
- `testdata/validation/enum.json`
- `testdata/validation/const.json`
- `internal/speccases/runner.go`
- `internal/speccases/runner_test.go`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — Exact JSON value model and strict decoder
2. **T2** — Deep JSON equality  _(after T1)_
3. **T3** — Load schemas and documents through the strict decoder  _(after T1)_
4. **T4** — type keyword  _(after T3)_
5. **T5** — enum and const keywords  _(after T2, T3)_
6. **T6** — Shared spec-case runner  _(after T3)_

## Risks

- The skeleton (story 178) may have defined the instance type in the evaluate signature as `any`. Changing it to jsonvalue.Value touches registry.go and eval.go, which every later keyword story depends on, so this must land before stories 181-185.
- encoding/json's Decoder.Token cannot see raw escapes, so lone-surrogate detection may need a small pre-scan or a hand-written string scanner.
- The number-cap flag overlaps with story 193's limit flags. Their names and option shape should line up.

Implementation constraints the coder must honor:

- Standard library only. go.mod stays without require lines.
- Check digit count and exponent size on the raw number text before any math/big conversion.
- Do not rely on encoding/json for UTF-8 or surrogate validation; it replaces bad text silently.
- Errors and keyword messages give locations only, never instance values.
- Tests read defaults from the exported constants instead of repeating numbers.
- Each keyword in its own file, registered via init(). No central keyword list or switch is edited.

- DB impact: no · API impact: yes: the library gains a number-cap option and returns parse/limit errors for broken input; the CLI gains a number-cap flag; the keyword evaluate signature may change to jsonvalue.Value · Migration needed: no

Breaking changes:

- If the skeleton's keyword evaluate signature used a different instance type, it changes to jsonvalue.Value.

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: go test ./internal/jsonvalue passes. It covers big integers, 0.1, 1.0 as an integer, duplicate names at depth, invalid UTF-8, \uD800, valid surrogate pairs, BOM, trailing values, and the number cap using the exported default (not a repeated literal).
- **T2**: go test ./internal/jsonvalue -run Equal passes, with positive and negative cases for every kind.
- **T3**: Library tests show that a broken schema, a broken document, a duplicate name and an over-cap number all return errors, and that the raised cap accepts a large number. CLI contract tests show exit code 2 for broken JSON in the schema file, a document file and stdin, and that the cap flag works.
- **T4**: go test ./internal/keywords/validation -run Type passes. It covers the testdata cases citing validation §6.1.1, including 1.0 as an integer and the array form.
- **T5**: go test ./internal/keywords/validation -run 'Enum|Const' passes. It covers the testdata cases citing validation §6.1.2 and §6.1.3, including object member order, 1 vs 1.0, null, and nested arrays.
- **T6**: The type, enum and const tests run through the runner. A case without a citation fails the runner. go test ./... passes.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
