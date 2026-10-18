# Release notes — US-0171: Rename the Go module path to github.com/jricardooliveira/jsonschema-2020-12

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: US-0171
- Status: proposed (pending Checkpoint 3 sign-off)

## What changed

The Go module path moves from github.com/neoface/jsonschema to github.com/jricardooliveira/jsonschema-2020-12. Users can now fetch and import the library from its real home. Package names, exported APIs, CLI flags and output are unchanged.

- go.mod: module line updated to github.com/jricardooliveira/jsonschema-2020-12; go version and require lines unchanged
- All Go source files: import paths starting with github.com/neoface/jsonschema renamed to github.com/jricardooliveira/jsonschema-2020-12 (subpaths, aliases, order and grouping preserved; package clauses unchanged)
- docs/library.md: package reference and go get example updated to github.com/jricardooliveira/jsonschema-2020-12

## How to verify

- go build ./... (tester confirmed: passes)
- go vet ./... (tester confirmed: passes)
- go test ./... (tester confirmed: passes, including golden comparison test against unmodified testdata/golden/)
- go test -race ./... (tester confirmed: passes)
- grep -rn 'github.com/neoface' . (tester confirmed: matches only in docs/work/, docs/releases/ and docs/architecture/adr/)

## Migration

Users must update import statements and go get target from github.com/neoface/jsonschema to github.com/jricardooliveira/jsonschema-2020-12. Example: import "github.com/jricardooliveira/jsonschema-2020-12". Package name remains jsonschema. No deprecation or migration warning is included in the library (operator decision).

## Rollback

Revert this commit. Users will fall back to importing from github.com/neoface/jsonschema if that path is still published or available in their go.mod and caches.

## Known limitations

- The old module path github.com/neoface/jsonschema is no longer valid. Code importing it will fail with a build error.
- No migration warning or note accompanies this release (operator chose to include none).

## Concerns

- Breaking change: the import path is part of the public API. Existing library consumers will see go build failures until they update their imports.
- Every Go file in the repository is touched, so any parallel story working on Go imports will require rebase after this release.
