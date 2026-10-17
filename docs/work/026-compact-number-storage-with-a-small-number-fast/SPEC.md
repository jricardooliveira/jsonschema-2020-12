# SPEC #26: Compact exact number storage in decoded documents (fix 5)

- Date: 2026-10-17
- Type: performance
- Verdict: pass
- Status: proposed (pending Checkpoint 1 sign-off)

## Problem

Every JSON number in a decoded document is stored with math/big, which allocates per number. This is a major part of the roughly 60x memory growth over file size seen on large documents. For example, a 5.5 MB document reaches about 330 MB peak in the CLI.

## Why this matters

This is the operator's fix 5 in the performance round, aimed at roughly matching tekuri on memory. Answers, reports and the library Result must stay identical, and number rules must stay exact.

## Acceptance criteria

- The decoded document keeps exactly today's Go shape: objects map[string]any, arrays []any, string, bool, nil and *jsonvalue.Number. No new Value struct and no ordered member slices. Only Number's internal representation changes.
- Number holds a small exact decimal inline: an int64 mantissa with no trailing zeros plus an int64 exponent. This makes 0.99, 0.01, 1.0 and 1e2 native and exact. A big.Int mantissa is used only when the digits do not fit in int64. The fast path allocates no big.Int, and float64 never decides any answer.
- Number's public methods keep their behaviour: Decimal, IsInteger, Rat, Canonical, String, Cmp and equality. The canonical and original number text are byte-identical in every report. No public function or type of the root package changes.
- Every guarantee from story 179 holds. Powers of 10 are never raised to a large exponent, Cmp aligns mantissas only within the digit cap, Rat() stays lazy and counted, and multipleOf keeps the exact decimal/modular method. The number digit and exponent caps and their LimitErrors are identical. The existing amplification and allocation-free-equality tests in internal/jsonvalue/number_test.go pass unchanged.
- Mixed fast/big operands compare, test equal and hash exactly, so 1, 1.0 and 1e0 are equal. uniqueItems, enum and const deep equality give the same answers as today.
- Each of the nine number rules has tests on the fast path, the exact big path and the boundary between them. The rules are type integer, multipleOf, maximum, exclusiveMaximum, minimum, exclusiveMinimum, enum, const and uniqueItems. Boundary cases include int64 max and max+1 mantissas, trailing-zero normalisation, and exponents at the caps. The tests include 0.3 multipleOf 0.1, exact comparisons on huge integers and 1.0 as an integer.
- Results are identical everywhere: the valid/invalid answer, the flag, basic, detailed and verbose reports, and the library Result. All existing tests pass unchanged.
- One regression test proves that every hostile-input limit triggers at exactly the same inputs as before, with the same error and exit code 2. The limits are nesting depth, input size, ref chain, evaluation-step charging, output units, report bytes and annotation count.
- The compiled Schema stays immutable and safe for concurrent use. A test covers concurrent validation of documents with fast-path and big-path numbers, and passes under go test -race.
- The project's benchmarks are run before and after on the same machine. Both sets go into docs/benchmarks.md, with ns/op, B/op, allocs/op and the retained-heap metric. The document scenarios show lower allocs/op or B/op and lower retained heap. No scenario's allocs/op or B/op increases, and no scenario is more than 10% slower.
- docs/benchmarks.md records the CLI peak memory before and after for the 0.5 MB, 5.5 MB and 0.33 MB documents, for information only.
- This story adds and changes no allocation-limit test.

## Non-goals

- Removing the per-value heap node or changing the document's Go shape (dropped by the operator).
- Changing verbose output's second evaluation pass.
- Any public API change for library callers.
- Allocation-limit tests (story 204).
- Using float64 to decide any answer.
- Preparing for any other number representation beyond what is described.

## Tasks

- **T1** — Compact Number representation
  - purpose: Replace Number's internals with an inline int64 mantissa and exponent, falling back to a big.Int mantissa. The representation stays canonical: trailing zeros are stripped, and a big mantissa that fits int64 is demoted to the fast form. Public methods, caps, LimitErrors, lazy counted Rat() and text output stay byte-identical. Decoding builds fast-path numbers with no big.Int allocation.
  - scope: internal/jsonvalue/number.go, internal/jsonvalue/number_canonical.go, internal/jsonvalue/decode.go, internal/jsonvalue/number_fastpath_test.go
  - done when: go test ./internal/jsonvalue/... passes with number_test.go unchanged. The new fast/big/boundary unit tests pass.
  - delivers: AC1, AC2, AC3
- **T2** — Exact mixed-representation equality and hashing
  - purpose: Make deep equality and the uniqueItems hash/key exact across fast and big operands, so 1, 1.0 and 1e0 are equal. Equality stays allocation-free where today's tests require it.
  - scope: internal/jsonvalue/equal.go, internal/jsonvalue/equal_fastpath_test.go
  - done when: New equality and hash tests over mixed operands pass. The existing equal_test.go and the allocation-free-equality tests pass unchanged.
  - depends on: T1
  - delivers: AC4, AC5
- **T3** — Per-rule fast/big/boundary tests and the multipleOf helper
  - purpose: Add one new test file per number rule, covering the fast path, the big path and the boundary between them. If the exact decimal/modular method sits in the keyword, switch multipleof.go to the faster exact jsonvalue helper. multipleof.go is the only keyword file that may be edited.
  - scope: internal/keywords/validation/multipleof.go, internal/keywords/validation/type_numberpath_test.go, internal/keywords/validation/multipleof_numberpath_test.go, internal/keywords/validation/maximum_numberpath_test.go, internal/keywords/validation/exclusivemaximum_numberpath_test.go, internal/keywords/validation/minimum_numberpath_test.go, internal/keywords/validation/exclusiveminimum_numberpath_test.go, internal/keywords/validation/enum_numberpath_test.go, internal/keywords/validation/const_numberpath_test.go, internal/keywords/validation/uniqueitems_numberpath_test.go
  - done when: All nine new test files pass, and every existing keyword test passes unchanged.
  - depends on: T1, T2
  - delivers: AC5, AC6, AC7
- **T4** — Limits regression and race test
  - purpose: Prove that every hostile-input limit triggers at the same inputs with the same error and exit code 2. Prove that concurrent validation of fast-path and big-path documents is race-free.
  - scope: number_limits_regression_test.go, number_race_test.go
  - done when: go test ./... and go test -race ./... pass.
  - depends on: T1, T2
  - delivers: AC7, AC8, AC9, AC12
- **T5** — Before/after benchmarks and CLI peak memory
  - purpose: Run the existing benchmarks on the old code and then on the new code, on the same machine. Record ns/op, B/op, allocs/op, retained heap and CLI peak memory in docs/benchmarks.md, and check the no-increase and no-more-than-10%-slowdown rules.
  - scope: docs/benchmarks.md
  - done when: docs/benchmarks.md holds both sets of numbers and the CLI peak memory. The document scenarios show lower allocs/op or B/op and lower retained heap, and no scenario regresses.
  - depends on: T3, T4
  - delivers: AC10, AC11, AC12
