# PLAN #20: README and docs/ that describe the finished validator, proven by tests

- Date: 2026-10-17
- Status: proposed (pending Checkpoint 2 sign-off)
- Design: see `ADR-020-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `cmd/jsonschema/run.go`
- `README.md`
- `docs/cli.md`
- `docs/limits.md`
- `docs/examples/`
- `docs/output-formats.md`
- `docs/scope.md`
- `docs/deviations.md`
- `docs/library.md`
- `example_test.go`
- `cmd/jsonschema/docs_examples_test.go`
- `cmd/jsonschema/docs_flags_test.go`

Modules the architecture expects to touch:

- `cmd/jsonschema/run.go`
- `README.md`
- `docs/cli.md`
- `docs/output-formats.md`
- `docs/limits.md`
- `docs/library.md`
- `docs/scope.md`
- `docs/deviations.md`
- `docs/examples/schema.json`
- `docs/examples/valid.json`
- `docs/examples/invalid.json`
- `docs/examples/bad-schema.json`
- `example_test.go`
- `cmd/jsonschema/docs_examples_test.go`
- `cmd/jsonschema/docs_flags_test.go`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — Fix the CLI usage line
2. **T2** — Write README.md and docs/cli.md, docs/limits.md with example input files  _(after T1)_
3. **T3** — Write docs/output-formats.md, docs/scope.md and docs/deviations.md  _(after T2)_
4. **T4** — Library docs with a compiled Example  _(after T2)_
5. **T6** — Docs flag-list test  _(after T2)_
6. **T5** — Docs example test  _(after T2, T3)_

## Risks

- Stories 195 and 196 are approved but not yet released. If either changes flag names or defaults after this story is designed, the docs and the flag test must follow, which is why this story depends on them.
- Output in the docs must match byte for byte. If output key order or whitespace is not deterministic, the example test will be flaky.
- Earlier stories may have created docs files for regex or formats whose paths are not in the visible inventory.

Implementation constraints the coder must honor:

- The only non-test Go change is the usage-text line in cmd/jsonschema/run.go. No other code in run.go changes.
- Read flag names and defaults from cmd/jsonschema/limitflags.go, cmd/jsonschema/refflags.go, cmd/jsonschema/run.go, options.go and numberlimits.go. In tests, ask the flag set and the options for values instead of repeating numbers.
- Read output shapes from internal/output/*.go and confirm them by running the examples. Read regex deviations from internal/ecmaregex/translate.go and properties.go. Read partly checked formats from the per-format files in internal/formats/.
- Never write to docs/work/, docs/architecture/adr/, docs/releases/, PROJECT_RULES.md or project-spec.json.
- No page mentions the official JSON Schema test suite or Bowtie.
- The docs tests use no network and no files outside the repository. Output contains no instance values: the documented examples must show this.
- go.mod is not touched.
- Do not delete dbg_test.go, internal/registry/dbg_tmp_test.go, internal/speccases or spec_applicator_test.go. Only list them in Known differences if they are still present.

- DB impact: no · API impact: no · Migration needed: no

Breaking changes:

- (none)

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: Running the tool with a bad flag prints a usage line naming all four formats. The existing cmd/jsonschema tests still pass.
- **T2**: Each flag in `jsonschema validate -h` appears in docs/cli.md with the same default. Each limit default matches the code.
- **T3**: Each JSON shape on the page is the real output of a command example. The mandatory list covers items (1) to (9) of the brief's correction. The deviations page matches internal/ecmaregex and internal/formats.
- **T4**: `go test -run Example ./` passes, and the code in docs/library.md is the same as the Example body.
- **T6**: The test passes. It fails when a flag is removed from docs/cli.md or a made-up flag is added to it.
- **T5**: The test passes. It fails when one character of a documented output or exit code is changed.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
