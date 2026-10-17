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
- **Candidates for evaluation:**
  - Round 1, the feature-complete validator: commit `09d8ec23da454a69fa5eb4d821403b14ca06e88a`
    (tag `evaluation-candidate`).
  - Round 2, the same behaviour made faster: commit `f35d9e4000a57a7c2bdda7a4f216b648de5265b4`
    (tag `performance-round-2`).
  - Both are the Factory's final integrated main for their round. The only commits in this
    repository not made by the Factory are the README updates, their merge and the module-path
    rename.

Known gaps are listed honestly in [`docs/deviations.md`](docs/deviations.md):

- Go's RE2 regex engine can't do lookaround or backreferences.
- Some formats are only partly checked.
- `verbose` output is only partly produced.

This public copy uses the Go module path `github.com/jricardooliveira/jsonschema-2020-12`. It was
renamed after the experiment, in one commit made outside the Factory. The Factory's own repository
and the two tagged candidates still use `github.com/neoface/jsonschema`, the path the Factory
chose; that is also why the historical evidence under `docs/work` and `docs/releases` shows the
old path.

```sh
go get github.com/jricardooliveira/jsonschema-2020-12
go install github.com/jricardooliveira/jsonschema-2020-12/cmd/jsonschema@latest
```

Or build from a clone as shown below.

## Round 2: performance

After round 1, the product owner benchmarked the validator against a widely used Go validator. It
was correct but about 2× slower and used about 3× the memory, and schema compilation was about 10×
slower. They amended the brief with those measurements. The Factory planned 7 more stories (#198–#204):

- reproducible in-repo benchmarks ([`docs/benchmarks.md`](docs/benchmarks.md));
- keyword locations rendered only when needed;
- the evaluated-property record kept only when an `unevaluated*` keyword can see it;
- annotations kept only when they can be observed;
- the meta-schema set built once per process;
- compact numbers with an inline `int64` fast path;
- allocation and byte-limit tests.

The rule was that correctness never regresses for speed. Every round-1 test still passes,
including under `go test -race`, and round 2 deleted no tests.

In-repo benchmarks (`go test -run '^$' -bench . -benchmem`) on an Apple M3 Max with Go 1.27.2:

| Scenario | Round 1 | Round 2 |
|---|---|---|
| Small document (~0.5 KB), per validation | ~49 µs · 75 KB · 1,582 allocs | ~27 µs · 42 KB · 785 allocs |
| 500 KB valid document | ~46 ms · 74 MB · 1.42 M allocs (~11 MB/s) | ~25 ms · 36 MB · 660 k allocs (~21 MB/s) |
| 330 KB document with 1,000 errors | ~32 ms · 54 MB | ~20 ms · 29 MB |
| 5.5 MB valid document | ~454 ms · 756 MB allocated · 315 MiB peak | ~252 ms · 369 MB allocated · 194 MiB peak |
| Schema compile (repeated) | ~963 µs · 1.04 MB · 28 k allocs | ~121 µs · 209 KB · 2.5 k allocs |

- The small-document, 500 KB and compile rows were re-measured independently at review. The
  other rows come from the operator's runs.
- An output-identity check compared the round-1 and round-2 binaries over 48 runs (12 cases × 4
  output formats); stdout, stderr and exit codes were byte-identical.
- Still open: large documents allocate about 1.5× more than the comparison validator.
  `multipleOf` still allocates on each evaluation. The retained size of decoded documents
  dropped only 8–11%.

## Conformance

| Candidate | Required (draft2020-12) | Optional | Optional `format/*` (assertion on) | Run by |
|---|---|---|---|---|
| v1 · `09d8ec2` · tag `evaluation-candidate` | **1,302 / 1,302** | — | — | the external evaluator (reported) |
| v1 · `09d8ec2` | **1,302 / 1,302** | 146 / 162 | 820 / 890 | post-experiment check |
| v1.1 · `f35d9e4` · tag `performance-round-2` | **1,302 / 1,302** | 146 / 162 | 820 / 890 | post-experiment check |

- The post-experiment check used the official [JSON-Schema-Test-Suite](https://github.com/json-schema-org/JSON-Schema-Test-Suite) at commit `55719286` (2026-10-14), `tests/draft2020-12`, through the library API.
  - The suite's `remotes/` were served through the offline registry as `http://localhost:1234/`.
  - Optional `format/*` ran with format assertion on.
  - A schema that fails to compile counted every test in its group as failed.
- It ran after both rounds and outside the Factory. Nothing from it was given back to the Factory.
- v1 and v1.1 fail exactly the same optional cases:

| Optional file | Passed |
|---|---|
| `cross-draft` | 0 / 1 |
| `dependencies-compatibility` | 22 / 36 |
| `format-assertion` | 3 / 4 |
| `format/hostname` | 41 / 64 |
| `format/idn-email` | 22 / 24 |
| `format/idn-hostname` | 55 / 98 |
| `format/uri-template` | 46 / 48 |

The public suite may have been in the models' training data. The Factory never accessed or ran it during the experiment.

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
