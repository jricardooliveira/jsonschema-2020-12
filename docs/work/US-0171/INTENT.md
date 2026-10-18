# INTENT US-0171

- Author: Joao Oliveira
- Date: 2026-10-18
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Rename the Go module path from github.com/neoface/jsonschema to github.com/jricardooliveira/jsonschema-2020-12: go.mod, every Go import, and the user docs that name the path (README.md, docs/*.md). The package name stays jsonschema. Leave the factory's past evidence (docs/work/**, docs/releases/**, docs/architecture/adr/**) unchanged. No behaviour change: every test still passes and is not modified beyond its import lines, go test -race is green, and the CLI's output (stdout, stderr, exit code) is byte-identical to the current main in every output format.

## Operator clarifications
- Some places outside README.md and docs/*.md may also name the old path, for example a CHANGELOG, a Makefile or script, the docs/examples folder, or comments in code. What should happen to them? — Only the files named in the request (option 2). I checked main 02e2ce1: outside the evidence folders the old path appears only in go.mod, the Go import lines and docs/library.md; README.md, docs/examples, comments and scripts do not name it. If the coder finds any other file that names it, do not change it: list it in the story's report for me to decide. OPERATOR DESIGN NOTE (applies to the whole story; the earlier run #170 stopped at gate 2 on the 32-file design cap). Do not fail the design, do not ask for a cap exception and do not split the story: the rename is one task done in one go. Stay within the cap by naming FOLDERS, not files, in modules_affected: the factory's scope check accepts a folder entry and covers every file under it, test files are always writable and go.mod is a manifest. Use exactly these entries: go.mod, docs/library.md, annotations.go, jsonschema.go, limits.go, numberlimits.go, options.go, output.go, registry_options.go, vocabulary_errors.go, cmd/jsonschema/, internal/ (12 entries). Inside them change ONLY the go.mod module line, Go import lines (the string "github.com/neoface/jsonschema..." becomes "github.com/jricardooliveira/jsonschema-2020-12...", import aliases and order unchanged), and the path in docs/library.md. Add NO new file, NO new test file and NO new test function: the test-function inventory must stay identical to main 02e2ce1. Acceptance is shown by go build, go vet, go test and go test -race passing on the renamed code, plus grep finding no old path outside docs/work, docs/releases and docs/architecture/adr; the operator re-checks this and the CLI byte-identity at Checkpoint 3.
- People who already use the library with the old path will get an error after this change until they switch to the new path. How should they find out? — No note — The docs simply show the new path. Old users find out when their build fails
- Some CLI output, test expectations or saved golden files might contain the module path, for example a version line or an error message. If renaming would change any of them, what should happen? — Stop and report it to you — The output must stay byte-identical, so the story is held for your decision and no golden file or test is edited
- How will you accept that the rename is complete? — Option 1: a search (grep -rn "github.com/neoface" over the repository) finds the old path only in docs/work, docs/releases and docs/architecture/adr, and go.mod's module line is the new path; go build ./..., go vet ./... and go test -race ./... pass; the CLI output matches current main in flag, basic, detailed and verbose. The search is a command run by the coder and the tester, NOT a new test: add no new test file or test function (the test-function inventory must stay identical to main 02e2ce1). The operator repeats the search and the CLI comparison at Checkpoint 3.

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
