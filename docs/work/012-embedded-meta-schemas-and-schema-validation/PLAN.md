# PLAN #12: Embedded 2020-12 meta-schemas and meta-schema validation of every loaded schema

- Date: 2026-10-17
- Status: proposed (pending Checkpoint 2 sign-off)
- Design: see `ADR-012-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `internal/registry/metaschemas/schema.json`
- `internal/registry/metaschemas/meta/core.json`
- `internal/registry/metaschemas/meta/applicator.json`
- `internal/registry/metaschemas/meta/unevaluated.json`
- `internal/registry/metaschemas/meta/validation.json`
- `internal/registry/metaschemas/meta/meta-data.json`
- `internal/registry/metaschemas/meta/format-annotation.json`
- `internal/registry/metaschemas/meta/format-assertion.json`
- `internal/registry/metaschemas/meta/content.json`
- `internal/registry/embedded.go`
- `internal/registry/embedded_test.go`
- `internal/registry/registry.go`
- `internal/compile/metaschema.go`
- `internal/compile/metaschema_test.go`
- `internal/compile/metacheck.go`
- `internal/compile/metacheck_test.go`
- `internal/compile/compile.go`
- `invalid_schema_error.go`
- `metaschema_test.go`
- `testdata/metaschema/schema_keyword.json`
- `testdata/metaschema/invalid_schemas.json`
- `testdata/metaschema/custom_metaschema.json`
- `testdata/metaschema/builtin_address.json`
- `cmd/jsonschema/metaschema_cli_test.go`
- `testdata/metaschema/cli/valid_schema.json`
- `testdata/metaschema/cli/invalid_schema.json`
- `testdata/metaschema/cli/draft07_schema.json`
- `testdata/metaschema/cli/custom_meta.json`
- `testdata/metaschema/cli/custom_meta_schema.json`
- `testdata/metaschema/cli/builtin_address_mismatch.json`
- `testdata/metaschema/cli/document.json`

Modules the architecture expects to touch:

- `internal/registry/metaschemas/schema.json`
- `internal/registry/metaschemas/meta/core.json`
- `internal/registry/metaschemas/meta/applicator.json`
- `internal/registry/metaschemas/meta/unevaluated.json`
- `internal/registry/metaschemas/meta/validation.json`
- `internal/registry/metaschemas/meta/meta-data.json`
- `internal/registry/metaschemas/meta/format-annotation.json`
- `internal/registry/metaschemas/meta/format-assertion.json`
- `internal/registry/metaschemas/meta/content.json`
- `internal/registry/embedded.go`
- `internal/registry/embedded_test.go`
- `internal/registry/registry.go`
- `internal/compile/metaschema.go`
- `internal/compile/metaschema_test.go`
- `internal/compile/metacheck.go`
- `internal/compile/metacheck_test.go`
- `internal/compile/compile.go`
- `invalid_schema_error.go`
- `metaschema_test.go`
- `testdata/metaschema/schema_keyword.json`
- `testdata/metaschema/invalid_schemas.json`
- `testdata/metaschema/custom_metaschema.json`
- `testdata/metaschema/builtin_address.json`
- `cmd/jsonschema/metaschema_cli_test.go`
- `testdata/metaschema/cli/valid_schema.json`
- `testdata/metaschema/cli/invalid_schema.json`
- `testdata/metaschema/cli/draft07_schema.json`
- `testdata/metaschema/cli/custom_meta.json`
- `testdata/metaschema/cli/custom_meta_schema.json`
- `testdata/metaschema/cli/builtin_address_mismatch.json`
- `testdata/metaschema/cli/document.json`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — Embed the official meta-schemas and index them in the registry
2. **T2** — Resolve '$schema' to a meta-schema  _(after T1)_
3. **T3** — Validate loaded schemas against their meta-schema at load time  _(after T1, T2)_
4. **T4** — Library tests with spec-derived valid and invalid schemas  _(after T3)_
5. **T5** — CLI contract tests for invalid schemas  _(after T3)_

## Risks

- If the existing evaluator cannot run with format assertion forced off independently of the caller's options, a small change to internal/eval would be needed, and that is a crowded file.
- Whether the CLI stderr message includes the schema file name depends on how main.go currently wraps load errors.
- The embedded meta-schema text must match the official $id and $dynamicAnchor layout exactly, or $dynamicRef '#meta' will not reach nested subschemas.

Implementation constraints the coder must honor:

- The meta-schema JSON files must be the official 2020-12 meta-schema texts, recreated from the specification. Nothing may be downloaded at build or test time.
- Do not edit internal/eval/eval.go or cmd/jsonschema/main.go (both crowded). Use the existing evaluator entry point and the CLI's existing error path.
- The meta-schema check never asserts format, whatever the compile options say.
- Error messages carry no schema or instance values.
- The compiled Schema stays immutable, and the meta-schema check runs entirely before Compile returns.
- Each test cites its 2020-12 core section. No network is used.

- DB impact: no · API impact: yes: new exported InvalidSchemaError type, and Compile now fails for schemas that fail the meta-schema check · Migration needed: no

Breaking changes:

- Schemas that were previously accepted but are invalid against the 2020-12 meta-schema now fail to compile. Existing tests that compile such schemas would break.

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: go test ./internal/registry/... passes. The tests show that every built-in URI resolves offline, that an equal supplied file is accepted, and that a different supplied file at a built-in address is an error naming the address and the file.
- **T2**: Unit tests cover the absent case, the 2020-12 URI with and without '#', draft-07 rejected, a custom meta-schema chain ending in 2020-12 accepted, and a chain that does not end in 2020-12 rejected.
- **T3**: Unit tests show that a nested invalid subschema is caught through $dynamicRef '#meta', that an unreferenced invalid registry schema is ignored, that the result is the same with format mode on or off, and that the error carries location and rule but no values.
- **T4**: go test -race ./... passes. Every case cites its core spec section.
- **T5**: The CLI tests pass. For invalid-schema cases they assert the exit code, empty stdout and a stderr message that does not contain the schema values. For the valid schema they assert a normal report.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
