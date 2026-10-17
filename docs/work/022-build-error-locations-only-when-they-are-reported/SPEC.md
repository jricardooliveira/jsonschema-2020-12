# SPEC #22: Build keyword locations only when a report or Result shows them

- Date: 2026-10-17
- Type: performance
- Verdict: pass
- Status: passed gate 1 with nothing for the operator to decide (run #157, 2026-10-17 10:53 UTC)

## Problem

Evaluation builds a keyword location string and an absolute keyword location string for every keyword it evaluates, including keywords that pass and that no report or library Result will ever show. Profiling on the order schema shows this costs about 30% of allocated memory on the validation scenarios.

## Why this matters

The goal for this round is to roughly match the operator's reference validator in memory and speed. Building locations lazily removes the largest single source of allocations without changing any result.

## Acceptance criteria

- During evaluation, a keyword location or absolute keyword location string is produced only when an output unit, error, annotation or library Result that will be shown needs it. Otherwise only a cheap reference to the schema position is kept, and the string is produced from it on demand. Locations are built while a document is checked, not when the schema is loaded.
- Results are identical everywhere: the same valid or invalid answer, the same flag, basic, detailed and verbose reports byte for byte, and the same library Result, including instance locations, keyword locations, absolute keyword locations and annotation locations. All existing tests pass unchanged.
- Verbose output stays correct and keeps its current second evaluation pass.
- One new regression test covers every hostile-input limit: nesting depth, ref chain, evaluation-step charging, output units, report bytes and annotation count. For each limit, a small input built by test code at exactly the limit still gets a normal valid or invalid answer, and one just over the limit gets the same error as today and exit code 2. Limit values are read from the code, not repeated as numbers.
- The compiled Schema and everything reachable from it stay immutable after Compile returns. Any location cache lives in the per-call evaluation context. go test -race ./... passes, including the existing concurrency tests.
- The project's benchmarks (go test -run '^$' -bench . -benchmem) are run on the same machine before and after the change. Each is run 10 times and the median is recorded. Both sets of numbers are added to docs/benchmarks.md for this story.
- On the validation scenarios, allocs/op or B/op goes down. No scenario's allocs/op or B/op goes up, compiling the schema included. ns/op on the validation scenarios is not worse beyond the ±5% run-to-run noise. Any speed gain is recorded but not required.
- This story adds and changes no allocation-limit test. internal/jsonvalue/number_test.go is unchanged.

## Non-goals

- Skipping evaluated-property or evaluated-item recording (story 200)
- Skipping or re-sorting annotation collection (story 201)
- Compile-time and meta-schema caching work (story 202)
- Compact document values or the number fast path (story 203)
- Allocation-limit tests (story 204)
- Removing verbose's second evaluation pass
- Precomputing location references at schema load if that would add any compile allocation

## Tasks

- **T1** — Record the before benchmark numbers
  - purpose: Run every project benchmark 10 times on the current code and record the medians in docs/benchmarks.md as this story's 'before' set.
  - scope: docs/benchmarks.md
  - done when: docs/benchmarks.md has a story-199 'before' table of median ns/op, B/op, allocs/op and MB/s for every scenario, with the command used.
  - delivers: AC6
- **T2** — Lazy location reference in the evaluator
  - purpose: Add a small value type in a new file. It holds a cheap reference to the schema position (parent reference plus keyword or segment, and the base URI of the resource) and renders the keyword location and absolute keyword location strings on demand, with exactly today's escaping. Any memoisation lives only in the per-call evaluation context.
  - scope: internal/eval/location.go, internal/eval/location_test.go
  - done when: Unit tests in location_test.go show the rendered strings equal today's strings for nested keywords, ~0/~1 escaping, $ref and $dynamicRef hops, and embedded $id resources.
  - delivers: AC1, AC5
- **T3** — Wire lazy locations into evaluation, units and annotations
  - purpose: Replace eager string building in the evaluation loop with the lazy reference. Render strings only where an output unit, error, annotation or Result is actually created, in both the normal pass and verbose's second pass. Keep step charging, unit counting and report-byte accounting at the same points and with the same values.
  - scope: internal/eval/eval.go, internal/eval/units.go, internal/eval/annotations.go
  - done when: go test ./... passes with no existing test edited, including the output, verbose, annotation, unevaluated* and CLI contract tests.
  - depends on: T2
  - delivers: AC1, AC2, AC3
- **T4** — Limits-at-the-edge regression test
  - purpose: Add one table-driven test that drives the CLI run entry point with small generated schemas and documents for each limit. One case sits exactly at the limit and gets a normal 0 or 1 answer. One case is just over it and gets today's error text and exit code 2. Limit values are taken from the code's defaults or set through the existing limit options.
  - scope: cmd/jsonschema/limits_edge_regression_test.go
  - done when: The test passes both before and after T3, which proves that no limit's trigger point shifted.
  - delivers: AC4
- **T5** — Race and immutability check
  - purpose: Confirm that no location cache was attached to the compiled Schema, and that concurrent validation renders identical locations.
  - scope: internal/eval/location_test.go
  - done when: go test -race ./... passes, and location_test.go includes a parallel render case.
  - depends on: T3
  - delivers: AC2, AC5
- **T6** — Record the after benchmark numbers and verify there is no regression
  - purpose: Re-run the benchmarks 10 times on the same machine, record the medians next to the before set, and check the acceptance rules: allocs/op or B/op lower on validation scenarios, none higher anywhere (compile included), and ns/op within ±5% or better.
  - scope: docs/benchmarks.md
  - done when: docs/benchmarks.md shows before and after medians side by side for story 199, with the reduction stated per validation scenario. No allocation-limit test file was added or changed.
  - depends on: T1, T3
  - delivers: AC6, AC7, AC8
