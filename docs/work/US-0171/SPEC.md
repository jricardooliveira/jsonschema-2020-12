# SPEC US-0171: Rename the Go module path to github.com/jricardooliveira/jsonschema-2020-12

- Date: 2026-10-18
- Type: tech-debt
- Verdict: pass
- Status: proposed (pending Checkpoint 1 sign-off)

## Problem

The Go module is published under github.com/neoface/jsonschema. The project's real home is github.com/jricardooliveira/jsonschema-2020-12, so go.mod, every internal Go import and the library docs name a path that is wrong for the repository.

## Why this matters

Users must be able to fetch the library with go get from its real location and import it under the documented path. The rename must change no behaviour, output, exit code or test.

## Acceptance criteria

- go.mod's module line reads exactly `module github.com/jricardooliveira/jsonschema-2020-12`. go.mod's go version line is unchanged and go.mod still has no require lines.
- Every Go import of the module in the repository uses github.com/jricardooliveira/jsonschema-2020-12 (with its original subpath, for example /internal/eval or /cmd/jsonschema) in place of github.com/neoface/jsonschema. Import aliases, import order and grouping are unchanged, and the root package name stays `jsonschema`.
- docs/library.md names github.com/jricardooliveira/jsonschema-2020-12 wherever it named github.com/neoface/jsonschema, including go get and import examples, and nothing else in it changes.
- Running `grep -rn "github.com/neoface"` over the repository finds the old path only under docs/work/, docs/releases/ and docs/architecture/adr/. Those evidence folders are byte-identical to main 02e2ce1.
- `go build ./...`, `go vet ./...`, `go test ./...` and `go test -race ./...` all pass on the renamed code.
- The set of _test.go files and test functions is identical to main 02e2ce1. No new file is added anywhere, and existing test files differ from main only in their import lines.
- The existing golden comparison test (cmd/jsonschema/golden_test.go) passes against the unmodified testdata/golden/ files, so the CLI's stdout, stderr and exit code match current main in flag, basic, detailed and verbose. No golden file, test expectation or testdata file is edited. If the rename would change any of them, the story stops and is held for the operator's decision.

## Non-goals

- Any behaviour, output, error message, limit, exit-code, performance or goroutine-safety change
- Editing the factory's evidence under docs/work/**, docs/releases/** or docs/architecture/adr/**
- Adding a migration or deprecation note for users of the old path (the operator chose no note)
- Changing README.md, docs/examples, CHANGELOG, scripts, Makefiles or code comments. The operator found no old path in them on main 02e2ce1. If the coder finds the old path in any file outside the 12 named scope entries, it is left unchanged and listed in the story's report for the operator to decide.
- Adding a test, test file or test function that checks for the old path. The grep is a command run by the coder and tester, not a test.
- Regenerating or editing any CLI golden file
- Splitting the rename into several tasks or stories

## Tasks

- **T-0001** — Rename the module path in go.mod, all Go import lines and docs/library.md
  - purpose: Move the module to github.com/jricardooliveira/jsonschema-2020-12 in one consistent change, so the build never sees a mix of old and new import paths.
  - scope: go.mod, docs/library.md, annotations.go, jsonschema.go, limits.go, numberlimits.go, options.go, output.go, registry_options.go, vocabulary_errors.go, cmd/jsonschema/, internal/
  - done when: In the scope entries, change only these things:
- the go.mod module line;
- import strings that start with github.com/neoface/jsonschema, replacing that prefix with github.com/jricardooliveira/jsonschema-2020-12 and keeping aliases and order;
- the path in docs/library.md.

Then run these commands and record their output in the story report:
- `go build ./...`
- `go vet ./...`
- `go test ./...`
- `go test -race ./...` (the golden comparison test must pass against the unmodified testdata/golden/)
- `grep -rn "github.com/neoface" .`, which must list matches only under docs/work/, docs/releases/ and docs/architecture/adr/
- `git diff --stat main`, which must show no added files, no changes under testdata/ or the evidence folders, and test files changed only on import lines.

Any other file found naming the old path is listed in the report and left unchanged.
  - delivers: AC1, AC2, AC3, AC4, AC5, AC6, AC7
