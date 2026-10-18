# SPEC US-0170: Rename the Go module path to github.com/jricardooliveira/jsonschema-2020-12

- Date: 2026-10-18
- Type: tech-debt
- Verdict: pass
- Status: proposed (pending Checkpoint 1 sign-off)

## Problem

The Go module is published under the old path github.com/neoface/jsonschema. go.mod, every Go import line and the user documentation in docs/library.md name that path, so the library cannot be fetched or imported under its intended address github.com/jricardooliveira/jsonschema-2020-12.

## Why this matters

Users must be able to `go get` and import the library under its real repository address. The change must be purely mechanical, so the validator's behaviour, output and guarantees stay exactly as they are on the current main.

## Acceptance criteria

- go.mod declares `module github.com/jricardooliveira/jsonschema-2020-12`, and its go directive and lack of require lines are unchanged. The root package name stays `jsonschema`.
- Every Go import of the module or one of its packages (root, internal/*, cmd/jsonschema) uses the github.com/jricardooliveira/jsonschema-2020-12 prefix. No .go file outside docs/work/**, docs/releases/** and docs/architecture/adr/** contains the string github.com/neoface/jsonschema. `go build ./...` and `go vet ./...` succeed.
- The user docs name only the new path: docs/library.md (and any other docs/*.md or README.md that names the module path) uses github.com/jricardooliveira/jsonschema-2020-12. No README.md or docs/*.md file outside the evidence folders contains github.com/neoface/jsonschema, and README.md has no note about the old address.
- Nothing under docs/work/**, docs/releases/** or docs/architecture/adr/** is changed, and no file in testdata/ (including testdata/golden/) is changed. `git diff --stat` against main 02e2ce1 shows only go.mod, .go files, docs/*.md files that named the old path, and the new guard test file.
- Existing _test.go files differ from main only in their import lines. No test is removed, skipped or weakened. `go test ./...` and `go test -race ./...` pass.
- The CLI's stdout, stderr and exit code are byte-identical to the current main in flag, basic, detailed and verbose. The existing golden comparison (cmd/jsonschema/golden_test.go against testdata/golden/) passes unchanged. If the old module path appears in any CLI output or golden file, the story stops and reports the exact place without changing that output or file.
- Any file outside go.mod, the Go code, README.md and docs/*.md that is found to name github.com/neoface/jsonschema is left unchanged and listed in the story's report for the operator to decide.

## Non-goals

- Changing the root package name (it stays `jsonschema`) or any package or directory layout
- Any change to validation behaviour, output reports, library Result, error messages, limits, exit codes, goroutine safety or performance
- Editing factory evidence: docs/work/**, docs/releases/**, docs/architecture/adr/**, PROJECT_RULES.md, project-spec.json, .claude/skills/
- Regenerating or editing any golden file in testdata/golden/, or any other testdata file
- Changing a CHANGELOG, build script, CI configuration or any other file outside go.mod, the Go code, README.md and docs/*.md that names the old path. Such files are reported only.
- A migration or deprecation note for users of the old address, or a redirect or compatibility shim module at the old path
- Adding any dependency or require line to go.mod

## Tasks

- **T-0001** — Switch go.mod and every Go import to the new module path
  - purpose: The module path change must be atomic. go.mod and every import line across the root package, internal/* and cmd/jsonschema have to change together, or the build breaks. This is a mechanical search-and-replace on import paths only, with no other edits to any .go file.
  - scope: go.mod, *.go (repository root, import lines only), internal/**/*.go (import lines only), cmd/jsonschema/*.go (import lines only)
  - done when: `grep -rn 'github.com/neoface/jsonschema' --include='*.go' --include='go.mod' . --exclude-dir=docs` returns nothing. `go build ./...`, `go vet ./...`, `go test ./...` and `go test -race ./...` pass. `git diff main -- '*_test.go'` shows only changed import lines. cmd/jsonschema golden tests pass with testdata/golden/ untouched.
  - delivers: AC1, AC2, AC4, AC5, AC6
- **T-0002** — Update the user docs to name the new module path
  - purpose: docs/library.md shows the module path in go get and import examples. It must name only the new address, with no note about the old one.
  - scope: docs/library.md
  - done when: `grep -rn 'github.com/neoface/jsonschema' README.md docs/*.md` returns nothing. docs/library.md shows github.com/jricardooliveira/jsonschema-2020-12 in its go get and import examples. README.md is unchanged. No file under docs/work, docs/releases or docs/architecture/adr appears in the diff. Existing doc tests (cmd/jsonschema/docs_examples_test.go, docs_flags_test.go) pass unchanged.
  - delivers: AC3, AC4
- **T-0003** — Add a guard test for the module path and report out-of-scope occurrences
  - purpose: Pins the rename with the smallest test that fails if the old path comes back. The test reads go.mod for the module line and scans the .go files, README.md and docs/*.md outside the evidence folders for the old path. The coder also searches the whole repository, outside the evidence folders, for any other file naming the old path, and lists those files in the story report without changing them.
  - scope: module_path_test.go
  - done when: New file module_path_test.go in the root package passes under `go test` and `go test -race`. It fails if go.mod's module line is not github.com/jricardooliveira/jsonschema-2020-12, or if a .go file, README.md or a docs/*.md file outside docs/work, docs/releases and docs/architecture/adr contains github.com/neoface/jsonschema. The story report lists every other file still naming the old path, or states that there are none.
  - delivers: AC1, AC2, AC3, AC7
