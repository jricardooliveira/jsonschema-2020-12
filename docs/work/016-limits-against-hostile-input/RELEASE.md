# Release notes — #16: Limits against hostile input and proof of the load-once guarantee

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #16
- Status: proposed (pending Checkpoint 3 sign-off)

## What changed

Limits against hostile input: nesting depth, input size, $ref/$dynamicRef recursion, work steps and pattern size are now configurable through Options and CLI flags with generous defaults. Exceeding a limit returns a distinct LimitError with exit code 2. Evaluation state lives per-call, so a compiled Schema is immutable and safe for concurrent Validate and Output calls. All existing tests pass; no changes to validation semantics beyond limit checks.

- limits.go: LimitError type (name and value), 11 limit constants, 5 new Option fields (MaxDepth, MaxInputBytes, MaxRefChain, MaxSteps, MaxPatternSize), 5 new default constants
- Options and Registry: MaxDepth, MaxInputBytes, MaxRefChain, MaxSteps, MaxPatternSize (0 means default, negative is rejected); Registry.MaxFileBytes=0 now means DefaultMaxInputBytes
- internal/eval/budget.go: per-call Limits struct and EvaluateWith function; callState tracks reference jumps, path depth and step budget
- internal/limits/limits.go: closed set of limit name strings and Resolve helper
- Keywords: proportional step charging before work (uniqueItems, enum, const, contains, pattern, patternProperties, maxLength, minLength, required, dependentRequired, dependentSchemas); compile-time size storage for enum and const
- internal/ecmaregex/translate.go: CompileBounded function checks pattern size before translation; returns LimitError
- cmd/jsonschema/limitflags.go: 12 limit flags (--max-depth, --max-input-bytes, --max-ref-chain, --max-steps, --max-pattern-size, --max-output-units, --max-report-bytes, --max-number-digits, --max-number-exponent, --max-annotations, --max-ref-file-bytes, --max-ref-total-bytes) with validation and error formatting
- CLI: limit errors give exit 2, empty stdout and one stderr line 'limit exceeded: <limit> <value> (raise with --<flag>)'; multi-document run marks error documents and worst result sets exit code
- Concurrency: race detector test (concurrency_race_test.go) runs 16+ goroutines Validate+Output on one Schema with limit-hitting documents; immutability test compares behavior and limits before/after 1000 validations

## How to verify

- Compile with Options.MaxDepth=5, validate a document with 6-level nesting: expect LimitError with Limit=LimitDepth, Value=5
- Compile with Options.MaxInputBytes=10, validate a document >10 bytes: expect LimitError with Limit=LimitInputSize, Value=10
- Run `jsonschema validate --max-depth=3 schema.json doc.json` where doc has 4 levels: exit code 2, stderr contains 'limit exceeded: nesting depth 3 (raise with --max-depth)'
- Create a schema with {items:{$ref:"#"}}, validate a 2000-level array with Options.MaxRefChain=DefaultMaxRefChain: passes (within default 20,000 jumps); with Options.MaxRefChain=50: LimitError with Limit=LimitRefChain
- Compile schema with {items:true}, validate 10000 elements with Options.MaxSteps=5000: LimitError (steps charged per keyword and per byte), passes at default (100,000,000)
- Pattern >DefaultMaxPatternSize bytes: LimitError, not schema error; raise with Options.MaxPatternSize
- Run concurrency_race_test under go test -race: no race detector warnings
- Call Result.Output multiple times on same Result: same output bytes (cached tree is rebuilt per call with fresh budget)

## Migration

Limit error handling: errors.As(err, &jsonschema.LimitError) now catches all exceeded limits (decode, registry, compile, eval, output). Limit names changed from internal wording (e.g. 'schema file bytes') to public names ('input size'); update any test that checks error messages. Registry.MaxFileBytes=0 no longer means unlimited; it defaults to Options.MaxInputBytes (64 MiB). Step counts are higher because steps are now proportional to work (per keyword + per byte); custom MaxSteps values may hit sooner. Pattern size over limit is now LimitError (exit 2) not schema error. No data migration needed; all limits are stateless.

## Rollback

Revert Options fields MaxDepth, MaxInputBytes, MaxRefChain, MaxSteps, MaxPatternSize to zero (or remove); revert LimitError type and limit constants; restore old limit error messages in decode, registry, eval, output and pattern; remove concurrency_race_test.go; remove internal/limits package; restore Registry.MaxFileBytes=0 to mean no per-file limit; restore step counting to per-node entry (one per EvaluatedItems.Step call). All changes are additive to the API (no removed public names); rollback removes new fields and types.

## Known limitations

- Limits cannot be disabled: 0 or less is rejected (by design, no setting switches a limit off)
- Very large MaxDepth (above stdlib json.Decoder limit of 10,000) is bounded by that decoder; setting MaxDepth > 10,000 still fails at 10,000 with LimitError
- Very large MaxDepth could exhaust goroutine stack in deeply recursive eval passes (documented in doc comments)
- $ref/$dynamicRef cycles that never terminate give LoopError matching LimitError via Unwrap; CLI hint 'raise with --max-ref-chain' does not help for true cycles (by design)
- Proportional charging makes ordinary documents ~1M steps per 64 MiB per pass through keywords; a document checked by multiple keywords at all levels can use significant portion of 100M default budget

## Concerns

- internal/limits package is separate from jsonvalue.LimitError (design decision A said reuse that error; actual code uses it and stores constants in internal/limits for closed-set validation). Minimal departure; error type is still shared.
- EvalContext.Charge() does nothing without a per-call record (only keywords with context that has Items as callState pay the cost). Cannot be reached in practice because evaluate() installs callState before any keyword runs; safe but opaque.
- Breaking change: limit error messages, Registry.MaxFileBytes=0 behavior, pattern size error type. Documented in migration_notes; all existing tests updated by story 193.
