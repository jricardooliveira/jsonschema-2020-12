# ADR #25: Faster schema loading: build the embedded 2020-12 meta-schema once and check schemas without annotations

- Date: 2026-10-17
- Story: #25
- Status: proposed (pending architecture sign-off)

## Decision
New file internal/compile/metacache.go holds an unexported builtinSet. It contains the compiled 2020-12 meta-schema, the compiled vocabulary meta-schemas and an immutable registry index of the embedded resources, and it is built by an unexported buildBuiltin() behind sync.Once on first use. Compile takes the shared set and builds a per-call registry view: caller resources are layered over the shared built-in index (copy-on-write or a two-level lookup) so the shared index is never written. Custom meta-schemas supplied through the registry are still compiled per call, as today, and may $dynamicRef into the shared vocabulary schemas. The meta-schema check in metacheck.go calls evaluation with an annotation-collection-off flag held in the per-call evaluation context (new internal/eval/collect_option.go, a small check in internal/eval/annotations.go). The check therefore still charges evaluation steps and depth exactly as before but never appends or sorts annotations. The first-load benchmark calls buildBuiltin() directly on each iteration, which avoids a reset hook on the sync.Once. compile.go (crowded) is touched only to replace the per-call build with a call to the shared set.

## Affected modules
- internal/compile/metacache.go
- internal/compile/metacache_test.go
- internal/compile/metaschema.go
- internal/compile/compile.go
- internal/compile/metacheck.go
- internal/compile/bench_test.go
- internal/registry/builtin.go
- internal/registry/registry.go
- internal/eval/annotations.go
- internal/eval/collect_option.go
- metacache_race_test.go
- compile_cache_limits_test.go
- docs/benchmarks.md

## Data / API impact
- DB impact: no
- API impact: no
- Migration needed: no

## Constraints for implementation
- Build the shared set lazily on first Compile with sync.Once (or sync.OnceValues). Never build it at package init or import.
- Nothing reachable from the shared set may be written after it is built. Per-call state (dynamic scope, limit counters, annotations, caller registry entries) stays in per-call structures.
- The shared build must not depend on caller Options. If any caller option (limits, format-assertion mode, base URI) affects how the embedded meta-schema is compiled or checked today, it must stay applied per call, not baked into the cache.
- The meta-schema check must charge evaluation steps, depth and ref-chain counters exactly as before. Only annotation collection and sorting are skipped.
- Load-error text and order must be byte-identical to today.
- Do not add or change any allocation-limit test (allocs_test.go).
- Standard library only, with no new go.mod require line.
- Tests read limit thresholds from the code's default values instead of repeating numbers.
- No timing assertions in ordinary tests. Speed is shown by benchmarks only.

## Risks
- Format-assertion mode may today apply to the meta-schema check (the meta-schema uses format, for example uri-reference). If it does, a cached check must still honour the per-call mode, or results change.
- If caller-supplied limits are applied while compiling the embedded meta-schema today, caching with fixed settings could move where a limit triggers. The regression test in T5 must catch this.
- A registry view that wraps the shared index could change lookup precedence when a caller registers a resource under a built-in URI. The existing precedence must be preserved exactly.
- Turning off annotation collection in the meta check could alter evaluation if any keyword in the meta-schema path relies on annotations, such as unevaluatedProperties or unevaluatedItems in a custom meta-schema. In that case, collection stays on for that check only.
- Story 201 also changes internal/eval/annotations.go. That is file overlap, not a build dependency.

## Breaking changes
- none

## Sensitivity
- security
