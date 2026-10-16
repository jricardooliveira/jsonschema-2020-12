# ADR #1: Walking skeleton: Go module, load-once check-many library for true/false/{}, keyword registry and minimal CLI

- Date: 2026-10-16
- Story: #1
- Status: proposed (pending architecture sign-off)

## Decision
Root package `jsonschema` (jsonschema.go): Compile(schema []byte) (*Schema, error), (*Schema).Validate(doc []byte) (Result, error), Result.Valid(). Errors are typed so the CLI can map them to exit code 2. internal/jsonvalue/decode.go: strict decoder built on encoding/json Decoder.Token with UseNumber. It detects duplicate member names per object and trailing tokens, and returns plain Go values (map[string]any, []any, json.Number, string, bool, nil); story 179 replaces the value model in this package. internal/keywords/registry.go: Keyword{Name, Vocabulary, Compile, Evaluate}, Register (panics on duplicates), Lookup. The compile and evaluate function types and their context types are defined here, so keyword packages import only internal/keywords and there are no import cycles. internal/keywords/all/all.go blank-imports the seven vocabulary packages, which only hold doc.go for now. internal/compile walks a schema object's members in sorted order, asks the registry for each name and ignores unknown names (annotation collection comes later, in story 191). internal/eval runs the compiled nodes against a document with a per-call context and never writes to the compiled schema. cmd/jsonschema/main.go calls os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)). run.go parses the `validate` subcommand with the flag package, reads the files or stdin, calls the root package and writes {"valid": bool} through encoding/json. Hubs: main.go/run.go (later CLI flags), internal/keywords/registry.go and all.go are written once here so later keyword stories only add files.

## Affected modules
- go.mod
- jsonschema.go
- jsonschema_test.go
- spec_test.go
- testdata/core/boolean-schemas.json
- testdata/core/empty-schema.json
- internal/jsonvalue/decode.go
- internal/jsonvalue/decode_test.go
- internal/keywords/registry.go
- internal/keywords/registry_test.go
- internal/keywords/all/all.go
- internal/keywords/core/doc.go
- internal/keywords/applicator/doc.go
- internal/keywords/unevaluated/doc.go
- internal/keywords/validation/doc.go
- internal/keywords/metadata/doc.go
- internal/keywords/formatannotation/doc.go
- internal/keywords/content/doc.go
- internal/compile/compile.go
- internal/compile/compile_test.go
- internal/eval/eval.go
- cmd/jsonschema/main.go
- cmd/jsonschema/run.go
- cmd/jsonschema/run_test.go

## Data / API impact
- DB impact: no
- API impact: yes: creates the public library API and the CLI contract
- Migration needed: no

## Constraints for implementation
- Standard library only. go.mod has no require lines.
- No central keyword list or switch: keywords are found only through internal/keywords registry.
- main.go contains argument wiring only. Testable logic lives in run.go.
- Error messages and output never include document values.
- Tests read case values from the code (for example, exit-code constants) rather than repeating literals where the code defines them.
- No network or extra file access in code or tests. CLI tests use t.TempDir files and in-memory stdin.

## Risks
- Duplicate-key detection through Decoder.Token must be applied recursively. A naive top-level check would miss nested duplicates.
- Story 179 rewrites internal/jsonvalue. Keep decode.go's surface small (one Decode function) to limit churn.
- Treating unbuilt keywords as {} gives silent 'valid' answers until each keyword lands. The operator accepted this.

## Breaking changes
- none

## Sensitivity
- security
