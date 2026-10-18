# Release notes — #31: Annotations collected in growing chunks, flattened lazily under sync.Once

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #31
- Status: approved by the operator at Checkpoint 3 (run #168, 2026-10-17 23:16 UTC)

## What changed

Annotations are now collected in growing chunks (capacities 16, 32, 64… up to 4,096 entries) instead of one expanding slice, then flattened and sorted on the first Annotations() or Output() call under sync.Once. Results are byte-identical to before; concurrent calls on the same Result are safe. Validate5_5MB allocations drop 34.4 MB (exceeding target 26.96 MB), ValidateErrors drops 2.49 MB, ValidateSmall improves in both bytes and allocs/op.

- internal/eval/annotation_chunks.go (new): Pending type—holds unflattened chunks, flattens and sorts lazily under sync.Once on first Sorted() call
- internal/eval/annotation_chunks_test.go (new): Tests for concurrent first-reader safety and chunk release inside Once
- internal/eval/annotation_chunks_random_test.go (new): Random tests verifying cross-boundary truncate and capped-chunk behavior
- internal/keywords/annotation_chunks.go (new): chunkStore with growing capacities (16…4,096), mark/truncate rollback, empty-chunk reuse
- internal/keywords/annotation_chunks_test.go (new): Unit tests for chunk growth, capacity caps, truncate across boundaries
- internal/keywords/annotation_chunks_random_test.go (new): 10,000-step differential test against plain slice, fixed seed 208 printed on failure
- annotations_once_race_test.go (new): 16 concurrent Annotations() and Output() calls on one shared Result under go test -race
- annotations.go: Added Pending field to Evaluation; Annotations field kept
- jsonschema.go: Result.annotations changed from Lazy to Pending
- internal/eval/annotations.go: EvaluateValidate and EvaluateLazy hand unflattened chunks to Pending
- internal/keywords/annotate.go: Uses chunkStore.Add() and exposes AnnotationChunks()
- internal/keywords/registry.go: Uses chunkStore.Mark() and Truncate() instead of slice operations
- internal/eval/annotation_sink_test.go: Updated to call ev.Pending.Sorted()
- allocs_test.go: Limits re-measured and lowered—Compile 1,640→1,325 (1,203), Small 615→432 (392), Medium 507,000→331,000 (300,035), Large 5,061,000→3,301,000 (3,000,069), Errors 353,000→266,000 (240,999), ErrorsBasic 359,000→272,000 (247,028), bytes limits similarly lowered
- docs/benchmarks.md: New section recording before/after medians from same-session benchmarks, computation of required drops against v1.1 values, proof all targets met
- cmd/jsonschema/golden_dbg_test.go (deleted with git rm): Stray file containing only 'package main', not referenced

## How to verify

- go test -race ./... passes; includes new concurrent-read and random tests
- go test -bench -count=10 . and compare medians in docs/benchmarks.md: Validate5_5MB 34.4 MB drop (required 26.96 MB), ValidateErrors 2.49 MB drop (required 1.88 MB), ValidateSmall +1.35 KB within allowed +318 B
- go test ./cmd/jsonschema and cmd/jsonschema/golden_test.go pass; golden_dbg_test.go no longer present
- cmd/jsonschema validate -o flag / -o basic / -o detailed / -o verbose against testdata/golden cases: output byte-identical to v1.1 commit f35d9e4 in all four formats
- annotations_once_race_test.go Annotations() and Output() call Pending.Sorted() concurrently on one Result; all 16 callers receive identical results under -race
- go test -run 'Chunk.*Random' internal/keywords under one second without -race: random test reaches cross-boundary truncate and a chunk at 4,096-entry cap with fixed seed 208

## Migration

None. Public API Result.Annotations() and Result.Output() unchanged and byte-identical. Annotations are stored unflattened in Pending internally and flattened on first read; this is transparent to callers.

## Rollback

Revert files listed in 'changes'. The public API is unchanged, so caller code requires no modification. If eval.Lazy (internal/eval/lazy.go) was removed earlier, restore it from a prior commit; otherwise it becomes used again.

## Known limitations

- Chunks created fresh per evaluation (per check); never pooled or passed between checks.
- Verbose output (second evaluation pass) not changed by this story; remains ~2× slower than basic.
- Annotation order, content and the Result's public signatures identical to prior release.

## Concerns

- Factory must name cmd/jsonschema/golden_dbg_test.go deletion in the release notes (story acceptance criterion 10).
