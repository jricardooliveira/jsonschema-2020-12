# Release notes — #13: Honour $vocabulary from embedded and registry meta-schemas

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #13
- Status: proposed (pending Checkpoint 3 sign-off)

**⚠️ the release-agent has concerns.**

## What changed

This change implements mandatory 2020-12 '$vocabulary' processing: when a schema's meta-schema declares vocabularies, the validator enforces the active set, refuses unknown required vocabularies with exit code 2, and ignores unknown optional ones. It also registers '$schema', '$vocabulary' and '$comment' as silent core keywords. All 12 acceptance criteria are tested through the library and CLI with custom meta-schemas supplied via the registry, and concurrency is verified with -race. One test coverage gap remains: the malformed-vocabulary error message is correctly implemented but not explicitly verified to include the bad entry in the CLI test.

- internal/compile/vocabulary.go: new ParseVocabulary() parses meta-schema '$vocabulary', enforces core required, rejects unknown-required, recognises eight standard 2020-12 vocabularies plus format-assertion
- internal/compile/vocabulary_apply.go: new vocabAt() and vocabFor() compute per-resource active set at compile time, inherited by embedded resources without their own '$schema'
- internal/compile/vocabulary_errors.go: new error types UnknownVocabularyError, MalformedVocabularyError, CoreVocabularyError, UnsupportedSchemaError (draft-specific)
- internal/compile/compile.go: added FormatAssertion flag on root Node, vocabs cache, and routing of inactive-vocabulary keywords to unknown-keyword path
- internal/compile/metaschema.go: new otherDraft() recognises draft-07 and other non-2020-12 URIs, UnsupportedSchemaError reported instead of generic ErrUnsupportedSchema
- internal/compile/refs.go: added vocabs and formatAssertion tracking across $ref resolution
- internal/keywords/vocabularies.go: new StandardVocabularies() and IsStandardVocabulary() expose the eight recognised URIs
- internal/keywords/core/schema.go, vocabulary.go: new silent keywords with no-op Compile and Evaluate
- vocabulary.go: new FormatAssertionRequested() accessor on Schema
- vocabulary_errors.go: exported error type aliases for caller detection with errors.As
- testdata/vocabulary/: three spec-derived case files (required_optional_unknown.json, unlisted_vocabulary.json, embedded_resource.json) and six custom meta-schemas
- vocabulary_test.go, vocabulary_apply_test.go, vocabulary_race_test.go, vocabulary_ref_test.go: library tests covering all 12 ACs
- cmd/jsonschema/vocabulary_cli_test.go: CLI contract tests with registry flags, exit codes 0/1/2, error message content
- testdata/metaschema/schema_keyword.json: updated draft-07 error message to 'unsupported $schema: draft-07; only 2020-12 is supported'
- cmd/jsonschema/metaschema_cli_test.go: updated draft-07 assertion to match new message

## How to verify

- Load a schema with a custom meta-schema (via registry) that lists only 'core' and 'applicator' vocabularies, then validate a document with 'minimum: 5' against value 3: it must pass (minimum ignored)
- Load a schema whose meta-schema requires an unknown vocabulary (e.g. 'https://example.com/vocab/money'): Compile must fail with exit code 2 and an error naming both the vocabulary and the meta-schema
- Load a schema whose meta-schema marks 'core' as optional (false): Compile must fail with 'core vocabulary must be required'
- Load a schema with a meta-schema that lists 'format-assertion' as required, call FormatAssertionRequested(): it must return true; 'format' stays annotation-only
- Load a schema with $schema naming 'http://json-schema.org/draft-07/schema#': Compile must fail with 'unsupported $schema: draft-07; only 2020-12 is supported'
- Load a schema and validate against it, then call s.Validate() and render to 'basic' output: the report must not contain any '/$schema', '/$vocabulary' or '/$comment' annotations
- Load a schema with an embedded resource that has its own $schema pointing to a meta-schema without 'validation': the outer schema checks 'minimum', the inner does not (both work as expected)

## Migration

none

## Rollback

Revert the commit. No database changes or stored state. Compiled schemas are immutable after load; rolling back does not affect any cached or in-use schema. A client that calls FormatAssertionRequested() will need to handle the method not existing (add a build tag or conditional compile for older versions if needed).

## Known limitations

- Format assertion is recognised and recorded, but 'format' keywords still produce annotations only, never assertions—a later story will act on the recorded flag
- Only 2020-12 meta-schemas are supported; schemas naming draft-07, draft-06 or 2019-09 are refused
- The 80-byte truncation of vocabulary URIs in error messages can split a UTF-8 character; the resulting text may be invalid UTF-8 (low severity, affects only error output)
- Unknown vocabularies in a meta-schema are only checked when the schema is loaded (compile time), not when the meta-schema itself is checked

## Concerns

- Test coverage gap (medium, AC8): the CLI test for a malformed '$vocabulary' entry checks only that the filename appears in the error, not that the entry details are included. The library test in vocabulary_test.go does verify the message format. Recommend: add an explicit CLI case that asserts the full message includes the bad entry (e.g. 'invalid '$vocabulary' entry "value-not-boolean": value is not a boolean')
- Two test files outside the originally planned scope were updated: testdata/metaschema/schema_keyword.json and cmd/jsonschema/metaschema_cli_test.go. Both changes are sound (the draft-07 error message must match the new format), but a reviewer should confirm the expected message change does not conflict with story 189's earlier handling.
- The malformed-vocabulary code correctly bounds long URIs to 80 bytes via bound(); callers detecting UnknownVocabularyError or MalformedVocabularyError will receive the bounded URI, not the original. This is intentional (keep error messages short and safe), but if an operator later needs the full URI in logs, the implementation would have to be revised.
