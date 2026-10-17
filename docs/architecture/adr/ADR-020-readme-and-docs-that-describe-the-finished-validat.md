# ADR #20: README and docs/ that describe the finished validator, proven by tests

- Date: 2026-10-17
- Story: #20
- Status: proposed (pending architecture sign-off)

## Decision
This is a docs story. It changes one usage-text line in Go code and adds three test files. All prose goes in new files under docs/ plus README.md. Pages: README.md (what it is, build, test, quick start, dependency rule, links), docs/cli.md, docs/output-formats.md, docs/limits.md, docs/library.md, docs/scope.md, docs/deviations.md (regex, partly checked formats, Known differences). Example inputs live in a new folder, docs/examples/. The docs tests sit in package main under cmd/jsonschema so they can call the existing run function and inspect its flag set without spawning a binary or touching the network. The Example function sits in the root package's external test package (jsonschema_test), so it uses only the public API. Every value in the docs (flag names, defaults, output shapes) is read from the built code, never copied from the brief. Where the two differ, the code wins and the difference goes in Known differences.

## Affected modules
- cmd/jsonschema/run.go
- README.md
- docs/cli.md
- docs/output-formats.md
- docs/limits.md
- docs/library.md
- docs/scope.md
- docs/deviations.md
- docs/examples/schema.json
- docs/examples/valid.json
- docs/examples/invalid.json
- docs/examples/bad-schema.json
- example_test.go
- cmd/jsonschema/docs_examples_test.go
- cmd/jsonschema/docs_flags_test.go

## Data / API impact
- DB impact: no
- API impact: no
- Migration needed: no

## Constraints for implementation
- The only non-test Go change is the usage-text line in cmd/jsonschema/run.go. No other code in run.go changes.
- Read flag names and defaults from cmd/jsonschema/limitflags.go, cmd/jsonschema/refflags.go, cmd/jsonschema/run.go, options.go and numberlimits.go. In tests, ask the flag set and the options for values instead of repeating numbers.
- Read output shapes from internal/output/*.go and confirm them by running the examples. Read regex deviations from internal/ecmaregex/translate.go and properties.go. Read partly checked formats from the per-format files in internal/formats/.
- Never write to docs/work/, docs/architecture/adr/, docs/releases/, PROJECT_RULES.md or project-spec.json.
- No page mentions the official JSON Schema test suite or Bowtie.
- The docs tests use no network and no files outside the repository. Output contains no instance values: the documented examples must show this.
- go.mod is not touched.
- Do not delete dbg_test.go, internal/registry/dbg_tmp_test.go, internal/speccases or spec_applicator_test.go. Only list them in Known differences if they are still present.

## Risks
- Stories 195 and 196 are approved but not yet released. If either changes flag names or defaults after this story is designed, the docs and the flag test must follow, which is why this story depends on them.
- Output in the docs must match byte for byte. If output key order or whitespace is not deterministic, the example test will be flaky.
- Earlier stories may have created docs files for regex or formats whose paths are not in the visible inventory.

## Breaking changes
- none

## Sensitivity
- none
