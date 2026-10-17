# ADR #16: Limits against hostile input and proof of the load-once guarantee

- Date: 2026-10-17
- Story: #16
- Status: proposed (pending architecture sign-off)

## Decision
This revises the previous design. I read the code first and adjusted to what is already there. Most limits already exist: number caps, registry per-file and total caps (Registry.MaxFileBytes/MaxTotalBytes, flags --max-ref-file-bytes/--max-ref-total-bytes), output units, annotations, report bytes, evaluation steps (eval.DefaultMaxEvalSteps, counted per node entry in refguard.go enterNode) and an evaluation-path backstop (eval.MaxEvalDepth). All of them already return *jsonvalue.LimitError. So the story unifies and exposes these limits rather than building new ones. Decisions:
(A) Shared internal error: reuse jsonvalue.LimitError{Limit, Max, ...}. There is no internal/limits package, because a second error type would duplicate it and break about 10 test files. internal/jsonvalue/limits.go gains the closed set of name constants. Root limits.go defines `type Limit string`, exported constants mirroring them, `type LimitError struct{Limit Limit; Value int}`, whose Error() is 'limit exceeded: <limit> <value>', the new Default* constants, a generic resolve(name, v, def) helper (0 means default, negative is an error 'options: <Field> must not be negative'), and asLimit(err). asLimit returns a top-level *jsonschema.LimitError (not wrapped) whenever errors.As finds a *jsonvalue.LimitError. It is called on the error return of compileWith, Schema.Validate and Result.Output, so LimitErrors from decode, registry, compile, meta-check and eval reach the caller unwrapped.
(B) The closed set is exactly these. Existing strings are kept where they exist, and the operator's own example message uses 'nesting depth':
- LimitDepth="nesting depth"
- LimitInputSize="input size" (was "schema file bytes" in the registry)
- LimitRegistrySize="registry size" (was "total registry bytes")
- LimitRefChain="reference chain"
- LimitSteps="evaluation steps"
- LimitNumberDigits="number digits"
- LimitNumberExponent="number exponent"
- LimitPatternSize="pattern size"
- LimitOutputUnits="output units"
- LimitAnnotations="annotations"
- LimitReportBytes="report bytes"
The old "evaluation depth" backstop now reports LimitDepth with the depth that was set.
(C) Options (the struct lives in numberlimits.go, so options.go is not touched) gains MaxDepth, MaxInputBytes, MaxRefChain, MaxSteps and MaxPatternSize. Existing fields keep their names. Registry.MaxFileBytes still overrides; when it is 0, the per-file registry cap is the resolved MaxInputBytes, so 'each input' shares one limit.
(D) Depth: jsonvalue.NumberLimits gains a MaxDepth field (0 means jsonvalue.MaxDepth). decode.go reads it instead of the constant. The same struct already reaches schema, registry and document decoding, so one setting covers schemas and documents. Compiling recurses only over decoded values, so it is bounded by the same check. The eval-path backstop becomes 4 x resolved MaxDepth, per call.
(E) Bounded reads: Compile and Validate reject len(input) > MaxInputBytes before decoding. Registry reads are already LimitReader-bounded. The CLI's readCapped takes the resolved cap.
(F) Per-call budget: internal/eval/budget.go defines eval.Limits{MaxUnits, MaxAnnotations, MaxSteps, MaxRefChain, MaxDepth} and EvaluateWith(n, inst, lim, tree bool). Each call builds a fresh callState (EvaluatedItems with maxSteps, a ref counter, a path cap). EvaluateFull, EvaluateTree and EvaluateLimited keep their signatures as default-limit wrappers, so existing tests do not change. Schema and Result store one immutable eval.Limits value. Validate, the meta-check and every Result.Output tree re-evaluation call EvaluateWith, so each gets a fresh budget.
(G) Reference chain: in refguard.go enterNode, when the parent's keyword is $ref or $dynamicRef, increment the call's ref count and decrement it in the pop func. This counts jumps on the current path. Past MaxRefChain it returns LimitRefChain. The existing exact-cycle detection stays. LoopError gains an Unwrap() returning a LimitRefChain LimitError, so a never-ending cycle is matched by errors.As as a limit error, and the existing LoopError tests keep passing. ref.go and dynamicref.go are not touched.
(H) Work steps, charged by all keywords: eval.go adds one Step per keyword evaluated in its dispatch (one line). A new EvalContext.Charge(n) in keywords/registry.go is for loops that can be super-linear: enum charges each candidate's size, precomputed at compile time and immutable, and uniqueItems charges one step per element. Node entries and the existing per-member object steps stay.
(I) Pattern size: one shared helper in ecmaregex/translate.go, CompileBounded(src, max), with 0 meaning the default. It returns LimitPatternSize before translating. pattern.go and patternproperties.go call it with CompileContext.MaxPatternBytes, which removes the copy noted in patternproperties.go. refs.go sets that field from a new CompileReachedWith(..., maxPattern). CompileReached stays as a wrapper. compile.go is not edited.
(J) CLI: cmd/jsonschema/limitflags.go owns every limit flag. The existing ones keep their exact names (--max-number-digits, --max-number-exponent, --max-annotations, --max-ref-file-bytes, --max-ref-total-bytes). New ones: --max-depth, --max-input-bytes, --max-ref-chain, --max-steps, --max-pattern-size, --max-output-units, --max-report-bytes. It also owns the existing explicit-value>0 Visit check and its message, the Options/Registry fill, and limitLine(err, compiling). limitLine maps Limit to a flag; an input-size error during compile can only come from a registry file, so it maps to --max-ref-file-bytes. The line is le.Error()+' (raise with --<flag>)'. run.go replaces its inline flag lines and Visit block with calls into limitflags.go, and maps errors through limitLine where it calls fail and where it builds the multi-document {document, error} line. It also drops the 'cannot render result' masking for LimitErrors. main.go is not touched. maxInputBytes moves to limitflags.go as an alias of jsonschema.DefaultMaxInputBytes, so run_test.go still compiles.
All items from the reviewer are in scope: pattern-size option and flag, output-units flag with its exit-2 test, steps option and flag charged by all keywords, report-bytes flag and name, and the concurrent Validate+Output race test. The design touches 30 files, within the 32 limit.

## Affected modules
- limits.go
- limits_test.go
- numberlimits.go
- jsonschema.go
- output.go
- registry_options.go
- concurrency_race_test.go
- internal/jsonvalue/limits.go
- internal/jsonvalue/decode.go
- internal/jsonvalue/depth_option_test.go
- internal/compile/refs.go
- internal/eval/budget.go
- internal/eval/budget_test.go
- internal/eval/annotations.go
- internal/eval/refguard.go
- internal/eval/refguard_test.go
- internal/eval/evaluated.go
- internal/eval/eval.go
- internal/keywords/registry.go
- internal/keywords/validation/enum.go
- internal/keywords/validation/uniqueitems.go
- internal/keywords/validation/pattern.go
- internal/keywords/applicator/patternproperties.go
- internal/ecmaregex/translate.go
- internal/registry/sources.go
- internal/registry/budget_test.go
- cmd/jsonschema/limitflags.go
- cmd/jsonschema/limitflags_test.go
- cmd/jsonschema/run.go
- cmd/jsonschema/parse_errors_test.go

## Data / API impact
- DB impact: no
- API impact: yes
- Migration needed: no

## Constraints for implementation
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

## Risks
- Files outside the reservation that the code requires: registry_options.go (where Registry and its caps actually live), internal/registry/sources.go (the real loader; registry.go does not need changes), internal/compile/refs.go (the real compile path builds CompileContext there, so compile.go is not edited), internal/eval/refguard.go and evaluated.go (where steps, path depth and the ref hop are counted), refguard_test.go (it asserts le.Max == eval.MaxEvalDepth, which changes to the depth that was set), validation/enum.go, validation/uniqueitems.go, validation/pattern.go and applicator/patternproperties.go. Dropped from the reservation as not needed: internal/limits/*, options.go, numberlimits_test.go, internal/compile/compile.go, internal/compile/limits*.go, internal/keywords/core/ref.go and dynamicref.go, cmd/jsonschema/main.go, internal/registry/registry.go.
- Meta-check of a schema nested near DefaultMaxDepth: the 2020-12 meta-schema makes about two $ref/$dynamicRef hops per schema level, so a schema like {items:{items:...}} close to 10,000 levels can reach the 20,000 ref-chain cap. It already meets the 40,000-node path backstop today. The coder must check that no existing test compiles such a schema. Ordinary schemas are far below this.
- The per-keyword Step changes step counts. Tests that search for a minimal cap (assertLookupCountsSteps) adapt automatically. Any test that uses a fixed small cap must keep passing; refguard_test passes its own max.
- Pure cycles now satisfy errors.As for LimitError through LoopError.Unwrap. The CLI hint '(raise with --max-ref-chain)' will not help for a true cycle; this is accepted to meet the acceptance criterion.
- Exceeded number caps no longer print line and column in the CLI line, because the mandated format is name and value only.
- Folding the depth into jsonvalue.NumberLimits makes the name inaccurate. Mark it with a ponytail comment; rename it to DecodeLimits when the file is next in scope.
- The step check runs on the hot path: one increment and one compare only.

## Breaking changes
- Public functions now return a top-level *jsonschema.LimitError for every exceeded limit, instead of a *jsonvalue.LimitError wrapped in 'schema:', 'document:' or 'registry:' text.
- Internal limit names change: 'schema file bytes' becomes 'input size', 'total registry bytes' becomes 'registry size', and 'evaluation depth' becomes 'nesting depth' carrying the depth that was set.
- An oversize pattern becomes a LimitError ('pattern size'), where it was a plain schema error with different wording.
- CLI stderr for every limit becomes 'limit exceeded: <limit> <value> (raise with --<flag>)', replacing '<limit> exceeds limit <n> ...'. parse_errors_test.go is updated.
- Registry.MaxFileBytes = 0 now follows Options.MaxInputBytes. The default value is unchanged (64 MiB).

## Sensitivity
- security
