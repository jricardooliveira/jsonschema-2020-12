# ADR #28: Golden CLI files from v1.1, then reuse evaluation path frames

- Date: 2026-10-17
- Story: #28
- Status: proposed (pending architecture sign-off)

## Decision
The golden harness lives in new test files in cmd/jsonschema: golden_cases_test.go enumerates cases, golden_capture_test.go is an opt-in capture behind a test flag that refuses to overwrite, and golden_test.go compares. It builds the CLI once with the Go toolchain into a temp dir and runs it with os/exec from the project folder using relative paths; generated perf documents come from internal/perfdata and are piped via stdin. The golden data is kept in two JSON files under testdata/golden/ (small outputs in full; large outputs and generated documents as size plus SHA-256), each with the provenance header. In internal/eval, a new pathstack.go holds the per-call []pathFrame indexed by depth, with lazyLoc embedded by value, plus enter/exit methods and the documented last-in first-out validity invariant. The stack is owned by the per-call evaluation state, allocated fresh per Validate and never pooled. refguard.go keeps enterNode as a thin wrapper over the stack so refguard_internal_test.go compiles; location.go keeps the lazyLoc type and constructor for lazyloc_test.go; dynamicscope.go adds an explicit exit and keeps push as a wrapper for dynamicscope_test.go. eval.go's evalNode switches to the explicit enter/exit calls; this crowded file is touched only at the call sites. The perf benchmarks and allocs_test.go come from stories 198 and 204.

## Affected modules
- cmd/jsonschema/golden_cases_test.go
- cmd/jsonschema/golden_capture_test.go
- cmd/jsonschema/golden_test.go
- testdata/golden/cli_v1_1.json
- testdata/golden/cli_v1_1_large.json
- internal/eval/pathstack.go
- internal/eval/pathstack_test.go
- internal/eval/refguard.go
- internal/eval/eval.go
- internal/eval/location.go
- internal/eval/dynamicscope.go
- verbose_nested_frames_test.go
- allocs_test.go
- docs/benchmarks.md

## Data / API impact
- DB impact: no
- API impact: no
- Migration needed: no

## Constraints for implementation
- Capture golden files before editing any production file; never edit or regenerate them afterwards.
- No git checkout or worktree; git is inspection-only.
- A frame stack is fresh per Validate call; no sync.Pool and no state shared between calls.
- No frame pointer may outlive its node's evaluation; frames are entered and left strictly last-in first-out through explicit exit calls, never closures.
- Existing test files stay byte-unchanged; compatibility entry points keep them compiling.
- Standard library only; go.mod unchanged.
- No timing or peak-memory assertions in tests.
- The golden test runs in default go test and under -race, in under about 60 s without -race.

## Risks
- Stack growth by slice append may move frames; any retained *pathFrame across a deeper enter could dangle. Use index-based access or guarantee no pointer is held across a child enter.
- Some testdata cases may produce nondeterministic output (map order), which would stop the story at capture.
- Golden test runtime over all testdata documents in four formats may exceed 60 s; grouping per schema may be needed.
- Verbose's second pass reuses the stack; locations must render identically.
- Targets measured on M3 Max may be missed on the build machine and require an operator hold.

## Breaking changes
- none

## Sensitivity
- none
