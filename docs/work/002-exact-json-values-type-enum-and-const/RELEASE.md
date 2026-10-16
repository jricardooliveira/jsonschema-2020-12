# Release notes — #2: Exact JSON value model, deep equality, and the type, enum and const keywords

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #2
- Status: proposed (pending Checkpoint 3 sign-off)

## What changed

This release adds exact JSON number handling with strict parsing, deep equality, and the first three validation keywords (type, enum, const). Schemas and documents are now parsed from an exact decimal number model that never loses precision on 0.1 or 12345678901234567890, and strict parsing rejects broken JSON—duplicate object keys, invalid UTF-8, unpaired Unicode surrogates, and trailing content—as errors with exit code 2, never as valid or invalid. Numbers can exceed a generous default cap; the caller raises it through library Options or CLI flags.

- internal/jsonvalue/: Exact number model (math/big.Int with exponent, no rounding), strict JSON decoder, UTF-8 and surrogate validation, deep equality, number digit and exponent cap enforcement
- internal/keywords/validation/type.go, enum.go, const.go: Three keywords registered from init(), type supports 7 JSON types with integer=zero fractional part, enum and const use deep equality
- jsonschema.go: CompileWithOptions(schema, Options) with MaxNumberDigits and MaxNumberExponent fields; Compile delegates with defaults
- numberlimits.go: Options struct, DefaultMaxNumberDigits and DefaultMaxNumberExponent constants (10000 each)
- cmd/jsonschema/run.go: Flags --max-number-digits and --max-number-exponent; both must be > 0 or usage error
- internal/speccases/runner.go: Shared test runner for spec-derived JSON cases; verifies failure messages never quote instances
- testdata/validation/type.json, enum.json, const.json: Spec-derived positive and negative test cases with 2020-12 validation citations
- docs/cli-contract.md, docs/limits.md: CLI interface contract and number cap defaults

## How to verify

- Big integer precision: `echo '12345678901234567890' | jsonschema validate schema.json` and check no digits were lost via library Rat()
- Duplicate key rejection: `echo '{"a":1,"a":2}' | jsonschema validate schema.json` exits 2 with error naming the second key's line and column
- Deep equality in const: Compile `{"const":{"a":1,"b":2}}` and validate `{"b":2,"a":1}` returns valid=true (member order ignored)
- Integer type matching: Compile `{"type":"integer"}`, validate 1.0 and -0 return valid=true, validate 1.5 returns valid=false
- Number cap enforcement: With default flags, validate `1e10001` exits 2 with limit error; with `--max-number-exponent=10001` it succeeds and value is kept exactly
- Strict UTF-8: `echo '"\xff"' | jsonschema validate schema.json` exits 2 with reason 'invalid UTF-8' naming the byte offset
- Unpaired surrogate: `echo '"\uD800"' | jsonschema validate schema.json` exits 2 with reason 'unpaired surrogate escape'

## Migration

Schemas and documents that were accepted before but contain invalid UTF-8, unpaired \u surrogates, duplicate object member names, or JSON content after the top-level value are now errors (exit code 2). A leading UTF-8 byte order mark is now accepted. Numbers exceeding the default cap (10000 digits or exponent magnitude) are now errors; raise the caps with Options.MaxNumberDigits / Options.MaxNumberExponent in the library or --max-number-digits / --max-number-exponent in the CLI (must be > 0). The type keyword is now enforced; schemas with unknown keywords collect them as annotations. Library callers must decode documents through DecodeLimited (with *jsonvalue.Number leaves), not Decode (which retains json.Number leaves for 178 compatibility). All existing 178 tests pass unchanged except TestValidateSchemas, where {"type":"string"} becomes {"x-unknown":"string"} to test unknown keywords.

## Rollback

Revert to the previous release if you must accept non-UTF-8 input, duplicate object member names, or numbers without size limits. Remove --max-number-digits and --max-number-exponent from scripts and CI automation; they do not exist in the previous release. Schemas and documents with broken JSON will be accepted by the previous release but must be fixed for this one.

## Known limitations

- type keyword only matches the seven standard JSON types (null, boolean, object, array, number, string, integer); custom types are not supported
- enum and const compare by deep value equality only; they do not accept schemas or perform assertions
- Number limit errors report location and the limit name (e.g., 'number exponent exceeds limit 10000') but never the number text itself
- Surrogate pair handling follows RFC 8259 §7 only; isolated surrogates (\uD800) and lone low surrogates (\uDC00) without high partners are rejected
- The 'integer' type matches any number with a zero fractional part mathematically (1.0, 1e2, 100e-2 are all integers); 1e-1000000 is not
- Integer precision is kept exactly but the ndig and exp fields are internal; the only safe operations are IsInteger(), Cmp and Equal, never math on the exponent

## Concerns

- Breaking change: Schemas and documents with non-UTF-8 bytes, duplicate object member names, unpaired Unicode surrogates, or trailing JSON values are now rejected as errors. Any existing input of these kinds must be fixed or the change must not be deployed.
- Error type change: LimitError is a new error type for exceeded caps (number digits, number exponent). Code that only checks for SyntaxError will miss these. Use errors.As(err, &jsonvalue.LimitError) to handle both parse and limit errors.
- Files outside declared task scope: docs/cli-contract.md and docs/limits.md document the new CLI flags and number cap defaults; they are necessary for users but were not in the original task files list. This is noted in the architecture review.
- Two new test files (prescan_test.go, decode_strict_test.go) in internal/jsonvalue demonstrate the strict parsing and UTF-8 validation behavior; they are not in the original scope but are required for the acceptance criteria.
- Later stories (180+) depend on this exact number model and deep equality. Any deviation from the Number format, Cmp behavior, or Equal semantics may break multipleOf, uniqueItems, unevaluatedProperties and const validation.
- The number cap defaults (10000 digits, 10000 exponent) are generous but hostile input with exponents near int64 limits (e.g., 1e-9223372036854775807) may still succeed at compile time and fail at validation time with a limit error; this is by design.
