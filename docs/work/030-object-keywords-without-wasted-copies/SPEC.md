# SPEC #30: Object keywords without wasted copies (properties, additionalProperties, patternProperties)

- Date: 2026-10-17
- Type: performance
- Verdict: pass
- Status: proposed (pending Checkpoint 1 sign-off)

## Problem

Three object keywords waste allocations. properties, additionalProperties and patternProperties copy each object's member names to the heap and sort them twice. The applied slice grows from a capacity of 1. The names are copied again for the annotation. The keyword location for the evaluated record is rendered even when recording is switched off.

## Why this matters

Round-3 story 3 cuts Validate allocations on object-heavy documents. Results stay byte-identical, so callers see no change. This story is built in parallel with #206, #210 and #211 after #205 is integrated, so it writes only its own files and new files.

## Acceptance criteria

- The three keywords get each object's member names in a sorted order identical to today's byte order (sort.Strings order). Objects with up to 8 members use a fixed-size stack buffer, and only objects with more than 8 members use the heap. The names are sorted once per object, not twice. A new unit test compares the helper against a sort.Strings reference for 0, 1, 8, 9 and many members, including names whose byte order differs from Unicode or case order (for example "B" < "a", "é" vs "z", "" and a shared prefix).
- The applied slice in properties, additionalProperties and patternProperties is pre-sized to its final bound instead of growing from capacity 1. The annotation is emitted without copying the names again. Annotation values and annotation order are unchanged.
- The keyword location for the evaluated-property record is rendered only when the record is kept, behind a Recording() check. When recording is off, no location string is built for it. unevaluatedProperties and unevaluatedItems keep recording and still see every evaluated property, including through $ref and $dynamicRef.
- A new boundary test validates objects with 0, 1, 8, 9 and many properties, including byte-order-sensitive names. It checks the valid/invalid result, the annotation values and their order, and the flag, basic, detailed and verbose reports against the pre-change outputs. Those outputs are captured by running the test on the unchanged code before any production edit. The test also confirms that unevaluatedProperties still sees every evaluated property. No existing test file is changed.
- The committed v1.1 golden CLI test from #205 passes byte for byte (stdout, stderr, exit code; flag, basic, detailed and verbose; perf fixtures, docs/examples, testdata cases). No golden file is regenerated or edited. If any output differs, the story stops as failed. No existing test is removed or weakened. go test -race ./... passes. go.mod is unchanged (standard library only).
- A new root-package file allocs_objectkeys_test.go holds this story's lowered allocation limits for the guarded scenarios. Each limit is the measured value plus about 10% headroom, with the measured value recorded next to it. It reuses the helpers and race gating of allocs_test.go and allocs_race_on/off_test.go without editing them. allocs_test.go stays untouched.
- A new file docs/benchmarks/r3-207-objectkeys.md records before medians (main after #205) and after medians (with this story) from go test -bench -count=10 -benchmem on the same machine, with B/op and allocs/op. It records the target verdict judged by amount saved. The saving required is the v1.1 Validate5_5MB B/op minus 345 MB, and the v1.1 allocs/op minus 6.0 M, compared with the code just before this story. If the saving falls short, the numbers are recorded and the story is held for the operator.
- The same file reports the speed check: medians of 10 runs before and after, with any scenario more than 5% slower flagged. It notes that timings taken during the parallel build are noisy and that the operator re-measures interleaved at Checkpoint 3 and judges the 5% rule there. The speed check is not asserted in any test.
- The story writes no file claimed by #206 (internal/keywords/registry.go, childctx*), #210 (internal/jsonvalue/number_inline*, internal/keywords/validation/*) or #211 (internal/registry/*, internal/compile/metacache.go). It also does not write allocs_test.go or docs/benchmarks.md.

## Non-goals

- Changing validation results, report shapes, error messages, limits, exit codes or goroutine safety
- Changing unevaluatedProperties/unevaluatedItems tracking beyond the Recording() guard on the three object keywords
- Touching propertyNames, dependentSchemas or array keywords
- Editing allocs_test.go, docs/benchmarks.md or any golden file
- Reusing child contexts (#206), chunked annotations (#208) or multipleOf changes (#210)
- Asserting timings or peak memory in tests

## Tasks

- **T1** — Capture pre-change behaviour and before numbers
  - purpose: Before any production edit, write the boundary test with expected outputs captured from the unchanged code, and record the before benchmark medians.
  - scope: objectkeys_boundary_test.go, docs/benchmarks/r3-207-objectkeys.md
  - done when: The boundary test passes on the unchanged code (main after #205). Before medians (count=10, B/op, allocs/op) are written to docs/benchmarks/r3-207-objectkeys.md.
  - delivers: AC4, AC7, AC8
- **T2** — Stack-buffer sorted member names helper
  - purpose: Add one helper that returns an object's member names in today's byte order, using an 8-entry stack buffer and the heap only above 8, sorted once.
  - scope: internal/keywords/applicator/membernames.go, internal/keywords/applicator/membernames_test.go
  - done when: The differential test against a sort.Strings reference passes for 0, 1, 8, 9 and many names, including byte-order-sensitive names. An AllocsPerRun check shows 0 heap allocations for 8 or fewer names.
  - depends on: T1
  - delivers: AC1
- **T3** — Recording() check and copy-free annotate entry point
  - purpose: Expose Recording() on the keyword EvalContext from recording.go (same package as registry.go, which is not edited) and on the evaluated tracker. Add an annotate entry point that takes the names without copying. Existing functions keep their signatures.
  - scope: internal/keywords/recording.go, internal/keywords/annotate.go, internal/eval/evaluated.go, internal/keywords/recording_objectkeys_test.go
  - done when: New tests show Recording() is false when the compiled schema needs no evaluated tracking and true when unevaluated* applies (including through $ref and $dynamicRef). The copy-free annotate produces the same annotation value as before. annotate_test.go and evaluated_test.go pass unchanged.
  - depends on: T1
  - delivers: AC2, AC3
- **T4** — Apply to properties, additionalProperties, patternProperties
  - purpose: Use the helper, pre-size the applied slice, annotate without a copy, and render the record's keyword location only when Recording() is true.
  - scope: internal/keywords/applicator/properties.go, internal/keywords/applicator/additionalproperties.go, internal/keywords/applicator/patternproperties.go
  - done when: The boundary test from T1, the golden CLI test, all applicator tests and the unevaluated* tests pass unchanged. go test -race ./... passes.
  - depends on: T2, T3
  - delivers: AC1, AC2, AC3, AC4, AC5
- **T5** — Lowered allocation limits in a new file
  - purpose: Re-measure the guarded scenarios and set limits at the measured value plus about 10%, reusing the existing helpers and race gating read-only.
  - scope: allocs_objectkeys_test.go
  - done when: go test ./... and go test -race ./... pass. Each limit has its measured value in a comment. allocs_test.go is unchanged (git diff empty).
  - depends on: T4
  - delivers: AC6
- **T6** — After numbers, verdicts and final gates
  - purpose: Record after medians, the saving-based target verdict and the speed report, and confirm that no foreign or shared file was written.
  - scope: docs/benchmarks/r3-207-objectkeys.md
  - done when: The doc has before and after tables, the saving required versus achieved for B/op and allocs/op, a pass or held verdict, and the 5% report with the parallel-noise note. git diff --name-only lists only this story's files. go.mod and testdata/golden/ are unchanged.
  - depends on: T5
  - delivers: AC5, AC7, AC8, AC9
