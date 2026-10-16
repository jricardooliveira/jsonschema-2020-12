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

- `internal/jsonvalue/decode.go`
- `internal/jsonvalue/limits.go`
- `internal/jsonvalue/number.go`
- `internal/jsonvalue/equal.go`
- `internal/jsonvalue/decode_test.go`
- `internal/jsonvalue/number_test.go`
- `internal/jsonvalue/equal_test.go`
- `jsonschema.go`
- `jsonschema_test.go`
- `numberlimits.go`
- `numberlimits_test.go`
- `cmd/jsonschema/run.go`
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

- Files outside the reservation: jsonschema.go is the real 'compiler.go'. cmd/jsonschema/run.go is where the flags live (main.go is never touched). jsonschema_test.go needs the one-entry fix, because implementing `type` makes 178's `{"type":"string"}` is-ignored expectation wrong; this is unavoidable and the only change to a 178 test. Reserved files that are dropped: compiler.go, internal/keywords/registry.go, internal/eval/eval.go, cmd/jsonschema/main.go, internal/jsonvalue/value.go and internal/speccases/* (spec_test.go already is the shared runner).
- Instead of the previous design's signature change to a jsonvalue.Value struct, the decoded `any` tree is the value model, because the struct would break 178's dispatch, compile, eval and decode tests (review point 5). Later keyword stories (181-185) must use jsonvalue.Rat, IsInteger and Equal, and never compare json.Number strings or call strconv.ParseFloat.
- Options is defined in numberlimits.go, while three other stories plan an options.go. Those stories (story 193 included) must add their fields to this Options struct and follow the Max<Thing> plus --max-<thing> naming. They must not declare a second Options type. The scheduler should order them after this story.
- Numbers are re-parsed into big.Rat on every comparison. The cost is bounded by the cap but repeated, so a large enum of numbers checked against many documents does redundant work. A follow-up can cache parsed values at compile time.
- When flag parsing fails, run.go's existing message names the first argument starting with '-', which can be a valid flag (for example `--max-number-digits=abc`). The exit code is still 2 with one line, but the message may mislead. The coder may narrow it with flags.Lookup if the change stays small.
- Raising the cap a long way (for example to 1e9) lets the caller opt into expensive big-number work. This is documented as the caller's choice, and the default remains safe.

Implementation constraints the coder must honor:

- Standard library only. go.mod stays without require lines.
- Do not change internal/keywords/registry.go, internal/eval/eval.go, internal/compile/compile.go or cmd/jsonschema/main.go. The instance type stays `any`, and its dynamic types are the jsonvalue tree described in data_model.
- Keep Decode(data []byte) (any, error) and its output shape (json.Number, map[string]any, []any). Every existing 178 test passes unchanged. The only exception is the single `{"type":"string"}` entry in jsonschema_test.go TestValidateSchemas, which becomes an unknown keyword.
- ADR-001 must not regress: the `len(stack)+1 > MaxDepth` check runs before descending into '[' or '{'; error text is offset plus line:column with a fixed phrase, never a JSON Pointer, member name or input bytes; readCapped in run.go stays the only input path.
- Check the number cap on the raw token text before any strconv or math/big call. Count mantissa digits. For the exponent, strip leading zeros and compare the digit-text length with len(strconv.Itoa(max)) before calling Atoi. Cover both positive and negative exponents, with tests for 1e999999999, 1e-999999999, an exponent longer than 20 digits, and leading-zero exponents that are within the cap.
- Do not rely on encoding/json for UTF-8 or surrogate validation. The pre-scan covers invalid bytes inside and outside strings, a lone high surrogate, a lone low surrogate, a high surrogate followed by a non-surrogate \u escape, and an escaped backslash before u (valid). A valid pair decodes to exactly one rune.
- Strip at most one leading BOM. Offsets in errors refer to the original input bytes, BOM included.
- Tests use DefaultMaxNumberDigits, DefaultMaxNumberExponent and MaxDepth by name: build inputs of length default and default+1, never literal numbers.
- Options: 0 means the default and a negative value makes CompileWithOptions return an error. CLI: an explicit flag value ≤ 0 is a usage error (one line, exit 2). The resolved limits are stored in the unexported Schema fields and never changed after Compile returns.
- Each keyword lives in its own file and registers from init() with the validation vocabulary URI. Bad keyword values (unknown type name, duplicate or empty type array, non-array enum) are compile errors, so Compile fails and the CLI exits 2. Keyword unit tests exercise the compile-error paths. The testdata files carry the valid/invalid cases through the existing spec_test.go runner, with `spec` citing validation §6.1.1/§6.1.2/§6.1.3 or RFC 8259 §6/§4.
- Keyword code must never put the instance into an error or message. Compile errors may name the keyword and the offending type name taken from the schema, but never document content.
- parse_errors_test.go covers CLI exit 2 for: broken schema file, broken document file and broken stdin; duplicate names at depth; invalid UTF-8; \uD800; trailing second value; an over-cap number with default flags; acceptance of that same number once the flags are raised; and a usage error for a flag value of 0 or a negative value. Stderr must never contain a planted secret.

- DB impact: no · API impact: yes · Migration needed: no

Breaking changes:

- The `type` keyword is now enforced. Schemas that used it were previously accepted as unknown-keyword no-ops, which changes 178's TestValidateSchemas expectation for {"type":"string"}.
- Input that 178 accepted is now rejected as broken JSON (exit 2): invalid UTF-8, unpaired \u surrogates, and numbers over the digit or exponent cap. A leading BOM is now accepted (178 rejected it).

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
