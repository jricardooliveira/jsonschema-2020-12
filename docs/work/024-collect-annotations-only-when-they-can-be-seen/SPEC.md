# SPEC #24: Collect and sort annotations only when they can be shown, without re-splitting JSON Pointers

- Date: 2026-10-17
- Type: performance
- Verdict: pass
- Status: proposed (pending Checkpoint 1 sign-off)

## Problem

Every validation collects every annotation and then sorts them. The sort splits each JSON Pointer into segments again on every comparison. Profiling of the order schema puts this among the main sources of allocations. Yet some annotations can never reach a caller: for example, annotations inside 'not' are always dropped. Annotations that are shown are often fully built before anything asks for them.

## Why this matters

The operator chose fixes 1 to 5 to bring speed and memory roughly in line with tekuri. This is fix 3. Results must stay identical everywhere. Today the library Result always carries annotations, so there are only two allowed moves: skip work that no Result accessor and no report could ever show, or build the same content in the same order on demand. Nothing in the library API changes.

## Acceptance criteria

- There is no new library option and no API change. Annotations are not collected or sorted in any part of an evaluation where no Result accessor and no output shape (flag, basic, detailed, verbose) could ever show them, for example inside a subschema whose annotations the 2020-12 rules always drop. Every caller keeps getting annotations in the Result without asking for them.
- Wherever annotations can be shown, every Result accessor and every output shape returns the same content in the same order as before this story. This holds whether the annotations are built during Validate or on demand on first access.
- Annotations are sorted without splitting a JSON Pointer on each comparison: the segments are computed once per annotation and compared directly. A unit test compares the new order with the old pointer-splitting order. It covers ~0/~1 escapes, percent-free and numeric segments, prefix pointers, the empty pointer and equal instance locations with different keyword locations, and the two orders match exactly.
- Results stay identical everywhere: the same valid or invalid answer, the same flag, basic, detailed and verbose reports, and the same library Result. Verbose keeps its second evaluation pass. All existing tests pass unchanged.
- Every hostile-input limit keeps working exactly as before. That covers nesting depth, ref chain, evaluation-step charging, output units, report bytes and annotation count. Each triggers at exactly the same inputs, with the same error and exit code 2. In particular, the annotation-count limit still fails Validate itself, even when annotation content is built later. One new regression test proves this through both the library and the CLI run entry point. For each limit it checks the input just under the limit and the input just over it.
- The compiled Schema stays immutable after Compile. A Result whose annotations are built on demand is immutable from the caller's view and safe to read from several goroutines at the same time. A new test reads annotations and renders every shape concurrently from the same Result and the same Schema, and `go test -race ./...` passes.
- The project's benchmarks are run on the same machine before and after the change, and both sets of numbers are added to docs/benchmarks.md. The targeted validation scenarios show lower allocs/op or B/op. No scenario's allocs/op or B/op increases. ns/op does not get worse beyond the normal run-to-run variation of about 5%.
- No allocation-limit test is added or changed. The number-amplification test in internal/jsonvalue/number_test.go stays unchanged.

## Non-goals

- Adding a library option or CLI flag to turn annotations off (the operator declined any API change).
- Removing the verbose second evaluation pass (left for a later round).
- Skipping annotation collection during the meta-schema check at Compile (story 202).
- Lazy keyword and absolute keyword locations (story 199) and evaluated-property recording (story 200).
- Setting allocation limits in tests (story 204).
- Preparing the annotation code for any other future use.

## Tasks

- **T1** — Sort annotations by precomputed pointer segments
  - purpose: Replace the comparator that splits JSON Pointers on every comparison with one that compares segments computed once per annotation, keeping exactly today's order.
  - scope: internal/eval/annotation_order.go, internal/eval/annotation_order_test.go, internal/eval/annotations.go
  - done when: annotation_order_test.go shows that the old comparator and the new one give the same order on the listed edge cases, and existing annotation tests pass unchanged.
  - delivers: AC3, AC4
- **T2** — Record annotations cheaply and build them on demand
  - purpose: During evaluation, keep a compact record per annotation and count it immediately, so the annotation-count limit still fires inside Validate. Build the public annotation values only when a Result accessor or a report first needs them, behind a sync.Once-style guard that keeps the Result safe for concurrent reads.
  - scope: internal/eval/annotations.go, internal/eval/annotation_record.go, annotations.go, jsonschema.go
  - done when: Existing tests for annotations in the root package, internal/eval, internal/keywords and the CLI pass unchanged, and the limit tests pass unchanged.
  - depends on: T1
  - delivers: AC2, AC4, AC5, AC6
- **T3** — Skip collection where no output can show annotations
  - purpose: Evaluate subschemas whose annotations are always dropped in a discard mode that neither records nor sorts annotations, while still charging evaluation steps and checking limits exactly as before.
  - scope: internal/eval/annotation_sink.go, internal/eval/eval.go, internal/eval/annotation_sink_test.go
  - done when: annotation_sink_test.go shows that discarded subtrees record nothing and charge the same evaluation steps, and all existing tests pass unchanged.
  - depends on: T2
  - delivers: AC1, AC4
- **T4** — Regression test: hostile-input limits unchanged
  - purpose: Prove that nesting depth, ref chain, evaluation-step charging, output units, report bytes and annotation count trigger at exactly the same inputs, with the same error and exit code 2, through the library and the CLI run entry point.
  - scope: cmd/jsonschema/annotations_limits_regression_test.go
  - done when: The test passes before and after T2 and T3, and it reads each limit value from the code instead of repeating the number.
  - depends on: T3
  - delivers: AC5
- **T5** — On-demand annotations: identical content and race-safety test
  - purpose: Show that the annotations and every output shape are identical to the eagerly built ones on spec-derived schemas, and that concurrent reads of one Result and concurrent Validate calls on one Schema are race-free.
  - scope: annotations_ondemand_test.go
  - done when: `go test -race ./...` passes, including the new test.
  - depends on: T3
  - delivers: AC2, AC4, AC6
- **T6** — Before and after benchmarks recorded
  - purpose: Run the story 198 benchmarks on the same machine before T1 and after T5, and record both sets of numbers in docs/benchmarks.md, without adding or changing any allocation-limit test.
  - scope: docs/benchmarks.md
  - done when: docs/benchmarks.md lists the before and after ns/op, B/op and allocs/op for every scenario. The targeted scenarios show a reduction, no scenario's allocs/op or B/op increases, and no allocs test changed.
  - depends on: T5
  - delivers: AC7, AC8
