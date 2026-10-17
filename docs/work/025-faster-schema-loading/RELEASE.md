# Release notes — #25: Faster schema loading: build the embedded 2020-12 meta-schema once and check schemas without annotations

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #25
- Status: approved by the operator at Checkpoint 3 (run #160, 2026-10-17 13:00 UTC)

**⚠️ the release-agent has concerns.**

## What changed

Schema compilation is now 70-82% faster on repeated loads: the embedded 2020-12 meta-schema and its vocabulary index are built once per process and reused, and the meta-schema check no longer collects unused annotations. All validation results and the public API are unchanged; every existing test passes unchanged. The verdict is warn because three test cases do not explicitly cover edge cases, though no functional issues were found.

- internal/compile/metacache.go: New shared cache for the embedded 2020-12 meta-schema and its index, built once per process via sync.Once
- internal/compile/metacheck.go: Updated to reuse the cached compiled meta-schema and run meta-schema checks without collecting annotations
- internal/eval/collect_option.go: New EvaluateNoAnnotations function runs evaluation without storing or sorting annotations, but still charges steps and enforces limits
- internal/registry/registry.go: Refactored BuildBuiltinIndex to build once and share as read-only pointers; AddBuiltins now layers shared resources into per-call registries
- docs/benchmarks.md: Added before-and-after benchmark numbers showing CompileOrder down 70-82% on ns/op, B/op and allocs/op
- internal/compile/bench_test.go: New BenchmarkFirstLoadOrder measures the one-shot cost of building the meta-schemas (~425 us, unchanged)
- compile_cache_limits_test.go: Regression test confirms all six limits (depth, ref-chain, steps, output-units, report-bytes, annotations) trigger at the same thresholds in a fresh CLI process and under the library cache
- jsonschema.go: metaCheck now calls EvaluateNoAnnotations instead of EvaluateWith

## How to verify

- Run go test -race ./... to confirm the shared meta-schema is safe for concurrent Compile and Validate calls, including the very first Compile
- Run go test ./compile_cache_limits_test.go to verify all six limits trigger at the same thresholds in a fresh process (CLI) and in the cache (library repeated Compile)
- Run go test -run '^$' -bench CompileOrder -benchmem ./. and confirm the repeated-compile benchmark is now ~125 us (was ~690 us)
- Run go test -run '^$' -bench FirstLoadOrder -benchmem ./internal/compile and confirm the first-load benchmark is still ~425 us
- Compile a schema twice in succession and confirm the second Compile is faster; then Validate the same document and confirm the result is identical to before the change

## Migration

None. All public APIs (Compiler, Options, Compile, Schema, Validate, Result) and validation results are unchanged.

## Rollback

If results differ on the operator's machine, revert these commits. The change is contained in internal packages and one call in jsonschema.go, so a revert has no public-API surface.

## Known limitations

- Custom meta-schemas that use unevaluatedProperties or unevaluatedItems could in theory depend on prior annotations; no such case exists in the test suite, but if one arises, the fix is to re-enable annotation collection for non-Draft2020 meta-schemas in metacheck.go
- When a caller registers a schema at a URI matching a built-in address, the behavior is unchanged but not explicitly tested; the test suite does not cover this overlap case
- The concurrency test does not guarantee a race on the very first sync.Once build in all timing conditions; -race enforces safety but may not always exercise the critical window

## Concerns

- The tester gave QA a 'warn' verdict for three medium test gaps: (1) no test of a custom meta-schema using unevaluated* keywords that depends on annotations, (2) no explicit test when a caller registers a schema at a built-in URI, and (3) the concurrency test may not reliably race the first build. None of these gaps cause a result change (all existing tests pass unchanged), but they leave specific edge cases without explicit coverage. If a future issue arises, the fix is localized to metacheck.go or registry.go.
- The benchmark improvement (70-82% on repeated Compile) applies only to repeated loads; the first Compile in a process still pays the full cost of building the meta-schemas, so a one-shot CLI run gets no benefit. The FirstLoadOrder benchmark is unchanged, which is correct.
- The embedded-resource index is built twice per process (once in registry.go and once in compile.go). Both are read-only after build, causing no correctness issue, but the redundancy was accepted as simpler than unifying the two caches.
