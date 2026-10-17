# Release notes — #12: Embedded 2020-12 meta-schemas and meta-schema validation of every loaded schema

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #12
- Status: proposed (pending Checkpoint 3 sign-off)

**⚠️ the release-agent has concerns.**

## What changed

The nine official 2020-12 meta-schemas are embedded in the binary and indexed by their $id. Every loaded schema (main and reached registry schemas) is validated against its meta-schema at load time: 2020-12 by default, or an unsupported-version error for other drafts, or a custom meta-schema chain that must end in 2020-12. Failures are reported as a new InvalidSchemaError type, with the schema location and keyword but no values, and exit code 2 — never as a validation result. The implementation uses $dynamicRef '#meta' to check nested subschemas, applies the check only to reached documents (not unreferenced supplied schemas), protects built-in addresses (equal allowed, different refused), and never asserts format regardless of compile options.

- internal/registry/metaschemas/{schema.json, meta/*.json}: Nine official 2020-12 meta-schema documents, embedded with go:embed
- internal/registry/embedded.go: Loads the embedded meta-schemas by $id, offline, into a shared read-only map
- internal/registry/registry.go: AddBuiltins() indexes the embedded meta-schemas and protects their addresses (accepts equal supplied file, refuses different one)
- internal/compile/metaschema.go: ResolveMetaSchema() normalizes '$schema' and picks the meta-schema URI (2020-12 default, unsupported-draft error, custom-chain validation to 2020-12)
- internal/compile/metacheck.go: MetaChecks() builds evaluations to validate each loaded schema against its meta-schema, including chains
- invalid_schema_error.go: New InvalidSchemaError type (Source, SchemaLocation as JSON Pointer, Keyword, message) with no schema or instance values
- jsonschema.go compileChecked() and metaCheck(): Call MetaChecks at load time after compilation succeeds; map failures to InvalidSchemaError
- cmd/jsonschema/metaschema_cli_test.go: Four CLI contract tests (invalid schema, unsupported draft, custom meta-schema failure, built-in address mismatch) assert exit 2, empty stdout, value-free stderr with file name and rule
- metaschema_test.go: Library tests with JSON case files covering '$schema' keyword, invalid schemas, custom meta-schemas, built-in addresses, each citing spec sections
- spec_test.go: Updated to skip (log) test groups whose schema fails meta-validation, since such a schema cannot be compiled

## How to verify

- Run `go test ./...` — the build, vet, race and all tests pass
- Compile a schema missing a required 'title' in a custom meta-schema: Compile returns InvalidSchemaError with location '/title', not a validation result
- Compile a schema with '$schema': 'http://json-schema.org/draft-07/schema#': Compile returns a plain error saying the version is not supported, not InvalidSchemaError
- Supply a built-in meta-schema address with different content via the registry: Compile returns an error naming the address and the file, quoting no values
- Compile a schema with a nested subschema under properties → items → allOf that is invalid: InvalidSchemaError points to that location via $dynamicRef traversal
- In the CLI, run `jsonschema validate invalid.json doc.json`: exit code 2, stdout empty, stderr says the schema file, location in schema and rule, no schema values
- Compile a schema that was previously accepted but is invalid against the 2020-12 meta-schema (e.g. `type: 'nope'`): now fails with InvalidSchemaError instead of compiling

## Migration

Schemas that previously compiled but are structurally invalid against the 2020-12 meta-schema (for example, 'type' with an invalid value, 'minLength' with a negative number, or 'enum' not an array) now fail at Compile time with InvalidSchemaError, exit code 2. Scripts using those schemas must either fix them or accept the new error. The default meta-schema is now enforced: schemas with no '$schema' are checked as 2020-12. Schemas declaring a different draft (e.g. draft-07) are rejected as unsupported at load time. Custom meta-schemas supplied through the registry are themselves validated, and their chain must end in the official 2020-12 meta-schema.

## Rollback

This change cannot be cleanly rolled back: it is a structural requirement of the public API (Compile now validates schemas against their meta-schema). Reverting the change would break any code that relies on InvalidSchemaError being returned, and schemas that currently fail would compile again (likely hiding bugs). The only practical rollback is a prior release.

## Known limitations

- Format assertion mode (format-assertion option) does not yet exist, so the requirement that meta-schema checking never asserts format (AC7) cannot be tested. When format-assertion is added in a later story, this gap will be closed.
- The '$vocabulary' keyword in a schema is not interpreted; it is only validated as a structure. Enforcing required vocabularies is story 190.
- Only 2020-12 is supported; other drafts are rejected at load time.

## Concerns

- AC7 (format assertion never asserts in meta-schema check) is not proven because format-assertion mode does not yet exist. This is a prerequisite gap: the feature must ship in a later story, and then this AC must be tested by verifying a schema with format errors still compiles.
- When a reached registry schema's compilation fails (not meta-schema validation, but parsing or reference errors), the error is a plain error that may echo schema content. The error safety boundary for schema values applies to InvalidSchemaError only. Verify that ResolveMetaSchema and MetaChecks never echo schema values in their error messages (they don't: both say 'not supported' and 'does not end in 2020-12', never quoting URIs). Plain compilation errors from other parts of the pipeline are outside this story's value-free contract.
- spec_test.go now silently skips test groups whose schema fails meta-validation. This weakens coverage of those spec sections (which can no longer be tested, since the schema does not compile). This is correct and expected: a schema that fails meta-validation is a load error, not a validation scenario. However, review existing test results to confirm no previously-passing spec group coverage is lost to an unexpected error.
