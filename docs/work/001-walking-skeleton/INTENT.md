# INTENT #1

- Author: Joao Oliveira
- Date: 2026-10-16
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Create the smallest end-to-end version of the JSON Schema Draft 2020-12 validator in Go. Set up a single Go module requiring Go 1.26 or newer, with no dependencies beyond the standard library. No existing JSON Schema validation library may ever be a dependency. Go packages pick up new files by themselves, so adding a file never requires editing a shared manifest. Provide a library package with a load-once, check-many API: the caller loads a schema once, then checks any number of JSON documents with it. At this stage the API accepts only the boolean schemas true and false and the empty object schema {}. Give the library ONE registration point for keywords: each schema keyword lives in its own file and registers itself with a central keyword registry, so later keywords add files instead of editing a central list or switch. Provide a non-interactive command-line tool whose main.go only wires flags to the library. It takes a schema from a file path and one document from a file path or from standard input. It prints a minimal JSON result saying valid or invalid. Exit codes are 0 for valid, 1 for invalid, and 2 for any usage error or unreadable file. Include tests for the library and for the CLI exit codes. The tests use no network.

## Operator clarifications
- At this first stage, what exactly should the tool print when it checks a document? Later the default will be the 'basic' report, which also lists each failing location. — The standard 'flag' report — Recommended. Prints exactly {"valid": true} or {"valid": false}. This is already one of the official 2020-12 report shapes, so scripts and evaluators can parse it now, and the 'basic' report can later add its error list without changing this part.
- At this stage the validator only understands true, false and the empty schema {}. Suppose someone gives it a schema with a rule that is not built yet, for example {"type": "string"}. What should happen? — Ignore the unbuilt rule and treat the schema like {} — This matches how the final product treats truly unknown words, but until each rule is built, wrong 'valid' answers come out silently.
- Programmers will type a name to bring the library into their own code (the Go module path, for example github.com/yourname/jsonschema). The tool also needs a command name. What should these be? — You choose placeholder names that are easy to rename — Recommended. For example the module path github.com/neoface/jsonschema and the command name jsonschema. These are cheap to change now, before anyone depends on them.
- How should a script tell the tool to read the document from standard input instead of a file? — No document path, or a single dash — Recommended. `jsonschema schema.json` with nothing after it, or `jsonschema schema.json -`, reads the document from standard input. This is the usual habit of command-line tools, so it fits scripts best.
- When the tool cannot check a document (exit code 2), for example because a file is missing or the JSON is broken, what should it print? — Short message on the error stream, nothing on normal output — Recommended. A one-line plain message such as 'cannot read schema.json: file not found' goes to the error stream (stderr). Normal output stays empty, so a script never mistakes it for a {"valid": ...} result.
- At this stage the tool checks only one document per run, and checking several comes in a later step. What should happen if someone passes two or more document paths now? — Stop with a usage error — Recommended. Exit code 2 and a one-line message on the error stream, such as 'only one document is supported for now'. Nothing is checked silently. A later step can lift this.
- A JSON file can be readable but still odd. For example, it might contain the same name twice in one object, like {"a": 1, "a": 2}, or have extra text after the JSON, like {} {}. How should the tool treat such a schema or document? — Treat both as broken JSON — Recommended. Exit code 2 with a short message. The result never depends on which of two values was picked, which is safer for files from untrusted sources.

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- CLI invocation form: the project spec names the command `jsonschema validate`. The operator's stdin clarification gives the examples `jsonschema schema.json` and `jsonschema schema.json -`, without a subcommand. This proposal follows the spec (`jsonschema validate schema.json [-]`). If the operator meant no subcommand, only run.go's argument parsing and its tests change.
- Output spacing: the operator wrote {"valid": true}. encoding/json emits {"valid":true}. Both are the same JSON object, so tests parse stdout instead of comparing bytes. If byte-exact spacing is required, run.go should write the literal text.
- Exact output API: the root package's Result rendering will be reshaped by story 180 (output-format registry). This story keeps rendering in run.go to avoid creating that registry early.
- Repository inventory was empty. This assumes a blank repository with no existing go.mod or files at these paths.

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
