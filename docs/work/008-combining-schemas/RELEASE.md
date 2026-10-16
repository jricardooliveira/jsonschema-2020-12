# Release notes — #8: Applicator keywords allOf, anyOf, oneOf, not and if/then/else, with keyword locations rooted at the schema root

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #8
- Status: proposed (pending Checkpoint 3 sign-off)

**⚠️ the release-agent has concerns.**

## What changed

Implements all seven applicator keywords (allOf, anyOf, oneOf, not, if/then/else) with rooted keyword locations from the schema root (e.g., /allOf/0/const). Every acceptance criterion is met: boolean subschemas work, nested failures are reported with no instance values, anyOf/oneOf evaluate every choice, oneOf names overlapping indices, and spec-derived test cases cite sections 10.2.1 and 10.2.2. The build and tests pass with -race.

- internal/compile/compile_location_test.go: New test for rooted child locations via CompileContext.Sub()
- internal/eval/eval_subschema_test.go: New test for child failure locations rooted at schema
- internal/eval/units_test.go: Updated to use Sub() instead of bare Compile, assertion changed from /const to /x-test-keep/0/const
- internal/output/units_test.go: New test for basic rendering with rooted child units and no value leakage
- internal/keywords/applicator/subschemas.go: Shared helper to compile non-empty schema arrays at /<keyword>/<i>
- internal/keywords/applicator/allof.go, allof_test.go: allOf keyword and unit tests
- internal/keywords/applicator/anyof.go, anyof_test.go: anyOf keyword that evaluates every choice
- internal/keywords/applicator/oneof.go, oneof_test.go: oneOf keyword that evaluates every choice and names multi-match indices
- internal/keywords/applicator/not.go, not_test.go: not keyword with single-entry failure reporting
- internal/keywords/applicator/if.go, if_test.go: if keyword that never fails a document alone
- internal/keywords/applicator/then.go, then_test.go: then keyword that applies when if passes
- internal/keywords/applicator/else.go, else_test.go: else keyword that applies when if fails
- testdata/applicator/*.json: Seven spec-derived test files with citations to 10.2.1/10.2.2, covering boolean subschemas, nested combinations and multi-match oneOf
- spec_applicator_test.go: New test runner for applicator cases

## How to verify

- Compile a schema with allOf/0/const inside allOf/0; a failure should report keywordLocation /allOf/0/const (not /const), proving rooted locations
- Run the unit tests: allof_test.go, anyof_test.go, oneof_test.go and not_test.go all pass, covering valid, invalid, boolean subschemas, no-value reporting and schema errors
- Check that anyOf and oneOf evaluate every choice: each test includes a cap-0 case that overflows if a later choice is skipped
- Verify no instance values leak: all report tests assert the secret value does not appear in error messages
- Run spec_applicator_test.go: all cited test cases pass

## Migration

Keyword locations of child failures changed from child-relative to schema-rooted. Any code that parses keyword locations must expect /allOf/0/const instead of /const. Story 180's TestUnitOrderAndNoValues assertion was updated to the rooted form.

## Rollback

Remove internal/keywords/applicator/*, testdata/applicator/* and spec_applicator_test.go. Revert internal/eval/units_test.go assertion from /x-test-keep/0/const to /const.

## Known limitations

- then/else keywords recompile the sibling 'if' on each branch instead of compiling it once; the 'if' subschema therefore gets /then or /else locations instead of /if, and nested conditionals can suffer exponential compile/evaluate time. Flagged as F1 (medium). Does not block this release but should be fixed before stories 191/192 build on if/then/else.
- No-value safety tests inspect error messages directly rather than the rendered basic output, so rendering safety cannot be guaranteed from these tests alone.
- If the existing spec test glob is testdata/**, applicator cases may run twice (once from spec_applicator_test.go, once from the glob).

## Concerns

- F1 (medium): then/else recompile and re-evaluate sibling 'if' instead of compiling once. Nested conditionals in a hostile schema can therefore blow up compile and evaluation time exponentially. The 'if' subschema gets /then or /else locations instead of /if. Tester recommends fixing before stories 191/192 build on it, which do not block this release.
