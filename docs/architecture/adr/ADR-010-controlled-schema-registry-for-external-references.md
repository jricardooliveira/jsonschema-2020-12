# ADR #10: Controlled schema registry for external references

- Date: 2026-10-16
- Story: #10
- Status: passed gate 2 with nothing for the operator to decide (run #145, 2026-10-16 21:56 UTC)

## Decision
New code goes into new files: internal/registry/sources.go (reading the listed sources), registry_options.go (public options) and cmd/jsonschema/refflags.go (flag parsing). The crowded files (internal/compile/compile.go, jsonschema.go, options.go and cmd/jsonschema/run.go) are touched only to wire the new pieces in. internal/registry/index.go is extended in two ways: it indexes external resources, and it walks $ref targets under unknown keywords (finding F1 from story 186). internal/compile/refs.go looks in the registry after the document's own index. The registry is built once per Compile and is read-only afterwards. testdata/registry/ is a new folder that holds only the fixture files listed. Symbolic-link cases are created at test time in t.TempDir and are not committed. --base-uri already exists in run.go (line 83, from story 186) and is reused; it is only given to the new default file:// base logic, not redeclared.

Rule for --ref and --ref-dir values (decided, from the reviewer): refflags.go splits a value at its FIRST '='. The value is a mapping <uri>=<path> only if the text before that '=' parses with url.Parse as an absolute URI, has a scheme of two or more characters and contains no whitespace. Otherwise the whole value is a path: for --ref, a file found by its own $id; for --ref-dir it is a usage error (exit 2), because a directory always needs a base URI. A path that looks like a URI can be prefixed with './'. A URI that contains '=' cannot be mapped with --ref; callers use the library option instead. The flag help text says all of this. Requiring a scheme of two or more characters keeps Windows drive paths such as C:\x.json as paths. refflags.go is one small parse function plus a flag.Value for repeatable values, with a table-driven test in registry_cli_test.go. The test covers './a=b.json' (a path), 'C:\x.json' (a path), 'https://x/a.json=./a.json' (a mapping) and a mapping with an empty path (usage error, exit 2).

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
- API impact: yes
- Migration needed: no

## Constraints for implementation
- Standard library only. No network access. Open only the paths the caller listed and the files found inside listed directories.
- Resolve symbolic links with filepath.EvalSymlinks and reject any whose target is outside the listed folder.
- Error messages name addresses and file paths, never instance values.
- Tests ask the code for limit values instead of repeating the numbers.
- Do not add code to crowded files beyond the wiring calls.
- Parse --ref and --ref-dir by the first-'=' rule only: mapping if the prefix is an absolute URI (scheme of two or more characters, no whitespace), else the whole value is a path. Document this in the flag help text, including the './' escape and the note that URIs containing '=' need the library option.
- Reuse the existing --base-uri flag in run.go; do not declare it again.

## Risks
- The duplicate-address exception for embedded meta-schemas only becomes testable once story 189 embeds them.
- Default file:// base on Windows paths needs careful RFC 8089 formatting.
- URIs containing '=' cannot be mapped through --ref. This is a deliberate limit, documented in the help text; the library option covers it.
- cmd/jsonschema/refs_cli_test.go asserts --base-uri behaviour. It is read only; if the new default file:// base changes its expected messages, it must be added as a write claim.

## Breaking changes
- none

## Sensitivity
- security
