# Release notes — #13: Honour $vocabulary from embedded and registry meta-schemas

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #13
- Status: proposed (pending Checkpoint 3 sign-off)

**⚠️ the release-agent has concerns.**

## What changed

Schemas with '$vocabulary' declarations are now properly processed per JSON Schema 2020-12. Unknown required vocabularies are rejected with exit code 2 and a clear error. Keywords from vocabularies a meta-schema does not list are treated as unknown. The format-assertion vocabulary is recognized and recorded for use by story 195. `$schema`, `$vocabulary` and `$comment` are now silent core keywords that produce no output annotations.

- internal/compile/vocabulary.go: new file — ParseVocabulary reads and validates a meta-schema's '$vocabulary' entry, fail-closed
- internal/compile/vocabulary_apply.go: new file — per-resource vocabulary set computed at compile time; inactive-vocabulary keywords routed to unknown-keyword handling
- internal/compile/vocabulary_errors.go: new file — UnknownVocabularyError, MalformedVocabularyError, CoreVocabularyError, UnsupportedSchemaError error types
- internal/compile/metaschema.go: recognize and reject draft-07 and other non-2020-12 $schema URIs by name; UnsupportedSchemaError returned for draft-07 with message 'unsupported $schema: draft-07; only 2020-12 is supported'
- internal/compile/refs.go: hold formatAssertion flag in shared state; initialize VocabularySet map on resolver; propagate flag to root node
- internal/compile/compile.go: Node.FormatAssertion field added to root node only
- internal/keywords/vocabularies.go: new file — define eight standard 2020-12 vocabulary URIs; StandardVocabularies() and IsStandardVocabulary() helpers
- internal/keywords/core/schema.go, vocabulary.go: new files — register $schema and $vocabulary as no-op core keywords that compile and evaluate to no effect and emit no annotation
- vocabulary.go: new file — Schema.FormatAssertionRequested() accessor reports whether any resource's meta-schema listed format-assertion as required
- vocabulary_errors.go: new file — exported error type aliases (UnknownVocabularyError, MalformedVocabularyError, CoreVocabularyError, UnsupportedSchemaError) for use with errors.As
- testdata/vocabulary/: new directory — JSON test cases for required/optional/unknown vocabularies, unlisted standard vocabularies, embedded resources with their own $schema, and draft-07 rejection
- testdata/metaschema/schema_keyword.json: draft-07 test case updated to expect specific 'unsupported $schema: draft-07' message instead of generic 'not supported'
- cmd/jsonschema/metaschema_cli_test.go: draft-07 error message updated to match new UnsupportedSchemaError output; removed assertion that error must not mention draft-07 (AC10 now requires the draft name)

## How to verify

- Compile a schema with a custom meta-schema listing only 'core' and 'applicator' vocabularies; validate a document with 'minimum: 5' in the schema — the minimum constraint is ignored (treated as unknown keyword)
- Compile a schema with a custom meta-schema that requires an unknown vocabulary — schema compilation fails with exit code 2 and error names both the vocabulary URI and the meta-schema URI
- Compile a schema with 'type': 'string' and '$schema': 'http://json-schema.org/draft-07/schema#' — compilation fails with exit code 2 and error message 'unsupported $schema: draft-07; only 2020-12 is supported'
- Compile a schema with '$schema' of 2020-12 and validate a document — the basic report contains no '/$schema' annotation (previously it would have)
- Compile a schema with a meta-schema that lists format-assertion as required — Schema.FormatAssertionRequested() returns true; validate a document with a malformed format value — validation passes (format still annotation-only)
- Compile a schema with a custom meta-schema that has a malformed '$vocabulary' entry (for example a value that is not a boolean, or a key that is not an absolute URI) — compilation fails with exit code 2 and error message names the meta-schema and the bad entry

## Migration

No schema data or external resources to migrate. Vocabulary processing is applied at compile time. Existing schemas continue to work: those without a declared '$schema' or '$vocabulary' are compiled and validated exactly as before. Custom meta-schemas must be supplied through the registry (CompileWithRegistry or --ref flag); they are never fetched from the network or from disk paths outside the registry.

## Rollback

Revert the commit. This change is applied during schema compilation only; there is no persisted state or database impact to undo. Schemas compiled under this version continue to work under a reverted version (they are immutable objects).

## Known limitations

- Only 2020-12 is supported. Schemas declaring draft-07, draft-06, draft-4, or any other version are rejected with exit code 2.
- Format assertion is recorded but not asserted; 'format' stays annotation-only. Story 195 will switch assertion on.
- No network or disk reads beyond the registry supplied by the caller. Meta-schemas must be in the embedded set or the registry; they are never fetched.
- Vocabulary error types exported for use with errors.As; however, errors wrapped through a $ref-resolved registry resource may not unwrap correctly (the error is still reported and fails with exit code 2, but errors.As may not find the underlying type).

## Concerns

- Breaking change: basic and detailed reports no longer contain '$schema', '$vocabulary' or '$comment' annotations. Tests and evaluators that expected these annotations will need updating.
- The change routes inactive-vocabulary keywords through the existing unknown-keyword path in compile.go, a crowded file. The edit is surgical (four lines in the keyword loop), but future changes to that loop must be aware of the new condition.
- Error message for malformed '$vocabulary' entries (the bound entry key) is asserted only at unit test level, not in CLI integration tests (F3). The message is generated and used correctly, but large entry keys longer than 80 characters are silently truncated in the error message.
