# SPEC #13: Honour $vocabulary from embedded and registry meta-schemas

- Date: 2026-10-17
- Type: feature
- Verdict: pass
- Status: proposed (pending Checkpoint 1 sign-off)

## Problem

The validator applies every known keyword regardless of what the schema's meta-schema declares in '$vocabulary'. An unknown required vocabulary is not refused. A custom meta-schema cannot switch off a standard vocabulary. '$schema', '$vocabulary' and '$comment' are treated as unknown keywords and show up as annotations, for example keywordLocation '/$schema' in a passing basic report.

## Why this matters

2020-12 core makes '$vocabulary' processing mandatory. The brief's mandatory scope requires that an unknown required vocabulary is an error (exit 2) and that only 2020-12 is supported. Story 195 needs the format-assertion vocabulary to be recognised and recorded so it only has to switch format checking on.

## Acceptance criteria

- When a schema's '$schema' names a meta-schema, from the embedded 2020-12 set or supplied through the registry, the validator reads that meta-schema's '$vocabulary' and fixes the set of active vocabularies for that schema resource at load time.
- A required ('true') vocabulary the validator does not recognise refuses the schema with exit code 2. The message names both the vocabulary URI and the meta-schema that required it, for example 'meta-schema https://example.com/my-meta requires unknown vocabulary https://example.com/vocab/money'. The library returns the same facts as a distinct exported error type that callers can detect with errors.As.
- An optional ('false') vocabulary the validator does not recognise is ignored. The schema loads, and documents are checked normally.
- Keywords that belong to a standard vocabulary which the meta-schema does not list are treated as unknown keywords. They have no validation effect and are collected as annotations. For example, under a meta-schema that lists core and applicator but not validation, 'minimum: 5' does not reject 3.
- A meta-schema without a '$vocabulary' entry activates all standard 2020-12 vocabularies, so the schema behaves exactly as under the official 2020-12 meta-schema.
- The format-assertion vocabulary (https://json-schema.org/draft/2020-12/vocab/format-assertion) is recognised. When it is required, the schema loads without error and the loaded schema records that format assertion was requested, readable through the library. 'format' stays annotation-only in this story.
- A meta-schema whose '$vocabulary' omits the core vocabulary, or marks it 'false', refuses the schema with a clear message and exit code 2.
- A malformed '$vocabulary' refuses the schema with exit code 2 and a message naming the meta-schema and the bad entry. Malformed means a value that is not a boolean, a key that is not an absolute URI, or '$vocabulary' that is not an object.
- An embedded schema resource with its own '$id' and '$schema' uses the vocabulary set of its own meta-schema. The outer resource keeps its own set. Example: the outer schema checks 'minimum' while the inner one, under a meta-schema without validation, does not.
- A '$schema' naming another draft's meta-schema, for example 'http://json-schema.org/draft-07/schema#', refuses the schema with exit code 2 and a message such as 'unsupported $schema: draft-07; only 2020-12 is supported'. The library returns the same facts as an error.
- '$schema', '$vocabulary' and '$comment' are registered core keywords with no validation effect and produce no annotation. A passing basic report for {"$schema":"https://json-schema.org/draft/2020-12/schema","type":"string"} against "x" contains no '$schema' annotation.
- Spec-derived tests, each citing the 2020-12 core section (8.1.2 and 8.1.1), cover required known, required unknown, optional unknown, unlisted standard vocabulary, missing '$vocabulary', missing or optional core, malformed entries, embedded-resource vocabularies, format-assertion recording and draft-07 refusal. They run through both the library and the CLI and check exit codes 0, 1 and 2.

## Non-goals

- Actually asserting 'format' values when the format-assertion vocabulary is required (story 195).
- Supporting any draft other than 2020-12.
- Fetching meta-schemas from the network or from disk paths the caller did not list.
- Adding new vocabularies, or a plug-in mechanism for third-party vocabularies.
- Changing the output format shapes.

## Tasks

- **T1** — Parse and check $vocabulary of a meta-schema
  - purpose: Read a meta-schema's '$vocabulary', check it, and compute the active vocabulary set. The checks are: an object, absolute-URI keys and boolean values; core present and required; unknown required vocabularies refused; unknown optional vocabularies dropped. With no '$vocabulary', every standard vocabulary is active. The set of recognised vocabularies is derived from the keyword registry's vocabulary URIs plus format-assertion.
  - scope: internal/compile/vocabulary.go, internal/compile/vocabulary_test.go, internal/keywords/vocabularies.go, internal/keywords/vocabularies_test.go
  - done when: go test ./internal/compile/ ./internal/keywords/ passes with tables for each '$vocabulary' case.
  - delivers: AC1, AC3, AC5, AC7, AC8
- **T2** — Distinct errors for unknown vocabulary and unsupported $schema
  - purpose: Add an exported error type carrying the vocabulary URI and the meta-schema URI, plus errors for a malformed '$vocabulary' and an unsupported draft. Refuse non-2020-12 '$schema' values such as draft-07 by name. All of these surface as schema errors that the CLI maps to exit code 2.
  - scope: internal/compile/vocabulary_errors.go, internal/compile/metaschema.go, vocabulary_errors.go
  - done when: errors.As finds the vocabulary error type on Compile's error, and the draft-07 test gives an 'unsupported $schema' message.
  - depends on: T1
  - delivers: AC2, AC10
- **T3** — Apply the active vocabulary set per schema resource during compilation
  - purpose: Fix the active set for each schema resource, so an embedded resource with its own '$schema' uses its own set. Keywords from inactive standard vocabularies compile as unknown keywords, so they become annotation-only. Record on the loaded schema whether format-assertion was required, and expose it through a read-only accessor in the root package.
  - scope: internal/compile/compile.go, internal/compile/vocabulary_apply.go, vocabulary.go
  - done when: Library tests show 'minimum' ignored under a meta-schema without validation, embedded resources following their own set, and the format-assertion flag recorded.
  - depends on: T1, T2
  - delivers: AC4, AC6, AC9
- **T4** — Register $schema, $vocabulary and $comment as silent core keywords
  - purpose: Register '$schema' and '$vocabulary' as core keywords whose evaluation does nothing and emits no annotation. Make sure '$comment' does the same. They then no longer fall through to the unknown-keyword annotation path.
  - scope: internal/keywords/core/schema.go, internal/keywords/core/vocabulary.go, internal/keywords/core/comment.go, internal/keywords/core/schema_keywords_test.go
  - done when: A basic report for the AC 11 schema and instance contains no '/$schema', '/$vocabulary' or '/$comment' annotation.
  - delivers: AC11
- **T5** — Spec-derived library and CLI tests for vocabularies
  - purpose: Add JSON test cases under testdata/vocabulary/ with custom meta-schemas supplied through the registry, a library test file that runs them, and CLI contract tests. The CLI tests use registry flags and assert exit codes 0, 1 and 2 and the error message content.
  - scope: testdata/vocabulary/required_optional_unknown.json, testdata/vocabulary/unlisted_vocabulary.json, testdata/vocabulary/embedded_resource.json, testdata/vocabulary/metaschemas/, vocabulary_test.go, cmd/jsonschema/vocabulary_cli_test.go
  - done when: go test -race ./... passes, and every case cites a 2020-12 core section.
  - depends on: T2, T3, T4
  - delivers: AC2, AC3, AC4, AC5, AC6, AC7, AC8, AC9, AC10, AC11, AC12
