# Release notes — #14: Annotation collection: metadata, format, content and unknown keywords, evaluated properties and items

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #14
- Status: proposed (pending Checkpoint 3 sign-off)

**⚠️ the release-agent has concerns.**

## What changed

The validator now collects JSON Schema 2020-12 annotations during validation: schema metadata (title, description, default, deprecated, readOnly, writeOnly, examples), format names, unknown keywords, and which properties/items each applicator evaluated. Annotations from passing subschemas are retained; from failing ones are discarded. Library users access them via Result.Annotations(); CLI users see them in passing 'basic' reports.

- internal/eval/annotations.go: new annotation collection, enforcement of 10-million default limit, per-subschema evaluation tree, deterministic ordering, and dropping of annotations from failing subschemas
- internal/keywords/annotate.go: new shared helper that registers annotation-only keywords (value from schema, never fail validation)
- internal/keywords/metadata/annotations.go: registers title, description, default, deprecated, readOnly, writeOnly, examples as annotation-only
- internal/keywords/formatannotation/format.go: registers format as annotation-only; unknown formats never error
- internal/keywords/content/content.go: registers contentEncoding, contentMediaType, contentSchema as annotation-only; contentSchema not applied to instance
- internal/compile/compile.go and refs.go: keep unknown keywords and collect them as annotations instead of discarding
- internal/keywords/applicator: properties, patternProperties, additionalProperties, prefixItems, items, contains now record which properties and items they evaluated, following 2020-12 annotation results
- internal/eval/eval.go: builds and populates per-subschema evaluation tree around each subschema evaluation
- internal/keywords/registry.go: adds Tree, Keyword and col fields to EvalContext; adds Annotate and SetMaxAnnotations methods
- internal/output/registry.go: new Budget capped writer and Report type; RegisterReport/RenderReport for report-bytes enforcement (256 MiB default)
- internal/output/basic.go: emits annotation units in passing reports; failing 'basic' and all 'flag' output unchanged
- annotations.go (root): new public Annotation type; Result.Annotations() accessor returns ordered list (empty on failing results)
- jsonschema.go: Schema stores annotation and report-byte caps; Result carries annotations and byte cap
- numberlimits.go: new Options.MaxAnnotations and Options.MaxReportBytes fields with resolver methods
- output.go: new DefaultMaxReportBytes constant; Result.Output calls RenderReport with annotations and cap
- cmd/jsonschema/run.go: new -max-annotations flag; passes through LimitError text for report-bytes cap
- testdata/annotations: five new JSON spec-cited test files (metadata, format, content, unknown, applicators)

## How to verify

- Library: compile schema with title and format keywords, validate a passing document, call result.Annotations() and verify list contains both with correct keyword locations
- Library: compile schema with unknown keyword x-custom, validate passing document, verify result.Annotations() includes it with the schema value
- Library: compile schema where first anyOf branch fails but has a title, validate a document that passes second branch, verify only second branch's annotations appear
- Library: compile with MaxAnnotations: 1, validate a document with two metadata keywords, verify LimitError with Limit 'annotations'
- CLI: `jsonschema validate schema.json doc.json` on a passing document with title/description, verify output JSON includes 'annotations' array
- CLI: run with -max-annotations 1 on multi-keyword schema, verify exit code 2 with error message mentioning 'annotations'
- CLI: run with -o flag (flag output), verify no 'annotations' field appears even for valid documents
- Concurrency: call Result.Annotations() and Result.Output() on same Schema from multiple goroutines under -race, verify deterministic output

## Migration

Library users: No changes needed. Result.Annotations() is new and optional; existing Validate calls are unaffected. CLI users: Passing 'basic' reports are now longer (include annotation units per 2020-12 standard). Scripts parsing 'basic' output should handle the 'annotations' array. Use -o flag with 'flag' output to avoid annotations (flag output never includes them).

## Rollback

Library: existing code ignoring Result.Annotations() works with prior versions unchanged. CLI: use -o flag for 'flag' output shape, or revert to prior release. Git: revert story 191 commit(s).

## Known limitations

- Report-bytes CLI flag (-max-report-bytes) not yet available; story 193 adds it. Library supports Options.MaxReportBytes; exceeding it returns LimitError with exit code 2.
- contentSchema is collected as annotation only and never applied to instance; content decoding and validation out of scope.
- Annotation collection costs time and memory; operator set default limit at 10 million per document and allows tuning.
- Empty stray file dbg_test.go (one line, package declaration only) should be deleted (low severity cosmetic issue).

## Concerns

- dbg_test.go is an empty stray file with only a package declaration; it should be deleted. Tester found this as F5 (low severity, persistent).
- Report-bytes error in run.go passes through but cannot be tested from CLI until story 193 adds the -max-report-bytes flag. Feature is covered at library level in reportbytes_test.go.
- Breaking change: passing 'basic' reports now include annotation units (longer output). Operator explicitly approved this as matching the 2020-12 standard shape. Scripts comparing before/after output will see a difference.
- Tester found no security issues. Build and go test -race pass. Every acceptance criterion (1-11) is covered by spec-derived tests.
