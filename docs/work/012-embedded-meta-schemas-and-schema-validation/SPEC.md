# SPEC #12: Embedded 2020-12 meta-schemas and meta-schema validation of every loaded schema

- Date: 2026-10-17
- Type: feature
- Verdict: pass
- Status: passed gate 1 with nothing for the operator to decide (run #148, 2026-10-17 00:11 UTC)

## Problem

Schemas are compiled without being checked against the 2020-12 meta-schema, so a malformed schema can be used to check documents and give a misleading valid or invalid answer. '$schema' is not honoured, and the official meta-schemas are not available offline.

## Why this matters

The brief's mandatory scope item (6) says a schema that fails the meta-schema is an error with exit code 2, never 'invalid'. Item (3) requires honouring '$schema' and validating the schema against its meta-schema. The tool must work fully offline, so the official meta-schemas must be embedded in the binary.

## Acceptance criteria

- The official 2020-12 meta-schema (https://json-schema.org/draft/2020-12/schema) and its eight vocabulary meta-schemas (meta/core, meta/applicator, meta/unevaluated, meta/validation, meta/meta-data, meta/format-annotation, meta/format-assertion, meta/content) are embedded in the binary with go:embed. They are indexed in the registry under their official $id and resolve with no network or disk access.
- A schema with no '$schema', or with '$schema' equal to https://json-schema.org/draft/2020-12/schema after URI normalisation (an empty fragment '#' is accepted), is validated against the embedded 2020-12 meta-schema at load time, before any document is checked.
- A schema whose '$schema' names a different version, for example http://json-schema.org/draft-07/schema#, is a schema error that says the declared version is not supported (exit code 2). Nothing is checked.
- A schema whose '$schema' points to a custom meta-schema supplied through the registry is validated against that meta-schema. The custom meta-schema is itself validated against its own meta-schema, and that chain must end in the official 2020-12 meta-schema, otherwise it is a schema error (exit code 2). '$vocabulary' contents are not interpreted by this story.
- Meta-schema validation uses '$dynamicRef': '#meta', so every nested subschema of the loaded schema is checked, including subschemas under properties, items, $defs, allOf and similar applicators.
- Meta-schema checking applies to the main schema and to every registry schema the main schema actually reaches through its references, all at load time. A supplied schema that is never referenced is not meta-schema-checked and never causes a meta-schema error. Parse errors in any supplied file remain exit code 2, as story 187 defines.
- The meta-schema check never asserts 'format', whether or not format-assertion mode is on, so enabling format assertion never makes a previously accepted schema fail.
- A registry file that uses a built-in meta-schema address is accepted only when it is deep-JSON-equal to the embedded meta-schema, and the embedded one is then used. A different schema at that address is a schema error (exit code 2) that names the address and the file. Built-in meta-schemas can never be replaced.
- The library returns a distinct invalid-schema error type for meta-schema failures. It carries the schema location (JSON Pointer) and the failing keyword or rule, and it never includes schema or instance values. It is never reported as a validation result of 'invalid'.
- CLI: when the schema fails its meta-schema check, the exit code is 2, stdout is empty, and stderr holds one clear message. The message names the schema file, the location in the schema and the failing rule, without quoting values.
- Tests cover valid and invalid schemas through both the library and the CLI. They include: no $schema, the 2020-12 $schema with and without '#', an unsupported draft, a custom registry meta-schema, a nested invalid subschema reached through $dynamicRef, an unreferenced invalid registry schema, format mode on and off, and a built-in address with equal and with different content. Each case cites the 2020-12 core section it derives from. No test uses the network.

## Non-goals

- Interpreting '$vocabulary' (required or optional vocabularies, unknown required vocabulary), which belongs to story 190
- Honouring the format-assertion vocabulary declared by a meta-schema, which belongs to story 195
- Any draft other than 2020-12
- Fetching meta-schemas from the network or reading any unlisted file
- Changes to the output-format shapes

## Tasks

- **T1** — Embed the official meta-schemas and index them in the registry
  - purpose: Ship the nine official 2020-12 meta-schema files in the binary and make them resolvable by their $id. The built-in addresses are protected: an equal registry file is accepted, a different one is refused.
  - scope: internal/registry/metaschemas/schema.json, internal/registry/metaschemas/meta/core.json, internal/registry/metaschemas/meta/applicator.json, internal/registry/metaschemas/meta/unevaluated.json, internal/registry/metaschemas/meta/validation.json, internal/registry/metaschemas/meta/meta-data.json, internal/registry/metaschemas/meta/format-annotation.json, internal/registry/metaschemas/meta/format-assertion.json, internal/registry/metaschemas/meta/content.json, internal/registry/embedded.go, internal/registry/embedded_test.go, internal/registry/registry.go
  - done when: go test ./internal/registry/... passes. The tests show that every built-in URI resolves offline, that an equal supplied file is accepted, and that a different supplied file at a built-in address is an error naming the address and the file.
  - delivers: AC1, AC8
- **T2** — Resolve '$schema' to a meta-schema
  - purpose: Normalise '$schema' and choose the meta-schema to check against: the 2020-12 default when it is absent, an unsupported-version error for other drafts, and a registry custom meta-schema whose chain must end in 2020-12.
  - scope: internal/compile/metaschema.go, internal/compile/metaschema_test.go
  - done when: Unit tests cover the absent case, the 2020-12 URI with and without '#', draft-07 rejected, a custom meta-schema chain ending in 2020-12 accepted, and a chain that does not end in 2020-12 rejected.
  - depends on: T1
  - delivers: AC2, AC3, AC4
- **T3** — Validate loaded schemas against their meta-schema at load time
  - purpose: Run the existing evaluator, with $dynamicRef dynamic scope and format never asserted, over the main schema and every reached registry schema before Compile returns. Map any failure to the invalid-schema error.
  - scope: internal/compile/metacheck.go, internal/compile/metacheck_test.go, internal/compile/compile.go, invalid_schema_error.go
  - done when: Unit tests show that a nested invalid subschema is caught through $dynamicRef '#meta', that an unreferenced invalid registry schema is ignored, that the result is the same with format mode on or off, and that the error carries location and rule but no values.
  - depends on: T1, T2
  - delivers: AC5, AC6, AC7, AC9
- **T4** — Library tests with spec-derived valid and invalid schemas
  - purpose: Prove the behaviour end to end through the public API, using JSON test cases.
  - scope: metaschema_test.go, testdata/metaschema/schema_keyword.json, testdata/metaschema/invalid_schemas.json, testdata/metaschema/custom_metaschema.json, testdata/metaschema/builtin_address.json
  - done when: go test -race ./... passes. Every case cites its core spec section.
  - depends on: T3
  - delivers: AC2, AC3, AC4, AC5, AC6, AC7, AC8, AC9, AC11
- **T5** — CLI contract tests for invalid schemas
  - purpose: Show that the CLI exits with code 2, keeps stdout empty and writes a value-free stderr message naming the file, location and rule.
  - scope: cmd/jsonschema/metaschema_cli_test.go, testdata/metaschema/cli/valid_schema.json, testdata/metaschema/cli/invalid_schema.json, testdata/metaschema/cli/draft07_schema.json, testdata/metaschema/cli/custom_meta.json, testdata/metaschema/cli/custom_meta_schema.json, testdata/metaschema/cli/builtin_address_mismatch.json, testdata/metaschema/cli/document.json
  - done when: The CLI tests pass. For invalid-schema cases they assert the exit code, empty stdout and a stderr message that does not contain the schema values. For the valid schema they assert a normal report.
  - depends on: T3
  - delivers: AC10, AC11
