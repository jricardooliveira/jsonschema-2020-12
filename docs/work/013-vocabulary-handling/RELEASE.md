# Release notes — #13: Honour $vocabulary from embedded and registry meta-schemas

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #13
- Status: proposed (pending Checkpoint 3 sign-off)

**⚠️ the release-agent has concerns.**

## What changed

The validator now reads '$vocabulary' from meta-schemas and activates only the declared vocabularies per schema resource. Keywords from unlisted vocabularies are treated as unknown (annotation-only). A required unknown vocabulary or missing core vocabulary is refused with exit code 2. The format-assertion vocabulary is recognised and recorded.

- internal/compile/vocabulary.go, vocabulary_apply.go, vocabulary_errors.go: new vocabulary parsing, application per resource, and error types
- internal/keywords/vocabularies.go: eight standard vocabulary URIs recognised
- internal/keywords/core/schema.go, vocabulary.go, comment.go: $schema, $vocabulary, $comment register as silent core keywords with no validation effect
- internal/compile/refs.go: wrap compilation errors with %w; route inactive-vocabulary keywords to unknown-keyword handling; store per-resource vocabulary set
- internal/compile/metaschema.go: detect non-2020-12 $schema by draft name (e.g. 'draft-07') and return UnsupportedSchemaError
- vocabulary.go, vocabulary_errors.go: exported error type aliases and FormatAssertionRequested() accessor
- testdata/vocabulary/: three case files with custom meta-schemas and metaschemas/ directory
- cmd/jsonschema/vocabulary_cli_test.go, vocabulary_test.go, vocabulary_apply_test.go, vocabulary_ref_test.go, vocabulary_race_test.go: library and CLI tests citing core 8.1.1 and 8.1.2; concurrent-use tests with -race
- cmd/jsonschema/metaschema_cli_test.go: draft-07 message changed from 'not supported' to 'unsupported $schema: draft-07; only 2020-12 is supported'

## How to verify

- Compile a schema with $schema pointing to a custom meta-schema that lists only core and applicator vocabularies; validate that 'minimum: 5' does not reject the value 3 (validation vocabulary omitted)
- Compile a schema with $schema pointing to a meta-schema that requires 'https://example.com/vocab/money'; confirm it is refused with exit code 2 and the error message names both the unknown vocabulary URI and the meta-schema
- Compile a schema with $schema = 'http://json-schema.org/draft-07/schema#'; confirm it is refused with 'unsupported $schema: draft-07; only 2020-12 is supported'
- Compile a schema with an embedded resource (own $id and $schema) using a restricted meta-schema; validate that the outer schema applies its vocabulary set and the inner applies its own set
- Compile a schema with $schema pointing to a meta-schema requiring the format-assertion vocabulary; confirm Schema.FormatAssertionRequested() returns true
- Validate a schema against a meta-schema with no $vocabulary entry; confirm all eight standard vocabularies are active
- Generate a basic report for {"$schema":"https://json-schema.org/draft/2020-12/schema","type":"string"} validated against "x"; confirm the report contains no /$schema annotation

## Migration

No data or schema changes. Breaking change: Basic and detailed validation reports no longer contain annotations for $schema, $vocabulary, or $comment keywords. Existing test assertions or log parsing that expects these annotations must be updated. Schemas are unaffected; only the report output changes.

## Rollback

This is a new feature with no database migration or irreversible schema changes. Rollback: revert the commits, and any tests or tooling expecting $schema/$vocabulary/$comment annotations will no longer fail.

## Known limitations

- The format-assertion vocabulary is recorded when required, but 'format' itself is not asserted (story 195 will implement that).
- An http:// URL naming a 2020-12 meta-schema (e.g. 'http://json-schema.org/draft/2020-12/schema#') will produce the message 'unsupported $schema: 2020-12; only 2020-12 is supported', which is confusing (should detect 2020-12 by version, not by whether URI is http vs https; medium impact, edge case).
- No test asserts that keywords from unlisted vocabularies appear in the annotation output (test coverage gap for AC4's annotation half; medium impact, feature works but not fully verified).
- Unknown vocabulary URIs in error messages are not bounded in length (could be arbitrarily long; low impact, edge case).
- Malformed '$vocabulary' entries in a meta-schema are caught by the meta-schema validation check before ParseVocabulary can report them end-to-end through the library (test coverage gap, not a functional problem; low impact).

## Concerns

- F1: An http:// 2020-12 $schema produces a contradictory message. The regex otherDraft() checks the URI pattern and extracts 'draft-07' etc., but an http:// URI matching 'draft/2020-12' would return 'draft-07'-style extraction. This is a real gap but rare (who uses http:// for 2020-12?). Consider: accept as known limitation, or fix before release to avoid confusing users.
- F3: Test coverage for AC4's annotation aspect (unlisted vocabulary keywords should be in annotations, e.g. 'minimum' as annotation when validation is omitted) does not appear to be in the test files. The feature works by routing inactive keywords to unknown-keyword handling, but no test explicitly asserts they show up in reports. Should add a test case.
- Breaking change: Reports no longer carry $schema/$vocabulary/$comment annotations. Any user tooling parsing these annotations will break. The architecture notes this, but it's significant enough to highlight prominently in migration.
