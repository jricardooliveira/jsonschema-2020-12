# SPEC #10: Controlled schema registry for external references

- Date: 2026-10-16
- Type: feature
- Verdict: pass
- Status: proposed (pending Checkpoint 1 sign-off)

## Problem

A schema can only reference resources inside its own document. Callers have no way to supply the external schemas that '$ref' points to, and the tool must never fetch them from the network or read unlisted files.

## Why this matters

The brief requires fully offline validation, with external references resolved only from schemas the caller lists. These come as files or directories mapped to base URIs, through both the library and the CLI. Anything else is an unresolved-reference error with exit code 2.

## Acceptance criteria

- A library option lets the caller supply extra schemas as individual files (mapped to a base URI, or with no URI and found by their own $id) and as directories mapped to a base URI. The registry is fixed when Compile returns and cannot change afterwards.
- For a directory mapped to a base URI, every .json file in it and its subfolders is loaded. Each file's address is the base URI plus the file's relative path (for example, base https://example.com/schemas/ makes common/address.json reachable as https://example.com/schemas/common/address.json). Files of other types are ignored.
- '$ref' to another resource is resolved after RFC 3986 resolution by that resource's $id, its embedded $ids, or its mapped URI. A mapped file whose own $id differs from its mapped URI can be reached by both addresses. Relative references inside it resolve against its own $id.
- A main schema without $id gets the file:// URI of its file location as its base by default. A library option and the CLI switch --base-uri set a different base. A file:// reference to a file the caller did not list is unresolved.
- The validator never accesses the network and never reads a file or folder the caller did not list. A symbolic link inside a listed folder that leads outside it is not followed and is reported as an error with exit code 2. A symbolic link that stays inside the folder works.
- A reference that cannot be found in the document or the registry is an unresolved-reference error with exit code 2, never 'invalid'.
- When two supplied schemas claim the same address, loading fails with exit code 2 and the error names the address and both sources. The exception is two schemas that are deep-JSON-equal: they are accepted once.
- Every supplied file is read and parsed at load time so its addresses can be indexed, and a file that is not valid JSON is a load error with exit code 2 that names the file. Every supplied resource the main schema reaches through references is resolved at load time, before any document is checked. Supplied schemas that are never reached are not checked beyond parsing.
- Each supplied file obeys the existing per-file size limit, and all supplied schemas together obey a total size cap. Going over either is an error with exit code 2 that names the file. Both limits can be raised through the existing limit option and CLI switch.
- The CLI accepts the repeatable switches --ref <uri>=<path> and --ref <path> for single files, and --ref-dir <uri>=<dir> for directories.
- Building the registry and the resource index also walks $ref targets reached under unknown keywords (for example '#/definitions/x', when the target is an object or boolean). An $id declared inside such a target, or below it, starts an embedded resource with its own base URI, and relative $refs inside it resolve against that base. A test covers this.
- Library tests and CLI contract tests cover file mapping, directory mapping and unresolved references, plus the exit codes for each error above. They run offline under go test -race.

## Non-goals

- Fetching any URI over the network, or resolving file:// addresses that were not listed
- Embedding the official 2020-12 meta-schemas and validating supplied schemas against them (story 189)
- $dynamicRef/$dynamicAnchor across registry resources (story 188)
- Custom vocabularies from registry meta-schemas (story 190)
- Following symbolic links that lead outside a listed folder

## Tasks

- **T1** — Load supplied files and directories safely
  - purpose: Read exactly the listed files and directories. This covers recursive .json discovery with relative-path addresses, rejecting symbolic links that escape a listed folder, the per-file size limit and the total size cap, and parse errors that name the file.
  - scope: internal/registry/sources.go, internal/registry/sources_test.go, testdata/registry/dir/schemas/root.json, testdata/registry/dir/schemas/common/address.json, testdata/registry/dir/schemas/notes.txt, testdata/registry/broken/bad.json
  - done when: sources_test.go passes under -race. It covers recursive discovery, ignoring non-JSON files, a symbolic link that escapes the folder and one that stays inside (both created in t.TempDir), the size limit, the total cap and a broken JSON file.
  - delivers: AC2, AC5, AC8, AC9
- **T2** — Index external resources and $ref targets under unknown keywords
  - purpose: Index each supplied resource under its mapped URI, its $id and its embedded $ids. Detect duplicate addresses, accepting deep-equal duplicates. Walk $ref targets reached under unknown keywords so that an $id inside them starts an embedded resource with its own base.
  - scope: internal/registry/index.go, internal/registry/external_index_test.go, internal/registry/unknown_ref_targets_test.go, testdata/registry/conflict/a.json, testdata/registry/conflict/b.json
  - done when: external_index_test.go and unknown_ref_targets_test.go pass. They cover both addresses for a remapped file, a duplicate address error naming both sources, deep-equal acceptance, and an $id under '#/definitions/x' with a relative $ref inside it.
  - depends on: T1
  - delivers: AC3, AC7, AC11
- **T3** — Resolve references through the registry at compile time
  - purpose: Let reference resolution look up the registry after the document itself. Give the main schema its default file:// base or the base the caller sets. Resolve every reached supplied resource at compile time. Report anything not found as an unresolved-reference error.
  - scope: internal/compile/refs.go, internal/compile/compile.go, internal/compile/registry_refs_test.go
  - done when: registry_refs_test.go passes. It covers resolution by $id and by mapped URI, relative resolution inside a remapped file, the default file:// base, an unlisted file:// reference left unresolved, and every reached resource resolved before Validate is called.
  - depends on: T2
  - delivers: AC3, AC4, AC6, AC8
- **T4** — Library options for the registry and the base URI
  - purpose: Expose the options for files, directories, the base URI and the total cap. Wire them into Compile through one registration call in options.go and jsonschema.go. Keep the registry fixed after Compile.
  - scope: registry_options.go, options.go, jsonschema.go, registry_option_test.go, testdata/registry/file/main.json, testdata/registry/file/person.json, testdata/registry/file/self_id.json, testdata/registry/dir/main.json, testdata/registry/unresolved/main.json
  - done when: registry_option_test.go passes under -race. It covers file mapping, file mapping without a URI, directory mapping, an unresolved reference, raising the limits and a compiled schema shared across goroutines.
  - depends on: T3
  - delivers: AC1, AC4, AC9, AC12
- **T5** — CLI flags --ref, --ref-dir and --base-uri
  - purpose: Parse the repeatable registry flags in their own file and pass them to the library options from run.go. Map every load or reference error to exit code 2.
  - scope: cmd/jsonschema/refflags.go, cmd/jsonschema/run.go, cmd/jsonschema/registry_cli_test.go
  - done when: registry_cli_test.go passes. It covers the exit codes and stdout for file mapping, address-less --ref, --ref-dir, --base-uri, unresolved references, a bad flag syntax, an escaping symbolic link, a duplicate address and a broken supplied file.
  - depends on: T4
  - delivers: AC5, AC6, AC10, AC12
