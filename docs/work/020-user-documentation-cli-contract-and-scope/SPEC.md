# SPEC #20: README and docs/ that describe the finished validator, proven by tests

- Date: 2026-10-17
- Type: docs
- Verdict: pass
- Status: approved by the operator at Checkpoint 1 (run #155, 2026-10-17 05:21 UTC)

## Problem

The validator has no user documentation. The CLI contract, the library API, the limits and their defaults, the line format for several documents, the split between mandatory and optional features, and the known deviations are recorded only in code and tests. Scripts, programmers and the outside evaluator have nothing to read. The CLI usage line in cmd/jsonschema/run.go still says '-o flag|basic', although detailed and verbose now exist.

## Why this matters

The brief requires a documented CLI contract, documented regex deviations and partly checked formats, and a dependency statement. The operator also asked for docs that match the real tool, are checked by automatic tests so they stay true, and are written after every other story so they describe the finished product.

## Acceptance criteria

- README.md states what the project is (a JSON Schema Draft 2020-12 validator: a Go library plus a CLI). It shows how to build with `go build ./cmd/jsonschema` and test with `go test ./...`, gives a quick-start command, and links to every page under docs/.
- README.md states the dependency rule: standard library only, go.mod has no require lines, and no JSON Schema validation library is a dependency.
- docs/cli.md documents every flag the built tool defines, each with its real name, meaning and default. It covers: schema from a file path; documents from file paths or standard input; registry flags mapping files and directories to base URIs; the output-format flag (flag, basic, detailed, verbose; default basic); the format-assertion flag; every limit flag; and the line format for several documents, including how the label wraps the unchanged output object.
- docs/cli.md states the exit codes 0 valid, 1 invalid, 2 error. It states that invalid schemas, unknown required vocabularies, unresolved references, unreadable or broken input and exceeded limits are errors (2), never invalid, and that with several documents the worst result sets the exit code.
- docs/output-formats.md shows the exact JSON shape the built tool prints for flag, basic, detailed and verbose. It states that reports carry instance location, keyword location and absolute keyword location, never instance values, and it names any part of detailed/verbose that is only partly produced.
- docs/limits.md lists every limit (CLI flag and library option) with its real default, read from the code. It states that exceeding a limit gives exit code 2.
- docs/library.md documents the load-once, check-many API (compile with registry, limits and format mode fixed at compile time, then validate many documents, safe for concurrent use) with a short example. The same example exists as a Go Example function in example_test.go, and `go test` checks its `// Output:`.
- docs/scope.md has a 'Mandatory 2020-12 behaviour' section covering every item of the brief's mandatory scope. It has a separate 'Optional features' section listing format-assertion mode (flag, option and vocabulary) and detailed/verbose output.
- docs/deviations.md lists the ECMA-262 regex versus RE2 deviations (what the translation step handles and what is rejected as a schema error), and every partly checked format with what it misses, including idn-hostname and idn-email. Any docs file an earlier story wrote about these subjects is merged into it and removed.
- docs/deviations.md has a 'Known differences' section that lists each place where the finished tool differs from the brief, or states that none remain. It lists the stray debug test files (dbg_test.go, internal/registry/dbg_tmp_test.go) and the duplicated spec runners (internal/speccases, spec_applicator_test.go) if they are still present. The same list is reported back in the story's completion notes.
- No page names or links to the official JSON Schema test suite or Bowtie. The docs say only that the tests are written from the 2020-12 specification text.
- An automatic test runs every CLI command example in README.md and docs/. Each example's printed output and exit code must match what the page shows, and the test fails if they differ.
- An automatic test compares the flag list in docs/cli.md with the flags the tool really defines, including every limit flag added by story 193 (--max-depth, --max-input-bytes, --max-ref-chain, --max-steps, --max-pattern-size, --max-output-units, --max-annotations, --max-report-bytes and the number and registry limit flags). It fails if either list has a flag the other lacks.
- The CLI usage text in cmd/jsonschema/run.go lists all four output formats (flag|basic|detailed|verbose). This one-line text fix is the only change to non-test Go code.
- `go build ./...`, `go vet ./...` and `go test -race ./...` pass with no network access.

## Non-goals

- Fixing any mismatch between the tool and the brief: each mismatch is listed in 'Known differences' and fixed in a separate story.
- Deleting the stray debug test files or merging the duplicated spec runners: they are only listed.
- Any behaviour change to the library or the CLI beyond the one-line usage text.
- Changes to go.mod or adding any dependency.
- Mentioning or referencing the official JSON Schema test suite or Bowtie.
- Writing into factory-owned docs paths (docs/work/, docs/architecture/adr/, docs/releases/).

## Tasks

- **T1** — Fix the CLI usage line
  - purpose: Make the usage text list flag|basic|detailed|verbose, so that the docs, the help text and the flags agree.
  - scope: cmd/jsonschema/run.go
  - done when: Running the tool with a bad flag prints a usage line naming all four formats. The existing cmd/jsonschema tests still pass.
  - delivers: AC14
- **T2** — Write README.md and docs/cli.md, docs/limits.md with example input files
  - purpose: Document build, test, quick start and the dependency rule. Document the complete CLI contract, the exit codes, the multi-document line format and every limit with its real default, read from limitflags.go, refflags.go, options.go and numberlimits.go. Add the small schema and document files the examples run against.
  - scope: README.md, docs/cli.md, docs/limits.md, docs/examples/
  - done when: Each flag in `jsonschema validate -h` appears in docs/cli.md with the same default. Each limit default matches the code.
  - depends on: T1
  - delivers: AC1, AC2, AC3, AC4, AC6, AC11
- **T3** — Write docs/output-formats.md, docs/scope.md and docs/deviations.md
  - purpose: Show the exact output shapes, separate mandatory scope from optional features, and gather regex deviations, partly checked formats and Known differences on one page. Merge and remove any docs file an earlier story wrote on these subjects.
  - scope: docs/output-formats.md, docs/scope.md, docs/deviations.md
  - done when: Each JSON shape on the page is the real output of a command example. The mandatory list covers items (1) to (9) of the brief's correction. The deviations page matches internal/ecmaregex and internal/formats.
  - depends on: T2
  - delivers: AC5, AC8, AC9, AC10, AC11
- **T4** — Library docs with a compiled Example
  - purpose: Write docs/library.md and the matching ExampleCompile in example_test.go, whose // Output: go test checks.
  - scope: docs/library.md, example_test.go
  - done when: `go test -run Example ./` passes, and the code in docs/library.md is the same as the Example body.
  - depends on: T2
  - delivers: AC7
- **T5** — Docs example test
  - purpose: Add a test that extracts every CLI example from README.md and docs/*.md, runs it through the CLI's run function with docs/examples/ as input, and compares stdout and exit code to what the page shows.
  - scope: cmd/jsonschema/docs_examples_test.go
  - done when: The test passes. It fails when one character of a documented output or exit code is changed.
  - depends on: T2, T3
  - delivers: AC12, AC15
- **T6** — Docs flag-list test
  - purpose: Add a test that collects the tool's real flags from its flag set and compares them both ways with the flag table in docs/cli.md, including every story-193 limit flag.
  - scope: cmd/jsonschema/docs_flags_test.go
  - done when: The test passes. It fails when a flag is removed from docs/cli.md or a made-up flag is added to it.
  - depends on: T2
  - delivers: AC13, AC15
