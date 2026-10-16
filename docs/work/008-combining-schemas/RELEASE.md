# Release notes — #8: Applicator keywords allOf, anyOf, oneOf, not and if/then/else, with keyword locations rooted at the schema root

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #8
- Status: proposed (pending Checkpoint 3 sign-off)

## What changed

Adds applicator keywords allOf, anyOf, oneOf, not, and if/then/else to the JSON Schema 2020-12 validator, making it capable of validating most real-world schemas. Keyword locations are now rooted at the schema root (e.g. /allOf/0/const instead of /const), enabling correct nested error reporting. Every applicator keyword self-registers and lives in its own file; no central list is edited.

- internal/keywords/registry.go: CompileContext gains CompileAt field to compile child schemas at absolute JSON Pointer locations
- internal/compile/compile.go: Exposes Sub and CompileAt seams so keywords can compile children with rooted locations
- internal/keywords/applicator/: Seven new keyword files (allof.go, anyof.go, oneof.go, not.go, if.go, then.go, else.go), each self-registering from init()
- internal/keywords/applicator/subschemas.go: Shared helper compileSchemaArray for non-empty schema arrays at /<keyword>/<i>
- internal/eval/: Child evaluation hook added (EvalChild) returning valid/invalid plus nested units
- internal/output/units_test.go: New test of nested child units with rooted locations and no leaked instance values
- testdata/applicator/: Seven JSON test case files (allOf.json, anyOf.json, oneOf.json, not.json, if-then-else.json, boolean-subschemas.json, nested.json) with spec citations
- spec_applicator_test.go: Test runner for applicator cases via shared speccases loader

## How to verify

- Run `go test -race ./internal/...` to verify thread-safety and all keyword functionality
- Check that a schema with allOf/anyOf/oneOf/not/if-then-else validates documents correctly: `echo '{"allOf":[{"type":"integer"},{"minimum":5}]}' | jsonschema validate schema.json`
- Verify keyword locations are rooted: compile a schema with allOf/0 and check error output contains `/allOf/0/type`, not `/type`
- Confirm every anyOf/oneOf choice is evaluated by setting output cap to 0 and seeing later choices still trigger a limit error
- Verify oneOf names multiple matches: validate with `{"oneOf":[true,true]}` and confirm error says 'choices 0 and 1 both matched'
- Confirm boolean schemas work as subschemas: `{"allOf":[true,false]}` is invalid

## Migration

Keyword locations in output units have changed from child-relative to schema-root-relative. A child `const` inside `allOf/0` now reports keywordLocation `/allOf/0/const` instead of `/const`. Callers parsing output must update location handling. Compiled schemas are immutable after compile-time, so recompile any saved schemas.

## Rollback

Revert to the prior commit and recompile. Keywords allOf, anyOf, oneOf, not, if, then, and else will no longer be available; schemas using them will fail at compile time with 'unknown keyword' errors. Evaluation of schemas without these keywords is unaffected.

## Known limitations

- The applicator spec cases may run twice if spec_test.go's glob includes testdata/applicator/ (tester found this; non-blocking).
- if/then/else do not validate their value when if is absent; a malformed then/else with no if only errors if checked explicitly (tester documented; not a criterion gap).
- Branch locations in if/then/else are built by hand (parent + '/' + name) rather than a location helper; correct but non-idiomatic (tester noted; low-priority refinement).

## Concerns

- Story 180 (output units and basic rendering of nested failures) is a dependency. This change assumes its Units type, KeywordLocation field, and output.Render are already available. The diff modifies internal/eval/units_test.go (x-test-keep helper and TestUnitOrderAndNoValues assertion); verify story 180's output package can accept rooted locations without further changes.
- CompileContext signature changed (added CompileAt field). Stories 181–184 (properties, patternProperties, additionalProperties, items) must be scheduled after this one and use the new rooted-location API.
- All seven new keywords register from init() calls. Verify internal/keywords/all/all.go blank-imports internal/keywords/applicator so they load automatically.
- Thread-safety verified by concurrent_test.go. Compiled schemas are immutable; per-evaluation state stays in the evaluation context. Confirm no per-schema mutable state was added.
