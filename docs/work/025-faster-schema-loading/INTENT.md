# INTENT #25

- Author: Joao Oliveira
- Date: 2026-10-17
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Make loading (compiling) a schema faster and lighter. Profiling shows four costs on every load: the registry index and built-in meta-schemas are rebuilt, the 2020-12 meta-schema is recompiled, and annotations are collected and sorted during the meta-schema check. Fix all four. Build the embedded 2020-12 meta-schema and its vocabulary meta-schemas, with their index, once per process, and share the compiled result. The shared result must be immutable and safe to share across goroutines. Check the schema against the meta-schema without collecting or sorting annotations, because nothing ever shows them. Caller-supplied registry schemas, custom meta-schemas and $vocabulary handling must behave exactly as before. A schema that fails its meta-schema is still an invalid-schema error with exit code 2. Results must stay identical everywhere: the same valid or invalid answer, the same flag, basic, detailed and verbose reports, the same library Result, and the same load errors. All existing tests must pass unchanged. Every hostile-input limit must keep working exactly as before. That covers nesting depth, ref chain, evaluation-step charging, output units, report bytes and annotation count. Each must trigger at exactly the same inputs, with the same error and exit code 2. Add one regression test that proves this. The compiled Schema must stay immutable and safe for concurrent use, checked with go test -race. Run the project's benchmarks on the same machine before and after the change, and add both sets of numbers to docs/benchmarks.md. The change must show a reduction in allocs/op, B/op and ns/op on the compile benchmark. No scenario's allocs/op or B/op may increase.

## Operator clarifications
- The project has tests that fail if loading a schema uses more memory allocations than an agreed limit. After this change, loading will use far fewer. What should happen to that limit? — No allocation-limit tests exist yet: story 'Allocation limits in tests' adds every limit once, after all five fixes, from the measured counts plus about 10% (the user's answer). This story adds or changes no allocation-limit test; it records its before and after benchmark numbers in docs/benchmarks.md.
- The story asks for a drop in time, memory and allocations when loading a schema. How big a drop counts as done? Today, loading your order schema takes 1.1 ms and 28,000 allocations. — Any clear drop, with the numbers recorded — Recommended. The story passes if all three numbers go down and none get worse. How close this gets to the other validator is judged by your own comparison outside the project.
- The built-in 2020-12 meta-schema will be prepared once and then reused. When should that one-time work happen? — The first time any schema is loaded — Recommended. A program that imports the library but never loads a schema pays nothing. The first load is a little slower, and every load after that is fast.
- The compile benchmark repeats the load many times, so it shows the fast reuse case. It does not show the one-time first load. Should docs/benchmarks.md also report that first load? — Yes, add a separate first-load benchmark — Recommended. It shows what a command-line run, which loads one schema once, actually pays. It also stops the gain from looking bigger than it is.

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- The inventory is truncated past internal/keywords/applicator, so the internal/registry file names (registry.go, builtin.go) and the exact location of the compile benchmark from story 198 (root bench_test.go or internal/compile/bench_test.go) are not confirmed. The coder must adjust these claims to the real files.
- It is not known whether the current meta-schema check applies format-assertion mode or caller limits. Either changes what may be cached, see risks.
- It is not known whether internal/eval already has a switch to turn annotation collection off. If story 201 lands first with such a switch, T3 should reuse it instead of adding internal/eval/collect_option.go.
- Story 198 (benchmarks and docs/benchmarks.md) is approved but not yet released. This story needs its compile benchmark and the order schema in testdata/perf for the before and after numbers.
- The before and after benchmark runs must be on the same machine. Whether the factory's build machine matches the operator's baseline Mac is unknown, so the numbers in docs/benchmarks.md may differ from the operator's 1.1 ms and 28k allocations baseline.

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
