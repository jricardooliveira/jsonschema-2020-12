# ADR #12: Embedded 2020-12 meta-schemas and meta-schema validation of every loaded schema

- Date: 2026-10-17
- Story: #12
- Status: approved by the operator at Checkpoint 2 (run #148, 2026-10-17 00:11 UTC)

## Decision
The design is unchanged from the previous proposal. The only change is that the folder entry testdata/metaschema/cli/ is replaced by the seven individual fixture files the CLI tests use. Embedded meta-schemas live in internal/registry/metaschemas and are loaded through embedded.go with go:embed. registry.go gains one hook so it indexes them under their $id and enforces the built-in-address rule. internal/compile/metaschema.go resolves '$schema'. internal/compile/metacheck.go runs the existing evaluator (through its public entry point, without editing internal/eval/eval.go) against the chosen meta-schema, with format assertion forced off. compile.go calls the check once, after resolving references and before returning the Schema. invalid_schema_error.go defines the root-level InvalidSchemaError type. The CLI already maps every load error to exit code 2 and stderr, so cmd/jsonschema/main.go is not edited and only a new test file is added.

## Affected modules
- internal/registry/metaschemas/schema.json
- internal/registry/metaschemas/meta/core.json
- internal/registry/metaschemas/meta/applicator.json
- internal/registry/metaschemas/meta/unevaluated.json
- internal/registry/metaschemas/meta/validation.json
- internal/registry/metaschemas/meta/meta-data.json
- internal/registry/metaschemas/meta/format-annotation.json
- internal/registry/metaschemas/meta/format-assertion.json
- internal/registry/metaschemas/meta/content.json
- internal/registry/embedded.go
- internal/registry/embedded_test.go
- internal/registry/registry.go
- internal/compile/metaschema.go
- internal/compile/metaschema_test.go
- internal/compile/metacheck.go
- internal/compile/metacheck_test.go
- internal/compile/compile.go
- invalid_schema_error.go
- metaschema_test.go
- testdata/metaschema/schema_keyword.json
- testdata/metaschema/invalid_schemas.json
- testdata/metaschema/custom_metaschema.json
- testdata/metaschema/builtin_address.json
- cmd/jsonschema/metaschema_cli_test.go
- testdata/metaschema/cli/valid_schema.json
- testdata/metaschema/cli/invalid_schema.json
- testdata/metaschema/cli/draft07_schema.json
- testdata/metaschema/cli/custom_meta.json
- testdata/metaschema/cli/custom_meta_schema.json
- testdata/metaschema/cli/builtin_address_mismatch.json
- testdata/metaschema/cli/document.json

## Data / API impact
- DB impact: no
- API impact: yes: new exported InvalidSchemaError type, and Compile now fails for schemas that fail the meta-schema check
- Migration needed: no

## Constraints for implementation
- The meta-schema JSON files must be the official 2020-12 meta-schema texts, recreated from the specification. Nothing may be downloaded at build or test time.
- Do not edit internal/eval/eval.go or cmd/jsonschema/main.go (both crowded). Use the existing evaluator entry point and the CLI's existing error path.
- The meta-schema check never asserts format, whatever the compile options say.
- Error messages carry no schema or instance values.
- The compiled Schema stays immutable, and the meta-schema check runs entirely before Compile returns.
- Each test cites its 2020-12 core section. No network is used.

## Risks
- If the existing evaluator cannot run with format assertion forced off independently of the caller's options, a small change to internal/eval would be needed, and that is a crowded file.
- Whether the CLI stderr message includes the schema file name depends on how main.go currently wraps load errors.
- The embedded meta-schema text must match the official $id and $dynamicAnchor layout exactly, or $dynamicRef '#meta' will not reach nested subschemas.

## Breaking changes
- Schemas that were previously accepted but are invalid against the 2020-12 meta-schema now fail to compile. Existing tests that compile such schemas would break.

## Sensitivity
- security
