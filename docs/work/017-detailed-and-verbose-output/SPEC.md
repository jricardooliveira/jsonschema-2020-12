# SPEC #17: Detailed and verbose output formats, with URI-safe absoluteKeywordLocation

- Date: 2026-10-17
- Type: feature
- Verdict: pass
- Status: proposed (pending Checkpoint 1 sign-off)

## Problem

The validator can report results only in the 'flag' and 'basic' shapes. Callers who need the nested 2020-12 report shapes, 'detailed' and 'verbose', cannot get them. These shapes show which subschema produced which errors and annotations. Story 188 also found that absoluteKeywordLocation is not always a valid URI: JSON Pointer fragments built from schema keys that contain spaces, '%' or non-ASCII characters are emitted without percent-encoding.

## Why this matters

The brief lists detailed and verbose as optional output features, to be delivered as far as practical, and the Mandatory scope item (7) names them. The spec also defines absoluteKeywordLocation as a URI, so an invalid one breaks any consumer that parses it. Every shape must stay log-safe (no document values) and bounded against hostile input (the report byte cap).

## Acceptance criteria

- A 'detailed' output shape lives in its own file and registers through the output-format registration point, with no edit to a central list. Both the library and the CLI output-format option can select it.
- A 'verbose' output shape lives in its own file and registers through the same registration point. Both the library and the CLI output-format option can select it.
- Every nested output unit in 'detailed' and 'verbose' has 'valid', 'keywordLocation' and 'instanceLocation', plus 'absoluteKeywordLocation' where one is known. Child units are nested under 'errors' or 'annotations' as the 2020-12 shapes describe. 'verbose' includes every evaluated keyword, including the ones that passed. 'detailed' keeps only the units that carry information, following the 2020-12 condensing rules.
- Annotations appear in 'detailed' and 'verbose' wherever the 2020-12 shapes include them. Annotations that list matched property names (properties, patternProperties, additionalProperties, unevaluatedProperties) show the names, treated like instanceLocation, and never show property values. No document value ever appears in either shape. A test proves this with a document whose values are distinctive secret strings.
- With one document, CLI stdout is exactly one 'detailed' or 'verbose' output object and nothing else. With several documents, each line holds the unchanged output object inside its label. The exit codes are unchanged: 0 valid, 1 invalid, 2 error, and the worst result wins.
- The existing rendered-report byte cap, DefaultMaxReportBytes, also bounds 'detailed' and 'verbose' output. A report that would exceed it stops with a clear error and exit code 2, never a truncated or silent result. The existing library option and CLI flag raise it. Tests read the cap from the code instead of repeating the number.
- absoluteKeywordLocation is a valid URI in every output shape (basic, detailed and verbose). Characters not allowed in a URI fragment are percent-encoded in the JSON Pointer fragment produced by Resource.AbsoluteLocation. Tests use schema keys that contain a space, '%' and non-ASCII characters.
- Code comments in the detailed and verbose files record every part of the 2020-12 shapes that is only partly produced.

## Non-goals

- New output shapes beyond flag, basic, detailed and verbose
- A new or separate size limit: the existing DefaultMaxReportBytes cap is reused
- Changing annotation collection semantics (story 191) or unevaluated* behaviour (story 192)
- Quoting any instance value in any report
- README/docs prose for the shapes (story 197)

## Tasks

- **T1** — Percent-encode the absoluteKeywordLocation fragment
  - purpose: Make absoluteKeywordLocation a valid URI in every shape. A helper in a new file percent-encodes the JSON Pointer fragment for characters not allowed in a URI fragment (RFC 3986). Resource.AbsoluteLocation in internal/compile/compile.go calls it with a one-line change.
  - scope: internal/compile/fragment.go, internal/compile/fragment_test.go, internal/compile/compile.go, cmd/jsonschema/abslocation_escape_test.go
  - done when: Unit tests cover keys with a space, '%' and non-ASCII characters, and the result parses with net/url. A CLI basic-shape test shows encoded absoluteKeywordLocation. Existing location tests still pass.
  - delivers: AC7
- **T2** — Shared nested-report builder bounded by the report byte cap
  - purpose: Build the nested 2020-12 unit tree from the evaluation result in one new shared file. It writes annotations without values from the document (property-name lists are shown as names) and enforces DefaultMaxReportBytes while building, so an oversized report fails with the existing limit error.
  - scope: internal/output/tree.go, internal/output/nested_limit_test.go
  - done when: Tests show that exceeding the cap gives the limit error (exit-2 class), that raising the option lets the report through, and that the cap value is read from the code.
  - depends on: T1
  - delivers: AC4, AC6
- **T3** — Detailed shape
  - purpose: Add internal/output/detailed.go, which registers 'detailed' through init() and condenses the tree per 2020-12. Comments document any partly produced parts.
  - scope: internal/output/detailed.go, internal/output/detailed_test.go
  - done when: Table tests on valid and invalid instances check nesting, locations, annotations and that secret values are absent.
  - depends on: T2
  - delivers: AC1, AC3, AC4, AC8
- **T4** — Verbose shape, including passing units
  - purpose: Add internal/output/verbose.go, which registers 'verbose' and emits every evaluated keyword unit. If the per-call unit recording keeps only failures, extend internal/eval/units.go to also keep passing units without touching eval.go. Comments document partly produced parts.
  - scope: internal/output/verbose.go, internal/output/verbose_test.go, internal/eval/units.go, internal/eval/units_verbose_test.go
  - done when: Tests show passing keywords present in verbose and absent from detailed, with no document values. go test -race passes.
  - depends on: T2
  - delivers: AC2, AC3, AC4, AC8
- **T5** — Library and CLI contract tests for detailed and verbose
  - purpose: Prove both shapes are selectable from the library and the CLI. One document gives exactly one object on stdout. Multi-document lines hold unchanged objects. Exit codes and the report cap error behave as specified. The encoded absoluteKeywordLocation appears in both nested shapes.
  - scope: output_nested_test.go, cmd/jsonschema/output_nested_test.go
  - done when: CLI tests parse stdout as a single JSON object for one document and per line for several documents, and assert exit codes 0, 1 and 2.
  - depends on: T3, T4
  - delivers: AC1, AC2, AC5, AC6, AC7
