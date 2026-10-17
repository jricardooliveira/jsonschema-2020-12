# PLAN #32: Render the report without copies (round 3, story 5)

- Date: 2026-10-17
- Status: passed gate 2 with nothing for the operator to decide (run #169, 2026-10-17 23:21 UTC)
- Design: see `ADR-032-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `docs/benchmarks.md`
- `internal/output/pointer.go`
- `internal/output/pointer_tilde_test.go`
- `internal/output/sizehint.go`
- `internal/output/sizehint_test.go`
- `internal/output/basic.go`
- `cmd/jsonschema/run.go`
- `cmd/jsonschema/reportbytes_boundary_test.go`
- `testdata/reportbytes/tilde_schema.json`
- `testdata/reportbytes/tilde_document.json`
- `allocs_test.go`

Modules the architecture expects to touch:

- `internal/output/pointer.go`
- `internal/output/pointer_tilde_test.go`
- `internal/output/sizehint.go`
- `internal/output/sizehint_test.go`
- `internal/output/basic.go`
- `cmd/jsonschema/run.go`
- `cmd/jsonschema/reportbytes_boundary_test.go`
- `testdata/reportbytes/tilde_schema.json`
- `testdata/reportbytes/tilde_document.json`
- `allocs_test.go`
- `docs/benchmarks.md`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — Measure the before numbers
2. **T2** — Decode a pointer segment only when it contains '~'  _(after T1)_
3. **T3** — Pre-size the basic report buffer  _(after T1)_
4. **T4** — Report-bytes limit boundary tests  _(after T2, T3)_
5. **T5** — Measure after, lower allocation limits, record and judge  _(after T4)_

## Risks

- The hint may over-reserve if the location-based estimate exceeds the real report size; the cap keeps it at or below the remaining limit, so the measured B/op decides
- The 30 MiB RSS target may be missed because stories 205 to 208 already lowered peak memory; recorded and held per criterion 8
- The BenchmarkCLI/5.5MB required drop (49,252,608 B/op, 1,902,785 allocs/op) comes from story 209 alone and may not be reachable from rendering changes; held for the operator if missed
- Golden output drift if the '~' fast path mishandles an edge case; the differential test and golden tests guard this

Implementation constraints the coder must honor:

- Standard library only; go.mod unchanged
- Never regenerate or edit testdata/golden/*; a golden mismatch fails the story
- The '~' fast path must produce identical output to the previous strings.Replacer decode, including the order of '~1' before '~0'
- Size hint = min(location-based size, remaining report-byte limit) + 1; basic shape only; no further cap
- Do not move or alter the report-byte counting or the limit check
- New tests go in new _test.go files; existing tests unchanged
- Tests derive limit and size values from the code or an unlimited run, never hard-coded
- No peak-memory or timing assertions in default test runs; TestCLIPeakMemory stays opt-in
- Record exact measured values and computations against the v1.1 f35d9e4 sources; no rounded or placeholder digits
- Touch crowded files (cmd/jsonschema/run.go, docs/benchmarks.md) minimally: one wiring line in run.go if the buffer is created there

- DB impact: no · API impact: no · Migration needed: no

Breaking changes:

- (none)

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: A docs/benchmarks.md story-209 'before' section with exact medians and all 5 RSS runs plus the TestCLIPeakMemory value
- **T2**: New tests pass, and the golden tests pass unchanged in all four shapes
- **T3**: sizehint_test proves the cap and the +1 byte, and the golden tests pass unchanged
- **T4**: New boundary tests pass, and the existing limit tests pass unchanged
- **T5**: A docs/benchmarks.md story-209 'after' section with exact computations, lowered limits recorded with measured values, and -race and golden tests green

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
