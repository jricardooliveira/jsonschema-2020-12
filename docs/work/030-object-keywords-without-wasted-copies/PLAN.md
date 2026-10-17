# PLAN #30: Object keywords without wasted copies (properties, additionalProperties, patternProperties)

- Date: 2026-10-17
- Status: passed gate 2 with nothing for the operator to decide (run #166, 2026-10-17 22:13 UTC)
- Design: see `ADR-030-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `objectkeys_boundary_test.go`
- `docs/benchmarks/r3-207-objectkeys.md`
- `internal/keywords/applicator/membernames.go`
- `internal/keywords/applicator/membernames_test.go`
- `internal/keywords/recording.go`
- `internal/keywords/annotate.go`
- `internal/eval/evaluated.go`
- `internal/keywords/recording_objectkeys_test.go`
- `internal/keywords/applicator/properties.go`
- `internal/keywords/applicator/additionalproperties.go`
- `internal/keywords/applicator/patternproperties.go`
- `allocs_objectkeys_test.go`

Modules the architecture expects to touch:

- `internal/keywords/applicator/membernames.go`
- `internal/keywords/applicator/membernames_test.go`
- `internal/keywords/applicator/properties.go`
- `internal/keywords/applicator/additionalproperties.go`
- `internal/keywords/applicator/patternproperties.go`
- `internal/keywords/recording.go`
- `internal/keywords/recording_objectkeys_test.go`
- `internal/keywords/annotate.go`
- `internal/eval/evaluated.go`
- `objectkeys_boundary_test.go`
- `allocs_objectkeys_test.go`
- `docs/benchmarks/r3-207-objectkeys.md`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — Capture pre-change behaviour and before numbers
2. **T2** — Stack-buffer sorted member names helper  _(after T1)_
3. **T3** — Recording() check and copy-free annotate entry point  _(after T1)_
4. **T4** — Apply to properties, additionalProperties, patternProperties  _(after T2, T3)_
5. **T5** — Lowered allocation limits in a new file  _(after T4)_
6. **T6** — After numbers, verdicts and final gates  _(after T5)_

## Risks

- Aliasing: annotating without a copy could expose the stack buffer or a slice later mutated, changing annotation values. Mitigation: annotate from the stable slice only, and the boundary test compares annotation values in all four shapes.
- Escape analysis may move the [8]string buffer to the heap when it is passed through interfaces, reducing the saving. Verify with -gcflags=-m and AllocsPerRun.
- Recording() could be false in a path where unevaluated* actually needs the record (through $ref or $dynamicRef). The unevaluated* tests and the new recording test cover this.
- Benchmark timings are noisy during the parallel build. B/op and allocs/op are exact.
- The saving target may be missed. The story then records the numbers and is held for the operator.

Implementation constraints the coder must honor:

- Sort order must equal sort.Strings (byte order) exactly. Stack buffer for 8 or fewer members, heap only above 8.
- Never write internal/keywords/registry.go, childctx*, internal/keywords/validation/*, internal/jsonvalue/number_inline*, internal/registry/*, internal/compile/metacache.go, allocs_test.go or docs/benchmarks.md.
- Never edit testdata/golden/*. If golden output differs, stop as failed.
- Copy-free annotation must not alias a buffer that is reused or mutated later. The stack buffer must not escape into a retained annotation. If escape analysis forces a heap move, record it rather than change semantics.
- unevaluatedProperties and unevaluatedItems must keep receiving every evaluated record whenever Recording() is true.
- Standard library only. go.mod unchanged. No existing test changed, removed or weakened.
- Timings are reported only, never asserted.

- DB impact: no · API impact: no · Migration needed: no

Breaking changes:

- (none)

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: The boundary test passes on the unchanged code (main after #205). Before medians (count=10, B/op, allocs/op) are written to docs/benchmarks/r3-207-objectkeys.md.
- **T2**: The differential test against a sort.Strings reference passes for 0, 1, 8, 9 and many names, including byte-order-sensitive names. An AllocsPerRun check shows 0 heap allocations for 8 or fewer names.
- **T3**: New tests show Recording() is false when the compiled schema needs no evaluated tracking and true when unevaluated* applies (including through $ref and $dynamicRef). The copy-free annotate produces the same annotation value as before. annotate_test.go and evaluated_test.go pass unchanged.
- **T4**: The boundary test from T1, the golden CLI test, all applicator tests and the unevaluated* tests pass unchanged. go test -race ./... passes.
- **T5**: go test ./... and go test -race ./... pass. Each limit has its measured value in a comment. allocs_test.go is unchanged (git diff empty).
- **T6**: The doc has before and after tables, the saving required versus achieved for B/op and allocs/op, a pass or held verdict, and the 5% report with the parallel-noise note. git diff --name-only lists only this story's files. go.mod and testdata/golden/ are unchanged.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
