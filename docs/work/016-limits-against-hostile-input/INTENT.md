# INTENT #16

- Author: Joao Oliveira
- Date: 2026-10-17
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Add built-in limits on nesting depth, input size, and $ref/$dynamicRef recursion that never terminates, so a crafted schema or document cannot hang or crash the tool. Exceeding a limit is a clear error with exit code 2, never a silent valid or invalid result. The defaults are generous enough that ordinary, spec-conformant schemas and documents are never affected. The caller can raise each limit through a library option and a CLI flag. Also prove the load-once guarantee: a loaded schema is never changed after loading and is safe to use from several goroutines at the same time. Cover this with tests run under the race detector. Add tests for each limit and CLI contract tests for the limit flags.

## Operator clarifications
- What starting limits should the tool use when the caller sets none? They should never get in the way of ordinary schemas and documents, but should stop a crafted file quickly. — Keep the defaults the skeleton (story 178) already ships, so nothing that works today starts failing: nesting depth 10,000 (jsonvalue.MaxDepth) and 64 MiB per input (maxInputBytes). This story turns those named constants into the configurable defaults. The $ref/$dynamicRef chain limit defaults to 2x the nesting depth (20,000 jumps on one evaluation path), because a legitimately deep document can need a jump per level. All defaults are documented and read from code in tests.
- The brief says callers can raise each limit. Should they also be able to lower a limit, or switch one off entirely? — Raise or lower, never switch off — Recommended. A service checking untrusted files can make the limits stricter, and a trusted batch job can make them looser, but there is always some limit, so a crafted file can never hang the tool.
- Some crafted schemas do not loop forever. Instead, each rule makes the next one check everything again, for example many nested 'anyOf' rules, so the work grows enormously and the check could take hours. Should there also be a limit on the total amount of checking work for one document? — Yes, a deterministic work limit (count of checking steps per document, never wall-clock time), configurable by library option and CLI flag like the others, exceeding it = exit 2. Default 100 million steps, so a legitimately large document within the 64 MiB input cap (for example a million elements checked by several keywords each) is never stopped, while exponential nested-anyOf blow-ups still end in seconds.
- When a limit is exceeded, what should the tool print? Scripts get exit code 2 either way. — Message on the error channel, nothing on the result output — Recommended. A short message such as 'limit exceeded: nesting depth 1000 (raise with --max-depth)' goes to stderr, and stdout stays empty for that document, so nothing on stdout can be mistaken for a valid or invalid report. When several documents are checked, that document's line marks it as an error instead of a report.
- The 64 MiB size limit: which inputs should it apply to? A run can read a schema file, extra schema files or folders for the registry, and one or more documents from files or piped in. — Each input, plus a total for the registry — Same as above, but a registry folder holding thousands of small files is also capped in total. Safer against a huge folder, but adds one more limit and flag.
- Should the nesting-depth limit (10,000 levels by default) also apply to the schema itself, for example a schema with thousands of 'properties' inside 'properties', and not only to the documents being checked? — Same limit for schemas and documents — Recommended. One depth flag covers both. A crafted schema gets the same protection as a crafted document.
- When a programmer uses the Go library and a limit is hit, should they be able to tell that apart from other problems such as a broken schema or a missing reference? — Yes, a distinct 'limit exceeded' error naming the limit — Recommended. The error says which limit (depth, size, reference chain or work steps) and the value that was set, so a program can decide to retry with a higher limit. The tool uses it to print the message with the right flag name.
- Answer 5 added a total size cap for all the extra schema files and folders loaded into the registry, but did not set its starting value. Each single file is already capped at 64 MiB. What should the total be when the caller sets nothing? Like the other limits, it can be raised or lowered with a library option and a CLI flag. — 256 MiB in total — Recommended. Room for four maximum-size schema files, or many thousands of normal ones, so no real registry folder is stopped. A huge folder pointed at by mistake or on purpose is still cut off quickly with exit code 2.
- Operator note on refining again — Add to this story (findings from story 180's review): (F3) a race-detector test that runs Validate and Result.Output concurrently from several goroutines on one compiled Schema; (F4) expose the output-unit cap (DefaultMaxOutputUnits, Options.MaxOutputUnits) as a CLI flag in the same style as the other limits, with a CLI contract test that exceeding it gives exit 2, empty stdout and one stderr line. Also make every existing limit (number digits/exponent, pattern size, output units, annotations) follow the same option/flag/LimitError style. Keep everything else in the accepted plan.

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- Several file paths are created by stories not yet in the inventory and must be confirmed before coding. The Options location (options.go vs numberlimits.go) and the output-unit cap file (assumed root output.go) depend on story 180. The pattern translation file (assumed internal/ecmaregex/translate.go) depends on 182. The $ref/$dynamicRef keyword files (assumed internal/keywords/core/ref.go and dynamicref.go) depend on 186/188. The registry loader (assumed internal/registry/registry.go) depends on 187. The annotation cap location (assumed internal/eval/annotations.go) depends on 191.
- Whether existing CLI flag names for the number caps already match the unified --max-* naming is unknown; renaming them would break TestNumberCapFlags and must be decided at build time.
- Where maxInputBytes lives today (cmd/jsonschema/run.go vs main.go) is not visible in the inventory.
- The immutability test design (compare behaviour/limits before and after many validations, since compiled nodes hold func values that reflect.DeepEqual cannot compare) may need adjusting to the final Node shape.
- The story is at the 6-task and ~31-file ceiling. If the paths from dependency stories differ or add files, split it: (a) limit model + depth/size/ref-chain/work limits + library tests; (b) unifying existing caps + CLI flags + race proof.
- Dependency on 191 (annotations cap) and 182 (pattern size cap) exists only because the operator asked to unify those caps. If those stories ship without a cap, that part of criterion 9 is empty for them.

## Proposed outcome

See `SPEC.md` in this directory — the spec-agent's acceptance criteria are
the machine-readable statement of the outcome, derived from the problem above.

## Affected users and systems

See `PLAN.md` → *Files that change*, and the project's `project-spec.json`
for the systems in scope.

## Constraints

The project's `PROJECT_RULES.md` and `project-spec.json` (stack, conventions,
NFRs and the `forbidden` list) are injected into every agent and bind this work.

## Open questions

Recorded in `SPEC.md` → *Open questions*. A non-empty list parks the run at
Checkpoint 1 for the operator rather than failing it.
