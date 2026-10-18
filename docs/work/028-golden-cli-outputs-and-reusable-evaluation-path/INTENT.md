# INTENT #28

- Author: Joao Oliveira
- Date: 2026-10-17
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

First, before any production code changes, capture golden files from the unchanged v1.1 code (commit f35d9e4). For each case, record the CLI's stdout, stderr and exit code in all four output formats: flag, basic, detailed and verbose. The cases are the perf fixtures in testdata/perf, the examples in docs/examples and the testdata cases. Keep the golden data in a small number of files so the story stays within the factory's file limits. Store the generated large documents only as their size plus SHA-256, not as content. Add a test that compares the current CLI output byte for byte against these golden files. Commit the golden files with this story. No later story may regenerate or edit them. If output differs from a golden file, the story has failed and must stop; the golden file is never updated. Then cut evaluation allocations by reusing path frames. Today enterNode in internal/eval/refguard.go allocates a pathFrame and returns a leave closure at every schema node. evalNode also allocates a lazyLoc, and DynamicScope.push returns a closure. Together these are about 27% of what Validate allocates. Replace them with a per-call stack of pathFrame indexed by depth, since the path is strictly last-in first-out. Embed lazyLoc in the frame, and replace the closures with an explicit exit call. Document the invariant that a frame is only valid while its node is being evaluated. Add a verbose-output test with nested $ref, if and allOf. Validation behaviour, output, exit codes, limits, error messages and goroutine safety must not change. Use the standard library only. Do not remove or weaken any existing test. go test -race must pass. Targets, measured on an Apple M3 Max with Go 1.27.2 (on other hardware, compare relative to the before numbers): BenchmarkValidate5_5MB at most 280 MB B/op and at most 4.8 M allocs/op; ValidateErrors at most 22.5 MB B/op; CompileOrder at most 1,600 allocs/op; no scenario more than 5% slower. Judge 'slower' by running each benchmark 10 times before and after with go test -count=10 and comparing medians by hand or with a small script, using only the standard toolchain. Report this result; do not assert it in a test. Record before and after benchmark numbers from the in-repo benchmarks on the same machine in docs/benchmarks.md. Re-measure the allocation limits in allocs_test.go and lower them on purpose, as docs/benchmarks.md describes. If a target is missed, record the numbers and hold the story for the operator's decision; do not mark it done silently.

## Operator clarifications
- On the 5.5 MB document, the 'verbose' and 'detailed' reports can be tens of megabytes each. How should the golden files keep the output of the large generated documents? — Size plus SHA-256 for large outputs — Recommended. Any output over about 64 KB is stored as its byte size and SHA-256, exactly like the large documents; small outputs are stored in full. Any change of even one byte is still caught, and the repository stays small. Cost: when a large output differs, the test says it differs but cannot show where.
- Comparing the CLI output on the large documents in all four report shapes takes a few seconds on every test run. When should that comparison run? — Always, in every normal test run — Recommended. The golden check is the guard every later story relies on, so it runs on every 'go test ./...' and also under -race. Cost: a few extra seconds per test run.
- Some error messages include a file's path, for example when a schema cannot be read. A path such as /Users/joao/... differs on every machine, so a golden file captured on your Mac could fail on the build machine even though nothing changed. How should the golden test handle file paths? — Run every case with paths relative to the project folder — Recommended. The test always runs the tool from the project folder with short relative paths such as testdata/perf/order.json. The output is the same on every machine, so it can be compared byte for byte with no rewriting.
- How should the golden test run the command-line tool for each case? — Build the real tool once and run it as a separate program — Recommended. The test builds the tool once per test run, then runs it for each case and records exactly what a script would see: stdout, stderr and exit code. This is the most faithful check. Cost: a few extra seconds to build.
- The listed cases each check one document from a file. The tool can also read a document piped in through standard input, and it can check several documents in one run (one report per line). Should the golden files also cover these two ways of running it? — Yes, add a few standard-input and several-document cases — Recommended. A handful of extra cases built from the existing fixtures, in all four report shapes. Later stories touch report writing, so these paths are guarded too. Cost: slightly larger golden data.
- The new stack of path frames can be made fresh for each check, or kept in a shared pool and reused across checks. A shared pool saves a little more memory on many small documents, but it adds sharing between checks that run at the same time. Which should this story use? — Fresh stack for each check — Recommended. Each check gets its own stack, which is thrown away afterwards. Nothing is shared between checks running at the same time, so the existing goroutine safety is clearly unchanged. Cost: small documents save slightly less.
- While capturing the golden files from v1.1, the story may find a case whose output changes from one run to the next, for example a list printed in a different order each time. A golden file for such a case would fail at random. What should the story do? — Stop and report it to you — Recommended. The story captures each case twice. If any output differs between the two runs, it stops before changing any code and lists the cases, so you decide. Nothing is hidden or quietly left out.
- Operator note on refining again — Operator note settling the open points (everything else in the plan stays):
1. Do NOT create a git worktree or check out f35d9e4: git in the coder session is inspection-only. The run's base d7198b7 has production code byte-identical to f35d9e4 (git diff f35d9e4 d7198b7 touches only docs/work/*.md and project-spec.json; internal/perfdata, docs/examples and testdata/perf all exist). So T1 captures in the run's own worktree BEFORE any production file is edited, and the golden files' header records "captured from base d7198b7, production code identical to v1.1 f35d9e4". The capture program refuses to overwrite existing golden files, as planned.
2. "The testdata cases" = every schema and every test document in the JSON case files under testdata/ (the spec-derived groups in annotations, applicator, core, metaschema, registry, unevaluated, validation, vocabulary), each document as its own single-document CLI call, in all four formats, plus testdata/perf and the generated perf documents through stdin as planned. Cases that need their original relative file name (refs, registry dirs) keep it under the temp folder. The whole golden test should stay under about 60 s without -race; if it would not, group the cases per schema but keep every document covered. Report the final case count in docs/benchmarks.md.
3. Keep the existing internal test files unchanged: keep thin enterNode / lazyLoc / DynamicScope.push compatibility entry points (or equivalent) so refguard_internal_test.go, lazyloc_test.go, dynamicscope_test.go and units_verbose_test.go compile and pass as they are. Only if that is impossible may a call site be adapted, every assertion kept, and the file named at Checkpoint 3.
4. If code outside internal/eval calls enterNode or DynamicScope.push, ask for that file at the gate; do not duplicate logic.
5. No existing test is removed or weakened; golden files are never regenerated after T1.

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- The inventory is truncated: it is not confirmed whether code outside internal/eval (for example internal/keywords/registry.go or keyword files) calls enterNode or DynamicScope.push. If so, that file must be requested at the gate and added to the claims.
- Whether lazyLoc is defined in internal/eval/location.go or in another internal/eval file is assumed, not verified.
- Whether internal/eval/units.go holds frame or lazyLoc pointers past node evaluation, for example for verbose units, is unknown; it may need a write claim.
- Some testdata cases may be nondeterministic; capture would then stop for an operator decision.
- The golden test runtime across all testdata documents in four formats is unmeasured; grouping per schema may be needed to stay under about 60 s.
- The 280 MB / 4.8 M allocs / 22.5 MB / 1,600 allocs targets come from an M3 Max; the build machine may differ, and a miss holds the story.
- CompileOrder at most 1,600 allocs/op may depend on compile-time paths outside internal/eval; whether frame reuse alone reaches it is unverified.

## Proposed outcome

See `SPEC.md` in this directory — the spec-agent's acceptance criteria are
the machine-readable statement of the outcome, derived from the problem above.

## Affected users and systems

See `PLAN.md` → *Files that change*, and the project's `project-spec.json`
for the systems in scope.

## Constraints

The project's `PROJECT_RULES.md` and `project-spec.json` (stack, conventions,
NFRs and the `forbidden` list) are injected into every agent and bind this work.

## Open questions

Recorded in `SPEC.md` → *Open questions*. A non-empty list parks the run at
Checkpoint 1 for the operator rather than failing it.
