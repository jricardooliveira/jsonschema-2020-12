# ADR #10: Controlled schema registry for external references

- Date: 2026-10-16
- Story: #10
- Status: proposed (pending architecture sign-off)

## Decision
New code goes into new files: internal/registry/sources.go (reading listed sources), registry_options.go (public options) and cmd/jsonschema/refflags.go (flag parsing). The crowded files (internal/compile/compile.go, jsonschema.go, options.go, cmd/jsonschema/run.go) are touched only to wire the new pieces in. internal/registry/index.go is extended so that it indexes external resources and walks $ref targets under unknown keywords (finding F1 from story 186). internal/compile/refs.go falls back to the registry after the document's own index. The registry is built once per Compile and is read-only afterwards. testdata/registry/ is a new folder holding only the fixture files listed. Symbolic-link cases are created at test time in t.TempDir, not committed.

## Affected modules
- internal/registry/sources.go
- internal/registry/sources_test.go
- internal/registry/index.go
- internal/registry/external_index_test.go
- internal/registry/unknown_ref_targets_test.go
- internal/compile/refs.go
- internal/compile/compile.go
- internal/compile/registry_refs_test.go
- registry_options.go
- options.go
- jsonschema.go
- registry_option_test.go
- cmd/jsonschema/refflags.go
- cmd/jsonschema/run.go
- cmd/jsonschema/registry_cli_test.go
- testdata/registry/file/main.json
- testdata/registry/file/person.json
- testdata/registry/file/self_id.json
- testdata/registry/dir/main.json
- testdata/registry/dir/schemas/root.json
- testdata/registry/dir/schemas/common/address.json
- testdata/registry/dir/schemas/notes.txt
- testdata/registry/unresolved/main.json
- testdata/registry/conflict/a.json
- testdata/registry/conflict/b.json
- testdata/registry/broken/bad.json

## Data / API impact
- DB impact: no
- API impact: yes: new library options and CLI flags (additive)
- Migration needed: no

## Constraints for implementation
- Standard library only. No network access. Open only the paths the caller listed and the files found inside listed directories.
- Resolve symbolic links with filepath.EvalSymlinks and reject any whose target is outside the listed folder.
- Error messages name addresses and file paths, never instance values.
- Tests ask the code for limit values instead of repeating the numbers.
- Do not add code to crowded files beyond the wiring calls.

## Risks
- The duplicate-address exception for embedded meta-schemas only becomes testable once story 189 embeds them.
- Default file:// base on Windows paths needs careful RFC 8089 formatting.

## Breaking changes
- none

## Sensitivity
- security
