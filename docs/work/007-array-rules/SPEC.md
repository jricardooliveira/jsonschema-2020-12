# SPEC #7: 2020-12 array keywords: prefixItems, items, contains, maxContains, minContains, maxItems, minItems, uniqueItems

- Date: 2026-10-16
- Type: feature
- Verdict: pass
- Status: proposed (pending Checkpoint 1 sign-off)

## Problem

The validator cannot yet check arrays. The 2020-12 array keywords are missing, so documents with arrays are not checked against their element, count and uniqueness rules. Nothing records which array elements were evaluated, and the later unevaluatedItems story needs that record.

## Why this matters

These keywords are part of the brief's mandatory 2020-12 scope (validation keywords maxItems, minItems, uniqueItems, maxContains, minContains and applicators prefixItems, items, contains). Recording the item annotations now, in the per-evaluation context, lets stories 191 and 192 build on them without rewriting these keyword files. The operator asked for that.

## Acceptance criteria

- Each of prefixItems, items, contains, maxContains, minContains, maxItems, minItems and uniqueItems lives in its own file internal/keywords/<vocabulary>/<keyword>.go. Each registers itself from init() into the keyword registry with its name, vocabulary URI, compile function and evaluate function. No central keyword list, switch or all.go entry is edited.
- All eight keywords apply only to array instances. For every non-array instance (null, boolean, number, string, object) each keyword passes and records no annotation.
- prefixItems validates element i against subschema i, for each index that both the array and the prefixItems list have. A shorter array is valid. When it applies to at least one element, prefixItems records in the per-evaluation record the largest index it applied to, or true when it applied to every element (2020-12 core 10.3.1.1).
- items validates only the elements after the ones covered by a sibling prefixItems, and every element when there is no prefixItems. The boolean schemas true and false work as its subschema, so items:false after prefixItems rejects extra elements. When items applies to at least one element it records the annotation true (2020-12 core 10.3.1.2).
- contains evaluates every element and counts the matches. Without minContains, at least one match is required. With minContains n, at least n matches are required, and minContains 0 makes the minimum always satisfied. When maxContains is present, at most that many matches are allowed. maxContains and minContains have no effect without a sibling contains. contains records the matching indices, or true when every element matched (2020-12 core 10.3.1.3, validation 6.4.4 and 6.4.5).
- When several elements fail prefixItems or items, the result lists every failing element: one error per failing position, with its instance location (for example /2, /5, /9) and keyword location. The whole array is checked even after the first failure, and no error contains an instance value.
- When contains fails, the result has exactly one error at the array's instance location. It names contains (or minContains or maxContains) and says how many elements matched and how many were required or allowed. It holds no instance values, and elements that do not match are not reported as errors.
- maxItems and minItems compare the element count with the keyword value using the exact number model, so a value written as 2.0 behaves as 2.
- uniqueItems:true fails when two elements are deep-JSON-equal: 1 and 1.0 are duplicates, and objects are compared regardless of key order. uniqueItems:false or absent has no effect. Each group of equal elements produces one error listing all of its positions (for example 'elements 1, 4 and 9 are equal'), never the values.
- The uniqueItems duplicate check first groups elements by a fingerprint that is consistent with deep equality, so 1 and 1.0 get the same fingerprint, then compares only inside a group. A deterministic test, with no wall-clock threshold, runs it on 100,000 mixed elements (numbers, strings, objects, with 1 and 1.0 duplicated). The test checks that the duplicate groups are exactly right and that the count of full deep-equality comparisons, returned per call, is far below n*n/100. The count is not global state.
- The item annotations from prefixItems, items and contains are stored in the per-evaluation record of evaluated locations, kept in the evaluation context and never in the compiled schema. They are tested for each keyword, and evaluating one compiled schema from several goroutines under go test -race shows no data race.
- Every keyword has spec-derived positive and negative test cases in testdata/<area>/<keyword>.json, run by the shared table-driven runner. Each case cites the 2020-12 core or validation section it derives from.

## Non-goals

- unevaluatedItems itself (story 192) and general annotation collection or output of annotations (story 191)
- Meta-schema validation of malformed keyword values such as a negative minItems (story 189)
- Any wall-clock performance threshold
- Preparing these keywords for drafts other than 2020-12

## Tasks

- **T1** — Per-evaluation evaluated-items record
  - purpose: Give the evaluation context a record of item annotations (prefixItems largest index or true, items true, contains indices or true) that prefixItems, items and contains write and story 192 can read. It follows the same style as the properties record from story 183 and is kept out of the compiled schema.
  - scope: internal/eval/evaluated_items.go, internal/eval/evaluated_items_test.go, internal/eval/evaluated.go, internal/eval/eval.go
  - done when: evaluated_items_test.go passes under go test -race. The record lives only in the per-call context, and eval.go changes by at most one wiring line.
  - delivers: AC11
- **T2** — prefixItems and items
  - purpose: Positional and remaining-element validation with every failing element reported and the item annotations recorded.
  - scope: internal/keywords/applicator/prefixitems.go, internal/keywords/applicator/prefixitems_test.go, internal/keywords/applicator/items.go, internal/keywords/applicator/items_test.go, testdata/applicator/prefixItems.json, testdata/applicator/items.json
  - done when: The spec-derived positive and negative cases pass, including items:false after prefixItems, non-array instances, multiple failing positions and annotation values.
  - depends on: T1
  - delivers: AC1, AC2, AC3, AC4, AC6, AC11, AC12
- **T3** — contains, minContains and maxContains
  - purpose: Count matches over the whole array, apply the min and max bounds, report a single array-level error, and record the matching indices.
  - scope: internal/keywords/applicator/contains.go, internal/keywords/applicator/contains_test.go, internal/keywords/validation/mincontains.go, internal/keywords/validation/mincontains_test.go, internal/keywords/validation/maxcontains.go, internal/keywords/validation/maxcontains_test.go, testdata/applicator/contains.json, testdata/validation/minContains.json, testdata/validation/maxContains.json
  - done when: Cases pass for minContains 0, maxContains bounds, min and max without contains, a single error with matched and required counts and no values, and the indices or true annotation.
  - depends on: T1
  - delivers: AC1, AC2, AC5, AC7, AC11, AC12
- **T4** — maxItems and minItems
  - purpose: Element-count limits using the exact number model.
  - scope: internal/keywords/validation/maxitems.go, internal/keywords/validation/maxitems_test.go, internal/keywords/validation/minitems.go, internal/keywords/validation/minitems_test.go, testdata/validation/maxItems.json, testdata/validation/minItems.json
  - done when: Spec-derived boundary cases pass, including 2.0 as a limit and non-array instances.
  - delivers: AC1, AC2, AC8, AC12
- **T5** — uniqueItems with a fingerprint-grouped duplicate check
  - purpose: Deep-equality uniqueness that stays near-linear on hostile input, with one error per duplicate group.
  - scope: internal/keywords/validation/uniqueitems.go, internal/keywords/validation/uniqueitems_test.go, testdata/validation/uniqueItems.json
  - done when: Cases pass for 1 and 1.0, key-order-independent objects, nested arrays and uniqueItems:false. Grouped errors list positions only. The 100,000-element test checks the exact groups and a comparison count far below n*n/100.
  - delivers: AC1, AC2, AC9, AC10, AC12
