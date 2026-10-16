# Release notes — #10: Controlled schema registry for external references

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #10
- Status: proposed (pending Checkpoint 3 sign-off)

**⚠️ the release-agent has concerns.**

## What changed

The jsonschema validator now supports a controlled schema registry for offline reference resolution. Callers supply external schemas as individual files or directories (each mapped to a base URI) through a library API or CLI flags. The tool resolves $ref by the schema's $id, its embedded $ids, or its mapped URI; never accesses the network; and reports unresolved references as exit code 2. All acceptance criteria met; all tests pass under -race. One low-severity issue: a stray debug file should be removed before release.

- internal/registry/sources.go: new Loader reads only listed files/folders, rejects symlinks escaping the folder, enforces per-file and total size limits, opens with O_NONBLOCK to avoid FIFO hangs
- internal/registry/sources_test.go: tests recursive .json discovery, non-JSON filtering, symlink escape detection, size limits, broken JSON naming
- internal/registry/sources_fifo_test.go: Unix-only tests that FIFO opens time out with O_NONBLOCK (not infinite hang)
- internal/registry/index.go: extended Build() to index external resources by URI/$id/embedded $ids, walk $ref targets under unknown keywords so embedded $ids start their own resources, detect duplicate addresses with deep-equal exception
- internal/registry/external_index_test.go: tests remapped file reachable by both addresses, file without URI found by $id, duplicate address error naming both sources, deep-equal duplicate accepted once
- internal/registry/unknown_ref_targets_test.go: tests $id inside $ref target under unknown keyword starts a resource, relative refs inside resolve to that base, unreferenced unknown keywords ignored
- internal/registry/dbg_tmp_test.go: stray debug file with only package declaration — should be removed
- internal/compile/refs.go: new CompileWithRegistry() looks up references in external registry after main document index, resolves all reached supplied resources at compile time, tracks shared state across main and supplied documents
- internal/compile/compile.go: updated to wire external registry through compilation
- internal/compile/registry_refs_test.go: tests resolution by $id and mapped URI, relative refs in remapped file use its own $id, default file:// base, unlisted file:// unresolved, reached resources resolved at compile time before any document
- internal/compile/refs_test.go: test calls updated to pass nil for external registry parameter
- registry_options.go: new RegistryFile (URI, Path), RegistryDir (BaseURI, Path), Registry (Files, Dirs, MaxFileBytes, MaxTotalBytes), CompileWithRegistry() function
- registry_option_test.go: end-to-end tests for file mapping, directory mapping, unresolved references, limit validation, limits raisable, registry immutable after Compile (concurrent access under -race)
- jsonschema.go: compileWith() helper loads external registry before schema compilation
- cmd/jsonschema/refflags.go: new parseRef() splits flag value at first '='; <uri>=<path> if prefix is absolute URI with scheme 2+ chars and no whitespace, else whole value is path; listFlag for repeatable values
- cmd/jsonschema/run.go: --ref, --ref-dir, --max-ref-file-bytes, --max-ref-total-bytes flags added; limit flags validated as positive; registry passed to CompileWithRegistry()
- cmd/jsonschema/registry_cli_test.go: contract tests for parseRef() table, file/dir mapping, base-uri, unresolved refs, bad syntax, escaping symlink, duplicate address, broken file, missing file, all error cases give exit code 2
- testdata/registry/file/{main,person,self_id}.json: fixtures for file mapping tests
- testdata/registry/dir/{main.json,schemas/{root.json,common/address.json,notes.txt}}: fixtures for directory mapping
- testdata/registry/{unresolved,conflict,broken}/: fixtures for error cases

## How to verify

- Library: compile with Registry{Files: []RegistryFile{{URI: "https://ex.com/s.json", Path: "./s.json"}}}; schema's $ref to that URI resolves to the file
- Library: compile with Registry{Dirs: []RegistryDir{{BaseURI: "https://ex.com/schemas/", Path: "./schemas"}}}; validate a document; subdirectories are walked, .json files are addressed as base+path, .txt and other types ignored
- CLI: jsonschema validate --ref https://ex.com/s.json=./s.json schema.json doc.json; $ref in schema resolves to mapped file
- CLI: jsonschema validate --ref-dir https://ex.com/s/=./schemas/ schema.json doc.json; recurses into ./schemas/, loads every .json
- Unresolved: compile {$ref: https://ex.com/missing.json} with no supplied schema; compilation fails with 'unresolved reference' error
- Duplicate address: supply two different schemas with same $id; load fails with error naming the address and both source paths
- Deep-equal: supply same schema (identical JSON) twice; both load successfully, first source recorded, no error
- Size limit exceeded: jsonschema validate --max-ref-file-bytes 100 --ref ./big.json schema.json doc.json where big.json > 100 bytes; exit code 2 naming big.json
- Symlink escape: create folder/link.json -> /etc/passwd in a listed folder; load fails with error naming the symlink and that it leads outside
- Default file:// base: compile schema at file:///work/main.json with {$ref: neighbor.json}; resolves only if neighbor.json listed as file:///work/neighbor.json
- Custom base URI: compile with Options{BaseURI: "https://custom/schemas/"} and {$ref: other.json}; resolves against custom base

## Migration

None. This is a new feature. Existing calls to Compile() and CompileWithOptions() continue unchanged and work as before.

## Rollback

To undo: stop using CompileWithRegistry(); switch back to Compile() or CompileWithOptions(). Remove --ref and --ref-dir flags from CLI commands. Existing schemas are unaffected.

## Known limitations

- URIs containing '=' cannot be mapped via --ref flag (keeps flag parsing simple); use the library Registry option instead for such URIs
- Windows absolute paths like C:\x.json are treated as file paths, not URIs (URI scheme must be 2+ characters)
- Embedded meta-schema duplicate detection with deep-equal exception is fully testable only once story 189 embeds the official 2020-12 meta-schemas
- RE2 regular expressions do not support ECMA-262 lookahead or backreferences (documented separately per project rules)
- idn-hostname and idn-email format checks are partial (project constraint: standard library only, no golang.org/x/net/idna)

## Concerns

- internal/registry/dbg_tmp_test.go is a stray debug file containing only a package declaration; it serves no purpose and should be removed before release (flagged by tester as F4, low severity). Decide: delete it or document why it is kept.
