# Release notes — #20: README and docs/ that describe the finished validator, proven by tests

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #20
- Status: proposed (pending Checkpoint 3 sign-off)

**⚠️ the release-agent has concerns.**

## What changed

Complete, verified documentation for the JSON Schema validator is now available: README with build, test and quick-start instructions; CLI contract covering every flag and exit code; output format examples with exact JSON shapes; library API with a compiled proof-of-concept; scope (mandatory 2020-12 behaviour vs optional features); and deviations (regex differences, partly checked formats, known differences). All CLI examples are validated by an automated test, the flag list is kept in sync by a test, and the library example is verified by go test.

- README.md: added (build, test, quick start, dependency rule, links to docs pages)
- docs/cli.md: added (complete flag documentation with examples, exit codes, multi-document format)
- docs/output-formats.md: added (exact JSON shapes for flag, basic, detailed, verbose; limitations of partly produced formats)
- docs/library.md: added (load-once, check-many API with example code)
- docs/scope.md: added (mandatory 2020-12 behaviour vs optional features)
- docs/deviations.md: added (ECMA-262 regex vs RE2 translation, partly checked formats, known differences from the brief)
- docs/limits.md: updated (all limits with flags and options; now includes story 193 limit flags: --max-depth, --max-input-bytes, --max-ref-chain, --max-steps, --max-pattern-size, --max-output-units, --max-annotations, --max-report-bytes, registry limits)
- docs/examples/: added (schema.json, valid.json, invalid.json, bad-schema.json for use in documentation examples)
- docs/cli-contract.md, docs/formats.md, docs/regex-deviations.md: converted to pointer pages (one line each linking to new locations)
- docs/references.md: updated (links to new docs, describes absoluteKeywordLocation and reference-chain limits)
- example_test.go: added (ExampleCompile in jsonschema_test package; output verified by go test)
- cmd/jsonschema/docs_examples_test.go: added (runs all CLI command examples from README.md and docs/, compares stdout and exit code to documented values)
- cmd/jsonschema/docs_flags_test.go: added (extracts flags from real flag set; cross-checks against flag table in docs/cli.md; includes all story-193 limits)
- cmd/jsonschema/run.go, line 73: changed usage text from '-o flag|basic' to '-o flag|basic|detailed|verbose' (only code change)

## How to verify

- Run `go build ./cmd/jsonschema` and invoke with a bad flag (e.g. `-bad`); usage line now lists all four output formats
- Run the example commands from README.md: `jsonschema validate docs/examples/schema.json docs/examples/valid.json` and `docs/examples/invalid.json`; output and exit codes match the page
- Run `go test ./cmd/jsonschema -run TestDocsCLIExamples` to verify all 5+ documented CLI examples produce exact output
- Run `go test ./cmd/jsonschema -run TestDocsFlag` to verify every flag in the code appears in docs/cli.md with no extras or gaps
- Run `go test -run Example ./` to verify ExampleCompile output matches the documented library example
- Run `go build ./...`, `go vet ./...` and `go test -race ./...` to confirm all tests pass and nothing is broken
- Manually spot-check 2-3 documented flag defaults against the code (e.g. `--max-depth` default 10000 vs `const DefaultMaxDepth = 10000` in the code)

## Migration

none

## Rollback

Delete or revert: README.md, docs/cli.md, docs/output-formats.md, docs/library.md, docs/scope.md, docs/deviations.md, docs/examples/, example_test.go, cmd/jsonschema/docs_examples_test.go, cmd/jsonschema/docs_flags_test.go. Restore docs/cli-contract.md, docs/formats.md, docs/regex-deviations.md to their earlier form. Revert the one-line change in cmd/jsonschema/run.go back to '-o flag|basic'. Restore original docs/limits.md if it existed. In practice, docs issues are fixed forward rather than rolled back.

## Known limitations

- Limit defaults in the docs are read from the code but not checked by an automated test; a future change to a default value in the code would require manual review of the docs to stay in sync (the flag test catches new or removed flags, but not changed defaults)
- Verbose output is only partly produced: a passing leaf assertion without an annotation has no node in the tree, so verbose may equal detailed for some documents (documented in output-formats.md)
- Four earlier docs pages (cli-contract.md, formats.md, regex-deviations.md) are now pointers to new pages; the old content cannot be deleted, only replaced (factory policy for docs merges)
- Stray debug test files (dbg_test.go, internal/registry/dbg_tmp_test.go) and duplicated spec runners (internal/speccases, spec_applicator_test.go) are listed in Known differences but not removed
- No mention of the official JSON Schema test suite or Bowtie, per the brief; docs state only that tests are written from the 2020-12 specification text

## Concerns

- Limit defaults are documented as read from the code, but there is no CI test that parses the docs and compares them to the actual code values. The docs are currently accurate (verified by manual inspection), but a future change to a default value (e.g. story 193's flags or a later story) could drift undetected. This is a low-severity issue because: (1) the flag test catches removed or added flags, (2) defaults are less likely to change than flags, and (3) the concern is noted for future stories. A test that extracts defaults from docs/limits.md and compares them to the actual option and constant values could be added in a follow-up story.
- The diff shows docs/limits.md was significantly rewritten (line count ~50 to ~60), but the exact state before this story is not visible. The story depends on stories 178–196, so all limits should be present. Verify that every constant in options.go, numberlimits.go, cmd/jsonschema/limitflags.go and cmd/jsonschema/refflags.go appears in the table.
- The docs_examples_test.go parses CLI examples as ```sh and ```json blocks in markdown. If a future page uses a code fence with trailing spaces or unusual formatting, the parser might fail silently (the test would see fewer examples than written). The test checks `if n == 0 { t.Fatal(...) }`, so it will error if no examples are found, but a single formatting error in one block is not caught. Consider a lint pass over the markdown if docs grow much larger.
