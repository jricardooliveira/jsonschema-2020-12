# PLAN #13: Honour $vocabulary from embedded and registry meta-schemas

- Date: 2026-10-17
- Status: approved by the operator at Checkpoint 2 (run #151, 2026-10-17 01:06 UTC)
- Design: see `ADR-013-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `internal/compile/vocabulary.go`
- `internal/compile/vocabulary_test.go`
- `internal/keywords/vocabularies.go`
- `internal/keywords/vocabularies_test.go`
- `internal/compile/vocabulary_errors.go`
- `internal/compile/metaschema.go`
- `vocabulary_errors.go`
- `internal/compile/compile.go`
- `internal/compile/vocabulary_apply.go`
- `vocabulary.go`
- `internal/keywords/core/schema.go`
- `internal/keywords/core/vocabulary.go`
- `internal/keywords/core/comment.go`
- `internal/keywords/core/schema_keywords_test.go`
- `testdata/vocabulary/required_optional_unknown.json`
- `testdata/vocabulary/unlisted_vocabulary.json`
- `testdata/vocabulary/embedded_resource.json`
- `testdata/vocabulary/metaschemas/`
- `vocabulary_test.go`
- `cmd/jsonschema/vocabulary_cli_test.go`

Modules the architecture expects to touch:

- `internal/compile/vocabulary.go`
- `internal/compile/vocabulary_test.go`
- `internal/compile/vocabulary_apply.go`
- `internal/compile/vocabulary_errors.go`
- `internal/compile/compile.go`
- `internal/compile/metaschema.go`
- `internal/keywords/vocabularies.go`
- `internal/keywords/vocabularies_test.go`
- `internal/keywords/core/schema.go`
- `internal/keywords/core/vocabulary.go`
- `internal/keywords/core/comment.go`
- `internal/keywords/core/schema_keywords_test.go`
- `vocabulary.go`
- `vocabulary_errors.go`
- `vocabulary_test.go`
- `cmd/jsonschema/vocabulary_cli_test.go`
- `testdata/vocabulary/required_optional_unknown.json`
- `testdata/vocabulary/unlisted_vocabulary.json`
- `testdata/vocabulary/embedded_resource.json`
- `testdata/vocabulary/metaschemas/core_applicator_only.json`
- `testdata/vocabulary/metaschemas/unknown_required.json`
- `testdata/vocabulary/metaschemas/unknown_optional.json`
- `testdata/vocabulary/metaschemas/no_core.json`
- `testdata/vocabulary/metaschemas/malformed.json`
- `testdata/vocabulary/metaschemas/format_assertion_required.json`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — Parse and check $vocabulary of a meta-schema
2. **T4** — Register $schema, $vocabulary and $comment as silent core keywords
3. **T2** — Distinct errors for unknown vocabulary and unsupported $schema  _(after T1)_
4. **T3** — Apply the active vocabulary set per schema resource during compilation  _(after T1, T2)_
5. **T5** — Spec-derived library and CLI tests for vocabularies  _(after T2, T3, T4)_

## Risks

- Routing keywords from inactive vocabularies to unknown-keyword handling changes the compile.go keyword loop, which is a crowded file.
- The official meta-schema validates itself with the meta-data, format-annotation and content vocabularies. All of them must be recognised, or the embedded meta-schemas would refuse themselves.
- Story 189 may already reject non-2020-12 '$schema' with a different message. Changing that message could break metaschema_cli_test.go.

Implementation constraints the coder must honor:

- Standard library only; go.mod is unchanged.
- Never fetch meta-schemas: resolve only from the embedded set and the caller-supplied registry.
- The compiled Schema stays immutable, and the vocabulary sets are not changed after Compile.
- Do not edit internal/keywords/registry.go or internal/keywords/all/all.go. Recognition reads the existing registry data from a new file.
- Error messages never quote instance values.
- Tests cite 2020-12 core sections and use no network.

- DB impact: no · API impact: yes: new exported error type and a read-only Schema accessor for format-assertion request · Migration needed: no

Breaking changes:

- Basic and detailed reports no longer contain '$schema', '$vocabulary' or '$comment' annotations. Existing tests that expected them would need updating.

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: go test ./internal/compile/ ./internal/keywords/ passes with tables for each '$vocabulary' case.
- **T4**: A basic report for the AC 11 schema and instance contains no '/$schema', '/$vocabulary' or '/$comment' annotation.
- **T2**: errors.As finds the vocabulary error type on Compile's error, and the draft-07 test gives an 'unsupported $schema' message.
- **T3**: Library tests show 'minimum' ignored under a meta-schema without validation, embedded resources following their own set, and the format-assertion flag recorded.
- **T5**: go test -race ./... passes, and every case cites a 2020-12 core section.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
