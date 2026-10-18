# SPEC #28: Golden CLI files from v1.1, then reuse evaluation path frames

- Date: 2026-10-17
- Type: performance
- Verdict: pass
- Status: passed gate 1 with nothing for the operator to decide (run #163, 2026-10-17 20:25 UTC)

## Problem

Round-3 performance stories need a fixed byte-for-byte guard of v1.1 CLI behaviour, and none exists yet. Separately, evaluation allocates a pathFrame plus a leave closure in enterNode (internal/eval/refguard.go) at every schema node. evalNode also allocates a lazyLoc, and DynamicScope.push returns a closure. Together these are about 27% of what Validate allocates.

## Why this matters

Every later round-3 story (206-211) is judged against the golden files captured here, so they must be taken from unchanged v1.1 production code before any edit and never regenerated. Path frames follow strict last-in first-out order, so a per-call stack indexed by depth with an embedded lazyLoc and explicit exit calls removes these allocations without changing results. This is the first and largest planned allocation cut.

## Acceptance criteria

- Before any production file is edited, golden files are captured in the run's own worktree (base d7198b7, production code identical to v1.1 f35d9e4; no git checkout or worktree is created). The golden header records 'captured from base d7198b7, production code identical to v1.1 f35d9e4'. Every case is captured twice; if any output differs between the two runs, the story stops before code changes and lists the cases. The capture program refuses to overwrite existing golden files.
- The golden data covers, in flag, basic, detailed and verbose, stdout, stderr and exit code for: testdata/perf fixtures; the generated perf documents (0.5 MB, 0.33 MB with 1,000 errors, 5.5 MB) piped through stdin; every docs/examples case; and every schema and test document in the JSON case files under testdata/ (annotations, applicator, core, metaschema, registry, unevaluated, validation, vocabulary), each as its own single-document CLI call. It also includes a few stdin and multi-document cases built from existing fixtures. All cases run from the project folder with relative paths; cases that need their original relative file name keep it under the temp folder. Outputs over about 64 KB and the generated documents are stored only as byte size plus SHA-256; smaller outputs are stored in full. The data lives in at most two files under testdata/golden/.
- A golden test builds the real CLI once per test run, runs every case as a separate process and compares stdout, stderr and exit code byte for byte. It runs in every normal go test ./... and under -race, takes under about 60 s without -race (grouping cases per schema if needed, with every document still covered), and never writes or updates a golden file.
- Evaluation uses a per-call stack of pathFrame indexed by depth, created fresh for each Validate call and never pooled or shared between calls. lazyLoc is embedded in the frame. enterNode/leave and DynamicScope.push use explicit exit calls instead of returned closures. The invariant that a frame is valid only while its node is being evaluated, and that frames are entered and left strictly last-in first-out, is documented next to the code.
- Thin enterNode, lazyLoc and DynamicScope.push compatibility entry points (or equivalent) keep refguard_internal_test.go, lazyloc_test.go, dynamicscope_test.go and units_verbose_test.go compiling and passing unchanged. No existing test is removed or weakened.
- A new verbose-output test with nested $ref, if and allOf checks the verbose report, including instance, keyword and absolute keyword locations.
- go test ./... and go test -race ./... pass, including the golden test. Validation results, reports, library Result, error messages, limits, exit codes and goroutine safety are unchanged. Only the standard library is used, and go.mod is unchanged.
- docs/benchmarks.md records before and after medians from go test -bench -count=10 on the same machine for all in-repo benchmarks, the per-scenario slowdown judgement (reported, not asserted), the final golden case count, and the targets: Validate5_5MB at most 280 MB B/op and 4.8 M allocs/op, ValidateErrors at most 22.5 MB B/op, CompileOrder at most 1,600 allocs/op, and no scenario more than 5% slower; on other hardware these are relative to the before numbers. If a target is missed, the numbers are recorded and the story is held for the operator's decision.
- The allocation limits in allocs_test.go are re-measured after the change and lowered on purpose to the measured count plus about 10% headroom, with the measured value recorded next to each limit and in docs/benchmarks.md.

## Non-goals

- Child EvalContext reuse (story 206), object keyword copies (207), chunked annotations (208), report rendering (209), multipleOf (210), built-in index (211)
- A shared sync.Pool of frame stacks across calls
- Changing verbose's second evaluation pass
- Regenerating or editing golden files after capture
- Any new dependency or comparison against another validator

## Tasks

- **T1** — Enumerate golden cases and capture v1.1 golden files
  - purpose: Before any production edit, build the CLI from the unchanged base, run every case twice, stop on any nondeterminism, and write the golden files with the provenance header. Refuse to overwrite existing files.
  - scope: cmd/jsonschema/golden_cases_test.go, cmd/jsonschema/golden_capture_test.go, testdata/golden/cli_v1_1.json, testdata/golden/cli_v1_1_large.json
  - done when: Golden files committed with the provenance header; the capture log shows two identical runs per case and the case count; no production file is modified at this point.
  - delivers: AC1, AC2
- **T2** — Golden comparison test
  - purpose: Build the real CLI once per test run, run every case and compare stdout, stderr and exit code byte for byte (size plus SHA-256 for large outputs). The test runs in every default and -race run.
  - scope: cmd/jsonschema/golden_test.go
  - done when: go test ./cmd/jsonschema -run Golden passes on the unchanged code in under about 60 s without -race, and also passes under -race.
  - depends on: T1
  - delivers: AC3
- **T3** — Per-call path frame stack with embedded lazyLoc
  - purpose: Replace the per-node pathFrame allocation, leave closure and lazyLoc allocation with a fresh per-call depth-indexed stack and an explicit exit call. Keep thin compatibility entry points and document the invariant.
  - scope: internal/eval/pathstack.go, internal/eval/refguard.go, internal/eval/eval.go, internal/eval/location.go, internal/eval/pathstack_test.go
  - done when: Existing internal/eval tests are unchanged and green; the new pathstack test covers last-in first-out enter/exit and per-call isolation; the golden test is green.
  - depends on: T2
  - delivers: AC4, AC5, AC7
- **T4** — DynamicScope explicit exit
  - purpose: Replace the closure returned by DynamicScope.push with an explicit exit call, keeping a compatibility push so dynamicscope_test.go compiles unchanged.
  - scope: internal/eval/dynamicscope.go, internal/eval/eval.go
  - done when: dynamicscope_test.go and dynamicscope_race_test.go pass unchanged; the golden test is green; go test -race ./... is green.
  - depends on: T3
  - delivers: AC4, AC5, AC7
- **T5** — Nested $ref/if/allOf verbose test
  - purpose: Add a verbose-output test whose nested $ref, if and allOf exercise frame reuse across depths.
  - scope: verbose_nested_frames_test.go
  - done when: The new test passes with and without -race.
  - depends on: T3, T4
  - delivers: AC6, AC7
- **T6** — Benchmarks, allocation limits and docs
  - purpose: Record before and after medians (count=10), the slowdown judgement, target status and golden case count. Lower the limits in allocs_test.go with the measured values recorded.
  - scope: docs/benchmarks.md, allocs_test.go
  - done when: docs/benchmarks.md has both sets of numbers, the target table and the case count; allocs_test.go limits are lowered and pass; any missed target is flagged for the operator.
  - depends on: T3, T4, T5
  - delivers: AC8, AC9
