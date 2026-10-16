# PLAN #1: Walking skeleton: Go module, load-once check-many library for true/false/{}, keyword registry and minimal CLI

- Date: 2026-10-16
- Status: proposed (pending Checkpoint 2 sign-off)
- Design: see `ADR-001-*.md` under `docs/architecture/adr/`

## Files that change

Declared task scope — the real git diff is checked against this list and
anything outside it is reported as a scope violation in the trust package:

- `go.mod`
- `internal/jsonvalue/decode.go`
- `internal/jsonvalue/decode_test.go`
- `internal/keywords/registry.go`
- `internal/keywords/registry_test.go`
- `internal/keywords/all/all.go`
- `internal/keywords/core/doc.go`
- `internal/keywords/applicator/doc.go`
- `internal/keywords/unevaluated/doc.go`
- `internal/keywords/validation/doc.go`
- `internal/keywords/metadata/doc.go`
- `internal/keywords/formatannotation/doc.go`
- `internal/keywords/content/doc.go`
- `internal/compile/compile.go`
- `internal/compile/compile_test.go`
- `internal/eval/eval.go`
- `jsonschema.go`
- `jsonschema_test.go`
- `spec_test.go`
- `testdata/core/boolean-schemas.json`
- `testdata/core/empty-schema.json`
- `cmd/jsonschema/main.go`
- `cmd/jsonschema/run.go`
- `cmd/jsonschema/run_test.go`

Modules the architecture expects to touch:

- `go.mod`
- `jsonschema.go`
- `jsonschema_test.go`
- `spec_test.go`
- `testdata/core/boolean-schemas.json`
- `testdata/core/empty-schema.json`
- `internal/jsonvalue/decode.go`
- `internal/jsonvalue/decode_test.go`
- `internal/keywords/registry.go`
- `internal/keywords/registry_test.go`
- `internal/keywords/all/all.go`
- `internal/keywords/core/doc.go`
- `internal/keywords/applicator/doc.go`
- `internal/keywords/unevaluated/doc.go`
- `internal/keywords/validation/doc.go`
- `internal/keywords/metadata/doc.go`
- `internal/keywords/formatannotation/doc.go`
- `internal/keywords/content/doc.go`
- `internal/compile/compile.go`
- `internal/compile/compile_test.go`
- `internal/eval/eval.go`
- `cmd/jsonschema/main.go`
- `cmd/jsonschema/run.go`
- `cmd/jsonschema/run_test.go`

## Order of work

Dependency-ordered (topological); the coder implements one task per call in
exactly this order:

1. **T1** — Module and strict JSON decoding
2. **T2** — Keyword registry and vocabulary packages  _(after T1)_
3. **T3** — Compiler, evaluator and root API  _(after T1, T2)_
4. **T4** — Spec-derived testdata runner  _(after T3)_
5. **T5** — Command-line tool  _(after T3)_

## Risks

- Duplicate-key detection through Decoder.Token must be applied recursively. A naive top-level check would miss nested duplicates.
- Story 179 rewrites internal/jsonvalue. Keep decode.go's surface small (one Decode function plus the two error types and MaxDepth) to limit churn.
- Treating unbuilt keywords as {} gives silent 'valid' answers until each keyword lands. The operator accepted this.
- encoding/json Decoder has its own internal depth limit; if our check is not strictly first, the user sees a json error instead of LimitError. Test at MaxDepth+1 guards this.
- The size-cap test allocates ~64 MiB; acceptable under -race but slow-ish. Keep it a single case; story 193 makes the cap configurable so tests can lower it later.
- Line/column computation must use the decoder's InputOffset, which for syntax errors is approximate; message is location-only either way, so imprecision is harmless.

Implementation constraints the coder must honor:

- Standard library only. go.mod has no require lines.
- No central keyword list or switch: keywords are found only through internal/keywords registry.
- main.go contains argument wiring only. Testable logic lives in run.go.
- Error messages and output never include document values.
- Tests read case values from the code (for example, exit-code constants) rather than repeating literals where the code defines them.
- No network or extra file access in code or tests. CLI tests use t.TempDir files and in-memory stdin.
- Nesting depth is checked before descending into '[' or '{' against jsonvalue.MaxDepth; exceeding it returns *LimitError, never panics or overflows the stack.
- All CLI input (schema file, document file, stdin) goes through one readCapped helper using io.LimitReader(r, maxInputBytes+1); more than maxInputBytes is a limit error, exit 2.
- Never surface encoding/json error strings: decode.go maps them to SyntaxError with a fixed Reason and line:column/offset only. Duplicate-member errors give only the location, never the member name.
- FlagSet uses flag.ContinueOnError and SetOutput(io.Discard); run.go writes the single stderr line for unknown flag, -h, missing argument, unknown subcommand and extra arguments.
- Required tests: decode_test — depth MaxDepth+1 gives *LimitError and MaxDepth nested arrays decode fine (built from the constant); duplicate-key and syntax errors whose Error() contains none of the input's distinctive fragments. run_test — 100000 nested '[' as document gives exit 2, empty stdout, exactly one stderr line; input of maxInputBytes+1 bytes (built from the constant) gives exit 2; a bad document containing a secret marker string (e.g. {"pw": "s3cr3t" !}) and a duplicate {"s3cr3t":1,"s3cr3t":2} give stderr not containing 's3cr3t'; `validate -bogus schema.json` gives exit 2, empty stdout, exactly one stderr line.

- DB impact: no · API impact: yes · Migration needed: no

Breaking changes:

- (none)

## Proof

Per-task completion evidence, plus the deterministic gates every task passes:

- **T1**: go build ./... passes. Decode tests show each broken-JSON case returns an error, and well-formed values decode with json.Number.
- **T2**: Registry tests: register and lookup work, and a duplicate registration panics. all.go compiles with every vocabulary imported.
- **T3**: Library tests show true and {} always valid, false always invalid, {"type":"string"} behaving like {}, compile errors for 5/"x"/null/[], broken-JSON errors, and a test keyword reached through the registry.
- **T4**: go test runs every case in testdata/**/*.json. Each case cites its 2020-12 core section.
- **T5**: In-process CLI tests cover exit codes 0, 1 and 2 for every listed error case, stdin via no path and via '-', empty stdout on errors, and one stderr line. go test -race ./... passes.

- `gate-build`: the materialized code compiles (and, unless test execution is
  switched off, the test bodies run and pass).
- `gate-test`: acceptance-criteria coverage, security and performance
  sub-verdicts from the tester.
