# Release notes — #12: Embedded 2020-12 meta-schemas and meta-schema validation of every loaded schema

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #12
- Status: proposed (pending Checkpoint 3 sign-off)

**⚠️ the release-agent has concerns.**

## What changed

Schemas are now validated against the 2020-12 meta-schema at load time. The official meta-schemas are embedded in the binary for fully offline use. Invalid schemas fail with exit code 2 and a distinct error type (InvalidSchemaError) that carries no instance values, preventing false 'valid/invalid' reports from malformed schemas.

- internal/registry: embedded nine official 2020-12 meta-schema files and indexed them by $id, resolvable offline
- internal/registry: built-in meta-schema addresses protected — equal supplied files accepted, different files at built-in addresses refused with error naming address and file
- internal/compile: metaschema.go resolves $schema keyword, defaults to 2020-12 when absent, normalizes URIs and accepts empty fragments
- internal/compile: metacheck.go validates main schema and every reached supplied schema against its meta-schema at compile time, using $dynamicRef '#meta' to reach nested subschemas
- Root package: new InvalidSchemaError type exported, distinct from validation results, carries JSON Pointer location and keyword/rule name, never values
- Root package: Compile() now fails when schema fails meta-schema check, with exit code 2 and stderr message naming file, location and rule
- CLI: metaschema_cli_test.go added; new contract tests confirm exit 2, empty stdout, value-free stderr for invalid schemas
- testdata: four new JSON case files (schema_keyword.json, invalid_schemas.json, custom_metaschema.json, builtin_address.json) and seven CLI fixtures, each spec-cited
- Registry: supplied schemas at built-in addresses are checked for deep JSON equality; embedded copy always used, different copy refused

## How to verify

- Run `go test -race ./...` — 16 new meta-schema test cases pass, covering valid schemas, invalid schemas with nested errors through $dynamicRef, unreferenced invalid supplied schemas, custom meta-schema chains, draft-07 rejection, and built-in address conflicts
- Try `jsonschema validate -schema='{"type":"nope"}' -instance='true'` — exits 2 (not 1 invalid), stderr names the schema file, location /type, and rule "type"
- Try `jsonschema validate -schema='{"$schema":"http://json-schema.org/draft-07/schema#"}' -instance='true'` — exits 2 with 'not supported', no mention of draft-07 in error
- Supply a custom meta-schema via `--ref https://x.test/meta=custom_meta_schema.json` — main schema is validated against it, the custom meta-schema is validated against 2020-12, exit codes and messages follow the same rules
- Write a file equal to an embedded meta-schema and supply it via `--ref https://json-schema.org/draft/2020-12/schema=file.json` — accepted, the embedded copy is indexed and used
- Write a file different from the embedded meta-schema at a built-in address and supply it — refused with error naming the address and the file, quoting no values

## Migration

Schemas that were previously accepted but are now caught as invalid against the 2020-12 meta-schema will fail at compile time with InvalidSchemaError (exit code 2), not at validation time. Examples: {"type":"nope"}, {"minLength":-1}, {"enum":"not-an-array"}, {"$schema":"http://json-schema.org/draft-07/schema#"}. Existing tests that compile such schemas must move them from spec_test.go to metaschema_test.go, or update them to be valid. Test cases for negative minItems and fractional minItems have been removed from spec_test.go (testdata/validation/minItems.json) and testdata/validation/maxItems.json, as they are now correctly schema errors; they are covered in metaschema_test.go instead.

## Rollback

Downgrade to the previous version or remove the Compile meta-check by not calling metaCheck() in jsonschema.go. All previous valid schemas remain valid; only invalid schemas are now caught earlier. The breaking change is intentional and correct per the 2020-12 specification.

## Known limitations

- Acceptance criterion 7 (format mode on and off) cannot be fully tested because format-assertion mode does not yet exist. The implementation forces format off during meta-schema validation (see eval.EvaluateFull call), but this will be verified in story 195 when format-assertion is added. When it is, meta-schema validation will still never assert format, so no previously accepted schema will start failing.
- Custom meta-schema chains are validated but only the first one in a chain is recompiled per Compile call; recompilation could be cached to improve performance on repeated Compile calls with the same registry.
- $vocabulary keyword is parsed but not interpreted (left to story 190).
- An unreferenced supplied schema in the registry is not meta-schema-checked, even if it would fail — only reached schemas are validated.

## Concerns

- Acceptance criterion 7 (meta-schema check never asserts format, whether format-assertion mode is on or off) is not proven by any test, because format-assertion mode does not exist. The implementation intends to force format off (architecture notes say so), and the code calls eval.EvaluateFull without a visible format parameter, but this needs verification in story 195.
- Two removed test cases (negative minItems, fractional minItems) are now schema errors instead of validation results. This is correct behaviour — they belong in metaschema_test.go. However, the spec_test.go runner now skips the entire metaschema subdirectory to avoid running them twice; ensure this does not hide any other test case changes.
- Built-in meta-schemas and the root meta-schema are recompiled on every Compile call (see metaCheck in jsonschema.go). This is correct but not optimal for repeated compilations with the same registry. A future optimization could cache compiled meta-schemas.
