# ADR #29: One reusable child evaluation context per depth

- Date: 2026-10-17
- Story: #29
- Status: proposed (pending architecture sign-off)

## Decision
The per-call collector gets a slice of *EvalContext indexed by child depth. EvalChild(depth d) takes stack[d], growing the stack only on a new maximum depth. It resets the context for the child (clearing errors, units, annotation marks and evaluated state while keeping capacity), evaluates, hands the units slice to the caller (detached so the next reset cannot overwrite it), and releases the context. Because evaluation is strictly last-in first-out, one context per depth is enough. The stack lives in per-call state only, so a compiled Schema stays immutable and goroutine-safe. A test-only poison step marks a released context as dead, so accessors panic if a keyword kept it; it is active only when testing.Testing() is true, which no CLI or golden output can observe. Any keyword that genuinely needs its child context afterwards is served a fresh context from registry.go, without editing that keyword's file.

## Affected modules
- internal/keywords/registry.go
- internal/keywords/childctx.go
- internal/keywords/childctx_test.go
- internal/keywords/childctx_poison.go
- internal/keywords/childctx_poison_test.go
- internal/eval/eval.go
- childctx_race_test.go
- allocs_childctx_test.go
- docs/benchmarks/r3-206-childctx.md

## Data / API impact
- DB impact: no
- API impact: no
- Migration needed: no

## Constraints for implementation
- Results, output, exit codes, limits, error messages and goroutine safety must not change.
- No golden file is edited. Any golden diff means the story fails and stops.
- Do not write allocs_test.go, docs/benchmarks.md, or any file claimed by #207, #210 or #211.
- Reset must clear every per-child field, so no state leaks between siblings.
- The returned units slice must not alias the reused context's buffer after the next reset.
- The poison check must not allocate and must not run in production binaries.
- Standard library only. go.mod is unchanged.

## Risks
- A keyword that keeps its child context (for example for verbose units) would produce a wrong report that tests may not notice. The poison check is the mitigation.
- If units are detached by copying, the B/op saving shrinks. Ownership must be handed over without an extra copy where possible.
- If testing.Testing() poisoning adds cost inside benchmarks, B/op and allocs/op are unaffected (it only zeroes a struct), but ns/op may shift slightly.
- The combined 200 MB target may be missed. In that case the story is held for the operator.

## Breaking changes
- none

## Sensitivity
- none
