# ADR #13: Honour $vocabulary from embedded and registry meta-schemas

- Date: 2026-10-17
- Story: #13
- Status: approved by the operator at Checkpoint 2 (run #151, 2026-10-17 01:06 UTC)

## Decision
New vocabulary logic lives in new files. internal/compile/vocabulary.go holds parsing and checks. internal/compile/vocabulary_apply.go holds the per-resource set and decides whether a keyword is active. internal/compile/vocabulary_errors.go holds the error types. Recognition of vocabularies sits in a new file, internal/keywords/vocabularies.go, which reads the vocabulary URIs that keywords already give the registry and adds the format-assertion URI, which has no keyword yet. registry.go is not edited. Crowded file compile.go is touched only to call the vocabulary step when a resource with '$schema' is entered, and to route inactive-vocabulary keywords to the existing unknown-keyword handling. The root package gets two new files: vocabulary_errors.go (exported error type alias or wrapper) and vocabulary.go (read-only accessor reporting format assertion was requested). jsonschema.go is not edited. '$schema' and '$vocabulary' get no-op keyword files in internal/keywords/core. Compile keeps reading '$schema' itself as before.

## Affected modules
- internal/compile/vocabulary.go
- internal/compile/vocabulary_test.go
- internal/compile/vocabulary_apply.go
- internal/compile/vocabulary_errors.go
- internal/compile/compile.go
- internal/compile/metaschema.go
- internal/keywords/vocabularies.go
- internal/keywords/vocabularies_test.go
- internal/keywords/core/schema.go
- internal/keywords/core/vocabulary.go
- internal/keywords/core/comment.go
- internal/keywords/core/schema_keywords_test.go
- vocabulary.go
- vocabulary_errors.go
- vocabulary_test.go
- cmd/jsonschema/vocabulary_cli_test.go
- testdata/vocabulary/required_optional_unknown.json
- testdata/vocabulary/unlisted_vocabulary.json
- testdata/vocabulary/embedded_resource.json
- testdata/vocabulary/metaschemas/core_applicator_only.json
- testdata/vocabulary/metaschemas/unknown_required.json
- testdata/vocabulary/metaschemas/unknown_optional.json
- testdata/vocabulary/metaschemas/no_core.json
- testdata/vocabulary/metaschemas/malformed.json
- testdata/vocabulary/metaschemas/format_assertion_required.json

## Data / API impact
- DB impact: no
- API impact: yes: new exported error type and a read-only Schema accessor for format-assertion request
- Migration needed: no

## Constraints for implementation
- Standard library only; go.mod is unchanged.
- Never fetch meta-schemas: resolve only from the embedded set and the caller-supplied registry.
- The compiled Schema stays immutable, and the vocabulary sets are not changed after Compile.
- Do not edit internal/keywords/registry.go or internal/keywords/all/all.go. Recognition reads the existing registry data from a new file.
- Error messages never quote instance values.
- Tests cite 2020-12 core sections and use no network.

## Risks
- Routing keywords from inactive vocabularies to unknown-keyword handling changes the compile.go keyword loop, which is a crowded file.
- The official meta-schema validates itself with the meta-data, format-annotation and content vocabularies. All of them must be recognised, or the embedded meta-schemas would refuse themselves.
- Story 189 may already reject non-2020-12 '$schema' with a different message. Changing that message could break metaschema_cli_test.go.

## Breaking changes
- Basic and detailed reports no longer contain '$schema', '$vocabulary' or '$comment' annotations. Existing tests that expected them would need updating.

## Sensitivity
- security
