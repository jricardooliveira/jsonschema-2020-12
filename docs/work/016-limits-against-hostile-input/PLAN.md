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
- `internal/keywords/validation/const.go`
- `internal/keywords/validation/uniqueitems.go`
- `internal/keywords/validation/pattern.go`
- `internal/keywords/validation/maxlength.go`
- `internal/keywords/validation/minlength.go`
- `internal/keywords/applicator/patternproperties.go`
- `internal/keywords/applicator/contains.go`
- `internal/ecmaregex/translate.go`
- `internal/registry/sources.go`
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

- Files outside the reservation that the code requires:
- registry_options.go: where Registry and its caps actually live.
- internal/registry/sources.go: the real loader; registry.go does not need changes.
- internal/compile/refs.go: the real compile path builds CompileContext there, so compile.go is not edited.
- internal/eval/refguard.go and evaluated.go: where steps, path depth and the ref hop are counted, and where Charge goes.
- refguard_test.go: it asserts le.Max == eval.MaxEvalDepth, which changes to the depth that was set.
- validation/enum.go, const.go, uniqueitems.go, pattern.go, maxlength.go and minlength.go, plus applicator/patternproperties.go and contains.go: the proportional charges the boundary review requires.
Dropped from the reservation as not needed: internal/limits/*, options.go, numberlimits_test.go, internal/compile/compile.go, internal/compile/limits*.go, internal/keywords/core/ref.go and dynamicref.go, cmd/jsonschema/main.go and internal/registry/registry.go.
- At the 32-file ceiling. To make room for the keyword files, the decoder-depth and registry-total tests live in the root limits_test.go instead of tests beside internal/jsonvalue and internal/registry. They test through the public Options/Registry, which also meets 'every limit has its own library test'. If one more file turns out to be needed, split the story: (a) model, limits and charging with library tests; (b) CLI flags and race proof.
- Byte-scaled charging makes step counts grow with string and value sizes. With BytesPerStep=64, a 64 MiB document costs about 1M steps per byte-charged pass, far below the 100M default. The coder must run the full suite, including any assertLookupCountsSteps-style tests that search for a minimal cap, and confirm that no existing test with a fixed small MaxSteps starts failing. refguard_test passes its own max.
- The uniqueItems bound depends on the fingerprint being unambiguous, so that equal fingerprints mean equal values and buckets hold one class. If a future change makes the fingerprint lossy, comparisons per bucket could grow. They are still charged per comparison, so the step cap holds, but it could trip on legitimate input. The findDuplicates doc comment records this invariant.
- Meta-check of a schema nested near DefaultMaxDepth: the 2020-12 meta-schema makes about two $ref/$dynamicRef hops per schema level, so a schema like {items:{items:...}} close to 10,000 levels can reach the 20,000 ref-chain cap. It already reaches the 40,000-node path backstop today. That is a limit error, never a silent result. The coder must confirm that no existing test or spec case compiles such a schema. Ordinary schemas are far below this.
- Pure cycles now satisfy errors.As for LimitError through LoopError.Unwrap. For a true cycle the CLI hint '(raise with --max-ref-chain)' will not help; this is accepted to meet the acceptance criterion.
- Exceeded number caps no longer print line and column in the CLI line, because the required format is limit name and value only.
- Folding the depth into jsonvalue.NumberLimits makes the name inaccurate. Mark it with a ponytail comment and rename it to DecodeLimits when the file is next in scope.
- The step check runs on the hot path: one add and one compare per charge only.

Implementation constraints the coder must honor:

- Binding boundary rules: (1) check input length against the resolved cap before decoding (library Compile/Validate, CLI readCapped, registry LimitReader); (2) the iterative decoder checks depth before descending, and enterNode checks the eval-path backstop (4 x depth) before pushing; (3) count ref chains per path: increment on entering through $ref/$dynamicRef, decrement in the pop func; (4) every keyword loop that can be super-linear charges steps in proportion to its work, and charges before doing that work (see the next constraint), plus one step per keyword in eval.go's dispatch; (5) every Validate, meta-check run and Result.Output tree re-evaluation builds a fresh budget through EvaluateWith, and nothing is cached on Schema or Result; (6) LimitErrors are wrapped only with %w inside the engine, and asLimit converts them into a top-level *jsonschema.LimitError at every public return.
- Proportional charging, all through keywords.EvalContext.Charge/ChargeBytes (ChargeBytes(n) = Charge(1 + n/keywords.BytesPerStep)):
- uniqueItems: ChargeBytes(len(fingerprint)) per element, and again per jsonvalue.Equal call. Keep the fingerprint-bucket algorithm and document in findDuplicates that it does about one comparison per element. Never fall back to pairwise comparison.
- enum: ChargeBytes(precomputed candidate size) before each Equal.
- const: ChargeBytes(precomputed size) before Equal.
- contains: Charge(1) per element.
- patternProperties: ChargeBytes(len(key)) per key and per (key, pattern) match, replacing the plain step calls.
- pattern: ChargeBytes(len(s)) before MatchString.
- maxLength/minLength: ChargeBytes(len(s)) through one shared runeCount(ctx, s) helper in maxlength.go, which minlength.go calls.
Compute candidate sizes at compile time with the existing canonical encoding (reuse uniqueitems.go's fingerprint; do not write a second encoder) and store them in the immutable compiled value.
- LimitError.Limit takes only the 11 exported constants; jsonvalue/limits.go holds the internal mirror. A test walks every internal producer (decode, registry, eval, keywords, output, ecmaregex) and checks that its name is in the set.
- Do not rename or remove any existing CLI flag, Options field, Registry field or exported constant. New flags use --max-<name>.
- Do not change the signatures of EvaluateFull, EvaluateTree, EvaluateLimited or CompileReached; they become default-limit wrappers. EvaluatedItems.Step keeps its signature and calls Charge(1).
- eval.go: add only the one-line Step per keyword in the dispatch. keywords/registry.go: add only the MaxPatternBytes field, the Charge and ChargeBytes methods and the BytesPerStep constant. jsonschema.go: resolve limits, run the length check, call asLimit and store eval.Limits; no other changes. Keyword files: add only the charge calls (and, for enum/const, the compile-time size). Validation results must not change.
- Tests read every value from exported or package constants (DefaultMax*, jsonvalue.MaxDepth, maxInputBytes, validation.DefaultMaxPatternSize, keywords.BytesPerStep), never literals.
- limits_test.go has a library test for each limit at, below and above the cap, using a lowered Options value, plus a negative-value error test. This includes nesting depth on both a schema and a document (moved from the dropped depth_option_test.go) and the registry total through Registry.MaxTotalBytes (moved from the dropped registry/budget_test.go).
- limitflags_test.go has a contract test for each limit flag: exit 2, empty stdout, one stderr line naming the flag; a negative or zero value is a usage error. It includes --max-output-units and a multi-document worst-result test.
- Hostile work tests, in internal/eval/budget_test.go. Each compiles a schema and runs EvaluateWith with a lowered MaxSteps. Each must end in an 'evaluation steps' LimitError, and runs under a generous time bound checked with testing deadlines or a timer.
(a) uniqueItems: an array of N distinct elements, each about S bytes. MaxSteps is set above what one step per element would charge (N plus base steps) but below N*S/BytesPerStep, so the test fails if charging goes back to per element.
(b) patternProperties x properties: an object with K long keys (each L bytes) and P patterns. MaxSteps is above the plain count of pairs (K*P plus base steps) but below K*P*L/BytesPerStep.
Also add a control: the same documents pass at the default MaxSteps, and a large ordinary document (many elements checked by several keywords) is not stopped.
- concurrency_race_test.go: 8+ goroutines run Validate and then Output for every shape on one Schema, mixing valid, invalid and limit-hitting documents (steps, ref chain, depth, output units). It must pass under go test -race. A second test compares Schema.lim, Validate results and Output bytes before and after 1,000 validations, comparing values, never func fields.
- Limit messages carry only the limit name and the value that was set; never an instance value or pattern text.
- Standard library only; go.mod unchanged.
- Doc comments state the defaults, that BytesPerStep scales the charge for strings and values, and that very large depth values can exhaust the goroutine stack.

- DB impact: no · API impact: yes · Migration needed: no

Breaking changes:

- Public functions now return a top-level *jsonschema.LimitError for every exceeded limit, instead of a *jsonvalue.LimitError wrapped in 'schema:', 'document:' or 'registry:' text.
- Internal limit names change: 'schema file bytes' becomes 'input size', 'total registry bytes' becomes 'registry size', and 'evaluation depth' becomes 'nesting depth', which carries the depth that was set.
- An oversize pattern becomes a LimitError ('pattern size'), where it was a plain schema error with different wording.
- CLI stderr for every limit becomes 'limit exceeded: <limit> <value> (raise with --<flag>)', replacing '<limit> exceeds limit <n> ...'. parse_errors_test.go is updated.
- Registry.MaxFileBytes = 0 now follows Options.MaxInputBytes. The default value is unchanged (64 MiB).
- Evaluation step counts rise, because steps are now charged per keyword and in proportion to the bytes compared or matched. A caller who set a low custom MaxSteps may hit it sooner.

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
