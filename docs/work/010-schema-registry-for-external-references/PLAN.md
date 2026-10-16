# PLAN #10: Controlled schema registry for external references

- Date: 2026-10-16
- Status: passed gate 2 with nothing for the operator to decide (run #145, 2026-10-16 21:56 UTC)
- Design: see `ADR-010-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `internal/registry/sources.go`
- `internal/registry/sources_test.go`
- `testdata/registry/dir/schemas/root.json`
- `testdata/registry/dir/schemas/common/address.json`
- `testdata/registry/dir/schemas/notes.txt`
- `testdata/registry/broken/bad.json`
- `internal/registry/index.go`
- `internal/registry/external_index_test.go`
- `internal/registry/unknown_ref_targets_test.go`
- `testdata/registry/conflict/a.json`
- `testdata/registry/conflict/b.json`
- `internal/compile/refs.go`
- `internal/compile/compile.go`
- `internal/compile/registry_refs_test.go`
- `registry_options.go`
- `options.go`
- `jsonschema.go`
- `registry_option_test.go`
- `testdata/registry/file/main.json`
- `testdata/registry/file/person.json`
- `testdata/registry/file/self_id.json`
- `testdata/registry/dir/main.json`
- `testdata/registry/unresolved/main.json`
- `cmd/jsonschema/refflags.go`
- `cmd/jsonschema/run.go`
- `cmd/jsonschema/registry_cli_test.go`

Modules the architecture expects to touch:

- `internal/registry/sources.go`
- `internal/registry/sources_test.go`
- `internal/registry/index.go`
- `internal/registry/external_index_test.go`
- `internal/registry/unknown_ref_targets_test.go`
- `internal/compile/refs.go`
- `internal/compile/compile.go`
- `internal/compile/registry_refs_test.go`
- `registry_options.go`
- `options.go`
- `jsonschema.go`
- `registry_option_test.go`
- `cmd/jsonschema/refflags.go`
- `cmd/jsonschema/run.go`
- `cmd/jsonschema/registry_cli_test.go`
- `testdata/registry/file/main.json`
- `testdata/registry/file/person.json`
- `testdata/registry/file/self_id.json`
- `testdata/registry/dir/main.json`
- `testdata/registry/dir/schemas/root.json`
- `testdata/registry/dir/schemas/common/address.json`
- `testdata/registry/dir/schemas/notes.txt`
- `testdata/registry/unresolved/main.json`
- `testdata/registry/conflict/a.json`
- `testdata/registry/conflict/b.json`
- `testdata/registry/broken/bad.json`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — Load supplied files and directories safely
2. **T2** — Index external resources and $ref targets under unknown keywords  _(after T1)_
3. **T3** — Resolve references through the registry at compile time  _(after T2)_
4. **T4** — Library options for the registry and the base URI  _(after T3)_
5. **T5** — CLI flags --ref, --ref-dir and --base-uri  _(after T4)_

## Risks

- The duplicate-address exception for embedded meta-schemas only becomes testable once story 189 embeds them.
- Default file:// base on Windows paths needs careful RFC 8089 formatting.
- URIs containing '=' cannot be mapped through --ref. This is a deliberate limit, documented in the help text; the library option covers it.
- cmd/jsonschema/refs_cli_test.go asserts --base-uri behaviour. It is read only; if the new default file:// base changes its expected messages, it must be added as a write claim.

Implementation constraints the coder must honor:

- Standard library only. No network access. Open only the paths the caller listed and the files found inside listed directories.
- Resolve symbolic links with filepath.EvalSymlinks and reject any whose target is outside the listed folder.
- Error messages name addresses and file paths, never instance values.
- Tests ask the code for limit values instead of repeating the numbers.
- Do not add code to crowded files beyond the wiring calls.
- Parse --ref and --ref-dir by the first-'=' rule only: mapping if the prefix is an absolute URI (scheme of two or more characters, no whitespace), else the whole value is a path. Document this in the flag help text, including the './' escape and the note that URIs containing '=' need the library option.
- Reuse the existing --base-uri flag in run.go; do not declare it again.

- DB impact: no · API impact: yes · Migration needed: no

Breaking changes:

- (none)

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: sources_test.go passes under -race. It covers recursive discovery, ignoring non-JSON files, a symbolic link that escapes the folder and one that stays inside (both created in t.TempDir), the size limit, the total cap and a broken JSON file.
- **T2**: external_index_test.go and unknown_ref_targets_test.go pass. They cover both addresses for a remapped file, a duplicate address error naming both sources, deep-equal acceptance, and an $id under '#/definitions/x' with a relative $ref inside it.
- **T3**: registry_refs_test.go passes. It covers resolution by $id and by mapped URI, relative resolution inside a remapped file, the default file:// base, an unlisted file:// reference left unresolved, and every reached resource resolved before Validate is called.
- **T4**: registry_option_test.go passes under -race. It covers file mapping, file mapping without a URI, directory mapping, an unresolved reference, raising the limits and a compiled schema shared across goroutines.
- **T5**: registry_cli_test.go passes. It covers the exit codes and stdout for file mapping, address-less --ref, --ref-dir, --base-uri, unresolved references, a bad flag syntax, an escaping symbolic link, a duplicate address and a broken supplied file.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
