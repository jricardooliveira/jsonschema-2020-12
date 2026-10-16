# Release notes — #8: Applicator keywords allOf, anyOf, oneOf, not and if/then/else, with keyword locations rooted at the schema root

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #8
- Status: proposed (pending Checkpoint 3 sign-off)

**⚠️ the release-agent has concerns.**

## What changed

Adds the seven mandatory 2020-12 applicator keywords (allOf, anyOf, oneOf, not, if/then/else) so schemas can combine subschemas and use conditional logic. Also roots keyword locations at the schema root (/allOf/0/const instead of /const) and updates story 180's test assertion. Each keyword registers from its own file; reports follow operator decisions (nested failures for allOf/anyOf/oneOf/then/else, matched indices for multi-match oneOf, single entry for not, silence for if). All anyOf/oneOf choices are always evaluated.

- internal/keywords/registry.go: CompileContext gains Location (current keyword location as JSON Pointer from root) and CompileAt (compile sibling schemas at absolute pointers, for if/then/else)
- internal/compile/compile.go: Added CompileAt closure to the CompileContext passed to keywords, enabling child compilation at rooted locations
- internal/compile/compile_location_test.go: New test verifying Sub() compiles children at rooted locations (/x-loc-sub/0/x-loc-sub/0)
- internal/eval/eval_subschema_test.go: New test verifying child failures report rooted keyword locations (/x-test-keep/0/const)
- internal/eval/units_test.go: Modified x-test-keep helper to use cc.Sub() instead of compile.Compile, so child units are rooted
- internal/output/units_test.go: New test verifying basic output contains no instance values and child units are rooted (/x-out-keep/0/const)
- internal/keywords/applicator/subschemas.go: New shared helper compiling a non-empty array of schemas at /<keyword>/<i>; rejects empty arrays and non-arrays as schema errors
- internal/keywords/applicator/allof.go, allof_test.go: allOf — valid when every subschema is valid; on failure reports each failing subschema's units then an allOf unit
- internal/keywords/applicator/anyof.go, anyof_test.go: anyOf — valid when any subschema is valid; always evaluates every choice; on failure reports each choice's units then an anyOf unit
- internal/keywords/applicator/oneof.go, oneof_test.go: oneOf — valid when exactly one subschema is valid; always evaluates every choice; when 2+ match, reports 'oneOf: choices 0 and 2 both matched' (with indices)
- internal/keywords/applicator/not.go, not_test.go: not — valid when its subschema is invalid; on failure reports exactly one unit 'not: document matched a schema it must not match' with no inner details
- internal/keywords/applicator/if.go, if_test.go: if — evaluates its subschema silently (units discarded, still counted against caps); then applies if passed, else if failed; alone, never fails the document
- internal/keywords/applicator/then.go, then_test.go: then — no evaluation of its own; 'if' keyword compiles and runs it when if passes; ignored without if
- internal/keywords/applicator/else.go, else_test.go: else — no evaluation of its own; 'if' keyword compiles and runs it when if fails; ignored without if
- testdata/applicator/allOf.json, anyOf.json, oneOf.json, not.json, if-then-else.json, boolean-subschemas.json, nested.json: Spec-derived JSON test cases citing 2020-12 core sections 10.2.1 and 10.2.2, including nested applicators and oneOf with 2+ matches
- spec_applicator_test.go: Runs testdata/applicator cases through the shared spec-cases runner, enforcing that each case cites a spec section

## How to verify

- Compile schema {"allOf":[{"const":1}]} and validate against 2; the error keywordLocation should be /allOf/0/const (rooted), not /const
- Compile schema {"anyOf":[true,{"type":"integer"}]} and validate against "x" with output cap 0; if the second choice is evaluated, it hits the cap and returns a limit error (proving no short-circuit)
- Compile schema {"oneOf":[{"type":"integer"},{"type":"number"}]} and validate against 5; error should be 'oneOf: choices 0 and 1 both matched' with no instance value
- Compile schema {"if":{"type":"string"},"then":{"minLength":10}} and validate against "abc"; error should report only the then failure at /then/minLength and /then, with no if entry
- Compile schema {"not":{"const":1}} and validate against 1; should report exactly one unit 'not: document matched a schema it must not match' at /not
- Compile schema {"allOf":[{"type":"integer"},{"type":"string"}]} and validate against true; should list both failures /allOf/0/type and /allOf/1/type, then /allOf
- Run go test ./... to verify all spec-derived applicator cases pass, including nested (e.g. allOf inside anyOf inside not) and multi-match oneOf
- Run go test -v ./internal/keywords/applicator... to verify 'every choice evaluated' tests catch short-circuit bugs via unit-count limits

## Migration

None. This is internal API only: keywords.CompileContext gets Location and CompileAt fields. The public jsonschema package API is unchanged. Compiled schemas remain immutable after Compile. Approved (not yet released) stories 181-184 that also modify registry.go and compile.go must use the new rooted-location seam when adding their keywords.

## Rollback

Remove internal/keywords/applicator/allof.go, anyof.go, oneof.go, not.go, if.go, then.go, else.go and subschemas.go. Revert internal/keywords/registry.go (remove CompileAt field from CompileContext) and internal/compile/compile.go (remove CompileAt closure). Revert the rooted-location test changes. Schemas using applicator keywords will then fail at compile time with 'unknown keyword' errors; users must use an older release or skip validation of applicators.

## Known limitations

- Race-test coverage unproven: code is immutable per design (per-call eval state, no mutation of compiled nodes) but no concurrent test with 'go test -race' exercises the applicator keywords, so thread-safety is asserted by code review only
- Spec cases may run twice if spec_test.go glob already covers testdata/applicator/; operator accepted this low-impact redundancy to avoid editing spec_test.go
- Empty allOf/anyOf/oneOf arrays are schema errors (per 2020-12 'must be non-empty array'), not validation failures; validates at compile time, not at validation time
- if/then/else branches use context.Siblings to find sibling keywords; if the schema parser ever stops exposing siblings, this will break

## Concerns

- F1 (medium, from tester): No concurrent test exercises these keywords with 'go test -race'. The code reads as immutable, but thread-safety is unproven. Recommendation: Before shipping to production, add a test that calls Schema.Validate on the same compiled schema from multiple goroutines with applicator keywords.
- F2 (low, from tester): spec_applicator_test.go added as a new file, but if spec_test.go's glob pattern already covers testdata/applicator/, the same cases run twice (harmless but redundant). Action: Check spec_test.go's glob; if it matches testdata/**, consider removing spec_applicator_test.go or narrowing the pattern.
- F3 (low, from tester): testdata/applicator/oneOf.json line 24 case is labeled 'none matches' (invalid expected) but tests oneOf of [string, boolean] against integer 12, which genuinely matches neither, so the label is accurate but confusing. Consider relabeling to 'both choices fail to match'.
- Story 180 dependency: Test assumes units_test.go is internal/output/units_test.go and child units precede parent units in the output. If story 180 changes the unit type, location, or report-building order, this story's rooted-location assertions and nested-failure reporting will need adjustment.
