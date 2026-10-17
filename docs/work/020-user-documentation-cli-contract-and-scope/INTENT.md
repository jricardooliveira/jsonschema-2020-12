# INTENT #20

- Author: Joao Oliveira
- Date: 2026-10-17
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Write README.md and a docs/ folder with no code changes. Cover how to build, test and run the project, using go build and go test ./.... Describe the complete CLI contract: every flag; schema and document input from files and from standard input; the registry flags that map files and directories to base URIs; the output-format flag and the exact JSON shapes for flag, basic, detailed and verbose; and the line format for multiple documents. State the exit codes: 0 valid, 1 invalid, 2 error, where invalid schemas, unresolved references, bad input and exceeded limits are errors, never invalid. List the limit flags and their defaults, and the format-assertion flag. Document the library API with a short example. Include a section that clearly separates MANDATORY 2020-12 behaviour from OPTIONAL features, which are format-assertion mode and vocabulary and detailed/verbose output. List the documented deviations: ECMA-262 regex versus RE2, and partly checked formats. Add the dependency statement: standard library only, no JSON Schema library. This story depends on these stories: Walking skeleton; Error reports: flag and basic output, multiple documents; Schema registry for external references; Embedded meta-schemas and schema validation; Limits against hostile input; Detailed and verbose output; Format assertion: mode, dates, email, hostnames, IP and UUID; Format assertion: URI, IRI, templates, pointers and regex.

## Operator clarifications
- This story changes no code, so the docs must describe the tool and library exactly as already built. Suppose the writer finds that the built tool differs from what the brief says, for example a limit flag with a different name or default. What should happen? — Document what the code does, list the mismatch — The docs match the real tool, and each difference from the brief is listed in a short 'known differences' note and reported back to you, to be fixed in a separate story
- How should we prove that the commands and examples in the docs actually work, for example that the example command really prints the documented JSON and exits with code 1? — Add an automatic test that runs the README examples — The examples stay correct in the future, but this adds a test file, which goes beyond 'no code changes'
- How should the docs/ folder be laid out? — Short README, one docs file per subject — The README covers what it is, building, testing and a quick start, and links to separate pages: CLI, output formats, library API, scope (mandatory vs optional), deviations, security limits. Easy to find things
- The automatic docs test from answer 2 can cover more or fewer examples. Which examples should it actually run and check? — Every command example in the docs — Each example command in the README and docs/ is run, and its printed output and exit code must match what the page shows. Strongest guarantee, slightly more work.
- Should the docs test also check that the list of flags in the docs matches the tool's real flags? — Yes, every real flag must be documented — The test fails if the tool has a flag that the CLI page does not mention, or if the page lists a flag the tool does not have. This keeps the 'complete CLI contract' true later.
- The brief says this project must not copy or reference the official JSON Schema test suite or Bowtie. Should the docs mention them at all, for example in a sentence like 'correctness is measured separately by an outside evaluator'? — Don't name them at all — The docs say only that the tests here are written from the 2020-12 spec text. This is the safest reading of the brief.
- Where should the 'known differences' note from answer 1 go, the list of places where the built tool differs from the brief? — Its own section on the deviations page — Sits next to the regex and partial-format deviations, so all the 'not exactly as specified' items are on one page.
- Operator note on refining again — Make this story depend on EVERY other story of the backlog (178 to 196) so it runs last and documents the finished product; 'Known differences' then lists only real remaining gaps. Merge any docs files earlier stories created (regex deviations, partly checked formats) into docs/deviations.md and claim them. Also compile the library example: add a Go Example function (for example ExampleCompile in example_test.go) whose output is checked by go test, so the library docs are proven like the CLI examples. Other open points accepted as stated.
- Operator note on refining again — Add to this story: the CLI usage line in cmd/jsonschema/run.go still says '-o flag|basic' although detailed and verbose exist; this story may make that one-line usage-text fix in run.go (the only code change allowed), and the flag-list docs test must cover every limit flag added by story 193 (--max-depth, --max-input-bytes, --max-ref-chain, --max-steps, --max-pattern-size, --max-output-units, --max-annotations, --max-report-bytes and the number/registry flags). Known low items to list honestly in 'Known differences' if still present: stray empty test files left by coders (dbg_test.go, internal/registry/dbg_tmp_test.go), duplicated spec runners (internal/speccases, spec_applicator_test.go). Keep everything else.

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- The inventory is truncated, so the full list of existing docs files is not visible. The spec convention suggests earlier stories may have written docs on regex deviations, partly checked formats, the CLI contract or default limits. Their exact paths could not be confirmed, so they are not claimed here. Before building, the claims must add each such existing docs file (outside the factory-owned paths) as a write, so it can be merged into the matching new page and removed.
- The internal/output and internal/formats file names (other than registry.go) are not visible in the truncated inventory. They are read-only here, but the coder needs them to document shapes and partial formats.
- Stories 195 and 196 are still only approved. Format-assertion flag names, format coverage and partial-check notes can only be documented after they are released, which is why they are dependencies.
- Byte-exact example checks assume the CLI output is deterministic (stable key order, fixed indentation). If detailed/verbose output order depends on map iteration, those examples need a normalised comparison (parse both as JSON and compare). That would be noted in Known differences.
- The exact set of 'number and registry' limit flags is known only from cmd/jsonschema/limitflags.go and refflags.go. The flag test derives it from the real flag set, so no name is assumed here.
- Running examples in-process through the run function assumes run takes args, stdin and stdout and returns an exit code, as the existing run_test.go suggests. If it does not, the example test must build and run the binary from a temp dir, still with no network.

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
