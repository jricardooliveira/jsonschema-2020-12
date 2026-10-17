# SPEC #15: unevaluatedProperties and unevaluatedItems (2020-12 unevaluated vocabulary)

- Date: 2026-10-17
- Type: feature
- Verdict: pass
- Status: passed gate 1 with nothing for the operator to decide (run #150, 2026-10-17 00:39 UTC)

## Problem

The validator cannot yet reject object properties or array items that no other rule covered. Doing so needs the evaluated-location annotations from every adjacent and nested applicator, including those reached through $ref and $dynamicRef, and it must ignore annotations from subschemas that failed.

## Why this matters

unevaluatedProperties and unevaluatedItems are part of the mandatory 2020-12 scope (brief, Mandatory scope item 4). Schemas that close an object or array across allOf, $ref or dynamic extension depend on them.

## Acceptance criteria

- 'unevaluatedProperties' lives in internal/keywords/unevaluated/unevaluatedproperties.go and registers itself from init() into the keyword registry under the 2020-12 unevaluated vocabulary URI (https://json-schema.org/draft/2020-12/vocab/unevaluated). It has its own test file, and no central keyword list or switch is edited.
- 'unevaluatedItems' lives in internal/keywords/unevaluated/unevaluateditems.go and registers itself the same way. It has its own test file.
- unevaluatedProperties applies its subschema to each instance property that no successful evaluation at the same instance location has evaluated. Evaluating sources are properties, patternProperties, additionalProperties, dependentSchemas, allOf, anyOf (every passing branch), oneOf, if/then/else, $ref, $dynamicRef and nested unevaluatedProperties. A property evaluated only inside a failed subschema (including a failed 'if', a failing anyOf/oneOf branch or a 'not' subschema) counts as unevaluated. Non-object instances are ignored.
- unevaluatedItems applies its subschema to each array item that no successful evaluation at the same instance location has evaluated. Evaluating sources are prefixItems, items, contains (only the items that matched), allOf, anyOf, oneOf, if/then/else, $ref, $dynamicRef and nested unevaluatedItems, with the same rule that failed subschemas count for nothing. Non-array instances are ignored.
- In any schema object, unevaluatedProperties and unevaluatedItems are evaluated after every sibling keyword of that schema object, whatever the key order in the schema text.
- When unevaluatedProperties or unevaluatedItems succeeds, every property or item it examined is recorded as evaluated, so an enclosing unevaluated keyword treats them as covered.
- When a document is rejected, there is one error per uncovered property or item. Each error carries that property's or item's own instance location (for example /nickname and /age), the keyword location and the absolute keyword location where known. No error quotes an instance value.
- Spec-derived positive and negative cases, each citing its 2020-12 core section (11.2 unevaluatedItems, 11.3 unevaluatedProperties, 7.7 annotation collection), are stored in testdata/unevaluated/unevaluatedProperties.json and testdata/unevaluated/unevaluatedItems.json and run by the shared runner. They cover adjacent and nested applicators, failed branches, nested unevaluated keywords, a $ref target and a $dynamicRef/$dynamicAnchor extension that adds properties or items.

## Non-goals

- Changing how other applicators collect annotations, beyond what this story needs (story 191 owns annotation collection).
- detailed/verbose output shapes (story 194).
- Limits on evaluation steps (story 193).

## Tasks

- **T1** — Evaluate after siblings
  - purpose: Give the keyword registry one flag that marks a keyword as depending on its siblings' annotations, and have the evaluator run flagged keywords last in each schema object, without naming any keyword.
  - scope: internal/keywords/registry.go, internal/eval/eval.go, internal/eval/order_test.go
  - done when: internal/eval/order_test.go shows that a flagged test keyword sees sibling annotations whatever the key order. The existing registry and eval tests still pass.
  - delivers: AC5
- **T2** — unevaluatedProperties keyword
  - purpose: Implement compile and evaluate for unevaluatedProperties using the per-call evaluated-properties annotations. Emit one error per uncovered property and record covered properties on success.
  - scope: internal/keywords/unevaluated/unevaluatedproperties.go, internal/keywords/unevaluated/unevaluatedproperties_test.go
  - done when: go test -race ./internal/keywords/unevaluated/ passes, including the per-property error locations and the no-value check.
  - depends on: T1
  - delivers: AC1, AC3, AC6, AC7
- **T3** — unevaluatedItems keyword
  - purpose: Implement compile and evaluate for unevaluatedItems using the per-call evaluated-items annotations (prefixItems/items index, contains matches, nested). Emit one error per uncovered item and record covered items on success.
  - scope: internal/keywords/unevaluated/unevaluateditems.go, internal/keywords/unevaluated/unevaluateditems_test.go
  - done when: go test -race ./internal/keywords/unevaluated/ passes, including the per-item error locations and the no-value check.
  - depends on: T1
  - delivers: AC2, AC4, AC6, AC7
- **T4** — Spec-derived cases
  - purpose: Write JSON cases citing core 11.2, 11.3 and 7.7 for every listed applicator, failed branches, nesting, $ref and $dynamicRef extension.
  - scope: testdata/unevaluated/unevaluatedProperties.json, testdata/unevaluated/unevaluatedItems.json
  - done when: go test ./... runs TestSpecCases over both files, and all cases pass.
  - depends on: T2, T3
  - delivers: AC3, AC4, AC8
