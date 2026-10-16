# SPEC #4: Numeric validation keywords: multipleOf, maximum, exclusiveMaximum, minimum, exclusiveMinimum

- Date: 2026-10-16
- Type: feature
- Verdict: pass
- Status: proposed (pending Checkpoint 1 sign-off)

## Problem

The validator cannot yet check number ranges or divisibility. Schemas using multipleOf, maximum, exclusiveMaximum, minimum or exclusiveMinimum are not enforced. Rounding shortcuts would give wrong answers for decimals such as 0.3 against multipleOf 0.1, and for huge integers.

## Why this matters

The mandatory 2020-12 validation scope requires these five keywords with exact JSON number semantics: 1.0 is an integer, multipleOf is exact on decimals, and big numbers keep full precision. The brief chose exact math/big arithmetic for this.

## Acceptance criteria

- multipleOf, maximum, exclusiveMaximum, minimum and exclusiveMinimum each live in their own file under internal/keywords/validation/. Each registers itself from init() into the keyword registry with its name, the 2020-12 validation vocabulary URI, a compile function and an evaluate function. No central keyword list or switch is edited.
- multipleOf is exact using the existing exact number model (math/big). 0.3 is a multiple of 0.1, 0.35 is not a multiple of 0.1, 4.5 is a multiple of 1.5, and big integers such as 12345678901234567890 are checked exactly against integer and decimal divisors. multipleOf is valid when instance / divisor is an integer.
- maximum and minimum pass when the instance is less than or equal to (maximum) or greater than or equal to (minimum) the limit. exclusiveMaximum and exclusiveMinimum pass only when the instance is strictly less than or strictly greater than the limit. Each comparison is exact by mathematical value, including integers beyond 64 bits and long decimals that differ only in the last digit.
- 1.0 and 1 are the same number for every keyword: maximum 1 accepts 1.0, exclusiveMaximum 1 rejects 1.0, and multipleOf 1 accepts 2.0.
- The five keywords apply only to number instances. Strings, booleans, null, arrays and objects are valid against them regardless of the keyword value. This includes numeric-looking strings such as "5".
- A keyword value that is not a JSON number is a schema error (exit code 2), never 'invalid'. So is a multipleOf value that is not strictly greater than 0.
- A failing keyword is reported with its instance location and keyword location through the existing output, without quoting the instance value.
- This story adds no number-size cap, no library option and no CLI flag. It relies on the number model and number-size limit owned by story 179.
- Spec-derived positive and negative JSON test cases exist in testdata/validation/<keyword>.json, one file per keyword. Each cites the 2020-12 validation spec section (6.2.1 to 6.2.5). They include decimal multipleOf, big-number and exclusive-bound edge cases. They run through the shared table-driven runner from a <keyword>_test.go per keyword, with no network.

## Non-goals

- A number-size cap, option, flag or limit error (owned by story 179).
- Meta-schema validation of keyword values beyond what the keyword's own compile needs (story 189).
- Annotation collection (story 191).
- Any change to the JSON decoder or to cmd/jsonschema/main.go.

## Tasks

- **T1** — Shared exact-number helpers for the validation vocabulary
  - purpose: Give the five keywords one in-package helper. It extracts the exact big.Rat from a jsonvalue number, recognises non-number instances, and reads a keyword's numeric value at compile time with a schema error when the value is not a number. The helper reuses the internal/jsonvalue model without editing it.
  - scope: internal/keywords/validation/numeric.go
  - done when: Package builds. The helpers are used by the keyword files in T2 and T3 and are covered by their tests.
  - delivers: AC2, AC3, AC4, AC5, AC6
- **T2** — multipleOf keyword
  - purpose: Implement exact multipleOf with big.Rat division and an integer check. Compile rejects a non-number or non-positive divisor as a schema error. Register through init().
  - scope: internal/keywords/validation/multipleof.go, internal/keywords/validation/multipleof_test.go, testdata/validation/multipleOf.json
  - done when: go test ./internal/keywords/validation/ -run MultipleOf passes. Cases include 0.3/0.1 valid, 0.35/0.1 invalid, big integers, 2.0/1 valid, non-number instances ignored, and divisor 0 or -1 as a schema error.
  - depends on: T1
  - delivers: AC1, AC2, AC4, AC5, AC6, AC7, AC9
- **T3** — maximum and minimum keywords
  - purpose: Implement inclusive bounds with exact big.Rat comparison, each in its own file registered through init().
  - scope: internal/keywords/validation/maximum.go, internal/keywords/validation/maximum_test.go, testdata/validation/maximum.json, internal/keywords/validation/minimum.go, internal/keywords/validation/minimum_test.go, testdata/validation/minimum.json
  - done when: go test for Maximum and Minimum passes. Cases include the boundary as valid, 1.0 vs 1, big integers differing by 1, long decimals differing in the last digit, and non-numbers ignored.
  - depends on: T1
  - delivers: AC1, AC3, AC4, AC5, AC6, AC7, AC9
- **T4** — exclusiveMaximum and exclusiveMinimum keywords
  - purpose: Implement strict bounds with exact big.Rat comparison, each in its own file registered through init().
  - scope: internal/keywords/validation/exclusivemaximum.go, internal/keywords/validation/exclusivemaximum_test.go, testdata/validation/exclusiveMaximum.json, internal/keywords/validation/exclusiveminimum.go, internal/keywords/validation/exclusiveminimum_test.go, testdata/validation/exclusiveMinimum.json
  - done when: go test for ExclusiveMaximum and ExclusiveMinimum passes. Cases include the boundary as invalid, 1.0 at limit 1 as invalid, big-number and decimal edges, and non-numbers ignored.
  - depends on: T1
  - delivers: AC1, AC3, AC4, AC5, AC6, AC7, AC9
- **T5** — Verify no cap, no flag and the race-free full suite
  - purpose: Confirm that no option, flag or decoder change was added and that the whole module passes.
  - scope: internal/keywords/validation/multipleof_test.go
  - done when: go test -race ./... passes. A git diff shows no change to internal/jsonvalue/decode.go, cmd/jsonschema/main.go, go.mod or the keyword registry.
  - depends on: T2, T3, T4
  - delivers: AC8
