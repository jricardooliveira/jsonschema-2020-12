# PLAN #25: Faster schema loading: build the embedded 2020-12 meta-schema once and check schemas without annotations

- Date: 2026-10-17
- Status: proposed (pending Checkpoint 2 sign-off)
- Design: see `ADR-025-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `internal/compile/bench_test.go`
- `docs/benchmarks.md`
- `internal/compile/metacache.go`
- `internal/compile/metaschema.go`
- `internal/compile/compile.go`
- `internal/registry/builtin.go`
- `internal/registry/registry.go`
- `internal/compile/metacheck.go`
- `internal/eval/annotations.go`
- `internal/eval/collect_option.go`
- `internal/compile/metacache_test.go`
- `metacache_race_test.go`
- `compile_cache_limits_test.go`

Modules the architecture expects to touch:

- `internal/compile/metacache.go`
- `internal/compile/metacache_test.go`
- `internal/compile/metaschema.go`
- `internal/compile/compile.go`
- `internal/compile/metacheck.go`
- `internal/compile/bench_test.go`
- `internal/registry/builtin.go`
- `internal/registry/registry.go`
- `internal/eval/annotations.go`
- `internal/eval/collect_option.go`
- `metacache_race_test.go`
- `compile_cache_limits_test.go`
- `docs/benchmarks.md`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — Baseline numbers and first-load benchmark
2. **T2** — Build the embedded meta-schemas and index once per process  _(after T1)_
3. **T3** — Meta-schema check without annotations  _(after T2)_
4. **T4** — Cache, concurrency and unchanged-behaviour tests  _(after T2, T3)_
5. **T5** — Hostile-input limit regression test  _(after T3)_
6. **T6** — After numbers and documentation  _(after T4, T5)_

## Risks

- Format-assertion mode may today apply to the meta-schema check (the meta-schema uses format, for example uri-reference). If it does, a cached check must still honour the per-call mode, or results change.
- If caller-supplied limits are applied while compiling the embedded meta-schema today, caching with fixed settings could move where a limit triggers. The regression test in T5 must catch this.
- A registry view that wraps the shared index could change lookup precedence when a caller registers a resource under a built-in URI. The existing precedence must be preserved exactly.
- Turning off annotation collection in the meta check could alter evaluation if any keyword in the meta-schema path relies on annotations, such as unevaluatedProperties or unevaluatedItems in a custom meta-schema. In that case, collection stays on for that check only.
- Story 201 also changes internal/eval/annotations.go. That is file overlap, not a build dependency.

Implementation constraints the coder must honor:

- Build the shared set lazily on first Compile with sync.Once (or sync.OnceValues). Never build it at package init or import.
- Nothing reachable from the shared set may be written after it is built. Per-call state (dynamic scope, limit counters, annotations, caller registry entries) stays in per-call structures.
- The shared build must not depend on caller Options. If any caller option (limits, format-assertion mode, base URI) affects how the embedded meta-schema is compiled or checked today, it must stay applied per call, not baked into the cache.
- The meta-schema check must charge evaluation steps, depth and ref-chain counters exactly as before. Only annotation collection and sorting are skipped.
- Load-error text and order must be byte-identical to today.
- Do not add or change any allocation-limit test (allocs_test.go).
- Standard library only, with no new go.mod require line.
- Tests read limit thresholds from the code's default values instead of repeating numbers.
- No timing assertions in ordinary tests. Speed is shown by benchmarks only.

- DB impact: no · API impact: no · Migration needed: no

Breaking changes:

- (none)

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: go test -run '^$' -bench . -benchmem ./... output before the change is recorded in docs/benchmarks.md, including the new first-load benchmark.
- **T2**: The existing metaschema, vocabulary, registry_refs and refs tests pass unchanged, and a new unit test shows the built-in build function runs once across many Compile calls.
- **T3**: internal/compile/metacheck_test.go and cmd/jsonschema/metaschema_cli_test.go pass unchanged, and a new test shows the meta check records no annotations.
- **T4**: go test -race ./... passes.
- **T5**: The new test passes, and it would fail if a limit triggered at a different input or returned a different error or exit code.
- **T6**: docs/benchmarks.md holds before and after tables for every scenario plus the first-load benchmark, with compile allocs/op, B/op and ns/op all lower.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
