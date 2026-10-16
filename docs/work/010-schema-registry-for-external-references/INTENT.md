# INTENT #10

- Author: Joao Oliveira
- Date: 2026-10-16
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Add a controlled schema registry. Through a library option and CLI flags, the caller supplies extra schemas as individual files or as whole directories, each mapped to a base URI. '$ref' to another resource is found by that resource's $id or by its mapped URI, after RFC 3986 resolution. The validator never accesses the network and never reads any file or folder the caller did not list, including file:// addresses. Anything not found in the registry is an unresolved-reference error with exit code 2. The registry is fixed when the schema is loaded. Add library tests and CLI contract tests for file mapping, directory mapping and unresolved references.

## Operator clarifications
- When the caller lists a whole folder of extra schemas, which files in it should be loaded, and what address does each one get? — Every .json file, including subfolders — Each file's address is the folder's base address plus its path inside the folder. For example, folder mapped to https://example.com/schemas/ makes common/address.json reachable as https://example.com/schemas/common/address.json. Other file types are ignored.
- What should happen when two supplied schemas claim the same address? For example, two files both declaring the same $id, or a file claiming the official 2020-12 meta-schema address. — Stop with an error (exit 2) when loading, naming the address and both sources, with one exception: if the two schemas claiming the same address are equal as JSON (deep JSON equality), it is the same schema and is accepted once. This covers a caller who supplies a copy of an official 2020-12 meta-schema that is also embedded. A different schema at an embedded meta-schema address is an error.
- A supplied schema may be broken: bad JSON, or not a valid 2020-12 schema. If the main schema never points to it, should that still be an error? — Split it: every supplied file is read and parsed at load time so its addresses ($id, embedded $id) can be indexed, and a file that is not valid JSON is a load error (exit 2) naming the file. But checking a supplied schema as a valid 2020-12 schema (meta-schema validation, its $schema) happens only for the schemas the main schema actually reaches through its references. Those are still all resolved and checked at load time, before any document, so results never depend on the document. An unreferenced supplied schema written for another draft is ignored, not an error.
- A listed folder may contain a shortcut (symbolic link) that leads to a file or folder outside it. For example, schemas/extra.json might actually point to /etc/secret.json. Should the shortcut be followed? — Never follow shortcuts that lead outside the listed folder — These are skipped and reported as an error (exit code 2). This keeps the promise that nothing unlisted is ever read. Shortcuts that stay inside the folder still work.
- How should a person tell the command-line tool about extra schemas? Scripts will type this, so the shape of the switches matters. — Two repeatable switches, address then path — --ref https://example.com/a.json=./a.json for one file, --ref-dir https://example.com/schemas/=./schemas for a folder. Each can be given many times. A file with its own $id can also be listed with no address (--ref ./a.json) and is found by that $id.
- The main schema may not declare its own address ($id), yet point to a neighbour, for example {"$ref": "address.json"}. What address should the main schema get, so that "address.json" can be looked up in the registry? — Its file location by default, with a switch to change it — A main schema at /work/main.json counts as file:///work/main.json, so "address.json" means file:///work/address.json. That is found only if the caller listed it. Nothing unlisted is read. A library option and a CLI switch (for example --base-uri) let the caller set a different address.
- The existing size limit stops a single huge document. A listed folder could still hold thousands of files or one enormous schema. Which limits should apply to the extra schemas? — Same per-file size limit, plus a total limit for everything loaded — Each supplied file obeys the existing size limit, and all supplied schemas together have a generous total cap. Going over either is a clear error (exit 2) naming the file. Both can be raised with the existing limit option and switch.
- A listed file may declare its own address ($id) that differs from the address you mapped it to. Example: you pass --ref https://example.com/a.json=./a.json, but inside the file it says "$id": "https://other.org/person.json". Which address should find it? — Both addresses find it — Recommended. A reference to https://example.com/a.json and one to https://other.org/person.json both reach this file. Inside the file, relative references are resolved from its own $id, as the 2020-12 rules require. This matches how most validators treat a mapped file.
- Operator note on refining again — Add to this story (low finding F1 from story 186's review): registry.Build and the resource index must also walk $ref targets under unknown keywords (e.g. '#/definitions/x', which story 186 follows when the target is an object or boolean), so an $id declared inside such a target (or below it) starts an embedded resource with its own base URI, and relative $refs inside it resolve against that base. Add a test. Keep everything else in the accepted plan.

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- baseuri_option_test.go suggests a base-URI option may already exist from story 186. If so, it is reused, and --base-uri is only wired in the CLI. That file is read, not changed, unless its calls break.
- cmd/jsonschema/refs_cli_test.go may already assert behaviour for unlisted references. If the new default file:// base changes its expected messages, it must be added as a write claim.
- The meta-schema check of reached supplied schemas and the deep-equal exception for embedded meta-schema addresses depend on story 189. This story only guarantees that reached resources are resolved at load time and that duplicate detection works among supplied schemas.
- The exact name of the existing limit option and switch for per-file size is not visible in the truncated inventory. The total cap is added next to it.

## Proposed outcome

See `SPEC.md` in this directory — the spec-agent's acceptance criteria are
the machine-readable statement of the outcome, derived from the problem above.

## Affected users and systems

See `PLAN.md` → *Files that change*, and the project's `project-spec.json`
for the systems in scope.

## Constraints

The project's `PROJECT_RULES.md` and `project-spec.json` (stack, conventions,
NFRs and the `forbidden` list) are injected into every agent and bind this work.

## Open questions

Recorded in `SPEC.md` → *Open questions*. A non-empty list parks the run at
Checkpoint 1 for the operator rather than failing it.
