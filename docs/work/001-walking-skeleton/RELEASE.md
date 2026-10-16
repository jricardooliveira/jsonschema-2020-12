# Release notes — #1: Walking skeleton: Go module, load-once check-many library for true/false/{}, keyword registry and minimal CLI

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #1
- Status: approved by the operator at Checkpoint 3 (run #134, 2026-10-16 16:48 UTC)

**⚠️ the release-agent has concerns.**

## What changed

This change creates the walking skeleton: a Go module with standard-library-only build, a load-once check-many library API for true/false/{} schemas, a single keyword registration point for future keywords, and a minimal CLI that reads schemas and documents from files or stdin, printing {"valid": true/false} with exit codes 0/1/2. All acceptance criteria are implemented; the tester confirms the code design is sound and no security or performance defect exists, but the required `go test -race ./...` verification is not established in the supplied evidence.

- go.mod: Module github.com/neoface/jsonschema with go 1.26, no dependencies.
- internal/jsonvalue/decode.go: Strict JSON decoder (MaxDepth=10000, LimitError on depth exceeded, SyntaxError with line/column only, no content in messages, duplicate member names rejected).
- internal/keywords/registry.go: Central registration point; Register panics on duplicate name; Lookup is the only way keywords are found.
- internal/keywords/all/all.go and vocabulary doc.go files: Seven vocabulary packages (core, applicator, validation, unevaluated, metadata, formatannotation, content) created as skeletons.
- internal/compile/compile.go: Compiles schemas; unknown keywords ignored, non-boolean/non-object schemas rejected with ErrNotSchema.
- internal/eval/eval.go: Evaluates documents; never mutates the compiled Schema.
- jsonschema.go: Root API—Compile(schema []byte) (*Schema, error), (*Schema).Validate(document []byte) (Result, error), Result.Valid() bool.
- cmd/jsonschema/main.go: Minimal—only calls run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr).
- cmd/jsonschema/run.go: Parses `jsonschema validate` subcommand, reads schema and document from files or stdin (capped at 64 MiB), outputs {"valid": bool}\n, exits 0 valid / 1 invalid / 2 error.
- Tests: dispatch_test.go (keyword registration reaches public API), spec_test.go (table-driven testdata runner with 2020-12 citations), run_test.go (CLI exit codes, stdin input, no content leaks), decode_test.go, registry_test.go, compile_test.go, eval_test.go (library behavior).
- testdata/core/boolean-schemas.json, testdata/core/empty-schema.json: Spec-derived test cases from 2020-12 core section 4.3.

## How to verify

- `go build ./...` succeeds with standard library only.
- `go vet ./...` passes.
- `go build ./cmd/jsonschema` produces a working binary.
- `go test ./...` passes all tests.
- `go test -race ./...` completes without race detector reports (REQUIRED by acceptance criterion 10 but NOT verified in supplied evidence—this is the tester's concern).
- CLI smoke test: `echo '{}' | ./jsonschema validate testdata/core/boolean-schemas.json -` exits 0 and prints `{"valid":true}` to stdout.
- Error case: `./jsonschema validate /nonexistent schema.json` exits 2 with exactly one line on stderr, nothing on stdout.
- Verify no external dependencies: `go list -m all` should list only the current module.

## Migration

None. This is a new project in a blank repository.

## Rollback

Delete all 24 new files and directories: go.mod; jsonschema.go, jsonschema_test.go, dispatch_test.go, spec_test.go; testdata/; internal/; cmd/.

## Known limitations

- Only true, false, {} schemas work; all other keywords are silently ignored (treated as {}) until built in later stories.
- Single document per run.
- Output is the 'flag' shape only: {"valid": true/false}.
- No $ref, $id, $anchor, meta-schema validation, format checking, or configuration of limits.
- Input size capped at 64 MiB and nesting depth at 10000—not configurable in this story.

## Concerns

- TESTER'S BLOCKING CONCERN: Acceptance criterion 10 requires `go test -race ./...` to pass with no network access. The tester confirms the code design is sound (immutable compiled schema, per-call evaluation context) and no race-related defect was identified, but the supplied evidence does not include output from running the race-enabled tests. Verify this passes before release. This is a confirmation step—the code appears correct, but the required test execution must be proven.
- Silent wrong answers for unbuilt keywords: a schema like {"type": "string"} compiles and accepts every JSON value until the 'type' keyword is implemented. The operator accepted this to allow parallel development of keywords, but make sure operators running this during development understand that 'valid' answers are incomplete until more keywords land.
