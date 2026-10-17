# ADR #16: Limits against hostile input and proof of the load-once guarantee

- Date: 2026-10-17
- Story: #16
- Status: proposed (pending architecture sign-off)

## Decision
New package internal/limits holds the limit names, the resolved Limits value (zero=default, negative=error) and an internal Error{Limit, Value}. Packages below the root (jsonvalue, compile, eval, registry, output, regex translation) return it. Root limits.go defines the public LimitError and Default* constants and converts the internal error, so callers use errors.As. Per-evaluation counters (steps, ref-chain depth) live in a budget struct created per Validate call in internal/eval/budget.go and reached through EvalContext. Nothing is stored on the compiled Node. Crowded files (compile.go, eval.go, registry.go, jsonschema.go, options.go, main.go, run.go) get only wiring lines; logic goes in new files. CLI flags live in cmd/jsonschema/limitflags.go, with one call from main.go and one error-mapping call from run.go.

## Affected modules
- internal/limits/limits.go
- internal/limits/limits_test.go
- limits.go
- limits_test.go
- options.go
- numberlimits.go
- numberlimits_test.go
- jsonschema.go
- output.go
- concurrency_race_test.go
- internal/jsonvalue/limits.go
- internal/jsonvalue/decode.go
- internal/jsonvalue/depth_option_test.go
- internal/compile/limits.go
- internal/compile/compile.go
- internal/compile/limits_test.go
- internal/eval/budget.go
- internal/eval/eval.go
- internal/eval/budget_test.go
- internal/eval/annotations.go
- internal/keywords/registry.go
- internal/keywords/core/ref.go
- internal/keywords/core/dynamicref.go
- internal/registry/registry.go
- internal/registry/budget_test.go
- internal/ecmaregex/translate.go
- cmd/jsonschema/limitflags.go
- cmd/jsonschema/limitflags_test.go
- cmd/jsonschema/main.go
- cmd/jsonschema/run.go
- cmd/jsonschema/parse_errors_test.go

## Data / API impact
- DB impact: no
- API impact: yes: new Options fields, LimitError type, Default* constants, new CLI flags
- Migration needed: no

## Constraints for implementation
- Work limit is a deterministic step count, never time-based.
- Limits cannot be disabled; zero means default and negative is rejected.
- Tests read limit values from exported constants or Options, never literal numbers.
- Limit error messages never quote instance values.
- Compiled Schema is never mutated; counters live per call.
- No new dependencies; go.mod unchanged.
- New tests go in new files; existing tests are edited only where a changed message or flag breaks them.

## Risks
- Step counting in the hot path could slow validation; keep it a single increment and compare.
- Ref-chain counting must count jumps on the current path (decrement on return), or deep but finite recursion will be miscounted.
- Deep recursion in compile/eval before the depth check could still overflow the goroutine stack; check before descending.
- Changing the stderr wording may break existing CLI tests from stories 178-180.

## Breaking changes
- Errors from existing number/depth caps become *jsonschema.LimitError; stderr wording for oversize inputs changes to the unified format.

## Sensitivity
- security
