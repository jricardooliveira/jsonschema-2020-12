# PLAN US-0171: Rename the Go module path to github.com/jricardooliveira/jsonschema-2020-12

- Date: 2026-10-18
- Status: approved by the operator at Checkpoint 2 (run #171, 2026-10-18 10:15 UTC)
- Design: see `ADR-US-0171-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `go.mod`
- `docs/library.md`
- `annotations.go`
- `jsonschema.go`
- `limits.go`
- `numberlimits.go`
- `options.go`
- `output.go`
- `registry_options.go`
- `vocabulary_errors.go`
- `cmd/jsonschema/`
- `internal/`

Modules the architecture expects to touch:

- `go.mod`
- `docs/library.md`
- `annotations.go`
- `jsonschema.go`
- `limits.go`
- `numberlimits.go`
- `options.go`
- `output.go`
- `registry_options.go`
- `vocabulary_errors.go`
- `cmd/jsonschema/`
- `internal/`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T-0001** — Rename the module path in go.mod, all Go import lines and docs/library.md

## Risks

- Root-level _test files (about 40, for example allocs_test.go, bench_test.go, example_test.go, childctx_race_test.go) need import edits but are not among the 12 named entries. The design relies on the rule that test files can always be written. If the scope check refuses them, the build breaks and the story has to be held.
- The story touches every Go file in the repository, including all the crowded files. Any story running in parallel will conflict on import lines, so schedule this one alone and rebase the others after it.
- Once the module is published under the new path, users importing the old path get a build failure. The operator chose to add no migration note.
- The golden harness builds the CLI with the Go toolchain. Nothing prints the module path (no build info, no version line), so the golden output should stay byte-identical. If it does not, the story stops for the operator.

Implementation constraints the coder must honor:

- Change only these: the module line in go.mod, import strings that start with github.com/neoface/jsonschema (swap only the prefix and keep the subpath, alias, order and grouping), and the path on line 3 of docs/library.md.
- go.mod keeps `go 1.26` and gets no require lines.
- Do not change any `package` clause. The root package stays `jsonschema`. Do not add aliases to imports that have none.
- Do all of the rename in one change, so the build never sees old and new paths mixed.
- Add no file, test file or test function. Do not edit testdata/, testdata/golden/, docs/work/, docs/releases/ or docs/architecture/adr/.
- Root-level _test.go files are changed only on their import lines. The other root .go files may not be touched beyond the 8 listed.
- Run gofmt -l afterwards to confirm nothing besides the import strings changed formatting.
- Verify with: go build ./..., go vet ./..., go test ./..., go test -race ./... (this includes cmd/jsonschema golden_test.go against the unmodified goldens), grep -rn "github.com/neoface" . (matches only under the three evidence folders), and git diff --stat main (no added files, no testdata changes).
- If the old path turns up in any other file, leave it unchanged and list it in the report.

- DB impact: no · API impact: yes · Migration needed: no

Breaking changes:

- The Go module import path changes from github.com/neoface/jsonschema (and its /internal/... and /cmd/jsonschema subpaths) to github.com/jricardooliveira/jsonschema-2020-12. Existing library consumers must update their imports and go get target. The package name and API are unchanged.

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T-0001**: In the scope entries, change only these things:
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

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
