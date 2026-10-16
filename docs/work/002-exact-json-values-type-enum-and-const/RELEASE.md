# Release notes — #2: Exact JSON value model, deep equality, and the type, enum and const keywords

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #2
- Status: proposed (pending Checkpoint 3 sign-off)

**⚠️ the release-agent has concerns.**

## What changed

The story implements an exact JSON value model with lossless number handling via math/big, strict parsing that rejects broken JSON as errors with location-based reporting, and three new validation keywords (type, enum, const) using deep JSON equality. Numbers retain full precision (12345678901234567890, 0.1), 1.0 is correctly identified as an integer, and malformed input (duplicate object keys, invalid UTF-8, mismatched surrogates, trailing content) exits with code 2 rather than silent invalidity.

- internal/jsonvalue: new package with Number type (exact decimal via math/big), DecodeLimited strict decoder with prescan for UTF-8 and surrogates, deep Equal, and NumberLimits with digit and exponent caps
- internal/keywords/validation: new package with type.go, enum.go, const.go keywords and spec-derived test data (type.json, enum.json, const.json)
- internal/speccases: new shared runner for spec-derived testdata-driven test cases
- jsonschema.go: added Options struct with MaxNumberDigits/MaxNumberExponent, CompileWithOptions, Schema stores and applies limits to Validate
- numberlimits.go: new file with Options and DefaultMaxNumberDigits/DefaultMaxNumberExponent constants
- cmd/jsonschema/run.go: added --max-number-digits and --max-number-exponent flags with improved error messages
- cmd/jsonschema/parse_errors_test.go: new CLI contract tests for broken JSON, duplicate names, invalid UTF-8, surrogates, over-cap numbers and flag errors
- docs/cli-contract.md, docs/limits.md: new documentation files (outside declared scope) describing the CLI contract and default limits

## How to verify

- go test ./... passes all spec-derived test cases in testdata/validation
- go test -race ./internal/jsonvalue passes, validating concurrent reads on shared big.Int
- Compile([]byte(`{"a":1,"a":2}`)) returns an error with location (not a Result with valid=false)
- Compile([]byte(fmt.Sprintf(`1e%d`, DefaultMaxNumberExponent+1))) returns a *LimitError
- CompileWithOptions(schema, Options{MaxNumberExponent: DefaultMaxNumberExponent+1}).Validate([]byte(`1e10001`)) succeeds
- echo '{} {}' | jsonschema validate schema.json - exits with code 2 (trailing data error)
- Schema with {"type":"integer"} accepts 1.0 because exp >= 0 makes it an integer
- Schema with {"enum":[{"a":1,"b":2}]} accepts {"b":2,"a":1} because deep equality ignores member order

## Migration

none

## Rollback

Remove internal/jsonvalue, internal/keywords/validation, and internal/speccases packages. Revert jsonschema.go to Decode(data) instead of DecodeLimited. Remove numberlimits.go and numberlimits_test.go. Restore cmd/jsonschema/run.go to pre-story flag handling. Update TestValidateSchemas in jsonschema_test.go to use {"type":"string"} instead of {"x-unknown":"string"}.

## Known limitations

- Number digit cap counts all written digits in the raw token, including leading and trailing zeros, before normalisation. A number written 0.0001 with extra zeros counts as written digits.
- Raising the number-size caps far above the default allows expensive Cmp alignment work when comparing numbers with very different exponents. The default caps are safe.
- The BOM is stripped before parsing; error offsets include the BOM bytes in the original input. A 3-byte BOM shifts all offsets in error messages.
- Regex deviations from ECMA-262 are documented separately; this story only enforces the number cap, not regex limits.

## Concerns

- Number-limit error reporting: The tester found that *LimitError diagnostics may not include the required location (byte offset, line, column) in the error message. Verify the error string reports location the same way *SyntaxError does.
- Schema-side cap regression test: The test asserting over-cap numbers in the schema are rejected may pass for the wrong reason. Explicitly verify: (1) Compile rejects a schema with 1e(DefaultMaxNumberExponent+1), (2) the error is specifically a *LimitError not a syntax error, (3) CompileWithOptions with a raised cap accepts the same number.
- Race-detector confirmation: The concurrent test in equal_test.go should be run with `go test -race ./...` but the factory's gate did not report whether -race was executed. Confirm TestEqualConcurrent passes under the race detector.
- Documentation scope variance: docs/cli-contract.md and docs/limits.md were created but not declared in task scope. While correct, they represent scope variance. Future documentation changes should be explicitly planned.
- The parse_errors_test.go uses a secret string to verify error messages do not leak instance values. Ensure the secret is distinct from any fixed error-message text in the implementation.
