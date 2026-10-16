# SPEC #2: Exact JSON value model, deep equality, and the type, enum and const keywords

- Date: 2026-10-16
- Type: feature
- Verdict: pass
- Status: passed gate 1 with nothing for the operator to decide (run #135, 2026-10-16 16:50 UTC)

## Problem

The validator must read schemas and documents without losing number precision, and it must reject broken or ambiguous JSON as an error, never report it as 'invalid'. The first value keywords (type, enum, const) need an exact value model and deep JSON equality to give correct answers.

## Why this matters

Later keywords depend on an exact, unambiguous view of the input: multipleOf, maximum, uniqueItems and the length rules all build on it. The operator treats input as possibly hostile, so damaged text, repeated names, hidden trailing values and huge numbers must stop with a clear error (exit code 2). They must never be guessed at or allowed to hang the tool.

## Acceptance criteria

- Schemas and documents are parsed into an exact JSON value model built only on the standard library. Numbers keep their written value exactly through math/big: 12345678901234567890 and 0.1 lose no precision, and 1.0, 1e2 and -0 are integers by mathematical value.
- Broken JSON in a schema or a document is an error, never 'invalid'. The library returns a parse error that names the location, and the CLI exits with code 2.
- An object that repeats the same member name (for example {"age": 5, "age": 9}), in a schema or a document, at any depth, is broken JSON (exit code 2). The error names the location of the repeat.
- Text that is not valid UTF-8, and a \u escape that leaves a surrogate without its partner (for example "\uD800"), is broken JSON (exit code 2). The error names the location of the bad text. Valid surrogate pairs decode to one code point.
- A single leading UTF-8 byte order mark is accepted. Whitespace around the value is accepted. Any non-whitespace content after the single top-level JSON value, including a second value, is broken JSON (exit code 2).
- Numbers have a default cap on digit count and exponent size, set far above real-world values and exposed as named defaults. A number over the cap is a limit error (exit code 2), never valid or invalid. The caller can raise the cap through a library option and a CLI flag. Exceeding the cap never triggers an unbounded big-number computation.
- Deep JSON equality is provided. Numbers compare by mathematical value (1 equals 1.0 and 1e0). Objects compare regardless of member order. Arrays compare element by element, in order. Values of different JSON types are never equal.
- The 'type' keyword, in its own file registered through the keyword registry, accepts a string or an array of unique type names. It supports null, boolean, object, array, number, string and integer. 'integer' matches any number with a zero fractional part, and 'number' matches every integer.
- The 'enum' keyword, in its own file registered through the keyword registry, passes only when the instance is deep-equal to at least one listed value.
- The 'const' keyword, in its own file registered through the keyword registry, passes only when the instance is deep-equal to its value. This includes null, objects in any member order, and numbers such as 1 against 1.0.
- Positive and negative tests are stored as spec-derived JSON cases in testdata/validation/type.json, enum.json and const.json, plus parser and equality tests. Each case cites the 2020-12 validation or RFC 8259 section it derives from. Failure messages from type, enum and const never include instance values. All tests pass with go test ./... and use no network.

## Non-goals

- Numeric, string, object, array and applicator keywords (stories 181-185)
- The 2020-12 output formats beyond the skeleton's minimal result (story 180)
- Other limits (nesting depth, input size, recursion) and the race-detector immutability proof (story 193). Only the number-size cap is added here, as the operator's clarification requires.
- Any third-party dependency, or any copy of or reference to the official JSON-Schema-Test-Suite

## Tasks

- **T1** — Exact JSON value model and strict decoder
  - purpose: Build internal/jsonvalue with a Value type for the seven JSON kinds and an exact Number backed by math/big. Add a strict decoder that rejects duplicate names, invalid UTF-8, lone surrogates and trailing content, accepts a leading BOM, and enforces the number digit and exponent cap before any big-number work. Errors name the location.
  - scope: internal/jsonvalue/value.go, internal/jsonvalue/number.go, internal/jsonvalue/decode.go, internal/jsonvalue/limits.go, internal/jsonvalue/number_test.go, internal/jsonvalue/decode_test.go
  - done when: go test ./internal/jsonvalue passes. It covers big integers, 0.1, 1.0 as an integer, duplicate names at depth, invalid UTF-8, \uD800, valid surrogate pairs, BOM, trailing values, and the number cap using the exported default (not a repeated literal).
  - delivers: AC1, AC3, AC4, AC5, AC6
- **T2** — Deep JSON equality
  - purpose: Add Equal(a, b) to internal/jsonvalue. Numbers compare by exact mathematical value, objects ignore member order, arrays compare in order, and different kinds are never equal.
  - scope: internal/jsonvalue/equal.go, internal/jsonvalue/equal_test.go
  - done when: go test ./internal/jsonvalue -run Equal passes, with positive and negative cases for every kind.
  - depends on: T1
  - delivers: AC7
- **T3** — Load schemas and documents through the strict decoder
  - purpose: Make the root library load schemas and check documents through jsonvalue: parse errors and number-cap errors come back as errors, never results. Add a library option that raises the number cap, and a CLI flag mapped to it. Broken JSON and an exceeded cap give exit code 2. Adapt the keyword evaluate signature to the jsonvalue.Value instance type if the skeleton used another type.
  - scope: compiler.go, numberlimits.go, internal/keywords/registry.go, internal/eval/eval.go, cmd/jsonschema/main.go, numberlimits_test.go, cmd/jsonschema/parse_errors_test.go
  - done when: Library tests show that a broken schema, a broken document, a duplicate name and an over-cap number all return errors, and that the raised cap accepts a large number. CLI contract tests show exit code 2 for broken JSON in the schema file, a document file and stdin, and that the cap flag works.
  - depends on: T1
  - delivers: AC2, AC6
- **T4** — type keyword
  - purpose: Add internal/keywords/validation/type.go, registered from init() with the validation vocabulary URI. It compiles the string or array form (an invalid type name is a schema error) and evaluates all seven types, with integer meaning a zero fractional part. Failure messages name the expected types only.
  - scope: internal/keywords/validation/type.go, internal/keywords/validation/type_test.go, testdata/validation/type.json
  - done when: go test ./internal/keywords/validation -run Type passes. It covers the testdata cases citing validation §6.1.1, including 1.0 as an integer and the array form.
  - depends on: T3
  - delivers: AC8, AC11
- **T5** — enum and const keywords
  - purpose: Add internal/keywords/validation/enum.go and const.go, each registered from init(). Both use jsonvalue.Equal. Failure messages never quote the instance.
  - scope: internal/keywords/validation/enum.go, internal/keywords/validation/const.go, internal/keywords/validation/enum_test.go, internal/keywords/validation/const_test.go, testdata/validation/enum.json, testdata/validation/const.json
  - done when: go test ./internal/keywords/validation -run 'Enum|Const' passes. It covers the testdata cases citing validation §6.1.2 and §6.1.3, including object member order, 1 vs 1.0, null, and nested arrays.
  - depends on: T2, T3
  - delivers: AC9, AC10, AC11
- **T6** — Shared spec-case runner
  - purpose: Add the table-driven runner that loads testdata/<area>/<keyword>.json cases (schema, instance, expected valid, spec citation) and runs them through the public library. It checks that no failure message contains the instance's serialised value.
  - scope: internal/speccases/runner.go, internal/speccases/runner_test.go
  - done when: The type, enum and const tests run through the runner. A case without a citation fails the runner. go test ./... passes.
  - depends on: T3
  - delivers: AC11
