# SPEC #1: Walking skeleton: Go module, load-once check-many library for true/false/{}, keyword registry and minimal CLI

- Date: 2026-10-16
- Type: feature
- Verdict: pass
- Status: proposed (pending Checkpoint 1 sign-off)

## Problem

Nothing exists yet. Every later story (keywords, output formats, registry, limits, formats) needs a module, a public library API, one place where keywords plug in, and a command-line tool with a fixed exit-code contract to build on.

## Why this matters

An end-to-end path from a schema file and a document to a valid or invalid answer and an exit code lets each later story add one keyword or feature as new files. The single keyword registration point means those stories never edit a shared list. Stories can then be built side by side.

## Acceptance criteria

- go.mod declares module github.com/neoface/jsonschema with `go 1.26` and no require lines. `go build ./...`, `go vet ./...` and `go build ./cmd/jsonschema` succeed with the standard library only.
- The root package exposes a load-once, check-many API. Compile(schema []byte) returns a *Schema for the schemas true, false and {}. That *Schema checks any number of JSON documents with Validate(document []byte), which returns a Result with Valid() bool. true and {} accept every well-formed JSON document. false rejects every one. Validate never changes the compiled Schema.
- A schema object with keywords that are not built yet (for example {"type": "string"} or {"foo": 1}) compiles and behaves exactly like {}, as the operator decided. A schema that is neither a boolean nor an object (for example 5, "x", null or []) is a compile error, never 'invalid'.
- Broken JSON in a schema or a document is an error returned by Compile or Validate, never an invalid result. Broken JSON covers: syntax errors, empty input, extra data after the value (for example `{} {}`), and the same member name twice in one object at any depth (for example {"a": 1, "a": 2}).
- internal/keywords/registry.go provides the ONE keyword registration point. A keyword file calls Register from init() with its name, vocabulary URI, compile function and evaluate function. Registering the same name twice panics. The compiler finds keywords only through the registry, with no central list or switch. internal/keywords/all/all.go blank-imports every 2020-12 vocabulary package (core, applicator, unevaluated, validation, metadata, formatannotation, content), and each package already exists. A test registers a test keyword and shows that compile and evaluate reach it through the registry.
- `jsonschema validate <schema-file> [<document-file>|-]` prints exactly one JSON object, {"valid": true} or {"valid": false}, followed by a newline on stdout. It exits 0 when the document is valid and 1 when it is invalid.
- When no document path is given, or the document path is a single `-`, the document is read from standard input.
- Every usage or input error exits 2, writes nothing to stdout and writes one plain line to stderr (for example 'cannot read schema.json: file not found'). This covers: a missing schema argument, an unknown flag or subcommand, two or more document paths ('only one document is supported for now'), an unreadable schema or document file, broken JSON in either input, and a schema that is not a boolean or an object. The stderr message never quotes document content.
- cmd/jsonschema/main.go only parses arguments and calls run.go, which in turn calls the root package. The CLI logic is an in-process function run(args, stdin, stdout, stderr) int, so the exit-code tests run without building or spawning a binary.
- Tests cover the library (true, false, {} and an unbuilt-keyword schema over several document types; compile errors; every broken-JSON case) and every CLI exit code (0, 1, and 2 for each error case above, including stdin input). Spec-derived cases live in testdata/core/*.json, read by a shared table-driven runner that cites the 2020-12 core section (4.3.2 boolean schemas). `go test -race ./...` passes with no network access.

## Non-goals

- Any real validation keyword (type, enum, properties, ...): later stories add them as new files.
- Exact number model and deep equality (story 179). This story only decodes with UseNumber and never inspects document values.
- The 'basic', 'detailed' and 'verbose' output formats, the output-format registry and several documents per run (stories 180, 194).
- Schema registry, $id/$ref, meta-schema validation, $vocabulary, limits and format assertion (stories 186-196).
- README and docs/ (story 197).
- Concurrency proof under the race detector for shared Schemas (story 193). This story only keeps Validate free of mutation.

## Tasks

- **T1** — Module and strict JSON decoding
  - purpose: Create go.mod and a strict decoder that rejects syntax errors, empty input, trailing data and duplicate member names at any depth.
  - scope: go.mod, internal/jsonvalue/decode.go, internal/jsonvalue/decode_test.go
  - done when: go build ./... passes. Decode tests show each broken-JSON case returns an error, and well-formed values decode with json.Number.
  - delivers: AC1, AC4
- **T2** — Keyword registry and vocabulary packages
  - purpose: Create the single keyword registration point and every vocabulary package, pulled in once by all.go.
  - scope: internal/keywords/registry.go, internal/keywords/registry_test.go, internal/keywords/all/all.go, internal/keywords/core/doc.go, internal/keywords/applicator/doc.go, internal/keywords/unevaluated/doc.go, internal/keywords/validation/doc.go, internal/keywords/metadata/doc.go, internal/keywords/formatannotation/doc.go, internal/keywords/content/doc.go
  - done when: Registry tests: register and lookup work, and a duplicate registration panics. all.go compiles with every vocabulary imported.
  - depends on: T1
  - delivers: AC5
- **T3** — Compiler, evaluator and root API
  - purpose: Compile true, false and {} (unknown keywords ignored, non-object or non-boolean schemas rejected) through the registry. Evaluate documents, and expose Compile, Schema.Validate and Result in the root package.
  - scope: internal/compile/compile.go, internal/compile/compile_test.go, internal/eval/eval.go, jsonschema.go, jsonschema_test.go
  - done when: Library tests show true and {} always valid, false always invalid, {"type":"string"} behaving like {}, compile errors for 5/"x"/null/[], broken-JSON errors, and a test keyword reached through the registry.
  - depends on: T1, T2
  - delivers: AC2, AC3, AC4, AC5
- **T4** — Spec-derived testdata runner
  - purpose: Add the shared table-driven runner and the first spec-derived JSON cases for boolean and empty schemas.
  - scope: spec_test.go, testdata/core/boolean-schemas.json, testdata/core/empty-schema.json
  - done when: go test runs every case in testdata/**/*.json. Each case cites its 2020-12 core section.
  - depends on: T3
  - delivers: AC10
- **T5** — Command-line tool
  - purpose: Add `jsonschema validate` with file or stdin input, {"valid": ...} output and exit codes 0, 1 and 2. main.go only wires arguments to run().
  - scope: cmd/jsonschema/main.go, cmd/jsonschema/run.go, cmd/jsonschema/run_test.go
  - done when: In-process CLI tests cover exit codes 0, 1 and 2 for every listed error case, stdin via no path and via '-', empty stdout on errors, and one stderr line. go test -race ./... passes.
  - depends on: T3
  - delivers: AC6, AC7, AC8, AC9, AC10
