# Release notes — #2: Exact JSON value model, deep equality, and the type, enum and const keywords

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #2
- Status: proposed (pending Checkpoint 3 sign-off)

**⚠️ the release-agent has concerns.**

## What changed

This change delivers all eleven acceptance criteria: an exact JSON number model using math/big with no precision loss (1.0 equals 1, 12345678901234567890 exact); strict JSON parsing that rejects duplicate object names, invalid UTF-8, unpaired surrogates and trailing content as exit code 2; deep JSON equality that compares numbers by value and objects regardless of key order; and the type, enum and const keywords with spec-derived test cases. The factory reports successful build and tests, but notes that two documentation files were created outside declared scope, CLI diagnostic messages may mislead on flag validation, and the deterministic gate did not run the race detector.

- internal/jsonvalue: new package with exact Number type (text, neg, mant/big.Int, exp, ndig), DecodeLimited() strict decoder, Decode() unchanged with json.Number leaves for compatibility, Equal() deep equality, Cmp() for number comparison, IsInteger() for fractional-part check, Rat() for lazy big.Rat expansion
- internal/jsonvalue/decode.go: strict parsing enforces number digit/exponent caps on raw tokens before big-number work, rejects duplicate object member names at any depth, detects invalid UTF-8, unpaired surrogate escapes (\uD800 without \uDC00..\uDFFF pair), accepts single leading BOM, rejects trailing non-whitespace after the JSON value; errors name byte offset, line and column, never quote values
- internal/jsonvalue/prescan.go: separate UTF-8 and surrogate validation called before JSON token decoding, handles edge cases (runs of backslashes before quotes, \u escapes split at input end, lone high/low surrogates, literal UTF-8-encoded surrogates ED A0 80)
- internal/jsonvalue/limits.go: NumberLimits struct with MaxDigits and MaxExponent (default 10000 each); checkNumber() enforces caps on raw token text before any big-number computation, counts written mantissa digits, strips leading exponent zeros before length check
- internal/jsonvalue/number.go: Number is immutable, built only from token text after cap check, with fields unexported; IsInteger() is exp >= 0; Cmp() aligns mantissas only when adjusted exponent (ndig + exp) is equal; Equal() uses text fast path; Rat() increments atomic counter for test visibility
- jsonschema.go: added NumberLimits field to Schema, new CompileWithOptions(schema, Options) function, Compile() now calls CompileWithOptions(schema, Options{}), Validate() uses DecodeLimited with stored limits; errors from decode remain wrap with 'schema:' or 'document:' prefix
- numberlimits.go: Options struct with MaxNumberDigits, MaxNumberExponent; exported constants DefaultMaxNumberDigits, DefaultMaxNumberExponent; zero field means default, negative field is error
- cmd/jsonschema/run.go: added flags --max-number-digits, --max-number-exponent; validates that explicit flag values are > 0 (≤ 0 is usage error, exit 2); passes options to CompileWithOptions
- internal/keywords/validation/type.go: registers with validation vocabulary; compiles string or array of type names (rejects unknown names, empty array, duplicates as schema errors); evaluates against null, boolean, object, array, string, number, integer; integer matches exp >= 0
- internal/keywords/validation/enum.go: registers with validation vocabulary; compiles array (rejects non-array as schema error); evaluates by deep equality to any array member
- internal/keywords/validation/const.go: registers with validation vocabulary; compiles any JSON value; evaluates by deep equality to that value
- internal/speccases/runner.go: loads testdata/<area>/<keyword>.json case files with schema, instance, expected valid, spec citation; runs through public library; checks that failure messages do not contain instance serialisation
- testdata/validation/type.json, enum.json, const.json: spec-derived test groups citing validation §6.1.1/6.1.2/6.1.3 or RFC 8259; positive and negative cases (1 vs 1.0, object member order, nested arrays, null)
- docs/cli-contract.md, docs/limits.md: new documentation of CLI contract (exit codes, flags, broken JSON definitions) and number cap defaults (added outside declared task scope)

## How to verify

- Run `go test ./...` and `go test ./... -race` to verify build, tests and concurrent safety of shared big.Int reads
- Validate a document with a schema containing type, enum or const keywords: `echo '{"a":1.0}' | jsonschema validate '{"type":"integer"}' -` should exit 0
- Submit broken JSON: `echo '{"a":1,"a":2}' | jsonschema validate 'true' -` should exit 2 with offset and line:column location of the repeat, not the value '1' or '2'
- Submit over-cap number with defaults: `echo '1e10001' | jsonschema validate 'true' -` should exit 2 with 'number exponent' limit error; with `--max-number-exponent 10001` should exit 0
- Verify deep equality: `echo '1.0' | jsonschema validate '{"const":1}' -` should exit 0; `echo '{"b":2,"a":1}' | jsonschema validate '{"const":{"a":1,"b":2}}' -` should exit 0
- Check that the type keyword is enforced: compiling `{"type":"unknown"}` should fail with schema error; `{"x-unknown":"string"}` (unknown keyword) should compile and validate normally
- Verify number precision: `echo '12345678901234567890' | jsonschema validate 'true' -` should preserve all digits exactly (test with big.Rat conversion)
- Test integer detection: `echo '1.0' | jsonschema validate '{"type":"integer"}' -` should exit 0; `echo '1.5' | jsonschema validate '{"type":"integer"}' -` should exit 1

## Migration

Schemas using the type keyword are now enforced by default (story 178's test `{"type":"string"}` becomes a schema error and is changed to `{"x-unknown":"string"}`). Exit codes change: malformed input (duplicate names, invalid UTF-8, unpaired surrogates, trailing content, over-cap numbers) now exits 2 instead of 1, so exit code 1 now means only 'document is invalid'. Input formerly reported as 'invalid' due to number precision loss (e.g. 1.0 not matching enum [1] in old validators) now validates correctly. The Compile() function still exists and is unchanged in signature; CompileWithOptions() is additive. The Validate() signature is unchanged. The Schema type gains an internal limits field (compatible change, not visible to callers). Library code should use jsonvalue.DecodeLimited() with exact Number leaves; jsonvalue.Decode() still returns json.Number for backwards compatibility with story 178's tests but the library itself now uses DecodeLimited. Unknown keywords in schemas are still allowed and collected as annotations (unchanged); only invalid required vocabularies and known keywords with incorrect schema values are errors.

## Rollback

This change is in new files (internal/jsonvalue, three validation keyword files, speccases runner, testdata, docs) and modifications to jsonschema.go, cmd/jsonschema/run.go and one test entry. Rollback requires reverting these files and restoring the original jsonschema.go (Compile without options, Validate using Decode), cmd/jsonschema/run.go (no flags), and the test entry `{"type":"string"}` → `{"type":"string"}` in TestValidateSchemas. The exact value model and strict parsing are foundational for stories 181-185 (numeric and length keywords); rolling back breaks their dependency chain. A partial rollback (keeping the number model but not strict parsing) is not recommended as it creates inconsistency (schema accepts broken JSON, documents reject it).

## Known limitations

- Race detector not explicitly verified to pass on this change by the supplied evidence, though architecture specifies it must run and tests include concurrent Equal/Cmp calls
- CLI diagnostic messages may be unclear when a flag is recognized but has an invalid value (e.g. `--max-number-digits=abc`); flag validation is improved from the initial implementation but could be more precise
- The spec-case runner in internal/speccases may overlap with existing spec-test infrastructure, not fully isolated to this change
- Integer overflow in normalised exponent (ndig + exp) is a limit error, not a silent value; with absurdly raised caps (e.g. MaxExponent = math.MaxInt64) some expressions that should fit may still error due to exp arithmetic bounds

## Concerns

- Two documentation files (docs/cli-contract.md, docs/limits.md) were created outside the declared task scope; while they document the new CLI and limit features correctly, they should be reviewed for consistency with other docs and against the project's documentation standards
- The change to TestValidateSchemas is explicitly allowed by the story (acceptance criterion 8 says type is now a keyword), but it changes a test that exists in story 178; verify that story 178's test expectations are updated and no other tests depend on the old entry
- CLI flag validation: when an explicit flag value is ≤ 0, the error is 'flag -<name> must be greater than 0' (exit 2, one line); this is correct, but the architecture notes a risk that when flags.Parse fails for other reasons (e.g. unknown flag or bad syntax like '=abc'), the message may mislead by echoing the first '-' argument, which could be a valid flag. The code mitigates this with flags.Lookup, but edge cases (e.g. `--max-number-digits=abc` being parsed as an invalid syntax error) may still produce unclear messages
- Instance values are never included in error messages (a security requirement), confirmed by tests that check for a planted secret; this is met
- The breaking change to exit codes (broken JSON now exits 2 instead of 1) is intentional and documented, but existing scripts or tools that treat exit code 2 as distinct from 1 will need review
- Story 181 (multipleOf, maximum, etc.) depends on the Number model and will need to use modular arithmetic and Cmp instead of Rat() for performance; the architecture notes this, but future reviewers should verify that 181's implementation does not expand the exponent unnecessarily
