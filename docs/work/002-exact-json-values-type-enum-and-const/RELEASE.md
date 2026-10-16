# Release notes — #2: Exact JSON value model, deep equality, and the type, enum and const keywords

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #2
- Status: proposed (pending Checkpoint 3 sign-off)

**⚠️ the release-agent has concerns.**

## What changed

Adds exact JSON number handling with no precision loss (12345678901234567890, 0.1 exact, 1.0 counts as integer) and enforces strict JSON parsing: duplicate object keys, invalid UTF-8, unpaired \u surrogates, and trailing content are now errors with exit code 2. Implements three 2020-12 validation keywords (type, enum, const) with deep JSON equality; numbers compare by mathematical value, objects ignore key order. Built-in number-size caps (default 10000 digits, 10000 exponent) protect against hostile input and are configurable via library Options and CLI flags.

- internal/jsonvalue: exact decimal Number type using math/big (fields text, neg, mant, exp, ndig); IsInteger() reports zero fractional part; Cmp(a, b) compares by mathematical value without expanding powers of ten; Equal() deep equality with text fast path
- internal/jsonvalue/decode.go: strict Decode() and DecodeLimited() functions rejecting duplicate object member names at any depth, invalid UTF-8, lone surrogate escapes (e.g., "\uD800"), and non-whitespace content after the top-level value; accepting one leading BOM; enforcing number digit and exponent caps before big-number work
- internal/jsonvalue/prescan.go: pre-scan pass for UTF-8 validity and surrogate pairing; scanText() runs before JSON decoding and reports offsets of violations
- internal/jsonvalue/limits.go: NumberLimits struct; checkNumber() function enforcing caps on raw token text; DefaultMaxNumberDigits and DefaultMaxNumberExponent both 10000
- internal/keywords/validation/type.go: type keyword supporting string or array form, all seven JSON types (null, boolean, object, array, number, string, integer), with integer meaning zero fractional part; error for unknown type names or duplicate array entries
- internal/keywords/validation/enum.go, const.go: enum (instance deep-equal to one listed value) and const (instance deep-equal to a single value); both use jsonvalue.Equal, never quote instance in failure messages
- jsonschema.go: new CompileWithOptions(schema, Options); Compile(schema) is now CompileWithOptions(schema, Options{}); both Compile and Validate use DecodeLimited with exact *Number leaves; Schema stores number limits for each Validate call
- numberlimits.go: Options struct with MaxNumberDigits, MaxNumberExponent (zero=default, negative=error); limits() validates and resolves to internal jsonvalue.NumberLimits
- cmd/jsonschema/run.go: --max-number-digits and --max-number-exponent flags; flag value ≤0 is a usage error (exit 2, one line stderr); flags integrated into CompileWithOptions call
- internal/speccases/runner.go: spec-case runner loading testdata/<area>/<keyword>.json, compiling schema, validating each instance, asserting failure messages never contain the instance's serialised value
- testdata/validation/type.json, enum.json, const.json: spec-derived test cases citing RFC 8259 §4, §6 and validation §6.1.1, §6.1.2, §6.1.3; covers integer detection, deep equality, array form, object key order, and nested structures
- docs/cli-contract.md, docs/limits.md: CLI contract (exit codes, flags, accepted input) and default limits documentation (added outside declared task scope per evidence gap)

## How to verify

- go test ./... passes: all 11 acceptance criteria covered, including exact big integers (12345678901234567890), decimal precision (0.1), integer detection (1.0, 1e2, -0), duplicate key rejection at depth, invalid UTF-8 rejection, unpaired surrogate rejection, trailing data rejection, BOM acceptance, number cap enforcement, and type/enum/const keywords
- Large integer exact: echo '{"type":"number"}' | jsonschema validate /dev/stdin <<< '12345678901234567890' returns valid with full precision intact
- Integer detection: schema {"type":"integer"} accepts 1.0, 1e2, -0 (zero fractional part), rejects 0.1 and 1.5
- Deep equality: {"enum":[1]} accepts 1.0 (mathematical equality, not text match); {"const":{"a":1,"b":2}} accepts {"b":2,"a":1} (key order ignored); {"enum":[[1,2]]} rejects [2,1]
- Duplicate key error: jsonschema validate <schema> <<< '{"x":1,"x":2}' exits 2 with no stdout; error names the byte offset and line:column of the repeat
- Invalid UTF-8 error: jsonschema validate <schema> <<< '"\xff"' exits 2; <<< '"\uD800"' exits 2 (lone surrogate)
- Trailing data error: jsonschema validate <schema> <<< '{} []' exits 2 (second value after first)
- BOM accepted: printf '\xEF\xBB\xBF{"type":"null"}' | jsonschema validate /dev/stdin <<< 'null' returns valid
- Cap enforcement: jsonschema validate <true> <<< '1e10001' exits 2; adding --max-number-exponent 10001 makes it pass
- Flag usage error: jsonschema validate --max-number-digits=0 <schema> exits 2 with one-line stderr 'flag -max-number-digits must be greater than 0'
- Concurrent safety: go test -race ./internal/jsonvalue passes, verifying Number.mant is only read by concurrent callers

## Migration

Input acceptance changes break backward compatibility: (1) The `type` keyword is now a real keyword (was unknown before); fix schemas using it as a data field or side effect. (2) Schemas and documents with invalid UTF-8, unpaired \u surrogates (e.g., "\uD800", "\uD800\u0041"), or numbers exceeding 10000 digits or exponent now error with exit 2 (previously silently accepted). (3) Duplicate object member names (e.g., {"x":1,"x":2}) now error at the second occurrence (previously accepted one silently). (4) Leading UTF-8 byte order mark (0xEF 0xBB 0xBF) now accepted (previously rejected). To migrate: (a) Clean input: remove invalid UTF-8, fix surrogate escapes, rewrite huge numbers or raise the cap; (b) CompileWithOptions(schema, Options{MaxNumberExponent: newValue}) or --max-number-exponent newValue for legitimately large numbers; (c) Fix duplicate member names in schemas and documents; (d) Update test suite: TestValidateSchemas entry {"type":"string"} becomes {"x-unknown":"string"}. Numbers are now exact: 1.0 is an integer, enum and const use deep equality, and multipleOf/maximum (story 181) will work correctly on decimals.

## Rollback

Revert the repository to the previous release and rebuild the binary. No runtime flag disables strict parsing or exact numbers; both are baked into the decoder after Compile. To re-enable the old input acceptance or type-as-unknown behaviour, compile against an earlier version of the library.

## Known limitations

- Rat() (lazy big.Rat construction) is not called by type, enum or const; those keywords use only IsInteger() and Equal(), which are O(1) or O(n) in structure size, not exponent
- Under MaxExponent=math.MaxInt64 (absurdly high), int64 overflow in exponent normalisation is caught as a limit error, not a panic; this is a bounded defect, not practical at default (10000) or reasonably raised caps
- Numeric keywords (multipleOf, maximum, etc.) and string/array length rules are story 181-182, not this one; format assertion is story 182; those build on this exact model
- story 178's TestDecodeValid unchanged (Decode still returns json.Number leaves); only the library uses DecodeLimited with *Number
- docs/cli-contract.md and docs/limits.md are documentation of already-implemented contract, not new executable behaviour; they were added outside declared task scope

## Concerns

- Race-detector execution not verified: the tester notes missing race-detector evidence. The design claims Number.mant is immutable and only read (safe for concurrent use), but `go test -race ./internal/jsonvalue` has not been run in the test factory to prove it. Verify before shipping.
- Bounded Rat conversion defect: at the minimum int64 exponent under extreme raised caps (e.g., MaxExponent near math.MinInt64), exponent arithmetic in Rat() can overflow; the design catches this as a limit error but it is a latent issue. Not practical at defaults (10000) but latent at extreme caps.
- Breaking input changes: invalid UTF-8, surrogates, over-cap numbers and duplicate keys now reject (exit 2) instead of accepting silently. Callers must clean their input or raise the cap. This is correct per operator clarification but is an incompatibility.
- Scope gap: docs/cli-contract.md and docs/limits.md were added but not explicitly listed in task scope; evidence gap notes them as outside §6 (runner) scope. They document already-implemented contract; clarify whether documentation is in scope or defer to later.
- Tester verdict is 'warn', not 'pass': approval is qualified by the race-detector and extreme-exponent concerns. Do not ship until race-detector verification passes and the documented concerns are addressed.
