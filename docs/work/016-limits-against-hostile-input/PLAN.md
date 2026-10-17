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

- `limits.go`
- `limits_test.go`
- `numberlimits.go`
- `jsonschema.go`
- `output.go`
- `registry_options.go`
- `concurrency_race_test.go`
- `internal/jsonvalue/limits.go`
- `internal/jsonvalue/decode.go`
- `internal/jsonvalue/depth_option_test.go`
- `internal/compile/refs.go`
- `internal/eval/budget.go`
- `internal/eval/budget_test.go`
- `internal/eval/annotations.go`
- `internal/eval/refguard.go`
- `internal/eval/refguard_test.go`
- `internal/eval/evaluated.go`
- `internal/eval/eval.go`
- `internal/keywords/registry.go`
- `internal/keywords/validation/enum.go`
- `internal/keywords/validation/uniqueitems.go`
- `internal/keywords/validation/pattern.go`
- `internal/keywords/applicator/patternproperties.go`
- `internal/ecmaregex/translate.go`
- `internal/registry/sources.go`
- `internal/registry/budget_test.go`
- `cmd/jsonschema/limitflags.go`
- `cmd/jsonschema/limitflags_test.go`
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

- Files outside the reservation that the code requires: registry_options.go (where Registry and its caps actually live), internal/registry/sources.go (the real loader; registry.go does not need changes), internal/compile/refs.go (the real compile path builds CompileContext there, so compile.go is not edited), internal/eval/refguard.go and evaluated.go (where steps, path depth and the ref hop are counted), refguard_test.go (it asserts le.Max == eval.MaxEvalDepth, which changes to the depth that was set), validation/enum.go, validation/uniqueitems.go, validation/pattern.go and applicator/patternproperties.go. Dropped from the reservation as not needed: internal/limits/*, options.go, numberlimits_test.go, internal/compile/compile.go, internal/compile/limits*.go, internal/keywords/core/ref.go and dynamicref.go, cmd/jsonschema/main.go, internal/registry/registry.go.
- Meta-check of a schema nested near DefaultMaxDepth: the 2020-12 meta-schema makes about two $ref/$dynamicRef hops per schema level, so a schema like {items:{items:...}} close to 10,000 levels can reach the 20,000 ref-chain cap. It already meets the 40,000-node path backstop today. The coder must check that no existing test compiles such a schema. Ordinary schemas are far below this.
- The per-keyword Step changes step counts. Tests that search for a minimal cap (assertLookupCountsSteps) adapt automatically. Any test that uses a fixed small cap must keep passing; refguard_test passes its own max.
- Pure cycles now satisfy errors.As for LimitError through LoopError.Unwrap. The CLI hint '(raise with --max-ref-chain)' will not help for a true cycle; this is accepted to meet the acceptance criterion.
- Exceeded number caps no longer print line and column in the CLI line, because the mandated format is name and value only.
- Folding the depth into jsonvalue.NumberLimits makes the name inaccurate. Mark it with a ponytail comment; rename it to DecodeLimits when the file is next in scope.
- The step check runs on the hot path: one increment and one compare only.

Implementation constraints the coder must honor:

- Binding boundary rules: (1) check input length against the resolved cap before decoding (library Compile/Validate, CLI readCapped, registry LimitReader); (2) depth is checked by the iterative decoder before descending, and the eval-path backstop (4 x depth) is checked in enterNode before pushing; (3) count ref chains per path: increment on entering through $ref/$dynamicRef, decrement in the pop func; (4) super-linear keyword loops charge steps in proportion to their work (enum: entry size precomputed at compile; uniqueItems: one step per element), plus one step per keyword in eval.go's dispatch; (5) every Validate, meta-check run and Result.Output tree re-evaluation builds a fresh budget through EvaluateWith, and nothing is cached on Schema or Result; (6) LimitErrors are wrapped only with %w inside the engine and converted by asLimit into a top-level *jsonschema.LimitError at every public return.
- LimitError.Limit takes only the 11 exported constants; jsonvalue/limits.go holds the internal mirror. A test walks every internal producer (decode, registry, eval, keywords, output, ecmaregex) and checks that its name is in the set.
- Do not rename or remove any existing CLI flag, Options field, Registry field or exported constant. New flags use --max-<name>.
- Do not change EvaluateFull, EvaluateTree, EvaluateLimited or CompileReached signatures; they become default-limit wrappers.
- eval.go: add only the one-line Step per keyword in the dispatch. keywords/registry.go: add only the MaxPatternBytes field and the Charge method. jsonschema.go: resolve limits, run the length check, call asLimit and store eval.Limits; no other changes.
- Tests read every value from exported or package constants (DefaultMax*, jsonvalue.MaxDepth, maxInputBytes, validation.DefaultMaxPatternSize), never literals. Each limit has a library test in limits_test.go at, below and above the cap via a lowered Options value, plus a negative-value error test. Each limit flag has a contract test in limitflags_test.go (exit 2, empty stdout, one stderr line naming the flag; a negative or zero value is a usage error), including --max-output-units and a multi-document worst-result test.
- concurrency_race_test.go: 8+ goroutines run Validate and then Output for every shape on one Schema, mixing valid, invalid and limit-hitting documents (steps, ref chain, depth, output units); it must pass under go test -race. A second test compares Schema.lim, Validate results and Output bytes before and after 1,000 validations, comparing values, never func fields.
- Limit messages carry only the limit name and the value set; never an instance value or pattern text.
- Standard library only; go.mod unchanged.
- Doc comments state the defaults, and that very large depth values can exhaust the goroutine stack.

- DB impact: no · API impact: yes · Migration needed: no

Breaking changes:

- Public functions now return a top-level *jsonschema.LimitError for every exceeded limit, instead of a *jsonvalue.LimitError wrapped in 'schema:', 'document:' or 'registry:' text.
- Internal limit names change: 'schema file bytes' becomes 'input size', 'total registry bytes' becomes 'registry size', and 'evaluation depth' becomes 'nesting depth' carrying the depth that was set.
- An oversize pattern becomes a LimitError ('pattern size'), where it was a plain schema error with different wording.
- CLI stderr for every limit becomes 'limit exceeded: <limit> <value> (raise with --<flag>)', replacing '<limit> exceeds limit <n> ...'. parse_errors_test.go is updated.
- Registry.MaxFileBytes = 0 now follows Options.MaxInputBytes. The default value is unchanged (64 MiB).

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
