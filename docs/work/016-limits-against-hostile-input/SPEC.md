# SPEC #16: Limits against hostile input and proof of the load-once guarantee

- Date: 2026-10-17
- Type: feature
- Verdict: pass
- Status: passed gate 1 with nothing for the operator to decide (run #152, 2026-10-17 01:41 UTC)

## Problem

A crafted schema or document can hang or crash the validator. The causes are deep nesting, huge inputs, $ref/$dynamicRef chains that never end, and exponential re-checking such as nested anyOf. The existing caps (nesting depth 10,000, 64 MiB per input, number digits/exponent, pattern size, output units, annotations) are spread across packages and do not share one option, flag or error style. The load-once guarantee (a compiled Schema is never changed and is safe for concurrent use) is not yet proven under the race detector.

## Why this matters

The brief treats schemas and documents as possibly hostile. Exceeding a limit must be a clear error with exit code 2, never a silent valid or invalid result. The defaults must never affect ordinary inputs. Library callers need to tell a limit error apart from other errors so they can retry with a higher limit. The concurrency guarantee is a stated non-functional requirement.

## Acceptance criteria

- Defaults are exported named constants with doc comments: nesting depth 10,000 (the existing jsonvalue.MaxDepth value), 64 MiB per input (the existing maxInputBytes value), $ref/$dynamicRef chain 2x the depth default (20,000 jumps on one evaluation path), 100,000,000 work steps per document, and 256 MiB total for registry inputs. Tests read these values from the code. All existing tests and spec cases pass unchanged under the defaults.
- Each limit can be raised or lowered through a library Options field and a CLI flag. An unset or zero value means the default. A negative value is rejected (library error, CLI exit 2). No setting switches a limit off.
- The nesting-depth limit applies to both schemas and documents. Exceeding it at compile or validate time returns a limit error. The CLI exits with code 2.
- The size limit applies to each input: the schema, each registry file, and each document from a file or stdin. A separate total limit applies to all registry files and folders together. Exceeding either returns a limit error, exit 2.
- A $ref/$dynamicRef chain that never terminates (for example a self-reference or a mutual cycle that consumes no instance) stops at the reference-chain limit with a limit error, exit 2. A legitimately deep recursive schema within the default is not stopped.
- A deterministic work-step counter (checking steps per document, never wall-clock time) stops an exponential blow-up such as deeply nested anyOf with a limit error, exit 2. A large ordinary document is not stopped.
- The library returns a distinct *jsonschema.LimitError, matchable with errors.As. It names the limit (depth, input size, registry size, reference chain, work steps, number digits, number exponent, pattern size, output units, annotations) and the value that was set. Schema, reference and syntax errors are never a LimitError.
- CLI with one document: a limit error gives exit 2, empty stdout, and exactly one stderr line of the form 'limit exceeded: <limit> <value> (raise with --<flag>)'. With several documents, that document's line marks it as an error instead of a report, and the worst result sets the exit code.
- The existing limits (number digits/exponent, pattern size, output units from story 180, annotations) follow the same style: Options field, CLI flag, LimitError. The output-unit cap (DefaultMaxOutputUnits, Options.MaxOutputUnits) gets a CLI flag. A contract test shows that exceeding it gives exit 2, empty stdout and one stderr line.
- Load-once proof under go test -race: several goroutines run Validate and Result.Output concurrently on one compiled Schema, including documents that hit a limit, with no race reported. A test shows that the compiled Schema's behaviour and limits are unchanged after many validations, because limit counters live only in the per-call evaluation context.
- Every limit has its own library test, and every limit flag has a CLI contract test.

## Non-goals

- No wall-clock timeouts; the work limit is a step count only.
- No way to disable a limit.
- README/docs prose is story 197; this story documents defaults in Go doc comments only.
- No change to validation semantics of any keyword beyond limit checks.

## Tasks

- **T1** — Shared limit model, defaults and LimitError
  - purpose: One internal limit type and counter used by every package, a public LimitError, exported default constants, and Options fields with zero=default and negative=rejected.
  - scope: internal/limits/limits.go, internal/limits/limits_test.go, limits.go, limits_test.go, options.go, numberlimits.go, numberlimits_test.go, jsonschema.go
  - done when: go test ./ ./internal/limits passes; tests read defaults from the constants
  - delivers: AC1, AC2, AC7, AC11
- **T2** — Depth and size limits for schemas, documents and the registry
  - purpose: Make the jsonvalue depth configurable, apply it to schema compilation, enforce the per-input size in the library, and enforce the total registry size.
  - scope: internal/jsonvalue/limits.go, internal/jsonvalue/decode.go, internal/jsonvalue/depth_option_test.go, internal/compile/limits.go, internal/compile/compile.go, internal/compile/limits_test.go, internal/registry/registry.go, internal/registry/budget_test.go, jsonschema.go, limits_test.go
  - done when: Depth, input-size and registry-total tests pass at, below and above each limit
  - depends on: T1
  - delivers: AC3, AC4, AC7, AC11
- **T3** — Reference-chain and work-step budget per evaluation
  - purpose: Per-call budget in the evaluation context that counts checking steps and $ref/$dynamicRef jumps on the current path, wired into eval and the ref keywords.
  - scope: internal/eval/budget.go, internal/eval/eval.go, internal/eval/budget_test.go, internal/keywords/registry.go, internal/keywords/core/ref.go, internal/keywords/core/dynamicref.go
  - done when: Self-ref and mutual-ref cycles and nested-anyOf blow-ups stop with LimitError quickly; deep legitimate recursion passes
  - depends on: T1
  - delivers: AC5, AC6, AC7, AC11
- **T4** — Bring existing limits into the same style
  - purpose: Number digits/exponent, pattern size, output units and annotation caps return the shared LimitError and are configurable by Options in the same way.
  - scope: internal/jsonvalue/decode.go, output.go, internal/ecmaregex/translate.go, internal/eval/annotations.go, limits_test.go
  - done when: Each existing cap returns a LimitError naming it and its set value
  - depends on: T1
  - delivers: AC7, AC9, AC11
- **T5** — CLI limit flags and limit-error reporting
  - purpose: One file registers every limit flag (including output units) and prints the single stderr line with the right flag name. Wired into main.go and run.go with one call each. Multi-document lines mark limit errors.
  - scope: cmd/jsonschema/limitflags.go, cmd/jsonschema/limitflags_test.go, cmd/jsonschema/main.go, cmd/jsonschema/run.go, cmd/jsonschema/parse_errors_test.go
  - done when: CLI contract tests per flag: exit 2, empty stdout, one stderr line; multi-document worst-result exit; negative value is a usage error
  - depends on: T2, T3, T4
  - delivers: AC2, AC8, AC9, AC11
- **T6** — Load-once proof under the race detector
  - purpose: Concurrent Validate plus Result.Output on one Schema, including limit-hitting documents, and an immutability test.
  - scope: concurrency_race_test.go
  - done when: go test -race ./... passes with the new test
  - depends on: T3, T4
  - delivers: AC10
