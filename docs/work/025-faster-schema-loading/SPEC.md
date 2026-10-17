# SPEC #25: Faster schema loading: build the embedded 2020-12 meta-schema once and check schemas without annotations

- Date: 2026-10-17
- Type: performance
- Verdict: pass
- Status: passed gate 1 with nothing for the operator to decide (run #160, 2026-10-17 12:41 UTC)

## Problem

Every Compile repeats four pieces of work. It rebuilds the registry index of the built-in meta-schemas, rebuilds the embedded 2020-12 meta-schema and its vocabulary meta-schemas, and recompiles the 2020-12 meta-schema. It then collects and sorts annotations while checking the schema against the meta-schema, although nothing ever shows them. Loading the order schema costs about 1.1 ms and 28k allocations.

## Why this matters

The operator's performance round aims at fast schema loading. The built-in meta-schemas never change within a process, so building them once and sharing an immutable result removes repeated work. The meta-schema check only needs a pass or fail answer, so collecting annotations there is waste. Results, load errors and limits must not change.

## Acceptance criteria

- The embedded 2020-12 meta-schema, all its vocabulary meta-schemas and their registry index are built and compiled at most once per process, lazily on the first Compile and not at package import. Every later Compile reuses that shared result. Caller-supplied registry schemas are indexed per Compile on top of the shared built-in index, exactly as before.
- The shared built-in result and every compiled Schema are immutable after they are built and safe for concurrent use. A new test that runs many concurrent Compile and Validate calls from separate goroutines, including the very first Compile in the process, passes under go test -race.
- The meta-schema check of a loaded schema runs without collecting or sorting annotations. It gives the same pass or fail answer as before, and a schema that fails its meta-schema is still an invalid-schema error with the same message and exit code 2.
- Caller-supplied registry schemas, custom meta-schemas (including ones that use $dynamicRef into the embedded vocabularies) and $vocabulary handling behave exactly as before: the same required and optional vocabulary decisions, the same unknown-required-vocabulary error and the same format-assertion vocabulary handling.
- Results stay identical everywhere: the same valid or invalid answer, the same flag, basic, detailed and verbose reports, the same library Result and the same load errors. All existing tests pass unchanged.
- One new regression test proves that every hostile-input limit (nesting depth, ref chain, evaluation-step charging, output units, report bytes and annotation count) triggers at exactly the same inputs as before, with the same error and exit code 2. It checks both the first Compile in the process and a repeated Compile that uses the cached meta-schema.
- The project's benchmarks are run on the same machine before and after the change, and both sets of numbers are added to docs/benchmarks.md. The compile benchmark shows a reduction in allocs/op, B/op and ns/op, and no scenario's allocs/op or B/op increases.
- A separate first-load benchmark measures building the built-in meta-schemas plus compiling the order schema. It reports what a one-shot CLI run pays, and its before and after numbers are recorded in docs/benchmarks.md next to the repeated-compile numbers.
- No allocation-limit test is added or changed by this story. Allocation limits are left to the story 'Allocation limits in tests'.

## Non-goals

- Adding or changing allocation-limit tests (story 204).
- Caching caller-supplied registry schemas or custom meta-schemas across Compile calls.
- Building the meta-schema at package import time.
- Changing validation-time annotation collection, keyword locations or evaluated-location tracking (stories 199, 200 and 201).
- Changing the verbose second evaluation pass.
- Any comparison with another validator inside the project.

## Tasks

- **T1** — Baseline numbers and first-load benchmark
  - purpose: Record the before numbers of every existing benchmark on this machine. Add a first-load benchmark that builds the built-in meta-schema set from scratch on every iteration through an unexported build function, then compiles the order schema against it.
  - scope: internal/compile/bench_test.go, docs/benchmarks.md
  - done when: go test -run '^$' -bench . -benchmem ./... output before the change is recorded in docs/benchmarks.md, including the new first-load benchmark.
  - delivers: AC7, AC8
- **T2** — Build the embedded meta-schemas and index once per process
  - purpose: Move construction of the embedded 2020-12 meta-schema, its vocabulary meta-schemas, their registry index and the compiled meta-schema into a lazily initialised, immutable shared value guarded by sync.Once. Compile reads it and layers caller registry resources over the shared index without mutating it. Custom meta-schemas still compile per call.
  - scope: internal/compile/metacache.go, internal/compile/metaschema.go, internal/compile/compile.go, internal/registry/builtin.go, internal/registry/registry.go
  - done when: The existing metaschema, vocabulary, registry_refs and refs tests pass unchanged, and a new unit test shows the built-in build function runs once across many Compile calls.
  - depends on: T1
  - delivers: AC1, AC4, AC5
- **T3** — Meta-schema check without annotations
  - purpose: Run the meta-schema check with annotation collection and sorting turned off through a per-call evaluation option. This keeps the same pass or fail answer, the same step charging and the same invalid-schema error.
  - scope: internal/compile/metacheck.go, internal/eval/annotations.go, internal/eval/collect_option.go
  - done when: internal/compile/metacheck_test.go and cmd/jsonschema/metaschema_cli_test.go pass unchanged, and a new test shows the meta check records no annotations.
  - depends on: T2
  - delivers: AC3, AC5
- **T4** — Cache, concurrency and unchanged-behaviour tests
  - purpose: Add tests that the shared built-in result is reused, never mutated by Compile (including Compile calls with registry schemas, custom meta-schemas and $vocabulary), and safe when the first and later Compile calls race with Validate under -race.
  - scope: internal/compile/metacache_test.go, metacache_race_test.go
  - done when: go test -race ./... passes.
  - depends on: T2, T3
  - delivers: AC2, AC4, AC5
- **T5** — Hostile-input limit regression test
  - purpose: Add one table-driven regression test that drives each limit (nesting depth, ref chain, evaluation-step charging, output units, report bytes, annotation count) exactly at and just over its threshold. Each threshold is read from the code's default limit values, and the test asserts the same error, and exit code 2 through the CLI run entry point, on a first and on a repeated Compile.
  - scope: compile_cache_limits_test.go
  - done when: The new test passes, and it would fail if a limit triggered at a different input or returned a different error or exit code.
  - depends on: T3
  - delivers: AC6
- **T6** — After numbers and documentation
  - purpose: Re-run every benchmark on the same machine after the change and add the after numbers next to the baseline in docs/benchmarks.md. Confirm that the compile benchmark dropped in all three numbers and that no scenario's allocs/op or B/op rose. Record that no allocation-limit test was touched.
  - scope: docs/benchmarks.md
  - done when: docs/benchmarks.md holds before and after tables for every scenario plus the first-load benchmark, with compile allocs/op, B/op and ns/op all lower.
  - depends on: T4, T5
  - delivers: AC7, AC8, AC9
