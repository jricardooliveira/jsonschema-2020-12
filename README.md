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
  - **v1.2**, round 3, faster again with about a third of v1.1's memory: tag `v1.2.0`. It is the Factory's main
    `b0202c8`, which includes the module-path rename story, merged here.
  - Each is the Factory's final integrated main for its round. The commits in this repository not
    made by the Factory are the README and documentation updates, the regrouping of root test
    files, the first module-path rename, and the merges.

Known gaps are listed honestly in [`docs/deviations.md`](docs/deviations.md):

- Go's RE2 regex engine can't do lookaround or backreferences.
- Some formats are only partly checked.
- `verbose` output is only partly produced.

The Go module path is `github.com/jricardooliveira/jsonschema-2020-12`. Since v1.2 the Factory's
own repository uses it too: a Factory story (US-0171) renamed it there. The two earlier tagged
candidates still use `github.com/neoface/jsonschema`, the path the Factory first chose. That is also
why the historical evidence under `docs/work`, `docs/releases` and the ADRs shows the old path.

```sh
go get github.com/jricardooliveira/jsonschema-2020-12
go install github.com/jricardooliveira/jsonschema-2020-12/cmd/jsonschema@latest
```

Or build from a clone as shown below.

## Round 2: performance (v1.1)

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
- Still open after round 2: large documents allocate about 1.5× more than the comparison
  validator, and `multipleOf` allocates on each evaluation. Round 3 addressed both.

## Round 3: memory and speed (v1.2)

- **The plan:** a profiling study of v1.1, done outside the Factory, found where the allocations
  were and prototyped each fix. The 7 changes with the highest confidence went to the Factory
  as one brief amendment, each with measured targets:
  - evaluation path frames reused per depth;
  - one child evaluation context per depth;
  - object keywords without copies;
  - annotations collected in growing chunks;
  - report rendering with a size hint;
  - `multipleOf` in integer arithmetic;
  - the built-in meta-schema index built once.
- **The build:** the Factory delivered them as 7 stories in about 4 hours.
- **The checks, for every story:**
  - all tests pass, and `go test -race` passes;
  - no test was removed;
  - the CLI's output is byte-identical to v1.1 in all four formats (1,052 runs per story).

Medians of 10 interleaved runs, same harness and documents, Apple M3 Max, Go 1.27.2:

| Per operation | v1.1 | v1.2 | santhosh-tekuri v6 | v1.2 vs tekuri |
|---|---|---|---|---|
| Small document | 25.8 µs · 41.7 KB | **19.8 µs · 17.9 KB** | 23.7 µs · 25.9 KB | 0.84× time · 0.69× memory |
| 500 KB valid | 22.5 ms · 35.8 MB | **16.4 ms · 12.2 MB** | 21.3 ms · 23.2 MB | 0.77× · 0.53× |
| 5.5 MB valid | 251 ms · 369 MB | **169 ms · 125 MB** | 223 ms · 242 MB | 0.76× · 0.52× |
| 1,000 errors | 17.0 ms · 29.3 MB | **13.0 ms · 11.6 MB** | 14.7 ms · 17.8 MB | 0.88× · 0.65× |
| Schema compile | 111 µs · 207 KB | **95 µs · 127 KB** | 89 µs · 117 KB | 1.07× · 1.09× |

- The CLI's peak memory on the 5.5 MB document fell from 198 MiB (v1.1) to 149 MiB.
- Still open:
  - schema compilation is 7% slower than tekuri;
  - the memory held after decoding a document is unchanged.

The method, the allocation counts and every caveat are in [`docs/performance.md`](docs/performance.md).

## Conformance

| Candidate | Required (draft2020-12) | Optional | Optional `format/*` (assertion on) | Run by |
|---|---|---|---|---|
| v1 · `09d8ec2` · tag `evaluation-candidate` | **1,302 / 1,302** | — | — | the external evaluator (reported) |
| v1 · `09d8ec2` | **1,302 / 1,302** | 146 / 162 | 820 / 890 | post-experiment check |
| v1.1 · `f35d9e4` · tag `performance-round-2` | **1,302 / 1,302** | 146 / 162 | 820 / 890 | post-experiment check |
| v1.2 · tag `v1.2.0` | **1,302 / 1,302** | 146 / 162 | 820 / 890 | post-experiment check (at `02e2ce1`, before the import-path rename) |

- The post-experiment check used the official [JSON-Schema-Test-Suite](https://github.com/json-schema-org/JSON-Schema-Test-Suite) at commit `55719286` (2026-10-14), `tests/draft2020-12`, through the library API.
  - The suite's `remotes/` were served through the offline registry as `http://localhost:1234/`.
  - Optional `format/*` ran with format assertion on.
  - A schema that fails to compile counted every test in its group as failed.
- It ran after both rounds and outside the Factory. Nothing from it was given back to the Factory.
- v1, v1.1 and v1.2 fail exactly the same optional cases:

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
