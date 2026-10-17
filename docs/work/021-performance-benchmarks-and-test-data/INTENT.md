# INTENT #21

- Author: Joao Oliveira
- Date: 2026-10-17
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Add in-project Go benchmarks so the validator's speed and memory can be measured before and after the performance work. Use only the standard library. Store this order schema exactly in testdata/perf: {"type": "object", "required": ["id", "items"], "properties": {"id": {"type": "integer", "minimum": 1}, "name": {"type": "string", "maxLength": 100}, "tags": {"type": "array", "items": {"type": "string"}, "uniqueItems": true}, "items": {"type": "array", "items": {"$ref": "#/$defs/item"}}}, "$defs": {"item": {"type": "object", "required": ["sku", "qty"], "properties": {"sku": {"type": "string", "pattern": "^[A-Z]{3}-[0-9]+$"}, "qty": {"type": "integer", "minimum": 0}, "price": {"type": "number", "multipleOf": 0.01}}, "additionalProperties": false}}}. Also store a small valid document of 10 items (492 bytes) in testdata/perf, built from the recipe below. Test code builds the large documents from a fixed recipe, so every run gets the same bytes and the repository stays small. Valid document recipe: {"id": 1, "name": "order", "tags": ["a", "b", "c"], "items": [...]}, where item i (counting from 0) is {"sku": "ABC-i", "qty": i, "price": i.99}. Build it with 10,000 items (about 0.5 MB) and with 100,000 items (about 5.5 MB). Error document recipe: {"id": 1, "items": [...]} with 10,000 items {"sku": "ABC-i", "qty": i}, except that every item whose i is divisible by 10 has sku "bad". This gives 1,000 pattern errors in about 0.33 MB. Write testing.B benchmarks, run with go test -run '^$' -bench . -benchmem. Cover: compiling the schema; validating the small, 0.5 MB and 5.5 MB valid documents; validating the 1,000-error document alone; and validating the 1,000-error document plus rendering the basic report. Each document benchmark includes parsing the JSON and calls b.SetBytes with the document size so MB/s is reported. Add in-process CLI benchmarks for the 0.5 MB and 5.5 MB documents that call the CLI's run entry point, with no subprocess. Report the retained heap of a decoded 0.5 MB and 5.5 MB document as a custom metric with b.ReportMetric, measured with runtime.ReadMemStats after runtime.GC, as the in-repo stand-in for peak memory. Write docs/benchmarks.md with the exact command, how to read the numbers, and the baseline numbers measured by this story before any fix. The benchmarks measure only this project: no other validator may appear in go.mod, go.sum or any test. No network is used. Do not change validator behaviour.

## Operator clarifications
- Your own measurements included two more command-line runs: the 1,000-error document (50 ms) and the 0.5 MB document with the verbose report (140 ms). The brief says the project's benchmarks should cover the same cases as your measurements, but this request lists command-line benchmarks only for the 0.5 MB and 5.5 MB documents. Should those two runs get command-line benchmarks too? — Yes, add both — Recommended. Every number in your measurements then has a matching benchmark in the project, so before and after can be compared one to one. It costs two more small benchmarks.
- When a command-line benchmark runs, how should the tool receive the document? — As a file path — Recommended. The test writes the generated document to a temporary file once, before timing starts. Each run then reads it like a script would, and the report is thrown away so printing does not distort the numbers.
- The request has the tool's run entry point called directly, with no separate program started. If the tool does not yet have an entry point that a test can call with its own input and output, what should this story do? — Add a thin entry point, no behaviour change — Recommended. The existing tool logic is moved behind a function that takes arguments, input and output. The real program calls that function, so exit codes and output stay identical, and all existing tests must still pass.
- Should ordinary tests (not benchmarks) check the benchmark inputs, so the measurements cannot silently change what they measure? — Yes, a few quick checks — Recommended. The tests check that the stored small document is exactly the 492 bytes the recipe produces. They also check that the generated valid documents pass and that the error document gives exactly 1,000 pattern errors. They run in a fraction of a second.
- The benchmark will report how much memory a decoded 0.5 MB and 5.5 MB document keeps. Which decoded form should it measure? — The validator's own form — Recommended. It measures the document as the validator holds it while checking. Fix 5 (how numbers are stored) changes exactly this, so the before and after numbers show its effect.
- docs/benchmarks.md will record baseline numbers from before any fix. This story measures them on the build machine, which is not your Mac, so its timings will differ from yours. What should the document contain? — This story's numbers plus yours — Recommended. It records this story's measured numbers with the machine, Go version and run count. Your Mac numbers from the brief go in a separate, labelled table, so both baselines are kept and never mixed up.
- Timings on one machine can move by about 5% from run to run. When this story records its baseline numbers in docs/benchmarks.md, how should it measure them so later rounds can be compared fairly? — Run each benchmark 10 times, record the middle value and the spread — Steadier numbers and shows how much they vary. Takes about twice as long to measure, mostly because of the 5.5 MB runs.

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- The inventory shows cmd/jsonschema/run.go and run_test.go, so a testable run entry point very likely exists already. That is why run.go and main.go are claimed read-only. If inspection shows the entry point cannot take arguments, input/output streams and return an exit code, the operator-approved thin entry point must be added. run.go and main.go then become write claims, and they are crowded files (8 and 5 stories), so the scheduler must be told before coding.
- Which function decodes into the validator's own form (internal/jsonvalue Decode or similar) and whether the root package exposes a way to parse plus validate in one call was not visible from the truncated inventory. The retained-heap benchmark assumes jsonvalue has an exported decode function.
- The basic report may also contain parent units (items, $ref, properties) next to the 1,000 pattern units. The check counts units whose keyword location ends in /pattern, which is how 'exactly 1,000 pattern errors' is interpreted.
- cmd/jsonschema/docs_examples_test.go and docs_flags_test.go may scan docs/*.md. If so, the command blocks in docs/benchmarks.md must be written in a form those tests accept, and this was not verified.
- The baseline numbers depend on the build machine. Later stories (199–203) must rerun on the same machine to compare, as their own requests require.
- The order schema is stored as given. Whether it is stored byte for byte or pretty-printed does not affect results, but 'exactly' is read as unchanged content. The coder should store it as the single-line text from the request.

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
