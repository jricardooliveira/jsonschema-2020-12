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
- `internal/jsonvalue/prescan.go`
- `internal/jsonvalue/limits.go`
- `internal/jsonvalue/number.go`
- `internal/jsonvalue/equal.go`
- `internal/jsonvalue/decode_test.go`
- `internal/jsonvalue/prescan_test.go`
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

- Files outside the reservation: jsonschema.go is the real compiler.go, cmd/jsonschema/run.go is where the flags live, and jsonschema_test.go needs the one accepted entry change. internal/jsonvalue/prescan.go and prescan_test.go are new files in the reserved internal/jsonvalue component. They keep the pre-scan separate so the differential test can target it. No other story holds them.
- Story 181 (multipleOf, maximum and the like) must work on the (neg, mant, exp) form. It should use Cmp for bounds and modular arithmetic on mantissas aligned to min(exp) for multipleOf, never Rat(), because Rat() of 1e10000 against 1e-10000 is cap-bounded but expensive. 181 will add a small read-only accessor to internal/jsonvalue/number.go. Alignment in multipleOf is bounded by |exp_a − exp_b| ≤ 2×MaxExponent + MaxDigits, which 181 must accept or short-circuit (for example, when exp_a ≥ exp_b, only mant_a × 10^(exp_a − exp_b) mod mant_b matters, and that can be reduced with modular exponentiation via big.Int.Exp(10, d, mant_b)).
- Two leaf shapes exist: Decode returns json.Number only to keep 178's TestDecodeValid unchanged, while the library uses DecodeLimited's *Number. Later stories (181-185) must decode through DecodeLimited and use Number.IsInteger, Cmp, Equal and (rarely) Rat, never json.Number or strconv.ParseFloat. Removing Decode's json.Number shape is a follow-up once 178's test may change.
- The memory test with 9 million tokens takes noticeable time and memory (an input of about 70 MB built in the test). It may skip under -short. TotalAlloc is coarse, so the bound is a generous per-byte constant rather than an exact figure.
- Differential test scope: the fuzz property uses encoding/json as the oracle only for inputs that both accept. Inputs that scanText rejects but encoding/json accepts are rejected by design: we are stricter.
- Options is defined in numberlimits.go while three other stories plan an options.go. They must add fields to this Options struct with the Max<Thing> / --max-<thing> naming and not declare a second type, so the scheduler should order them after this story.
- When flag parsing fails, run.go's existing message names the first argument starting with '-', which can be a valid flag (for example `--max-number-digits=abc`). The exit code is still 2, but the message may mislead. The coder may narrow it with flags.Lookup if the change stays small.
- Raising the cap a long way lets the caller opt into expensive big-number work in Cmp alignment and Rat(). That is the caller's choice, and the default stays safe.

Implementation constraints the coder must honor:

- Standard library only. go.mod stays without require lines.
- Do not change internal/keywords/registry.go, internal/eval/eval.go, internal/compile/compile.go or cmd/jsonschema/main.go. The instance type stays `any`.
- Every existing 178 test passes unchanged. The single exception is the `{"type":"string"}` entry in TestValidateSchemas, which becomes `{"x-unknown":"string"}`. Decode keeps returning json.Number leaves; jsonschema.go calls only DecodeLimited.
- Decode and DecodeLimited share one internal decode(data, lim, exact) routine. Every strict check is identical in both paths; only the number leaf type differs.
- Number is a normalised compact decimal: an unexported big.Int mant with no trailing zeros, an int64 exp, a neg flag, a cached ndig and the original text. Zero has exactly one form, so -0 is not negative. newNumber strips zeros by string slicing before mant.SetString and never uses big.Rat or big.Int.Exp. No *big.Rat exists anywhere in the decode path.
- IsInteger is `exp >= 0`. Equal on two *Number returns true at once when their texts match, and otherwise compares neg, exp and mant without allocating. Cmp compares the sign, then adj = ndig + exp, and multiplies by a power of ten only when adj is equal (then |exp difference| ≤ MaxDigits). None of the three calls Rat().
- Rat() is the only function that builds a big.Rat or raises 10 to a power for an exponent. It returns a fresh value every call, caches nothing, and increments ratBuilds (sync/atomic). The shared big.Int is only read. Number has no exported fields and no setters.
- number_test.go must show the following. (1) Decoding 1e10000 and 1e-10000 (built from DefaultMaxNumberExponent, not literals), and a cap-sized digit string, leaves ratBuilds at 0. Each Number's retained size (ndig, plus len(text)) is bounded by its text length. (2) A document of 9,000,000 such tokens, for example `[1e10000,1e-10000,...]`, decodes through DecodeLimited with ratBuilds still 0 and runtime.MemStats TotalAlloc growth below a small constant × len(input). The constant is a named test value, and the test may skip under -short. (3) Equal, Cmp and IsInteger called many times (for example 10000) on cap-scale values leave ratBuilds at 0, including Equal on a cap-sized number written two ways (different text, same value: 1e10000 vs 10e9999, 0.1 vs 1e-1). (4) Correctness across very different scales: 1e10000 > 9…9 (DefaultMaxNumberDigits nines), 1e-10000 < 1e-9999, -1e10000 < -1, 1e-10000 > 0 > -1e-10000, 0 == -0 == 0.0 == 0e5, 1 == 1.0 == 1e0 == 10e-1, 100 == 1e2, 0.1 != 0.10000000000000001, 12345678901234567890 exact, and IsInteger true for 1.0, 1e2, -0 and 100e-2 but false for 1e-10000 and 0.1. (5) Rat() matches big.Rat.SetString on the same text for a sample of values, and increments ratBuilds.
- equal_test.go runs Equal and Cmp from several goroutines on shared values, so `go test -race` covers concurrent reads of the shared big.Int.
- ADR-001 must not regress: the `len(stack)+1 > MaxDepth` check runs before descending into '[' or '{'. Error text is offset plus line:column with a fixed phrase, never a JSON Pointer, member name or input bytes. readCapped stays the only input path in run.go.
- Check the number cap on the raw token text before newNumber. Count mantissa digits. For the exponent, strip leading zeros and compare the digit-text length with len(strconv.Itoa(max)) before parsing it. Test 1e999999999, 1e-999999999, an exponent of more than 20 digits, and leading-zero exponents that are within the cap. An int64 overflow in exp arithmetic is a "number exponent" LimitError.
- prescan.go: scanText(data, onString) is the only UTF-8 and surrogate check, and decode() calls it with nil. It must handle runs of backslashes (an even count before a quote ends the string, an odd count does not), \u escapes split at the end of the input, a lone high surrogate, a lone low surrogate, a high surrogate followed by a non-surrogate \u, a high surrogate followed by an escaped backslash and then `uDC00`, invalid bytes inside and outside strings, and UTF-8-encoded surrogates (ED A0 80), which are invalid UTF-8.
- prescan_test.go has a table test and `func FuzzScanTextAgreesWithEncodingJSON(f *testing.F)`. Its seeds are added with f.Add only (no binary corpus files) and cover every case in the previous constraint plus valid pairs, a literal U+FFFD and \uFFFD. The property is checked whenever both scanText and json.Valid accept: the reported spans, decoded with json.Unmarshal, match the Decoder's string tokens in count, order and value; each one is valid UTF-8; and strings.Count(s, "\uFFFD") equals the reported fffd. A second assertion: when Decode succeeds, every string in its tree satisfies the same FFFD invariant.
- Tests name DefaultMaxNumberDigits, DefaultMaxNumberExponent and MaxDepth instead of writing the numbers: build inputs of length default and default+1.
- Strip at most one leading BOM. Error offsets refer to the original input bytes, BOM included.
- Options: 0 means the default, and a negative value makes CompileWithOptions return an error. CLI: an explicit flag value ≤ 0 is a usage error (one line, exit 2). The resolved limits sit in unexported Schema fields and never change after Compile returns.
- Each keyword lives in its own file and registers from init() with the validation vocabulary URI. Bad keyword values (an unknown type name, a duplicate or empty type array, a non-array enum) are compile errors. testdata cases cite validation §6.1.1, §6.1.2 or §6.1.3, or RFC 8259 §6 or §4.
- Keyword code never puts the instance into an error or message. type, enum and const use only IsInteger and Equal, never Rat().
- parse_errors_test.go covers CLI exit 2 for: a broken schema, a broken document file and broken stdin; duplicate names at depth; invalid UTF-8; \uD800; a trailing second value; an over-cap number with default flags; acceptance of that same number once the flags are raised; and a usage error for a flag value of 0 or a negative value. Stderr never contains a planted secret.

- DB impact: no · API impact: yes · Migration needed: no

Breaking changes:

- The `type` keyword is now enforced, which changes 178's TestValidateSchemas expectation for {"type":"string"}; that entry becomes an unknown keyword.
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
