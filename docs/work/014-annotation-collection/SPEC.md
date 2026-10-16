# SPEC #14: Annotation collection: metadata, format, content and unknown keywords, evaluated properties and items

- Date: 2026-10-16
- Type: feature
- Verdict: pass
- Status: passed gate 1 with nothing for the operator to decide (run #144, 2026-10-16 21:28 UTC)

## Problem

The validator only answers valid or invalid. The 2020-12 core spec also requires it to collect annotations during evaluation, each with its instance and keyword locations. Annotations from any failing subschema must be dropped. Several keywords produce only annotations: title, description, default, deprecated, readOnly, writeOnly and examples; 'format' in its default mode; contentEncoding, contentMediaType and contentSchema; and unknown keywords. Applicators must record which properties and items they evaluated. The output formats must show annotations where the 2020-12 shapes include them.

## Why this matters

Annotation collection is part of the mandatory 2020-12 scope (brief item 4, plus items 5, 8 and 9). unevaluatedProperties and unevaluatedItems (story 192) build on the evaluated-location annotations. Detailed and verbose output (story 194) needs the per-subschema evaluation tree. Library users need annotations such as 'default' without having to parse JSON.

## Acceptance criteria

- During every evaluation the engine collects each annotation with its instance location (JSON Pointer), keyword location, absolute keyword location where known, keyword name and value. An annotation produced inside a subschema that fails is discarded, together with every annotation nested below it. Examples are a failing anyOf/oneOf branch, the subschema under 'not', and an 'if' that fails. Annotations from passing subschemas are kept. Collection always runs, with one code path.
- title, description, default, deprecated, readOnly, writeOnly and examples each produce an annotation whose value is the keyword's value from the schema, and none of them ever changes the valid/invalid result. They are registered through ONE shared annotate-only helper. The registrations live in internal/keywords/metadata/annotations.go.
- 'format' (internal/keywords/formatannotation/format.go) is annotation-only by default. It records the format name as its annotation and never fails validation. An unknown format name is never an error.
- contentEncoding, contentMediaType and contentSchema (internal/keywords/content/content.go) are annotation-only through the same helper. Each records its schema value. contentSchema is never applied to the instance, and none of these keywords ever fails validation.
- A keyword that is not registered in the keyword registry is ignored for validation and collected as an annotation whose value is the keyword's schema value.
- properties, patternProperties and additionalProperties each record the set of property names they evaluated. prefixItems records the largest index it applied to, or true when it applied to every item. items records true when it applied to any item. contains records the indices of the items that matched, or true when all matched. All of these follow the 2020-12 core annotation results.
- The library Result has an Annotations() accessor. On a passing result it returns a list of annotations. Each entry has the instance location, keyword location, absolute keyword location where known, keyword and value. On a failing result the list is empty. The list is always ordered by instance location, then by keyword location, so the same input always gives exactly the same list.
- An internal API exposes the per-subschema evaluation tree. For each evaluated subschema it gives the keyword location, instance location, validity, errors and kept annotations, including passing parts inside a failing evaluation. Dropped annotations appear nowhere in this tree.
- The number of collected annotations per document has a built-in limit. The default is a named constant of 10,000,000. The caller can raise or lower it through a library Options field and a CLI flag, and it can never be turned off: a non-positive value is a usage error with exit code 2. Exceeding the limit returns a LimitError with a new limit name, which gives exit code 2 and never a valid or invalid result. Tests read the default from the constant.
- In a passing 'basic' report from the library and the CLI, each annotation appears as a 2020-12 output unit. The unit carries valid true, instanceLocation, keywordLocation, absoluteKeywordLocation where known, and the annotation value, in the same order as Result.Annotations(). A failing 'basic' report lists only errors. 'flag' never lists annotations. Annotation values come only from the schema, and no instance value ever appears in any report.
- Spec-derived tests cover each criterion above. The JSON cases live under testdata/annotations/ and each cites its 2020-12 core or validation section. There are also CLI contract tests for passing and failing basic output, flag output and the annotation-limit flag. go test -race ./... passes with no network access.

## Non-goals

- Showing dropped annotations in any report: the operator chose to leave them out of verbose as well.
- The detailed and verbose output shapes. Story 194 builds them on the internal evaluation tree.
- Format assertion mode and any format checks (stories 195 and 196).
- unevaluatedProperties and unevaluatedItems (story 192).
- Decoding or validating content described by contentEncoding, contentMediaType or contentSchema.
- Making collection optional or lazy: the operator chose to always collect.
- Any coupling to story 193: the annotation limit follows the existing limit style from stories 178 and 179.

## Tasks

- **T1** — Annotation collection core and evaluation tree in internal/eval
  - purpose: Record annotations during evaluation, discard them when a subschema fails, keep a per-subschema evaluation tree, order the results deterministically and enforce the annotation-count limit. Also give keywords a way to record annotations through the evaluation context, and keep unknown keywords so they are collected as annotations.
  - scope: internal/eval/annotations.go, internal/eval/annotations_test.go, internal/eval/eval.go, internal/keywords/registry.go, internal/compile/compile.go
  - done when: go test -race ./internal/eval/... ./internal/keywords/ ./internal/compile/... passes. The tests show that annotations are dropped under a failing subschema, that the ordering holds, that the tree contains passing parts of a failing evaluation, that the limit error fires, and that an unknown keyword is collected.
  - delivers: AC1, AC5, AC8, AC9
- **T2** — Shared annotate-only helper with metadata, format and content keywords
  - purpose: One helper registers keywords whose only effect is to annotate their own schema value. Three files, one per vocabulary, use it.
  - scope: internal/keywords/annotate.go, internal/keywords/annotate_test.go, internal/keywords/metadata/annotations.go, internal/keywords/metadata/annotations_test.go, internal/keywords/formatannotation/format.go, internal/keywords/formatannotation/format_test.go, internal/keywords/content/content.go, internal/keywords/content/content_test.go
  - done when: Unit tests show that each keyword annotates its value and never fails, that unknown formats are accepted, and that contentSchema is not applied to the instance.
  - depends on: T1
  - delivers: AC2, AC3, AC4
- **T3** — Evaluated-location annotations from object and array applicators
  - purpose: properties, patternProperties, additionalProperties, prefixItems, items and contains record the annotation results that 2020-12 core defines for them.
  - scope: internal/keywords/applicator/properties.go, internal/keywords/applicator/patternproperties.go, internal/keywords/applicator/additionalproperties.go, internal/keywords/applicator/prefixitems.go, internal/keywords/applicator/items.go, internal/keywords/applicator/contains.go
  - done when: The spec cases in testdata/annotations/applicators.json pass.
  - depends on: T1
  - delivers: AC6
- **T4** — Library API: Result.Annotations() and the annotation-limit option
  - purpose: Expose the ordered annotation list on passing results and add the Options field for the annotation limit.
  - scope: annotations.go, jsonschema.go, options.go, annotations_test.go
  - done when: Root package tests show an ordered list on a passing result, an empty list on a failing result, and a LimitError when the limit is lowered and exceeded.
  - depends on: T1
  - delivers: AC7, AC9
- **T5** — Annotations in basic output and the CLI limit flag
  - purpose: A passing 'basic' report lists annotation units, a failing one lists only errors, and 'flag' never lists annotations. Add the CLI flag for the annotation limit.
  - scope: internal/output/basic.go, cmd/jsonschema/run.go, cmd/jsonschema/annotations_test.go
  - done when: CLI contract tests check passing and failing basic output, flag output, the limit flag (exit code 2 when exceeded and for non-positive values), and that no instance values are printed.
  - depends on: T4
  - delivers: AC9, AC10
- **T6** — Spec-derived annotation test cases
  - purpose: Add JSON cases that cite the 2020-12 sections for metadata, format, content, unknown keywords, applicator results and dropping on failure, run by a table-driven runner in the root annotations test.
  - scope: testdata/annotations/metadata.json, testdata/annotations/format.json, testdata/annotations/content.json, testdata/annotations/unknown.json, testdata/annotations/applicators.json, annotations_test.go
  - done when: go test -race ./... passes, and every case cites a spec section.
  - depends on: T2, T3, T4
  - delivers: AC1, AC2, AC3, AC4, AC5, AC6, AC11
