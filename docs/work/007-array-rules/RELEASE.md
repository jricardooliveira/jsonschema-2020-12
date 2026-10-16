# Release notes — #7: 2020-12 array keywords: prefixItems, items, contains, maxContains, minContains, maxItems, minItems, uniqueItems

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #7
- Status: proposed (pending Checkpoint 3 sign-off)

## What changed

The validator now checks arrays against 2020-12 positional rules (prefixItems, items), element counts (minItems, maxItems), specific elements (contains with minContains/maxContains bounds), and uniqueness (with deep JSON equality: 1 and 1.0 are duplicates, objects compare key-order-independently). All eight keywords apply only to array instances and pass silently on non-arrays. Item annotations (largest index validated, matched indices) are recorded in the evaluation context for later use by unevaluatedItems.

- internal/keywords/applicator: prefixItems (positional validation, records largest index or true), items (remaining elements, records true), contains (matching count, records indices or true)
- internal/keywords/validation: minContains, maxContains (read by sibling contains, compile-only), maxItems, minItems (count limits, exact number model: 2.0 behaves as 2), uniqueItems (deep-equality duplicates, fingerprint-grouped for near-linear performance)
- internal/eval: EvaluatedItems record (per-call storage of item annotations, Mark/Rollback for failed branches), attached once per evaluation context
- internal/keywords: CompileContext extended with Location, Siblings map, and Sub callback; EvalContext extended with Items field; EvalChild now rolls back child annotations on failure
- internal/compile: Full CompileContext built for each keyword with location and sibling access
- internal/jsonvalue: Number.Canonical() for fingerprinting (normalizes 1, 1.0, 1e0 to same string)
- testdata: 8 JSON files (prefixItems, items, contains, minContains, maxContains, maxItems, minItems, uniqueItems) with spec-derived positive and negative cases

## How to verify

- Run `go test ./...` to execute all keyword and integration tests; all pass under -race
- Compile and validate a schema with `{"prefixItems":[{"type":"string"},{"type":"integer"}],"items":{"type":"boolean"}}` against `["a",1,true,false]` (valid: first two elements match positional rules, rest match items)
- Validate `{"contains":{"type":"integer"},"minContains":2}` against `[1,"x",2,"y"]` (valid: 2 integers >= minContains 2)
- Validate `{"uniqueItems":true}` against `[1,1.0,{"a":1}]` (invalid: 1 and 1.0 are equal; error lists positions only, no values)
- Verify no instance values appear in error output: validate `{"maxItems":1}` against `["secret","secret2"]` and confirm message shows count and limit, not the strings
- Run 100,000-element array through uniqueItems to confirm no hangs (deterministic test checks comparison count stays near-linear, far below n²/100)

## Migration

No data migration needed. The keywords are new and do not affect existing schemas or compiled validators. Internal eval API extended by one field (Items) in EvalContext, but this is internal and backward-compatible (new keyword implementations pass it, existing ones never read it).

## Rollback

Remove the eight keyword implementation files (prefixitems.go, items.go, contains.go in applicator; mincontains.go, maxcontains.go, maxitems.go, minItems.go, uniqueitems.go in validation), the evaluated_items record (evaluated_items.go), their test files, testdata JSON files, and revert CompileContext and EvalContext changes in internal/keywords/registry.go. Revert the one-line attachItems() call in internal/eval/eval.go, the CompileContext build in internal/compile/compile.go, and the Canonical() method in internal/jsonvalue/number_canonical.go. No external dependencies changed.

## Known limitations

- unevaluatedItems (story 192) is not implemented; item annotations recorded here are read by that later story
- Annotation collection and output (story 191) is not implemented; item annotations are stored internally but not yet exposed in standard output formats
- Meta-schema validation of malformed keyword values, such as negative minItems, is not checked (story 189)
- No wall-clock performance assertions; only 100k-element uniqueItems test validates comparison count stays near-linear
- 2020-12 draft only; these keywords are not prepared for other JSON Schema drafts

## Concerns

- This release adds array validation, a significant new capability. While all keywords are spec-compliant and fully tested (including under -race), the interaction between per-element evaluation, annotation recording, Mark/Rollback on failure, and sibling keyword access (e.g., items reading prefixItems length at compile time) is complex. Verify that rollback correctly drops annotations from failed branches: tests confirm it works, but this is critical for correctness of later stories (191, 192).
- The uniqueItems duplicate detection uses a fingerprint-grouped bucket approach to avoid O(n²) comparisons. The fingerprint must normalize exact number values (1, 1.0, 1e0 to the same key) or duplicates fall into different buckets. Tests verify the 100k-element case is near-linear and groups are exact, but the Number.Canonical() string representation is new and complex. Code review should confirm it is deterministic and collision-free.
- prefixItems, items and contains evaluate every array element with no short-circuit. On arrays with millions of elements and a failing schema, all elements will be evaluated and their units accumulated until the output cap is hit (which happens later in the pipeline). This is intentional (per acceptance criteria) but means very large arrays may feel slow until story 193 adds evaluation-step limits. The current behavior is correct per the spec.
- Sibling keyword access: items reads prefixItems length, contains reads minContains and maxContains values, all from ctx.Siblings map at compile time. If the skeleton's CompileContext does not provide siblings to every keyword, this will fail silently (keywords will not find siblings and will behave as if they are absent). Tests cover the expected cases (items without prefixItems, contains without min/max) but integration tests should verify a schema with all three keywords together works correctly.
