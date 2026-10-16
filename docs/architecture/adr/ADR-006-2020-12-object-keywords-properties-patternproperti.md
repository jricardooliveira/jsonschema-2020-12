# ADR #6: 2020-12 object keywords: properties, patternProperties, additionalProperties, propertyNames, dependentSchemas, maxProperties, minProperties, required, dependentRequired

- Date: 2026-10-16
- Story: #6
- Status: proposed (pending architecture sign-off)

## Decision
There are nine new keyword files, split by 2020-12 vocabulary. The applicator package gets properties, patternProperties, additionalProperties, propertyNames and dependentSchemas. The validation package gets maxProperties, minProperties, required and dependentRequired. Each file registers itself from init() through keywords.Register, with VocabularyURI, the same way items.go and pattern.go do. all.go and registry.go are not edited.

Ordering (this revision, as the reviewer asked). The decoder is not touched, and an instance object stays a Go map[string]any. Every keyword that walks object members walks them in sorted key order: byte-wise order of the UTF-8 member names, which is what sort.Strings gives on Go strings. This applies to properties, patternProperties, additionalProperties and propertyNames. dependentSchemas and dependentRequired walk their trigger names in sorted order. required and dependentRequired report missing names in the order the schema array lists them. Each keyword sorts locally (keys := sorted copy of the map's keys). There is no shared helper; it is two lines.

Within one schema object, the engine already evaluates keywords in sorted keyword-name order, because compile.go sorts the names and evalNode follows Node.Items. So the full order of units for one object is: keyword order first, then sorted member name inside each per-member keyword, then schema array order for missing names. For the operator's case (wrong fields under properties/patternProperties/additionalProperties plus missing required fields), this gives exactly 'by sorted member name, then missing required names in schema order', because 'required' sorts after every per-member keyword. The rule goes in a short package doc comment in properties.go. It is tested by running one schema against a document with many keys (more than 20, inserted in non-sorted order) twice through Result.Output("basic") and comparing the bytes. A second test checks that the unit order equals the sorted names followed by the missing names in schema order.

additionalProperties reads its sibling properties and patternProperties at compile time through CompileContext.Siblings, the same way items.go reads prefixItems. It must not read them from the evaluated record: compile sorts keywords, so additionalProperties runs before patternProperties and properties. This also limits it to the same schema object, as the 2020-12 spec requires. additionalProperties compiles its own copy of the sibling patternProperties regexes through the same helper, so it never depends on the order keywords are compiled.

The evaluated-property record. The only per-call handle that child contexts inherit is EvalContext.Items, which holds *eval.EvaluatedItems; EvalChild calls Mark/Rollback on it. The property record therefore goes into that same append-only log, so a failed subschema rolls back its property annotations together with its item annotations, and registry.go and eval.go are not touched. evaluated_items.go gets one small edit: itemEntry gains a names []string field and a kind flag, and At() skips property entries. The new file evaluated.go adds (*EvaluatedItems).RecordProperties(instLoc, kwLoc string, names []string), which copies the slice, and PropertiesAt(instLoc string) map[string][]string. properties, patternProperties and additionalProperties type-assert ctx.Items to a small local interface {RecordProperties(...)} declared in the applicator package. Go interfaces are structural, so no hub changes. A keyword records the names it applied a subschema to, in sorted order, whether or not the keyword itself passed. Dropping annotations of failed subschemas is already done by the EvalChild rollback, and anything more is left to story 191/192.

I dropped the crowded internal/eval/eval.go from the plan: reading the code showed that evalNode already attaches the record through attachItems, so this story needs no wiring there.

patternProperties regexes go through the same steps as validation/pattern.go: the size check against validation.DefaultMaxPatternSize, then ecmaregex.Translate, then regexp.Compile. Any failure is a compile error, which gives exit code 2. Messages are value-free. Failures under properties/patternProperties/additionalProperties are the child's units at instance location parent+'/'+escaped key (~0/~1), followed by one wrapping unit. propertyNames evaluates each key as a string instance at the object's own location, per core 10.3.2.4, so the key never appears in the message. required and dependentRequired report 'missing required property: <name from schema>'. Flag output may stop early, as it does today.

## Affected modules
- internal/eval/evaluated.go
- internal/eval/evaluated_test.go
- internal/eval/evaluated_items.go
- internal/keywords/applicator/properties.go
- internal/keywords/applicator/properties_test.go
- internal/keywords/applicator/patternproperties.go
- internal/keywords/applicator/patternproperties_test.go
- internal/keywords/applicator/additionalproperties.go
- internal/keywords/applicator/additionalproperties_test.go
- internal/keywords/applicator/propertynames.go
- internal/keywords/applicator/propertynames_test.go
- internal/keywords/applicator/dependentschemas.go
- internal/keywords/applicator/dependentschemas_test.go
- internal/keywords/validation/maxproperties.go
- internal/keywords/validation/maxproperties_test.go
- internal/keywords/validation/minproperties.go
- internal/keywords/validation/minproperties_test.go
- internal/keywords/validation/required.go
- internal/keywords/validation/required_test.go
- internal/keywords/validation/dependentrequired.go
- internal/keywords/validation/dependentrequired_test.go
- testdata/applicator/properties.json
- testdata/applicator/patternproperties.json
- testdata/applicator/additionalproperties.json
- testdata/applicator/propertynames.json
- testdata/applicator/dependentschemas.json
- testdata/validation/maxproperties.json
- testdata/validation/minproperties.json
- testdata/validation/required.json
- testdata/validation/dependentrequired.json

## Data / API impact
- DB impact: no
- API impact: no
- Migration needed: no

## Constraints for implementation
- Go standard library only, and go.mod is not edited.
- The compiled Schema stays immutable. Compiled data (subschema nodes, *regexp.Regexp, property-name sets, required lists) is built once in Compile and only read afterwards. Evaluated names live only in ctx.Items. This is verified with go test -race.
- Messages never contain instance values or keys. Keys appear only in the JSON Pointer instance location, with ~0/~1 escaping.
- patternProperties (and additionalProperties for its sibling patterns) must call ecmaregex.Translate and regexp.Compile after the size check against validation.DefaultMaxPatternSize, and must not reimplement translation. An untranslatable or uncompilable pattern is a compile error (exit code 2).
- Member iteration is in sorted key order (sort.Strings on a copy of the map keys, which is byte-wise UTF-8 order), never Go map range order. Trigger names of dependentSchemas and dependentRequired are also walked in sorted order. Missing required and dependentRequired names are reported in schema array order. The decoder (internal/jsonvalue/decode.go) is not touched.
- Within one schema object, the unit order is the engine's existing keyword order (sorted keyword names from compile.go), then the per-keyword order above. Document this in a doc comment in properties.go.
- additionalProperties takes its siblings from CompileContext.Siblings, never from the evaluated record (it is evaluated before properties and patternProperties).
- Keywords reach the record only through a local interface type-asserted from ctx.Items. registry.go and eval.go are not edited.
- Tests read limits and messages from the code where defined (for example DefaultMaxPatternSize), not from repeated literals. The determinism test renders basic output twice for a document with more than 20 keys inserted out of order and compares the bytes.

## Risks
- internal/eval/evaluated_items.go is outside the original reservation. It must gain a property-entry field so one Mark/Rollback log covers both item and property annotations. Without that, a failed subschema would leak property annotations, or registry.go (crowded) would need a second record. The edit is small, and this file is not crowded.
- internal/eval/eval.go was dropped from the plan: attachItems already wires the record. The reservation can release it.
- Sorted keyword evaluation means dependentRequired failures (keyword 'dependentRequired' sorts before 'properties') come before the per-member failures of properties in the same object. The operator's 'members, then missing names' order holds exactly for required and is documented per keyword for dependentRequired. Changing the order across keywords would require editing the crowded compile.go/eval.go and is not done here.
- Sorted key order replaces the earlier document-order answer. The operator has revised it, and the story text that says 'order the keys appear in the document' is read as sorted key order.
- patternProperties repeats the size-check, Translate and Compile steps from validation/pattern.go (about 6 lines), and the duplicated code is internal/keywords/validation/pattern.go and internal/keywords/applicator/patternproperties.go. Moving them into ecmaregex would change the translation package, which is a non-goal. Recorded as a follow-up: share a CompilePattern helper.
- Story 191 may define its own annotation mechanism. It should build on the EvaluatedItems log (evaluated_items.go and evaluated.go) rather than duplicate it.

## Breaking changes
- none

## Sensitivity
- pii
