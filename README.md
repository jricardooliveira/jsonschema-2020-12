# jsonschema-2020-12

A JSON Schema Draft 2020-12 validator written from scratch in Go: a Go library
(load a schema once, check many documents, safe for concurrent use) plus a
non-interactive command-line tool. It works fully offline. The tests are
written from the 2020-12 specification text.

## Build and test

```sh
go build ./cmd/jsonschema
go test ./...
```

Go 1.26 or newer is required. Nothing is fetched at build, test or run time.

## Quick start

Check a document against a schema. The result is one JSON report on stdout and
the exit code says the outcome (0 valid, 1 invalid, 2 error).

```sh
jsonschema validate docs/examples/schema.json docs/examples/invalid.json
```

```json
{"valid":false,"errors":[{"instanceLocation":"/age","keywordLocation":"/properties/age/minimum","error":"value is less than minimum"},{"instanceLocation":"/name","keywordLocation":"/properties/name/minLength","error":"string is shorter than minLength"}]}
```

Exit code: 1

Reports give locations and rules, never the checked values.

## Dependency rule

Standard library only. `go.mod` has no `require` lines, and no JSON Schema
validation library is a dependency. Adding any dependency needs an explicit
operator decision.

## Documentation

- [CLI](docs/cli.md): flags, input, exit codes, several documents.
- [Output formats](docs/output-formats.md): flag, basic, detailed, verbose.
- [Limits](docs/limits.md): every limit, its flag, option and default.
- [Library](docs/library.md): compile once, validate many.
- [Scope](docs/scope.md): mandatory 2020-12 behaviour and optional features.
- [Deviations](docs/deviations.md): regex differences, partly checked formats, known differences.
- [References](docs/references.md): base URIs, `$ref` resolution, loops.

Earlier pages that now only point to the pages above:
[cli-contract.md](docs/cli-contract.md), [formats.md](docs/formats.md),
[regex-deviations.md](docs/regex-deviations.md).
