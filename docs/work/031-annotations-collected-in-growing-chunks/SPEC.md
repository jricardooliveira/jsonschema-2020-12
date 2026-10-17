# SPEC #31: Annotations collected in growing chunks, flattened lazily under sync.Once

- Date: 2026-10-17
- Type: performance
- Verdict: pass
- Status: proposed (pending Checkpoint 1 sign-off)

## Problem

Annotate appends every annotation to one slice that keeps growing. Its regrowth copies are the biggest single allocations during evaluation. Flattening and sorting work is also not tied to the first reader of the Result.

## Why this matters

Round-3 story 4. It cuts Validate memory with no change to any answer, report, error, limit, exit code or goroutine-safety guarantee. Fixed 512-entry chunks were measured to make small documents allocate 114% more, so chunk sizes must grow.

## Acceptance criteria

- Annotations are collected per check in chunks whose capacities grow 16, 32, 64, … and are capped at 4,096 entries, never fixed-size. Each check creates its own chunks; nothing is pooled or passed between checks.
- Rollback works by mark and truncate, including across chunk boundaries. Chunks emptied by a truncate are kept and reused for later annotations in the same check, and are not reallocated.
- Flattening and sorting happen only on the first Annotations() or Output() call on a Result, inside one sync.Once. The chunks are released inside that same Once. A new test calls Annotations() and Output() concurrently on one shared Result under go test -race and gets identical answers.
- A new randomised test runs about 10,000 add/mark/truncate/flatten steps against a plain slice and checks they agree. It uses a fixed starting seed that is printed in any failure message. It asserts that at least one truncate crossed a chunk boundary and that a chunk at the 4,096-entry cap was reached; a fixed growth sequence past about 8,200 entries is added if the random walk does not get there. It runs well under one second without -race.
- Annotations() and Output() stay byte-identical. All existing tests pass unchanged, including the golden CLI test against testdata/golden in flag, basic, detailed and verbose. No golden file is regenerated or edited. Any output difference fails the story, and the story stops.
- go test -race ./... passes, standard library only, and go.mod is unchanged.
- docs/benchmarks.md records go test -bench -count=10 medians from the in-repo benchmarks, measured just before and just after this change on the same machine. It shows the computation against v1.1 (f35d9e4, round 2 final), using exact measured values: Validate5_5MB drops by at least 26,958,349 B/op, ValidateErrors drops by at least 1,882,834 B/op, ValidateSmall rises by at most 318 B/op, and ValidateSmall allocs/op does not rise.
- The limits in allocs_test.go are re-measured and lowered on purpose: measured value plus about 10%, with the measured value recorded next to each limit, as docs/benchmarks.md describes. If any limit would have to rise, that is a missed target. allocs_<topic>_test.go files and docs/benchmarks/r3-*.md files are left untouched.
- If any target is missed, the numbers are recorded and the story is held for the operator's decision. It is never marked done silently.
- cmd/jsonschema/golden_dbg_test.go (only 'package main', referenced by nothing) is removed with git rm. No other file is deleted, and the removal is named in the release notes.

## Non-goals

- Reusing or pooling chunks between checks (operator chose fresh chunks per check).
- Changing the verbose second evaluation pass.
- Changing annotation content, order, sort keys or the sink's public method signatures.
- Editing golden files, allocs_<topic>_test.go or docs/benchmarks/r3-*.md.
- Report rendering changes (story 209).

## Tasks

- **T1** — Chunked annotation store with mark/truncate and reuse
  - purpose: Add a self-contained chunk store: growing capacities 16…4,096, Add, Mark, Truncate (keeping and reusing emptied chunks), Len and Flatten into one slice. Document the invariant next to the code. Add unit tests and the randomised differential test.
  - scope: internal/eval/annotation_chunks.go, internal/eval/annotation_chunks_test.go, internal/eval/annotation_chunks_random_test.go
  - done when: go test -race ./internal/eval -run 'Chunk' passes. The random test reports reaching a cross-boundary truncate and a capped 4,096 chunk. Without -race it runs under 1 s.
  - delivers: AC1, AC2, AC4
- **T2** — Wire the chunk store into the annotation sink
  - purpose: Replace the single growing slice behind Annotate and the sink's mark/rollback with the chunk store. Keep every method signature so eval.go and the keywords do not change. Create chunks fresh in the per-call evaluation context.
  - scope: internal/eval/annotation_sink.go, internal/eval/annotations.go
  - done when: Existing internal/eval annotation, order and sink tests pass unchanged under -race.
  - depends on: T1
  - delivers: AC1, AC2, AC5
- **T3** — Lazy flatten and sort under sync.Once, release chunks
  - purpose: Flatten and sort on the first Annotations() or Output() call inside one sync.Once, using the existing sort order, and drop the chunk references inside the same Once. Add a concurrent -race test on one shared Result.
  - scope: annotations.go, jsonschema.go, annotations_once_race_test.go
  - done when: go test -race -run 'Annotations|Once' . passes. The existing annotations_test.go and annotations_ondemand_test.go tests pass unchanged.
  - depends on: T2
  - delivers: AC3, AC5
- **T4** — Remove stray golden_dbg_test.go
  - purpose: Run git rm on cmd/jsonschema/golden_dbg_test.go, which holds only 'package main'. Delete nothing else.
  - scope: cmd/jsonschema/golden_dbg_test.go
  - done when: git status shows exactly one deleted file. go test ./cmd/jsonschema passes, including the golden test.
  - delivers: AC10
- **T5** — Measure, lower allocation limits, record numbers
  - purpose: Run benchmarks with -count=10 just before T1 and after T3 on the same machine. Compute the required drops against the v1.1 values supplied by the operator. Re-measure and lower the allocs_test.go limits. Record everything in docs/benchmarks.md, and hold the story on any miss or rising limit. Confirm that go test -race ./... and the golden test pass.
  - scope: allocs_test.go, docs/benchmarks.md
  - done when: docs/benchmarks.md shows the before and after medians, the computations with sources and the verdict per target. allocs_test.go limits are lowered, with measured values. go test -race ./... is green.
  - depends on: T3, T4
  - delivers: AC5, AC6, AC7, AC8, AC9
