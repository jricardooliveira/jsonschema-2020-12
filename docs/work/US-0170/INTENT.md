# INTENT US-0170

- Author: Joao Oliveira
- Date: 2026-10-18
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Rename the Go module path from github.com/neoface/jsonschema to github.com/jricardooliveira/jsonschema-2020-12: go.mod, every Go import, and the user docs that name the path (README.md, docs/*.md). The package name stays jsonschema. Leave the factory's past evidence (docs/work/**, docs/releases/**, docs/architecture/adr/**) unchanged. No behaviour change: every test still passes and is not modified beyond its import lines, go test -race is green, and the CLI's output (stdout, stderr, exit code) is byte-identical to the current main in every output format.

## Operator clarifications
- Some files outside go.mod, the Go code, README.md and docs/*.md might also name the old address. Examples are a CHANGELOG, build scripts, CI settings or example commands in testdata. What should happen to those? — Only the files named in the request (option 2). I checked main 02e2ce1: outside the evidence folders the old path appears only in go.mod, the Go import lines and docs/library.md; README.md does not name it, and no CHANGELOG, script, CI file or testdata file names it. If the coder finds any other file that names it, do not change it: list it in the story's report for me to decide.
- The tool's output should stay byte-identical. But the old address could appear in what the tool prints, for example a version line or a saved golden file. If that happens, what should the story do? — Stop and report it to you — Safest. No output and no golden file is changed. The story is held and names the exact place, and you decide.
- People who already use the library under the old address will find that it stops working. Should the README tell them about the move? — No note — The README shows only the new address, with no mention of the old one.

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
