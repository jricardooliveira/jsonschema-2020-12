# PLAN #27: Allocation and memory limit tests for the order scenarios, plus routed test-only fixes

- Date: 2026-10-17
- Status: passed gate 2 with nothing for the operator to decide (run #162, 2026-10-17 13:42 UTC)
- Design: see `ADR-027-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `allocs_test.go`
- `allocs_race_on_test.go`
- `allocs_race_off_test.go`
- `docs/benchmarks.md`
- `metaschema_custom_strict_test.go`
- `cmd/jsonschema/metaschema_custom_strict_cli_test.go`
- `metaschema_uri_resource_test.go`
- `compile_cache_limits_test.go`
- `internal/compile/metacache_test.go`
- `first_compile_concurrency_test.go`
- `cmd/jsonschema/peakmem_test.go`
- `cmd/jsonschema/peakmem_other_test.go`
- `cmd/jsonschema/peakmem_flags_test.go`

Modules the architecture expects to touch:

- `allocs_test.go`
- `allocs_race_on_test.go`
- `allocs_race_off_test.go`
- `docs/benchmarks.md`
- `metaschema_custom_strict_test.go`
- `cmd/jsonschema/metaschema_custom_strict_cli_test.go`
- `metaschema_uri_resource_test.go`
- `first_compile_concurrency_test.go`
- `compile_cache_limits_test.go`
- `internal/compile/metacache_test.go`
- `cmd/jsonschema/peakmem_test.go`
- `cmd/jsonschema/peakmem_other_test.go`
- `cmd/jsonschema/peakmem_flags_test.go`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — Race-mode detection and allocation-count limit tests
2. **T4** — Story 202 follow-up tests: strict custom meta-schema, built-in URI resource, F4
3. **T5** — Reliable very-first-Compile concurrency test via re-exec
4. **T6** — Make the CLI peak-memory test portable
5. **T2** — Total-memory limits for the three large documents  _(after T1)_
6. **T3** — Record the limits in docs/benchmarks.md and fix the Go-version sentence  _(after T1, T2)_

## Risks

- Allocation counts differ between darwin and linux and between Go versions. The limits are measured on the build machine, so they may fail on the operator's Mac if its counts are over 10% higher. Per the operator's answer, that is handled by re-measuring on purpose.
- TotalAlloc wobbles more than counts do, which could make the byte limits flaky. The 10% headroom is the operator's choice.
- The 5.5 MB scenario adds seconds to every go test ./... run (accepted by the operator).
- A new 202 follow-up test may reveal a real defect that this story is not allowed to fix.

Implementation constraints the coder must honor:

- No non-test .go file may change. If a new test exposes a defect, stop and report it rather than fixing production code in this story.
- Standard library only, no network, and nothing added to go.mod or go.sum.
- Start every measurement from raw bytes. Schema compile includes loading the built-in meta-schema, and document scenarios include decoding.
- Keep the run counts small for the 5.5 MB scenario so the extra test time stays at a few seconds. Never skip it under -short.
- Measure memory as the TotalAlloc delta from runtime.ReadMemStats around the operation, after a warm-up run, averaged or taken over a fixed run count that is recorded in the docs.
- Limits are named constants equal to the measured value times about 1.1, with the measured value in a comment, and the docs repeat the same numbers.
- The re-exec child must be selected by an environment variable plus -test.run on os.Args[0], and must not touch Compile anywhere before the 32 goroutines are released.
- The peak-memory flags must be defined in a file without build constraints. Only the syscall.Getrusage use moves behind //go:build unix.
- No timing-based assertions anywhere.

- DB impact: no · API impact: no · Migration needed: no

Breaking changes:

- (none)

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: go test -run Alloc -v . shows six passing tests, each logging its count against its limit. go test -race -run Alloc -v . shows them skipped with the race message.
- **T4**: The new tests pass on the unchanged code. The F4 assertion fails when the behaviour it names is deliberately broken locally, and the code is then reverted.
- **T5**: go test -race -run FirstCompile -v . passes and shows that the child process ran. The parent fails if the child exits non-zero or prints 'WARNING: DATA RACE'.
- **T6**: GOOS=windows go vet ./... and GOOS=linux go vet ./... both pass, and go test ./cmd/jsonschema passes on darwin.
- **T2**: go test -run Alloc -v . logs the bytes and the limit for the three large scenarios and passes.
- **T3**: The docs section lists six count limits and three byte limits that match the constants in allocs_test.go. The baseline sentence names Go 1.27.2 as the same version as the build machine.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
