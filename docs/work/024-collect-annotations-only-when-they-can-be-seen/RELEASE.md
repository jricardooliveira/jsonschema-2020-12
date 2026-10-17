# Release notes — #24: Collect and sort annotations only when they can be shown, without re-splitting JSON Pointers

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #24
- Status: proposed (pending Checkpoint 3 sign-off)

**⚠️ the release-agent has concerns.**

## What changed

Annotations are now sorted using precomputed JSON Pointer segments instead of re-splitting pointers on each comparison. Annotations are built on first access to Result.Annotations() or when rendering output, behind a sync.Once guard that keeps the Result immutable and race-safe. Annotations inside 'not' keywords are not collected (as the 2020-12 spec always drops them) but still count toward the annotation limit. Every validation result, report shape and limit trigger stays identical to before.

- internal/eval: added annotation_order.go (segment-based sorting), annotation_record.go (Lazy on-demand building), annotation_sink.go (discard-mode for 'not')
- internal/eval/annotations.go: removed pointer-splitting comparator, now delegates to segment-based sort
- internal/eval/eval.go: added discard-mode switch before 'not' keyword evaluation
- internal/keywords/annotate.go: added SetDiscard() to skip recording but still count toward limit
- jsonschema.go: Result.annotations field changed from []Annotation to *Lazy, wrapped in NewLazy() at Validate
- annotations.go: Result.Annotations() now calls r.annotations.Sorted()
- output.go: renders annotations via r.annotations.Sorted()
- docs/benchmarks.md: added before/after benchmark measurements on same machine

## How to verify

- Run `go test -race ./...` to confirm concurrent access is race-free (annotations_ondemand_test.go)
- Run `go test ./...` to confirm all existing tests pass unchanged, including segment-based sort oracle test and limit regression test
- Run benchmarks with `go test -run '^$' -bench . -benchmem -count=5 . ./internal/jsonvalue ./cmd/jsonschema` and compare allocs/op and B/op to before figures in docs/benchmarks.md; CompileOrder, ValidateSmall, Validate0_5MB and Validate5_5MB should show improvements
- Validate a document against a schema with 'not' keyword and confirm Result.Annotations() excludes any annotations from inside the 'not'
- Set max-annotations to include all annotations (including those in 'not'), then one less, and confirm the limit error fires at the same edge as before

## Migration

None. Result.Annotations() returns identical content in identical order. Annotations are built on demand rather than during Validate, but callers see no difference.

## Rollback

Revert to the previous commit. Result API, annotation content, order and all output shapes are unchanged.

## Known limitations

- Annotations inside 'not' are not recorded (by 2020-12 spec) but still count toward the annotation-count limit.
- Discard-mode skipping applies only to non-tree evaluation; detailed and verbose shapes always record 'not' annotations because they evaluate the full tree.
- The on-demand test compares lazy-built annotations with themselves rather than against an eagerly-built comparison, so it does not independently verify both paths produce identical results under all conditions.

## Concerns

- A panic inside the sync.Once guard during first annotation access would not propagate to the caller (the once prevents re-entry). This is accepted as very unlikely but limits error observability.
- ValidateErrors benchmark shows a 3-allocation rise (510,077 to 510,080) attributed to noise, but sits just outside the before range (510,076–510,079). Monitor to confirm this does not grow further in follow-up work.
- Segment-based sorting precomputes all pointer tokens upfront, which saves re-splitting on comparison but consumes more temporary memory during the sort. Net benefit is positive on large annotation sets.
- eval.go and annotations.go are also touched by stories 199 and 200; later rebases may cause merge conflicts.
