# ADR-US-0171: Rename the Go module path to github.com/jricardooliveira/jsonschema-2020-12

- Date: 2026-10-18
- Story: US-0171
- Status: proposed (pending architecture sign-off)

## Decision
This is a mechanical rename of the module path, done as one task. It changes no structure, behaviour or files beyond the rename. I checked main as it is now. Outside docs/work, docs/releases and docs/architecture/adr, the old path appears in exactly three kinds of place: (1) line 1 of go.mod, `module github.com/neoface/jsonschema`; (2) Go import strings in about 228 .go files, both production and _test files, at the repository root, under cmd/jsonschema/ and under internal/; (3) line 3 of docs/library.md (Package `github.com/neoface/jsonschema`). The old path is not in any comment, string literal, JSON testdata file or golden file. No code reads the module path at run time: there is no debug.ReadBuildInfo and no PkgPath use. The only runtime.Caller use, perfdata.Fixture, works from the file's location on disk, not the module path. So nothing the CLI prints can change. The rename is a plain prefix swap: `github.com/neoface/jsonschema` becomes `github.com/jricardooliveira/jsonschema-2020-12`, and any subpath after it stays the same. The new last path element `jsonschema-2020-12` is not a valid Go identifier, so it differs from the package name. Go then takes the package name from the `package jsonschema` clause, which does not change. Files that already write the alias `jsonschema "github.com/neoface/jsonschema"` (for example vocabulary_ref_test.go) keep it with the new path. Files that import without an alias still bind the name `jsonschema`, so callers compile unchanged and no alias needs adding. Imports are rewritten in place, keeping their order and grouping. gofmt does not reorder lines inside a group, and the new prefix sorts the same way against the standard-library group. go.mod keeps `go 1.26` and gets no require lines. The design uses the operator's 12 folder and file entries. Root-level _test files (allocs_test.go, bench_test.go, example_test.go and others) are outside the named root .go files. They are covered by the factory rule that test files can always be written, and only their import lines change. No crowded file gets more than its import line changed: jsonschema.go, options.go, internal/eval/eval.go, internal/compile/compile.go, cmd/jsonschema/run.go, internal/keywords/registry.go and internal/eval/annotations.go. Even so, this story touches every Go file. To keep merges simple it should run alone. Stories waiting on those files will need their import lines rebased.

## Affected modules
- go.mod
- docs/library.md
- annotations.go
- jsonschema.go
- limits.go
- numberlimits.go
- options.go
- output.go
- registry_options.go
- vocabulary_errors.go
- cmd/jsonschema/
- internal/

## Data / API impact
- DB impact: no
- API impact: yes
- Migration needed: no

## Constraints for implementation
- Change only these: the module line in go.mod, import strings that start with github.com/neoface/jsonschema (swap only the prefix and keep the subpath, alias, order and grouping), and the path on line 3 of docs/library.md.
- go.mod keeps `go 1.26` and gets no require lines.
- Do not change any `package` clause. The root package stays `jsonschema`. Do not add aliases to imports that have none.
- Do all of the rename in one change, so the build never sees old and new paths mixed.
- Add no file, test file or test function. Do not edit testdata/, testdata/golden/, docs/work/, docs/releases/ or docs/architecture/adr/.
- Root-level _test.go files are changed only on their import lines. The other root .go files may not be touched beyond the 8 listed.
- Run gofmt -l afterwards to confirm nothing besides the import strings changed formatting.
- Verify with: go build ./..., go vet ./..., go test ./..., go test -race ./... (this includes cmd/jsonschema golden_test.go against the unmodified goldens), grep -rn "github.com/neoface" . (matches only under the three evidence folders), and git diff --stat main (no added files, no testdata changes).
- If the old path turns up in any other file, leave it unchanged and list it in the report.

## Risks
- Root-level _test files (about 40, for example allocs_test.go, bench_test.go, example_test.go, childctx_race_test.go) need import edits but are not among the 12 named entries. The design relies on the rule that test files can always be written. If the scope check refuses them, the build breaks and the story has to be held.
- The story touches every Go file in the repository, including all the crowded files. Any story running in parallel will conflict on import lines, so schedule this one alone and rebase the others after it.
- Once the module is published under the new path, users importing the old path get a build failure. The operator chose to add no migration note.
- The golden harness builds the CLI with the Go toolchain. Nothing prints the module path (no build info, no version line), so the golden output should stay byte-identical. If it does not, the story stops for the operator.

## Breaking changes
- The Go module import path changes from github.com/neoface/jsonschema (and its /internal/... and /cmd/jsonschema subpaths) to github.com/jricardooliveira/jsonschema-2020-12. Existing library consumers must update their imports and go get target. The package name and API are unchanged.

## Sensitivity
- none
