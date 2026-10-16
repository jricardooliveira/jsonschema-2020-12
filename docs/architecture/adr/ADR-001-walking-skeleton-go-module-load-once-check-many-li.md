# ADR #1: Walking skeleton: Go module, load-once check-many library for true/false/{}, keyword registry and minimal CLI

- Date: 2026-10-16
- Story: #1
- Status: approved by the operator at Checkpoint 2 (run #134, 2026-10-16 16:38 UTC)

## Decision
Root package `jsonschema` (jsonschema.go): Compile(schema []byte) (*Schema, error), (*Schema).Validate(doc []byte) (Result, error), Result.Valid(). Errors are typed so the CLI can map them to exit code 2. internal/jsonvalue/decode.go: strict decoder built on encoding/json Decoder.Token with UseNumber. It detects duplicate member names per object and trailing tokens, and returns plain Go values (map[string]any, []any, json.Number, string, bool, nil); story 179 replaces the value model in this package. internal/keywords/registry.go: Keyword{Name, Vocabulary, Compile, Evaluate}, Register (panics on duplicates), Lookup. The compile and evaluate function types and their context types are defined here, so keyword packages import only internal/keywords and there are no import cycles. internal/keywords/all/all.go blank-imports the seven vocabulary packages, which only hold doc.go for now. internal/compile walks a schema object's members in sorted order, asks the registry for each name and ignores unknown names (annotation collection comes later, in story 191). internal/eval runs the compiled nodes against a document with a per-call context and never writes to the compiled schema. cmd/jsonschema/main.go calls os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)). run.go parses the `validate` subcommand with the flag package, reads the files or stdin, calls the root package and writes {"valid": bool} through encoding/json. Hubs: main.go/run.go (later CLI flags), internal/keywords/registry.go and all.go are written once here so later keyword stories only add files.

Revision for reviewer feedback (three fixes, nothing else changed):
(1) Hostile-input limits. decode.go exports `const MaxDepth = 10000` and a typed `*LimitError` (fields: Limit name, Value, plus Offset). The token loop keeps an explicit depth counter (iterative or recursive, but the check `depth+1 > MaxDepth` runs BEFORE descending into any '[' or '{'), so 100000 nested '[' fails fast with no stack growth. Note: encoding/json's own Decoder also has an internal depth cap (10000); our check is at the same value and runs first, so our typed error is what callers see. run.go exports `const maxInputBytes = 64 << 20` (64 MiB) and reads schema, document file and stdin through one helper `readCapped(r io.Reader, name string) ([]byte, error)` using io.LimitReader(r, maxInputBytes+1); reading more than maxInputBytes returns a limit error. The root package re-exports nothing new beyond letting errors.As find *jsonvalue.LimitError through Compile/Validate wrapping (wrap with %w). Every limit error maps to exit 2, one stderr line (e.g. 'document: nesting depth exceeds limit 10000' / 'cannot read doc.json: input exceeds limit 67108864 bytes'), empty stdout. run() recovers nothing on purpose: the design guarantees no panic path rather than hiding one. Story 193 makes these configurable; here they are named constants.
(2) No content in error text. decode.go converts every encoding/json error into its own `*SyntaxError{Offset int64, Line, Column int, Reason string}` whose Error() is location-only, e.g. 'invalid JSON at line 3, column 7 (offset 41)'. Reason is a fixed phrase chosen by the decoder ('syntax error', 'unexpected end of input', 'empty input', 'trailing data after value', 'duplicate member name'), never the encoding/json message (which quotes the offending character/token). Line/column are computed from the input bytes and Decoder.InputOffset(). The duplicate-member error carries the JSON Pointer of the containing object (pointer built from array indices and member names on the path is NOT used, since member names are content) — it reports only the byte offset and line:column of the duplicate key. Errors from io (file not found) are rendered as 'cannot read <path>: <os error without the path repeated>' using errors.Is(fs.ErrNotExist) → 'file not found', else 'read failed'; path is caller-supplied, not document content.
(3) Flag parsing. run.go creates flag.NewFlagSet("validate", flag.ContinueOnError) with fs.SetOutput(io.Discard) so the flag package prints nothing. On a parse error, a missing schema argument, an unknown subcommand, or more than one document path, run.go writes exactly one line itself ('unknown flag -x', 'missing schema argument', 'unknown subcommand', 'only one document is supported for now') and returns exitError. -h/-help (flag.ErrHelp) is treated as a usage error the same way (one line, exit 2). The unknown-flag line names the flag as the user typed it (argument, not document content).

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
- API impact: yes
- Migration needed: no

## Constraints for implementation
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

## Risks
- Duplicate-key detection through Decoder.Token must be applied recursively. A naive top-level check would miss nested duplicates.
- Story 179 rewrites internal/jsonvalue. Keep decode.go's surface small (one Decode function plus the two error types and MaxDepth) to limit churn.
- Treating unbuilt keywords as {} gives silent 'valid' answers until each keyword lands. The operator accepted this.
- encoding/json Decoder has its own internal depth limit; if our check is not strictly first, the user sees a json error instead of LimitError. Test at MaxDepth+1 guards this.
- The size-cap test allocates ~64 MiB; acceptable under -race but slow-ish. Keep it a single case; story 193 makes the cap configurable so tests can lower it later.
- Line/column computation must use the decoder's InputOffset, which for syntax errors is approximate; message is location-only either way, so imprecision is harmless.

## Breaking changes
- none

## Sensitivity
- security
