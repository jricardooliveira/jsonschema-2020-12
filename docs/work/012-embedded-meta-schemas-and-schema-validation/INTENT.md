# INTENT #12

- Author: Joao Oliveira
- Date: 2026-10-17
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Embed the official 2020-12 meta-schema and all of its vocabulary meta-schemas, under base https://json-schema.org/draft/2020-12/schema, in the binary so they work offline. Honour '$schema'. Before checking any document, validate every loaded schema against its meta-schema. A schema that fails is an invalid-schema error with exit code 2, never 'invalid'. This story depends on dynamic references: the official meta-schemas reach every subschema through "$dynamicRef": "#meta", so validation against them is incomplete without $dynamicRef. Only 2020-12 is supported. Write tests that use valid and invalid schemas, run through both the library and the CLI.

## Operator clarifications
- Some schemas don't say which version of the rules they follow (they have no "$schema" line). How should those be treated? — Treat them as 2020-12 — Checked against the built-in 2020-12 meta-schema as if they had declared it. This tool supports only 2020-12, so this is the safe default.
- What if a schema says it follows a different version, for example "$schema": "http://json-schema.org/draft-07/schema#"? — Schema error, exit code 2: the tool says the declared version is not supported and checks nothing. The 2020-12 meta-schema URI is matched after normalisation, so https://json-schema.org/draft/2020-12/schema and the same URI with an empty fragment (#) both count as 2020-12. A custom meta-schema from the registry whose own meta-schema chain ends in 2020-12 also counts as 2020-12.
- A schema can name its own custom meta-schema instead of the official one, supplied through the schema registry. Should this story handle that, or only the official 2020-12 meta-schema? — Also custom meta-schemas from the registry — If "$schema" points to a meta-schema the caller supplied, the schema is validated against it. That meta-schema is itself validated against the official one. Its "$vocabulary" rules are left to the vocabulary story.
- When a schema fails its meta-schema check, what should the command-line tool print? Scripts and evaluators may read its standard output. — Nothing on stdout, a clear message on stderr — Exit code 2. The message names the schema file, where in the schema it is wrong and which rule failed, without quoting the values. Stdout stays reserved for validation reports, so it is never mistaken for an 'invalid' result.
- The caller can hand the tool extra schema files and folders up front (the schema registry). Some of these may never be used by the main schema. Should every supplied file be checked against its meta-schema before any document is checked, or only the ones actually used? — Same rule as agreed for story 187: every supplied file is parsed at load time (not valid JSON = exit 2), but meta-schema checking applies to the main schema and to every supplied schema the main schema actually reaches through its references, all at load time before any document. An unreferenced supplied schema is not meta-schema-checked and never causes an error.
- The official meta-schema uses 'format' for some schema fields, for example that "$id" looks like a web address. If the operator turns on format checking, should those format rules also apply when a schema is checked against its meta-schema? — No, the schema check never asserts formats — The meta-schema check behaves the same in every mode. Format checking only affects documents, so turning it on never makes a previously accepted schema fail.
- What if a file supplied by the caller uses the same address as one of the built-in official meta-schemas (https://json-schema.org/draft/2020-12/...)? — Same rule as agreed for story 187: if the supplied file is equal as JSON (deep JSON equality) to the built-in official meta-schema at that address, it is accepted and the built-in one is used. A different schema at a built-in meta-schema address is refused as a schema error, exit 2, naming the address and the file. The built-in meta-schemas can never be replaced.

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- The previous full proposal was not supplied with the send-back. Only its modules_affected list was, so the spec and tasks above are reconstructed around that unchanged design. The only change is that the folder testdata/metaschema/cli/ is replaced by seven named fixture files.
- The exact fixture set for the CLI tests may need adjusting, for example a separate registry fixture for the unreferenced-invalid case. Any extra fixture must be added as a named file, and the total stays within 32.
- It is unverified that the existing evaluator in internal/eval can be invoked with format assertion forced off, without editing internal/eval/eval.go (crowded).
- It is unverified that main.go already prefixes load errors with the schema file name. If it does not, satisfying the stderr criterion needs an edit to the crowded cmd/jsonschema/main.go, which must then be claimed.
- The repository inventory is empty, so the actual file names in internal/registry and internal/compile (registry.go, compile.go) are assumed from the project conventions.
- Existing tests that compile schemas which are invalid against the meta-schema would start failing. None can be identified without the inventory.

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
