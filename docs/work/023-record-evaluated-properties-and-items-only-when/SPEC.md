# SPEC #23: Record evaluated properties and items only when something can use the record

- Date: 2026-10-17
- Type: feature
- Verdict: pass
- Status: proposed (pending Checkpoint 1 sign-off)

## Problem

Every applicator always records which properties and items it evaluated. Profiling shows this costs about 15% of allocated memory. Only unevaluatedProperties and unevaluatedItems, or a report that shows the record, ever use it. The order schema has neither, so flag and basic checks pay for work that nothing reads.

## Why this matters

This is fix 2 of the five performance fixes the operator chose to roughly match the operator's external comparison in memory. The saving must change no result: the same valid or invalid answer, the same reports in every shape and the same library Result. Story 199's release review also found dead location code (tester finding F1), which the operator routed into this story.

## Acceptance criteria

- When a schema is loaded, the compiled Schema records once whether an unevaluatedProperties or unevaluatedItems keyword is reachable. Reachable means in the root schema or in any schema that a $ref or $dynamicRef could land on, counting every registry resource a $dynamicRef could jump to (the cautious rule). The decision covers the whole schema at once. It is fixed when Compile returns and is never changed afterwards.
- During a check, applicators skip recording evaluated properties and items only when two things hold: the load-time decision says no unevaluated keyword is reachable, and the check produces no report shape and no library Result view that shows the record. When either the load-time decision or the check needs it (for example a detailed or verbose report), recording happens exactly as today.
- Results stay identical everywhere: the same valid or invalid answer, the same flag, basic, detailed and verbose reports, and the same library Result. All existing tests pass unchanged, including every unevaluatedProperties and unevaluatedItems test through $ref and $dynamicRef.
- One new regression test proves that every hostile-input limit triggers at exactly the same inputs as before, with the same error and exit code 2, on a schema where recording is skipped and on a schema where it is kept. The limits covered are nesting depth, ref chain, evaluation-step charging, output units, report bytes and annotation count.
- The compiled Schema stays immutable and safe for concurrent use. A concurrent test validates one compiled schema from several goroutines with mixed report shapes, on schemas that record and schemas that skip recording. go test -race ./... passes.
- The unused nodeLoc and kwLoc types in internal/eval/location.go are removed together with their tests in internal/eval/location_test.go. If a file cannot be deleted, only its package clause and a one-line comment remain. A new unit test for lazyLoc proves that its keyword location and absolute keyword location equal the old eagerly built strings in five cases: through $ref, through $dynamicRef, inside if/then/else, with ~0/~1 escaping and percent-encoding, and with an empty absolute location before any reference is crossed.
- The project's benchmarks (go test -run '^$' -bench . -benchmem) are run on the same machine before and after the change. Both sets of numbers are added to docs/benchmarks.md. allocs/op or B/op drop on the targeted scenarios, which are flag and basic validation of the order documents. No scenario's allocs/op or B/op rises. ns/op is recorded for both runs, and a change within the ±5% noise is acceptable.
- No allocation-limit test is added or changed by this story, and internal/jsonvalue/number_test.go stays unchanged.

## Non-goals

- Making the recording decision rule by rule instead of for the whole schema.
- Skipping or reordering annotation collection and sorting (story 201).
- Compile-time meta-schema caching (story 202) and the compact number model (story 203).
- Adding allocation-limit tests (story 204).
- Changing the verbose second evaluation pass.
- Preparing seams for other future optimisations.

## Tasks

- **T1** — Load-time reachability analysis for unevaluated keywords
  - purpose: Compute once at Compile whether unevaluatedProperties or unevaluatedItems is reachable from the root through subschemas, $ref and every registry resource a $dynamicRef could land on. Store the answer as an immutable field of the compiled Schema. The analysis goes in a new file, and compile.go gets only a one-line call.
  - scope: internal/compile/evaluated_need.go, internal/compile/evaluated_need_test.go, internal/compile/compile.go
  - done when: evaluated_need_test.go covers six cases: no unevaluated keyword (false), a direct keyword (true), a keyword only behind $ref (true), a keyword only in a registry resource reachable by $dynamicRef (true), a cycle of $refs (terminates), and boolean subschemas.
  - delivers: AC1
- **T2** — Skip evaluated-property/item recording when not needed
  - purpose: Give the per-call evaluation context one switch, set when the check starts from the Schema's load-time decision combined with what this check's report shape or Result needs. Guard the recording helpers in evaluated.go and evaluated_items.go with that switch, so keyword files stay unchanged. Evaluation-step charging and every limit counter run exactly as before.
  - scope: internal/eval/eval.go, internal/eval/evaluated.go, internal/eval/evaluated_items.go, internal/eval/evaluated_skip_test.go, jsonschema.go
  - done when: evaluated_skip_test.go shows that recording is off for flag and basic checks on a schema with no unevaluated keyword, and on for detailed and verbose checks and for schemas with unevaluated keywords. go test ./... passes with no existing test changed.
  - depends on: T1
  - delivers: AC2, AC3
- **T3** — Limits regression test and concurrency test
  - purpose: Add the single regression test for all hostile-input limits on schemas that skip recording and schemas that keep it. Add a race-detector test that shares one compiled Schema across goroutines with mixed report shapes.
  - scope: evaluated_tracking_limits_test.go, evaluated_tracking_race_test.go
  - done when: The limits test asserts the same error and exit code 2 at the same boundary inputs for each limit. go test -race ./... passes.
  - depends on: T2
  - delivers: AC4, AC5
- **T4** — Remove unused location types and test lazyLoc
  - purpose: Delete nodeLoc and kwLoc and their tests (tester finding F1 from story 199). Add a lazyLoc unit test comparing its output against the old eagerly built strings.
  - scope: internal/eval/location.go, internal/eval/location_test.go, internal/eval/lazyloc_test.go
  - done when: nodeLoc and kwLoc no longer exist in the package. lazyloc_test.go covers $ref, $dynamicRef, if/then/else, ~0/~1 escaping with percent-encoding, and an empty absolute location before any reference is crossed. go vet ./... and go test ./internal/eval pass.
  - delivers: AC6
- **T5** — Before and after benchmarks
  - purpose: Run the benchmark command on the same machine before T2 lands and after all tasks. Record both sets of numbers in docs/benchmarks.md, and confirm the drops and that nothing rises. Add no allocation-limit test.
  - scope: docs/benchmarks.md
  - done when: docs/benchmarks.md has a story-200 section with before and after ns/op, B/op, allocs/op and MB/s for every scenario. It shows lower allocs/op or B/op on flag and basic order-document validation and no rise in any scenario.
  - depends on: T2, T3, T4
  - delivers: AC7, AC8
