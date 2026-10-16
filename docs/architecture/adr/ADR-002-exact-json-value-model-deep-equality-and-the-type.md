# ADR #2: Exact JSON value model, deep equality, and the type, enum and const keywords

- Date: 2026-10-16
- Story: #2
- Status: proposed (pending architecture sign-off)

## Decision
Revision of the previous design. It maps the assumed file names to what story 178 actually built and settles the five review points. File mapping: compiler.go is jsonschema.go. The CLI's flags and error handling live in cmd/jsonschema/run.go, so main.go (a crowded file) is not touched. The shared spec-case runner already exists in spec_test.go: it globs testdata/*/*.json and fails a group that has no `spec` citation. T6 therefore reduces to adding the three testdata files, and internal/speccases is dropped.

One review point needed a decision: the evaluate signature. Changing EvaluateFunc's instance type to a new jsonvalue.Value struct would break existing 178 tests. dispatch_test.go uses `func(*keywords.EvalContext, any, any)` literals and `c.(bool)`. TestDecodeValid compares Decode's output against map[string]any and json.Number with reflect.DeepEqual. Review point 5 (every 178 test keeps passing unchanged) overrides an allowed-but-optional change. So the value model is the tree that ADR-001's Decode already returns, made exact and strict: nil, bool, string (always valid UTF-8), json.Number (the number's exact written text, already checked against the cap), []any, and map[string]any (duplicate names rejected at decode, so the map is never ambiguous). The registry, internal/eval/eval.go (crowded) and internal/compile/compile.go (crowded) stay unchanged. Exactness comes from internal/jsonvalue/number.go: Rat(json.Number) *big.Rat calls big.Rat.SetString on the checked text and is exact for decimals and exponents, and IsInteger(json.Number) is Rat(n).IsInt(), so 1.0, 1e2 and -0 are integers. ponytail: numbers are re-parsed on each call, with a bound because the cap is enforced at decode time. Cache a *big.Rat at compile time in the numeric keywords (story 181) if profiling shows a need.

Strict decoder (decode.go, extending ADR-001's Decode and keeping MaxDepth checked before descending):
(a) Strip one leading UTF-8 BOM (EF BB BF). All reported offsets still count bytes of the original input. A second BOM is a syntax error.
(b) Before tokenising, one linear pre-scan checks the text. It runs utf8.DecodeRune over the whole input, so the first invalid byte gives a SyntaxError with Reason "invalid UTF-8" at that offset. Inside string literals it tracks backslash escapes (so `\\uD800` stays plain text) and checks \u escapes: a high surrogate D800-DBFF must be followed directly by a \u low surrogate DC00-DFFF, and a lone low surrogate is an error. Either failure is a SyntaxError with Reason "unpaired surrogate escape" at the offset of the backslash. Valid pairs are then decoded by encoding/json into one code point.
(c) Each json.Number token is checked on its raw text before any math/big work. The mantissa digit count is all digits in the integer and fraction parts, sign and '.' excluded. The exponent check first strips leading zeros from the exponent digit text and compares that text's length with len(strconv.Itoa(max)). Only when the length fits is strconv.Atoi called, and then |exp| is compared with max, so 1e999999999999999999999 and 1e-999999999 are both rejected with no wrap-around. Going over the cap returns ADR-001's *LimitError with Limit "number digits" or "number exponent", Max and Offset. On review point 1: the over-cap case is a *LimitError, not a new SyntaxError Reason, because the acceptance criteria call it a limit error. The only new error values are the two SyntaxError Reasons above and the two LimitError Limit names. There is no ParseError type.
(d) Trailing content keeps ADR-001's "trailing data after value".
Error text stays as offset plus line:column with a fixed phrase. It never includes a JSON Pointer or a member name.

Limits (internal/jsonvalue/limits.go): `type NumberLimits struct{ MaxDigits, MaxExponent int }`, `const DefaultMaxNumberDigits = 10000` and `const DefaultMaxNumberExponent = 10000`, far above float64's 17 digits and exponent of 308, but small enough that 10^10000 is about 4 KB. `DecodeLimited(data []byte, lim NumberLimits) (any, error)` is the strict decoder, and `Decode(data)` is `DecodeLimited(data, defaults)`, so every existing caller and test keeps working.

Library (numberlimits.go): `type Options struct{ MaxNumberDigits, MaxNumberExponent int }`, with root constants DefaultMaxNumberDigits and DefaultMaxNumberExponent re-exported from jsonvalue. A zero field means the default. A negative field makes Compile return an error (exit 2) and never means 'no limit'. jsonschema.go gains `CompileWithOptions(schema []byte, opts Options) (*Schema, error)`. Compile(schema) calls it with Options{}. The resolved NumberLimits are stored in the unexported, immutable Schema struct and applied when decoding the schema and every document passed to Validate.

CLI (run.go): flags.Int("max-number-digits", jsonschema.DefaultMaxNumberDigits) and flags.Int("max-number-exponent", ...). Go's flag package accepts both -x and --x. A value of zero or less, which can only be given explicitly, is a usage error: one stderr line and exit 2. The values are passed to CompileWithOptions. Broken JSON and an over-cap number already reach exit 2 through run.go's existing `fail(err.Error())` path.

Equality (equal.go): `Equal(a, b any) bool` does a type switch over the six dynamic types. Numbers compare with Rat Cmp, arrays element by element in order, and objects by equal size plus every key present and equal, so member order does not matter. Values of different kinds are never equal.

Keywords: internal/keywords/validation/type.go, enum.go and const.go each register from init() with Vocabulary "https://json-schema.org/draft/2020-12/vocab/validation". all.go already blank-imports this package, so no hub is edited.
- type: compile accepts a string, or a non-empty array of unique strings, from {null, boolean, object, array, number, string, integer}. Anything else is a schema error. Evaluate maps the instance's Go type to its JSON type: a json.Number matches "number", and also "integer" when IsInteger is true.
- enum: compile requires an array. Evaluate passes when Equal holds for any element.
- const: compile stores the value, null included. Evaluate is Equal.
Evaluate returns only a bool, and keyword messages arrive with story 180, so no instance value can leak. Constraint: when messages are added, they name only the expected types or the keyword.

One existing test must change. jsonschema_test.go's TestValidateSchemas expects `{"type":"string"}` to accept every document, because type was an unknown keyword in 178. Now that type is implemented, that expectation is wrong. The one-entry fix is to replace it with another unknown keyword (`{"x-unknown":"string"}`). Nothing else in that file changes.

## Affected modules
- internal/jsonvalue/decode.go
- internal/jsonvalue/limits.go
- internal/jsonvalue/number.go
- internal/jsonvalue/equal.go
- internal/jsonvalue/decode_test.go
- internal/jsonvalue/number_test.go
- internal/jsonvalue/equal_test.go
- jsonschema.go
- jsonschema_test.go
- numberlimits.go
- numberlimits_test.go
- cmd/jsonschema/run.go
- cmd/jsonschema/parse_errors_test.go
- internal/keywords/validation/type.go
- internal/keywords/validation/enum.go
- internal/keywords/validation/const.go
- internal/keywords/validation/type_test.go
- internal/keywords/validation/enum_test.go
- internal/keywords/validation/const_test.go
- testdata/validation/type.json
- testdata/validation/enum.json
- testdata/validation/const.json

## Data / API impact
- DB impact: no
- API impact: yes
- Migration needed: no

## Constraints for implementation
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

## Risks
- Files outside the reservation: jsonschema.go is the real 'compiler.go'. cmd/jsonschema/run.go is where the flags live (main.go is never touched). jsonschema_test.go needs the one-entry fix, because implementing `type` makes 178's `{"type":"string"}` is-ignored expectation wrong; this is unavoidable and the only change to a 178 test. Reserved files that are dropped: compiler.go, internal/keywords/registry.go, internal/eval/eval.go, cmd/jsonschema/main.go, internal/jsonvalue/value.go and internal/speccases/* (spec_test.go already is the shared runner).
- Instead of the previous design's signature change to a jsonvalue.Value struct, the decoded `any` tree is the value model, because the struct would break 178's dispatch, compile, eval and decode tests (review point 5). Later keyword stories (181-185) must use jsonvalue.Rat, IsInteger and Equal, and never compare json.Number strings or call strconv.ParseFloat.
- Options is defined in numberlimits.go, while three other stories plan an options.go. Those stories (story 193 included) must add their fields to this Options struct and follow the Max<Thing> plus --max-<thing> naming. They must not declare a second Options type. The scheduler should order them after this story.
- Numbers are re-parsed into big.Rat on every comparison. The cost is bounded by the cap but repeated, so a large enum of numbers checked against many documents does redundant work. A follow-up can cache parsed values at compile time.
- When flag parsing fails, run.go's existing message names the first argument starting with '-', which can be a valid flag (for example `--max-number-digits=abc`). The exit code is still 2 with one line, but the message may mislead. The coder may narrow it with flags.Lookup if the change stays small.
- Raising the cap a long way (for example to 1e9) lets the caller opt into expensive big-number work. This is documented as the caller's choice, and the default remains safe.

## Breaking changes
- The `type` keyword is now enforced. Schemas that used it were previously accepted as unknown-keyword no-ops, which changes 178's TestValidateSchemas expectation for {"type":"string"}.
- Input that 178 accepted is now rejected as broken JSON (exit 2): invalid UTF-8, unpaired \u surrogates, and numbers over the digit or exponent cap. A leading BOM is now accepted (178 rejected it).

## Sensitivity
- security
