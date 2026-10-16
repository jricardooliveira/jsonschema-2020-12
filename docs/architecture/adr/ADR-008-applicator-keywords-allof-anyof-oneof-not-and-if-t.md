# ADR #8: Applicator keywords allOf, anyOf, oneOf, not and if/then/else, with keyword locations rooted at the schema root

- Date: 2026-10-16
- Story: #8
- Status: approved by the operator at Checkpoint 2 (run #142, 2026-10-16 20:40 UTC)

## Decision
Keywords live in internal/keywords/applicator, one file per keyword, each registering from init(). The applicator package is already blank-imported by internal/keywords/all/all.go, so that file is only read. The crowded hubs (internal/keywords/registry.go, internal/compile/compile.go, internal/eval/eval.go) are touched once, in T1, and only to add the location-aware child compile/evaluate seam that every later applicator reuses. No keyword-specific code goes there. A shared helper, applicator/subschemas.go, compiles schema arrays at /<kw>/<i>. 'if' reads its sibling then/else through the compile context; then.go and else.go register so they are known keywords but contribute no evaluation on their own. anyOf/oneOf never short-circuit, per the operator's decision, so story 192 gets complete information.

## Affected modules
- internal/keywords/registry.go
- internal/compile/compile.go
- internal/compile/compile_location_test.go
- internal/eval/eval.go
- internal/eval/eval_subschema_test.go
- internal/output/units_test.go
- internal/keywords/applicator/subschemas.go
- internal/keywords/applicator/allof.go
- internal/keywords/applicator/allof_test.go
- internal/keywords/applicator/anyof.go
- internal/keywords/applicator/anyof_test.go
- internal/keywords/applicator/oneof.go
- internal/keywords/applicator/oneof_test.go
- internal/keywords/applicator/not.go
- internal/keywords/applicator/not_test.go
- internal/keywords/applicator/if.go
- internal/keywords/applicator/if_test.go
- internal/keywords/applicator/then.go
- internal/keywords/applicator/then_test.go
- internal/keywords/applicator/else.go
- internal/keywords/applicator/else_test.go
- testdata/applicator/allOf.json
- testdata/applicator/anyOf.json
- testdata/applicator/oneOf.json
- testdata/applicator/not.json
- testdata/applicator/if-then-else.json
- testdata/applicator/boolean-subschemas.json
- testdata/applicator/nested.json
- spec_applicator_test.go

## Data / API impact
- DB impact: no
- API impact: no
- Migration needed: no

## Constraints for implementation
- Reports never include instance values; only instanceLocation, keywordLocation and error text.
- No central keyword list or switch; registration only via init() in each keyword's file.
- No new dependency; go.mod unchanged.
- anyOf and oneOf evaluate every choice (operator decision for unevaluated*).
- Compiled nodes and their locations are immutable after Compile returns; per-call state stays in the evaluation context.
- Tests ask the code for messages and locations where defined, rather than repeating literals unnecessarily.
- Each JSON case cites a 2020-12 spec section (the speccases loader rejects missing citations).

## Risks
- Story 180 is still in progress: the output-unit type, its units_test.go path and how basic renders nested child failures may change before this story builds.
- Adding location to CompileContext changes the signature seen by the type/enum/const keywords and by approved stories 181-184, which all write registry.go and compile.go.
- 'if' needing sibling access may require exposing the parent schema object in CompileContext, if it does not already.

## Breaking changes
- Keyword locations of nested units change from child-relative to schema-root-relative; story 180's TestUnitOrderAndNoValues assertion is updated accordingly.

## Sensitivity
- security
