# PLAN #16: Limits against hostile input and proof of the load-once guarantee

- Date: 2026-10-17
- Status: proposed (pending Checkpoint 2 sign-off)
- Design: see `ADR-016-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `internal/limits/limits.go`
- `internal/limits/limits_test.go`
- `limits.go`
- `limits_test.go`
- `options.go`
- `numberlimits.go`
- `numberlimits_test.go`
- `jsonschema.go`
- `internal/jsonvalue/limits.go`
- `internal/jsonvalue/decode.go`
- `internal/jsonvalue/depth_option_test.go`
- `internal/compile/limits.go`
- `internal/compile/compile.go`
- `internal/compile/limits_test.go`
- `internal/registry/registry.go`
- `internal/registry/budget_test.go`
- `internal/eval/budget.go`
- `internal/eval/eval.go`
- `internal/eval/budget_test.go`
- `internal/keywords/registry.go`
- `internal/keywords/core/ref.go`
- `internal/keywords/core/dynamicref.go`
- `output.go`
- `internal/ecmaregex/translate.go`
- `internal/eval/annotations.go`
- `cmd/jsonschema/limitflags.go`
- `cmd/jsonschema/limitflags_test.go`
- `cmd/jsonschema/main.go`
- `cmd/jsonschema/run.go`
- `cmd/jsonschema/parse_errors_test.go`
- `concurrency_race_test.go`

Modules the architecture expects to touch:

- `internal/limits/limits.go`
- `internal/limits/limits_test.go`
- `limits.go`
- `limits_test.go`
- `options.go`
- `numberlimits.go`
- `numberlimits_test.go`
- `jsonschema.go`
- `output.go`
- `concurrency_race_test.go`
- `internal/jsonvalue/limits.go`
- `internal/jsonvalue/decode.go`
- `internal/jsonvalue/depth_option_test.go`
- `internal/compile/limits.go`
- `internal/compile/compile.go`
- `internal/compile/limits_test.go`
- `internal/eval/budget.go`
- `internal/eval/eval.go`
- `internal/eval/budget_test.go`
- `internal/eval/annotations.go`
- `internal/keywords/registry.go`
- `internal/keywords/core/ref.go`
- `internal/keywords/core/dynamicref.go`
- `internal/registry/registry.go`
- `internal/registry/budget_test.go`
- `internal/ecmaregex/translate.go`
- `cmd/jsonschema/limitflags.go`
- `cmd/jsonschema/limitflags_test.go`
- `cmd/jsonschema/main.go`
- `cmd/jsonschema/run.go`
- `cmd/jsonschema/parse_errors_test.go`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — Shared limit model, defaults and LimitError
2. **T2** — Depth and size limits for schemas, documents and the registry  _(after T1)_
3. **T3** — Reference-chain and work-step budget per evaluation  _(after T1)_
4. **T4** — Bring existing limits into the same style  _(after T1)_
5. **T5** — CLI limit flags and limit-error reporting  _(after T2, T3, T4)_
6. **T6** — Load-once proof under the race detector  _(after T3, T4)_

## Risks

- Step counting in the hot path could slow validation; keep it a single increment and compare.
- Ref-chain counting must count jumps on the current path (decrement on return), or deep but finite recursion will be miscounted.
- Deep recursion in compile/eval before the depth check could still overflow the goroutine stack; check before descending.
- Changing the stderr wording may break existing CLI tests from stories 178-180.

Implementation constraints the coder must honor:

- Work limit is a deterministic step count, never time-based.
- Limits cannot be disabled; zero means default and negative is rejected.
- Tests read limit values from exported constants or Options, never literal numbers.
- Limit error messages never quote instance values.
- Compiled Schema is never mutated; counters live per call.
- No new dependencies; go.mod unchanged.
- New tests go in new files; existing tests are edited only where a changed message or flag breaks them.

- DB impact: no · API impact: yes: new Options fields, LimitError type, Default* constants, new CLI flags · Migration needed: no

Breaking changes:

- Errors from existing number/depth caps become *jsonschema.LimitError; stderr wording for oversize inputs changes to the unified format.

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: go test ./ ./internal/limits passes; tests read defaults from the constants
- **T2**: Depth, input-size and registry-total tests pass at, below and above each limit
- **T3**: Self-ref and mutual-ref cycles and nested-anyOf blow-ups stop with LimitError quickly; deep legitimate recursion passes
- **T4**: Each existing cap returns a LimitError naming it and its set value
- **T5**: CLI contract tests per flag: exit 2, empty stdout, one stderr line; multi-document worst-result exit; negative value is a usage error
- **T6**: go test -race ./... passes with the new test

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
