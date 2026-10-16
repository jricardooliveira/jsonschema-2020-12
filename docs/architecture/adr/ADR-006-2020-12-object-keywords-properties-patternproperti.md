# ADR #6: 2020-12 object keywords: properties, patternProperties, additionalProperties, propertyNames, dependentSchemas, maxProperties, minProperties, required, dependentRequired

- Date: 2026-10-16
- Story: #6
- Status: proposed (pending architecture sign-off)

## Decision
There are nine new keyword files, split by 2020-12 vocabulary. The applicator package gets properties, patternProperties, additionalProperties, propertyNames and dependentSchemas. The validation package gets maxProperties, minProperties, required and dependentRequired. Each file registers itself from init() through keywords.Register, with VocabularyURI, the same way items.go and pattern.go do. all.go and registry.go are not edited.

Ordering rule. The operator has decided this rule, it replaces the story's 'document order' wording, and it must be recorded in the ADR as THE rule:
(a) Keywords are evaluated and reported in sorted keyword-name order. The engine already does this: compile.go sorts the names and evalNode follows Node.Items.
(b) Within one keyword, member failures come in sorted member-name order. That is byte-wise order of the UTF-8 names, which sort.Strings gives on Go strings. This covers properties, patternProperties, additionalProperties and propertyNames. dependentSchemas and dependentRequired also walk their trigger names in sorted order.
(c) Missing required and dependentRequired names are reported in the order the schema array lists them.
The only requirement is determinism: the same input gives byte-identical output, and a test proves it.

With this rule, dependentRequired failures come before properties failures in the same object, because 'dependentRequired' sorts before 'properties'. That is accepted, not a deviation. The decoder is not touched, and an instance object stays a Go map[string]any. Each keyword sorts locally (keys := a sorted copy of the map's keys). There is no shared helper, because it is two lines. The rule goes in a short package doc comment in properties.go.

Two tests cover it. The first runs one schema against a document with more than 20 keys, inserted in non-sorted order. It renders Result.Output("basic") twice and compares the bytes. The second checks that the unit order is keyword order, then sorted member names, then missing names in schema order.

additionalProperties reads its sibling properties and patternProperties at compile time through CompileContext.Siblings, the same way items.go reads prefixItems. It must not read them from the evaluated record: compile sorts keywords, so additionalProperties runs before patternProperties and properties. Reading siblings at compile time also limits it to the same schema object, as 2020-12 requires. additionalProperties compiles its own copy of the sibling patternProperties regexes through the same steps, so it never depends on the order keywords are compiled.

The evaluated-property record. The only per-call handle that child contexts inherit is EvalContext.Items, which holds *eval.EvaluatedItems, and EvalChild calls Mark/Rollback on it. The property record therefore goes into that same append-only log. A failed subschema then rolls back its property annotations together with its item annotations, and registry.go and eval.go are not touched.
- evaluated_items.go gets one small edit: itemEntry gains a names []string field and a kind flag, and At() skips property entries.
- The new file evaluated.go adds (*EvaluatedItems).RecordProperties(instLoc, kwLoc string, names []string), which copies the slice, and PropertiesAt(instLoc string) map[string][]string.
- properties, patternProperties and additionalProperties type-assert ctx.Items to a small local interface {RecordProperties(...)} declared in the applicator package. Go interfaces are structural, so no hub changes.
- A keyword records the names it applied a subschema to, in sorted order, whether or not the keyword itself passed. Dropping the annotations of failed subschemas is already done by the EvalChild rollback; anything more is left to stories 191/192.

The crowded internal/eval/eval.go is not in the plan: evalNode already attaches the record through attachItems.

patternProperties regexes go through the same steps as validation/pattern.go: the size check against validation.DefaultMaxPatternSize, then ecmaregex.Translate, then regexp.Compile. Any failure is a compile error (exit 2).

Messages are value-free:
- Failures under properties, patternProperties and additionalProperties are the child's units at instance location parent + '/' + escaped key (~0/~1), followed by one wrapping unit.
- propertyNames evaluates each key as a string instance at the object's own location (core 10.3.2.4), so the key never appears in the message.
- required and dependentRequired report 'missing required property: <name from schema>'.
- Flag output may stop early, as it does today.

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
- patternProperties, and additionalProperties for its sibling patterns, must run the size check against validation.DefaultMaxPatternSize, then call ecmaregex.Translate and regexp.Compile. They must not reimplement translation. An untranslatable or uncompilable pattern is a compile error (exit code 2).
- Ordering rule (operator decision, record it in the ADR): keywords are evaluated and reported in sorted keyword-name order. Within one keyword, member failures come in sorted member-name order (sort.Strings on a copy of the map keys, never Go map range order), and dependentSchemas/dependentRequired walk their trigger names in sorted order. Missing required and dependentRequired names are reported in schema array order. The decoder (internal/jsonvalue/decode.go) is not touched.
- Document the ordering rule in a doc comment in properties.go.
- additionalProperties takes its siblings from CompileContext.Siblings, never from the evaluated record, because it is evaluated before properties and patternProperties.
- Keywords reach the record only through a local interface type-asserted from ctx.Items. registry.go and eval.go are not edited.
- Tests read limits and messages from the code where they are defined (for example DefaultMaxPatternSize), not from repeated literals.
- The determinism test renders basic output twice for a document with more than 20 keys inserted out of order and compares the bytes. A second test asserts the unit order: keyword order, then sorted member names, then missing names in schema order.

## Risks
- internal/eval/evaluated_items.go is outside the original reservation. It must gain a property-entry field so that one Mark/Rollback log covers both item and property annotations. Without that, a failed subschema would leak property annotations, or the crowded registry.go would need a second record. The edit is small, and the file is not crowded.
- internal/eval/eval.go is no longer needed, because attachItems already wires the record. The reservation can release it.
- The acceptance criterion and T2/T5 wording ('order the keys appear in the document') are superseded by the operator's ordering rule (sorted keyword names, then sorted member names, then schema order for missing names). Reviewers should check against the ADR rule, not the original story text.
- patternProperties repeats about 6 lines of size check, Translate and Compile from internal/keywords/validation/pattern.go, in internal/keywords/applicator/patternproperties.go. Moving them into ecmaregex would change the translation package, which is a non-goal. Follow-up: share a CompilePattern helper.
- Story 191 may define its own annotation mechanism. It should build on the EvaluatedItems log (evaluated_items.go and evaluated.go) rather than duplicate it.

## Breaking changes
- none

## Sensitivity
- pii
