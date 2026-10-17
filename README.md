# jsonschema-2020-12

A **JSON Schema Draft 2020-12 validator** written from scratch in Go (a library and a
command-line tool), built end to end by an **AI Software Factory** as an engineering
experiment.

> **Status: candidate for independent evaluation.** The internal tests pass, but that does not
> mean the validator conforms to Draft 2020-12. Conformance will be decided by an external
> evaluator running the official JSON Schema Test Suite and private tests. Neither was available
> to the Factory or its agents during development.

## The experiment

The goal was to find out how well an AI Software Factory delivers a complex, specification-driven
product through its full development lifecycle, rather than only whether a working program comes
out. The Factory is a deterministic pipeline that drives AI agents (interview → backlog → spec →
architecture → boundary/security review → implementation → testing/review → release →
integration) through gates and human checkpoints, one user story at a time.

What it was asked to build:

- A validator for **JSON Schema Draft 2020-12** ([specification](https://json-schema.org/draft/2020-12)),
  covering every mandatory behaviour: types and validation keywords, objects, arrays and tuples,
  composition and conditionals, local and external `$ref`, `$id`/`$anchor`, `$dynamicRef`,
  annotation collection with `unevaluatedProperties`/`unevaluatedItems`, invalid-schema handling,
  standard output formats, `$vocabulary`, and `format` as an annotation, with
  **format-assertion as an explicit opt-in**.
- A Go library API plus a **non-interactive CLI** with a documented contract (exit codes 0 valid /
  1 invalid / 2 error), so that an external evaluator can drive it.
- An **offline schema registry** for external references. Nothing is fetched from the network,
  and the 2020-12 meta-schemas are embedded.
- **No existing JSON Schema library** may do the validation; the engine is the project's own.

The rules for keeping the evaluation independent:

- Nobody in the experiment (the operator, the controller or the Factory's agents) searched for,
  downloaded or used the official JSON Schema Test Suite or Bowtie.
- The Factory's agents had no network access; their tests are written from the specification
  text.
- When the interview agent proposed vendoring the official suite, the operator refused, and the
  ban is recorded in the brief (`docs/work/INTERVIEW.md`, `project-spec.json`).

## How it was built

- **Language and architecture:** Go was the product owner's only technical constraint. The
  architecture was the Factory's own decision: standard library only, library-first, one
  immutable compiled schema checking many documents. It is recorded in 20 ADRs under
  [`docs/architecture/adr`](docs/architecture/adr).
- **Backlog:** the Factory planned 20 user stories, from the walking skeleton through
  `$dynamicRef`, the meta-schemas, `$vocabulary`, `unevaluated*`, hostile-input limits, output
  formats and format assertion. All 20 went through spec, design, security review,
  implementation, review, a human release checkpoint and integration.
- **Evidence:** each story's plan, spec, review and trust package (the evidence behind each
  release) is in [`docs/work`](docs/work) and [`docs/releases`](docs/releases). Commits prefixed
  `factory:` were made by the Factory.
- **Tests:** 466 Go test and fuzz functions and 84 JSON test-data files, all written from the
  specification. Every candidate passed `go build`, `go vet` and `go test -race`.
- **Candidate for evaluation:** commit `09d8ec23da454a69fa5eb4d821403b14ca06e88a` is the
  Factory's final integrated main. The only commit after it in this repository adds this
  README section.

Known gaps are listed honestly in [`docs/deviations.md`](docs/deviations.md):

- Go's RE2 regex engine can't do lookaround or backreferences.
- Some formats are only partly checked.
- `verbose` output is only partly produced.

The Go module path is `github.com/neoface/jsonschema` (the Factory chose it). Build from a clone
as shown below.

---

## Product README (written by the Factory)

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
