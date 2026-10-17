# PLAN #28: Golden CLI files from v1.1, then reuse evaluation path frames

- Date: 2026-10-17
- Status: proposed (pending Checkpoint 2 sign-off)
- Design: see `ADR-028-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `cmd/jsonschema/golden_cases_test.go`
- `cmd/jsonschema/golden_capture_test.go`
- `testdata/golden/cli_v1_1.json`
- `testdata/golden/cli_v1_1_large.json`
- `cmd/jsonschema/golden_test.go`
- `internal/eval/pathstack.go`
- `internal/eval/refguard.go`
- `internal/eval/eval.go`
- `internal/eval/location.go`
- `internal/eval/pathstack_test.go`
- `internal/eval/dynamicscope.go`
- `verbose_nested_frames_test.go`
- `docs/benchmarks.md`
- `allocs_test.go`

Modules the architecture expects to touch:

- `cmd/jsonschema/golden_cases_test.go`
- `cmd/jsonschema/golden_capture_test.go`
- `cmd/jsonschema/golden_test.go`
- `testdata/golden/cli_v1_1.json`
- `testdata/golden/cli_v1_1_large.json`
- `internal/eval/pathstack.go`
- `internal/eval/pathstack_test.go`
- `internal/eval/refguard.go`
- `internal/eval/eval.go`
- `internal/eval/location.go`
- `internal/eval/dynamicscope.go`
- `verbose_nested_frames_test.go`
- `allocs_test.go`
- `docs/benchmarks.md`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — Enumerate golden cases and capture v1.1 golden files
2. **T2** — Golden comparison test  _(after T1)_
3. **T3** — Per-call path frame stack with embedded lazyLoc  _(after T2)_
4. **T4** — DynamicScope explicit exit  _(after T3)_
5. **T5** — Nested $ref/if/allOf verbose test  _(after T3, T4)_
6. **T6** — Benchmarks, allocation limits and docs  _(after T3, T4, T5)_

## Risks

- Stack growth by slice append may move frames; any retained *pathFrame across a deeper enter could dangle. Use index-based access or guarantee no pointer is held across a child enter.
- Some testdata cases may produce nondeterministic output (map order), which would stop the story at capture.
- Golden test runtime over all testdata documents in four formats may exceed 60 s; grouping per schema may be needed.
- Verbose's second pass reuses the stack; locations must render identically.
- Targets measured on M3 Max may be missed on the build machine and require an operator hold.

Implementation constraints the coder must honor:

- Capture golden files before editing any production file; never edit or regenerate them afterwards.
- No git checkout or worktree; git is inspection-only.
- A frame stack is fresh per Validate call; no sync.Pool and no state shared between calls.
- No frame pointer may outlive its node's evaluation; frames are entered and left strictly last-in first-out through explicit exit calls, never closures.
- Existing test files stay byte-unchanged; compatibility entry points keep them compiling.
- Standard library only; go.mod unchanged.
- No timing or peak-memory assertions in tests.
- The golden test runs in default go test and under -race, in under about 60 s without -race.

- DB impact: no · API impact: no · Migration needed: no

Breaking changes:

- (none)

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: Golden files committed with the provenance header; the capture log shows two identical runs per case and the case count; no production file is modified at this point.
- **T2**: go test ./cmd/jsonschema -run Golden passes on the unchanged code in under about 60 s without -race, and also passes under -race.
- **T3**: Existing internal/eval tests are unchanged and green; the new pathstack test covers last-in first-out enter/exit and per-call isolation; the golden test is green.
- **T4**: dynamicscope_test.go and dynamicscope_race_test.go pass unchanged; the golden test is green; go test -race ./... is green.
- **T5**: The new test passes with and without -race.
- **T6**: docs/benchmarks.md has both sets of numbers, the target table and the case count; allocs_test.go limits are lowered and pass; any missed target is flagged for the operator.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
