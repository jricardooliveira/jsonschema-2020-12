# ADR #16: Limits against hostile input and proof of the load-once guarantee

- Date: 2026-10-17
- Story: #16
- Status: approved by the operator at Checkpoint 2 (run #152, 2026-10-17 01:51 UTC)

## Decision
This revises the previous design. Only the work-step charging (H), the tests the reviewer asked for and the file list change; everything else is as reviewed. I read the code first and adjusted to what is already there. Most limits already exist: number caps, registry per-file and total caps (Registry.MaxFileBytes/MaxTotalBytes, flags --max-ref-file-bytes/--max-ref-total-bytes), output units, annotations, report bytes, evaluation steps (eval.DefaultMaxEvalSteps, counted per node entry in refguard.go enterNode) and an evaluation-path backstop (eval.MaxEvalDepth). All of them already return *jsonvalue.LimitError. So the story unifies and exposes these limits rather than building new ones. Decisions:
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
(C) Options (the struct lives in numberlimits.go, so options.go is not touched) gains MaxDepth, MaxInputBytes, MaxRefChain, MaxSteps and MaxPatternSize. Existing fields keep their names. Registry.MaxFileBytes still overrides. When it is 0, the per-file registry cap is the resolved MaxInputBytes, so 'each input' shares one limit.
(D) Depth: jsonvalue.NumberLimits gains a MaxDepth field (0 means jsonvalue.MaxDepth), and decode.go reads it instead of the constant. The same struct already reaches schema, registry and document decoding, so one setting covers schemas and documents. Compiling recurses only over decoded values, so the same check bounds it. The eval-path backstop becomes 4 x the resolved MaxDepth, per call.
(E) Bounded reads: Compile and Validate reject len(input) > MaxInputBytes before decoding. Registry reads are already LimitReader-bounded. The CLI's readCapped takes the resolved cap.
(F) Per-call budget: internal/eval/budget.go defines eval.Limits{MaxUnits, MaxAnnotations, MaxSteps, MaxRefChain, MaxDepth} and EvaluateWith(n, inst, lim, tree bool). Each call builds a fresh callState (EvaluatedItems with maxSteps, a ref counter, a path cap). EvaluateFull, EvaluateTree and EvaluateLimited keep their signatures as default-limit wrappers, so existing tests do not change. Schema and Result store one immutable eval.Limits value. Validate, the meta-check and every Result.Output tree re-evaluation call EvaluateWith, so each gets a fresh budget.
(G) Reference chain: in refguard.go enterNode, when the parent's keyword is $ref or $dynamicRef, increment the call's ref count and decrement it in the pop func. This counts jumps on the current path. Past MaxRefChain it returns LimitRefChain. The existing exact-cycle detection stays. LoopError gains an Unwrap() returning a LimitRefChain LimitError, so errors.As matches a never-ending cycle as a limit error and the existing LoopError tests keep passing. ref.go and dynamicref.go are not touched.
(H) REVISED. Work steps are charged in proportion to the work done.
- Base charges: eval.go adds one Step per keyword evaluated in its dispatch (one line). Node entries and the existing object-member steps stay.
- Seam: evaluated.go gains EvaluatedItems.Charge(n int) error, which adds n to the same counter and returns the same 'evaluation steps' LimitError, and Step() becomes Charge(1). keywords/registry.go gains EvalContext.Charge(n) (it forwards to ctx.Items through a small interface, like required.go's step helper does today), a constant BytesPerStep = 64, and ChargeBytes(n), which charges 1 + n/BytesPerStep.
- Why 64: a 64 MiB string costs about 1M steps per pass. An ordinary large document checked by several keywords stays far below the 100M default, but repeating a pass over a huge string (many patterns, nested re-checking) still reaches the cap.
- Per keyword, each a few lines in its own file:
  - uniqueItems keeps its sub-quadratic algorithm. findDuplicates buckets elements by an unambiguous canonical fingerprint, so equal fingerprints mean equal values and there is about one deep-equality comparison per element. A doc comment records this as the documented algorithm. It charges ChargeBytes(len(fingerprint)) per element, which covers building the fingerprint, and ChargeBytes(len(fingerprint)) again per Equal comparison inside a bucket. Total work is about O(array bytes), and every byte is charged. Nested arrays that each carry uniqueItems are charged at every level.
  - enum: each candidate's canonical size is precomputed at compile time and stored immutably. ChargeBytes(size) is charged before each Equal. Equal stops at the first difference, so its cost is bounded by the candidate's size.
  - const: the same, with a single precomputed size.
  - contains: Charge(1) per element before EvalChild, on top of the node-entry step.
  - patternProperties keeps its per-key and per-(key, pattern) steps. Each step becomes ChargeBytes(len(key)), because each RE2 match is linear in the key.
  - pattern: ChargeBytes(len(s)) before MatchString.
  - maxLength and minLength: ChargeBytes(len(s)) before counting code points. maxlength.go gains one small runeCount(ctx, s) helper that charges and counts, and minlength.go calls it, so there is no copy.
- The ECMA-262 translation and RE2 compile run at compile time and are bounded by the pattern-size limit (I).
(I) Pattern size: one shared helper in ecmaregex/translate.go, CompileBounded(src, max), with 0 meaning the default. It returns LimitPatternSize before translating. pattern.go and patternproperties.go call it with CompileContext.MaxPatternBytes, which removes the copy noted in patternproperties.go. refs.go sets that field from a new CompileReachedWith(..., maxPattern), and CompileReached stays as a wrapper. compile.go is not edited.
(J) CLI: cmd/jsonschema/limitflags.go owns every limit flag.
- Existing flags keep their exact names: --max-number-digits, --max-number-exponent, --max-annotations, --max-ref-file-bytes, --max-ref-total-bytes.
- New flags: --max-depth, --max-input-bytes, --max-ref-chain, --max-steps, --max-pattern-size, --max-output-units, --max-report-bytes.
- limitflags.go also owns the existing explicit-value>0 Visit check and its message, the Options/Registry fill, and limitLine(err, compiling). limitLine maps each Limit to its flag. An input-size error during compile can only come from a registry file, so it maps to --max-ref-file-bytes. The line is le.Error()+' (raise with --<flag>)'.
- run.go replaces its inline flag lines and Visit block with calls into limitflags.go. It maps errors through limitLine where it calls fail and where it builds the multi-document {document, error} line, and it drops the 'cannot render result' masking for LimitErrors. main.go is not touched.
- maxInputBytes moves to limitflags.go as an alias of jsonschema.DefaultMaxInputBytes, so run_test.go still compiles.
All the reviewer's items are in scope: pattern-size option and flag, output-units flag with its exit-2 test, steps option and flag charged in proportion to the work, report-bytes flag and name, and the concurrent Validate+Output race test. The design touches 32 files, the factory's limit. To make room for the six keyword files: the decoder-depth and registry-total tests move into the root limits_test.go (through Options.MaxDepth and Registry.MaxTotalBytes), and internal/jsonvalue/depth_option_test.go and internal/registry/budget_test.go are dropped.

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
- internal/keywords/validation/const.go
- internal/keywords/validation/uniqueitems.go
- internal/keywords/validation/pattern.go
- internal/keywords/validation/maxlength.go
- internal/keywords/validation/minlength.go
- internal/keywords/applicator/patternproperties.go
- internal/keywords/applicator/contains.go
- internal/ecmaregex/translate.go
- internal/registry/sources.go
- cmd/jsonschema/limitflags.go
- cmd/jsonschema/limitflags_test.go
- cmd/jsonschema/run.go
- cmd/jsonschema/parse_errors_test.go

## Data / API impact
- DB impact: no
- API impact: yes
- Migration needed: no

## Constraints for implementation
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

## Breaking changes
- Public functions now return a top-level *jsonschema.LimitError for every exceeded limit, instead of a *jsonvalue.LimitError wrapped in 'schema:', 'document:' or 'registry:' text.
- Internal limit names change: 'schema file bytes' becomes 'input size', 'total registry bytes' becomes 'registry size', and 'evaluation depth' becomes 'nesting depth', which carries the depth that was set.
- An oversize pattern becomes a LimitError ('pattern size'), where it was a plain schema error with different wording.
- CLI stderr for every limit becomes 'limit exceeded: <limit> <value> (raise with --<flag>)', replacing '<limit> exceeds limit <n> ...'. parse_errors_test.go is updated.
- Registry.MaxFileBytes = 0 now follows Options.MaxInputBytes. The default value is unchanged (64 MiB).
- Evaluation step counts rise, because steps are now charged per keyword and in proportion to the bytes compared or matched. A caller who set a low custom MaxSteps may hit it sooner.

## Sensitivity
- security
