# Release notes — #13: Honour $vocabulary from embedded and registry meta-schemas

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #13
- Status: proposed (pending Checkpoint 3 sign-off)

**⚠️ the release-agent has concerns.**

## What changed

Feature-complete implementation of 2020-12 $vocabulary processing (story US-0190, 12/12 acceptance criteria). Each schema resource's meta-schema vocabulary declarations are read at compile time, unknown required vocabularies are rejected with exit code 2 and distinct error types detectable via errors.As, optional unknown vocabularies are ignored, unlisted standard vocabulary keywords become annotation-only, and format-assertion requirement is recorded via FormatAssertionRequested(). Spec-derived test coverage (library and CLI) with -race concurrency check passes. Two low tester findings: unreferenced $defs vocabularies are not checked (edge case, no validation happens there anyway), and one malformed case is primarily unit-tested.

- internal/compile/vocabulary.go: new; ParseVocabulary reads and validates $vocabulary from meta-schemas, fails closed on any malformation, computes active vocabulary set per core 8.1.2
- internal/compile/vocabulary_apply.go: new; vocabAt and vocabFor determine the active vocabulary set per schema resource, allowing embedded resources to inherit or override
- internal/compile/vocabulary_errors.go: new; UnknownVocabularyError, MalformedVocabularyError, CoreVocabularyError, UnsupportedSchemaError with spec-compliant messages
- internal/compile/compile.go: Node gets FormatAssertion bool field, compile() calls vocabAt() and routes inactive-vocabulary keywords to annotation-only path
- internal/compile/metaschema.go: otherDraft() regex detects non-2020-12 drafts by name, ResolveMetaSchema returns UnsupportedSchemaError instead of ErrUnsupportedSchema
- internal/compile/refs.go: shared struct tracks formatAssertion across resources, resolver caches computed vocabulary sets in vocabs map, root.FormatAssertion is set before return
- internal/keywords/vocabularies.go: new; StandardVocabularies() and IsStandardVocabulary() expose the eight standard 2020-12 URIs including format-assertion
- internal/keywords/core/schema.go: new; registers $schema as a silent (no-op) core keyword
- internal/keywords/core/vocabulary.go: new; registers $vocabulary as a silent core keyword
- internal/keywords/core/schema_keywords_test.go: new; verifies $schema, $vocabulary, $comment produce no annotations in passing basic reports
- vocabulary.go: new root package; FormatAssertionRequested() read-only accessor on Schema
- vocabulary_errors.go: new root package; type aliases for exported error types
- vocabulary_test.go, vocabulary_apply_test.go, vocabulary_ref_test.go, vocabulary_race_test.go: new library tests covering all ACs, embedded resources, error types, concurrency
- cmd/jsonschema/vocabulary_cli_test.go: new; CLI tests with --ref registry flags, verifying exit codes 0/1/2 and error message content
- internal/keywords/vocabularies_test.go: new unit tests for vocabulary recognition
- internal/compile/vocabulary_test.go: new unit tests for ParseVocabulary (all $vocabulary cases)
- testdata/vocabulary/*.json and testdata/vocabulary/metaschemas/*.json: new test data (case files and custom meta-schemas for registry testing)
- Modified: internal/compile/metaschema_test.go (added comment clarifying unsupported-draft message intentionally names the draft), cmd/jsonschema/metaschema_cli_test.go (updated draft-07 assertion), testdata/metaschema/schema_keyword.json (updated draft-07 error message), spec_test.go (skip vocabulary testdata folder like metaschema)

## How to verify

- Run: go test -race ./... and verify all vocabulary tests pass (vocabulary_test.go, vocabulary_apply_test.go, vocabulary_ref_test.go, vocabulary_race_test.go, cmd/jsonschema/vocabulary_cli_test.go, internal/compile/vocabulary_test.go, internal/keywords/vocabularies_test.go, schema_keywords_test.go)
- Verify AC 1: compile a schema with $schema pointing to a custom meta-schema in the registry; its vocabulary set is read and applied
- Verify AC 2: compile a schema whose meta-schema requires an unknown vocabulary (e.g., https://example.com/vocab/money); expect UnknownVocabularyError with both the vocabulary URI and meta-schema URI in the message, exit code 2
- Verify AC 3: compile a schema whose meta-schema marks an unknown vocabulary as optional (false); expect success, schema loads
- Verify AC 4: compile a schema under a meta-schema listing only core and applicator; 'minimum: 5' is ignored for validation (passes 3), but appears as an annotation
- Verify AC 5: compile a schema under a meta-schema without $vocabulary; all eight standard vocabularies are active
- Verify AC 6: compile a schema whose meta-schema requires the format-assertion vocabulary; FormatAssertionRequested() returns true, format validation stays annotation-only
- Verify AC 7: compile a schema whose meta-schema omits core or marks it false; expect CoreVocabularyError, exit code 2
- Verify AC 8: compile a schema whose meta-schema has a malformed $vocabulary entry (non-boolean value or non-absolute-URI key); expect MalformedVocabularyError, exit code 2, message names the meta-schema and the bad entry
- Verify AC 9: compile a schema with an embedded resource having its own $schema; the outer resource validates with its vocabulary set, the inner with its own
- Verify AC 10: compile a schema with $schema: http://json-schema.org/draft-07/schema#; expect UnsupportedSchemaError with message 'unsupported $schema: draft-07; only 2020-12 is supported'
- Verify AC 11: compile a schema with $schema, $vocabulary and/or $comment; run Validate and output basic report; verify no annotations for /$schema, /$vocabulary or /$comment
- Verify AC 12: library tests run shared testdata/vocabulary/*.json case files; CLI tests use vocabulary_cli_test.go with registry flags; each case cites a core 8.1.* section; exit codes 0/1/2 are checked

## Migration

None. Schemas without $schema continue to use all standard vocabularies (unchanged). Schemas with $schema now correctly restrict vocabularies per the declared meta-schema (new correct behavior). Custom meta-schemas supplied through --ref or Registry are now honored for their $vocabulary declarations. All existing valid schemas remain valid.

## Rollback

Revert the commit. The new vocabulary implementation is isolated in new files (internal/compile/vocabulary*.go, internal/keywords/vocabularies.go, vocabulary*.go in root). Compile.go changes are surgical (Node.FormatAssertion field, vocabAt lookup, keyword routing logic). After revert: $schema, $vocabulary, $comment are treated as unknown keywords again (annotations visible in reports); unknown required vocabularies and non-2020-12 $schema are not rejected; FormatAssertionRequested() method does not exist.

## Known limitations

- Unreferenced $defs entries (not reached by any $ref) are never validated for vocabulary errors because they are never evaluated. This is by design: vocabulary incompatibility is only discovered if the resource is actually used. Minor edge case with no practical impact.
- The malformed $vocabulary case for relative URIs (e.g., 'vocab/money' instead of absolute) is covered by unit test in vocabulary_test.go (TestParseVocabulary, 'key not absolute' case) and is also end-to-end tested in vocabulary_test.go via CompileWithRegistry with a registry containing the malformed meta-schema.
- Format-assertion vocabulary is recognized and recorded by this story (FormatAssertionRequested() returns true), but format values are NOT enforced. Story 195 will implement the actual format assertion. Until then, 'format' is annotation-only regardless of the vocabulary requirement.

## Concerns

- AC 10: The error message for non-2020-12 $schema changed from a generic 'not supported' to a specific 'unsupported $schema: draft-07; only 2020-12 is supported'. This is intentional and helpful, but metaschema_cli_test.go was updated. Verify the new message is acceptable to stakeholders.
- AC 11: $schema, $vocabulary and $comment are now silent core keywords with no annotation output. This is a breaking change for any test, tool or workflow that expects these keywords in basic/detailed reports. All existing affected tests (metaschema_cli_test.go, schema_keyword.json) have been updated, but verify no downstream code depends on these annotations.
- Vocabulary sets are immutable after Compile. If a custom meta-schema is updated (e.g., removing the validation vocabulary), any schema referencing it must be recompiled to pick up the change. This is correct (immutable compiled schema), but verify alignment with your caching/recompilation strategy.
- Tester's two low findings (non-blockers): (1) Unreferenced $defs entries are not validated for vocabulary errors until referenced. (2) One malformed case (relative key) is primarily unit-tested. Both acceptable but worth reviewer awareness.
