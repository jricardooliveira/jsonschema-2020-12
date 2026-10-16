# SPEC #6: 2020-12 object keywords: properties, patternProperties, additionalProperties, propertyNames, dependentSchemas, maxProperties, minProperties, required, dependentRequired

- Date: 2026-10-16
- Type: feature
- Verdict: pass
- Status: passed gate 1 with nothing for the operator to decide (run #141, 2026-10-16 20:16 UTC)

## Problem

The validator cannot yet check JSON objects. None of the 2020-12 object applicators or object validation keywords exist, and the later unevaluatedProperties story needs a record of which property names these applicators checked.

## Why this matters

Object keywords are part of the mandatory 2020-12 scope. unevaluatedProperties (story 192) relies on the evaluated-property record these applicators produce, and the operator asked for that record to exist now so story 192 never has to reopen these files.

## Acceptance criteria

- Each of the nine keywords lives in its own file internal/keywords/<vocabulary>/<keyword>.go with its own <keyword>_test.go, and registers itself from init() through the keyword registry. No central keyword list, switch or internal/keywords/all/all.go is edited. properties, patternProperties, additionalProperties, propertyNames and dependentSchemas are in the applicator vocabulary package. maxProperties, minProperties, required and dependentRequired are in the validation vocabulary package.
- Every one of the nine keywords has no effect on instances that are not JSON objects (string, number, integer, boolean, null, array): such instances pass these keywords.
- 'properties' applies each subschema to the instance member with the same name. 'patternProperties' applies each subschema to every member whose name matches the regex, compiled through the existing ECMA-262-to-RE2 translation step. A patternProperties regex that the translation step or RE2 cannot handle is a schema error (exit code 2), never 'invalid'. Boolean subschemas true and false work in both keywords.
- 'additionalProperties' applies its subschema only to members whose names are neither listed in 'properties' nor matched by any 'patternProperties' regex in the same schema object. Keywords in other schema objects (for example inside allOf or $ref) have no effect on which members count as additional.
- 'propertyNames' validates every member name as a string instance. 'dependentSchemas' applies the whole subschema to the object when the named property is present, and has no effect when it is absent.
- 'maxProperties' and 'minProperties' compare the number of members, inclusively. 'required' fails for each listed name that is absent. 'dependentRequired' fails for each listed name that is absent when its trigger property is present. Each missing-field error names the missing field, taken from the schema, for example 'missing required property: email'.
- 'properties', 'patternProperties' and 'additionalProperties' each record the set of property names they applied a subschema to, as their 2020-12 annotation. The record is kept in the per-evaluation context, never in the compiled Schema, so a later keyword in the same schema object can read it.
- Every error gives the failing instance location as a JSON Pointer, with ~0/~1 escaping, plus the keyword location. A failing key appears only inside the instance location path, for example /users/alice@example.com. The message text never repeats the key and never quotes any instance value.
- When one object breaks several of these rules, the basic, detailed and verbose reports list every failure. The flag report may stop at the first failure.
- Failures always come out in the same order: first in the order the keys appear in the document, then missing required and dependentRequired fields in the order the schema lists them. Running the same schema and document twice gives byte-identical output.
- Each keyword has spec-derived positive and negative test cases in testdata/<vocabulary>/<keyword>.json. The cases run through the shared table-driven runner, cite the 2020-12 core or validation section they derive from, and pass under go test -race with no network.

## Non-goals

- unevaluatedProperties itself (story 192). This story only records the evaluated property names.
- General annotation collection, dropping annotations from failing subschemas, and showing annotations in output (story 191).
- Detailed and verbose output shapes themselves (story 194). The ordering and all-failures rules here apply to them once they exist.
- Array, applicator-combinator, reference and format keywords (stories 184-196).
- Any change to the regex translation step beyond calling it.

## Tasks

- **T1** — Per-evaluation record of evaluated property names
  - purpose: Let properties, patternProperties and additionalProperties record, for the current instance location and schema object, which member names they evaluated, so additionalProperties and the later unevaluatedProperties can read them. The record lives in the per-call evaluation context and the compiled Schema stays immutable.
  - scope: internal/eval/evaluated.go, internal/eval/evaluated_test.go, internal/eval/eval.go
  - done when: evaluated_test.go shows names recorded and read back per instance location and schema object, with no shared state between two concurrent evaluations (go test -race).
  - delivers: AC7
- **T2** — properties, patternProperties, additionalProperties
  - purpose: Add the three property applicators as self-registering files. patternProperties compiles through the translation step and reports untranslatable patterns as a schema error. additionalProperties considers only names not covered by properties or patternProperties in the same schema object. All three record evaluated names via T1, and iterate members in document order.
  - scope: internal/keywords/applicator/properties.go, internal/keywords/applicator/properties_test.go, internal/keywords/applicator/patternproperties.go, internal/keywords/applicator/patternproperties_test.go, internal/keywords/applicator/additionalproperties.go, internal/keywords/applicator/additionalproperties_test.go, testdata/applicator/properties.json, testdata/applicator/patternproperties.json, testdata/applicator/additionalproperties.json
  - done when: Keyword tests and testdata cases pass, including non-object instances, boolean subschemas, an untranslatable pattern giving a schema error, and additionalProperties ignoring names matched in the same object only.
  - depends on: T1
  - delivers: AC1, AC2, AC3, AC4, AC7, AC8, AC11
- **T3** — propertyNames and dependentSchemas
  - purpose: Add propertyNames, which validates each key as a string instance and puts the key only in the instance location, and dependentSchemas, which applies its subschema only when the trigger property is present.
  - scope: internal/keywords/applicator/propertynames.go, internal/keywords/applicator/propertynames_test.go, internal/keywords/applicator/dependentschemas.go, internal/keywords/applicator/dependentschemas_test.go, testdata/applicator/propertynames.json, testdata/applicator/dependentschemas.json
  - done when: Tests show a rejected key appearing only in the instance location and never in the message, and dependentSchemas applying only when its trigger is present.
  - delivers: AC1, AC2, AC5, AC8, AC11
- **T4** — maxProperties, minProperties, required, dependentRequired
  - purpose: Add the four object validation keywords. Missing-field errors name the field from the schema, in schema order, and never quote instance data.
  - scope: internal/keywords/validation/maxproperties.go, internal/keywords/validation/maxproperties_test.go, internal/keywords/validation/minproperties.go, internal/keywords/validation/minproperties_test.go, internal/keywords/validation/required.go, internal/keywords/validation/required_test.go, internal/keywords/validation/dependentrequired.go, internal/keywords/validation/dependentrequired_test.go, testdata/validation/maxproperties.json, testdata/validation/minproperties.json, testdata/validation/required.json, testdata/validation/dependentrequired.json
  - done when: Tests cover boundary counts, several missing fields reported in schema order, a message such as 'missing required property: email', and non-object instances passing.
  - delivers: AC1, AC2, AC6, AC8, AC11
- **T5** — All failures and deterministic order
  - purpose: Prove that one object breaking several of these rules reports every failure in basic output, in document key order followed by missing fields in schema order, and that two runs give identical output. The flag output may short-circuit.
  - scope: internal/keywords/validation/required_test.go, internal/keywords/applicator/properties_test.go
  - done when: A test with two wrong fields and one missing field shows three errors in the required order in basic output, and compares the rendered output from two runs byte for byte.
  - depends on: T2, T4
  - delivers: AC9, AC10
