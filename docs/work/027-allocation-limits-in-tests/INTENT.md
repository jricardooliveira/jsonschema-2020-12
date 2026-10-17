# INTENT #27

- Author: Joao Oliveira
- Date: 2026-10-17
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Add ordinary Go tests, not benchmarks, that fail when checking a document or loading a schema uses more memory allocations than an agreed limit. Allocation counts are stable, unlike timings. Measure with testing.AllocsPerRun or runtime.MemStats, using the project's order schema and documents from testdata/perf and the fixed generation recipe. Guard these scenarios: compiling the schema; validating the small document; validating the 0.5 MB valid document; validating the 5.5 MB valid document; validating the 1,000-error document; and validating the 1,000-error document plus rendering the basic report. Set each limit from the count the improved code actually uses, plus about 10% headroom. Record each limit, and how it was measured, in docs/benchmarks.md. The tests use only the standard library, use no network, and run with go test ./.... Do not change validator behaviour.

## Operator clarifications
- Should the allocation checks count from the raw document text, so reading the JSON is included, or only the checking step on a document that is already read? — Include reading the document — Counts what a programmer actually pays per document, including how numbers are stored (fix 5). Catches slowdowns in either part.
- The 5.5 MB document makes these tests take noticeably longer on every 'go test ./...' run. Should the large-document checks always run? — Always run — Every run guards every scenario, as the request says. Costs a few extra seconds per test run.
- Some Go test modes add their own allocations and would break the limits. For example, the race checker ('go test -race') counts extra memory use that the validator did not cause. What should the allocation tests do in such modes? — Skip the limit check in those modes — Under -race the tests skip and say why. A normal 'go test ./...' still enforces every limit.
- Allocation counts can shift slightly between Go versions or between Mac and Linux. If a newer Go release changes the count, what should happen? — One limit for all; record the Go version measured — docs/benchmarks.md states the Go version and machine used. If a newer Go goes over the limit, the test fails and the limit is re-measured on purpose.
- These tests will count how many separate pieces of memory the validator asks for. They will not measure how much memory that adds up to. Your biggest complaint was memory: the 5.5 MB document needs about 330 MB. A change could keep the count the same but make each piece much bigger, and these tests would not notice. Should the tests also set a limit on the total memory used? — Total memory only for the large documents — Adds a memory limit only for the 0.5 MB, 5.5 MB and 1,000-error documents, where memory growth matters most. Small cases keep counts only.
- For the memory limit on the three large documents, which kind of memory should the test measure? — All memory asked for along the way — Adds up every piece requested, even pieces freed straight away, like the 74 MB figure in your baseline. This is steadier from run to run, but it does not show the peak directly.
- Memory totals wobble a little more between runs than allocation counts do. How much room should the memory limits have? — About 10%, the same as the counts — Recommended. One rule for every limit. If a limit proves flaky, it gets re-measured on purpose and docs/benchmarks.md records why.
- For the 'compiling the schema' check, should the count start from the raw schema text, so reading the schema file is included, as you chose for documents? — Yes, start from the raw schema text — Recommended. This is consistent with answer 1 and counts what a programmer actually pays to load a schema. Loading the built-in meta-schema is included too.
- Operator note on refining again — Also include these test-only items routed from earlier release reviews (no non-test Go file changes; results unchanged). From story 202 (tester F1-F4): (a) a test with a custom meta-schema supplied through the registry that uses unevaluatedProperties: false over an allOf $ref to the 2020-12 schema: a schema with an unknown keyword at the root, or nested under properties, is an invalid-schema error (exit code 2 through the CLI), and a schema using only listed keywords loads; (b) a test for a caller resource at the built-in 2020-12 meta-schema URI: a JSON-equal copy is accepted, a different document gives today's duplicate-address error; (c) a reliable very-first-Compile concurrency test: the test re-executes its own test binary in a fresh process (environment variable plus -test.run filter) where 32 goroutines released together call Compile then Validate on the order schema before anything else built the shared meta-schema set, and the parent fails on a non-zero exit or a race report under go test -race; (d) replace the vacuous assertion story 202's tester reported as F4 with a real one. From story 198: in docs/benchmarks.md the operator-baseline table says the operator's Mac uses 'a different machine and a different Go version'; the brief says Go 1.27.2, the same version as the build machine, so correct that sentence. Add the new test files to the claims.
- Operator note on refining again — One more test-only item, routed from story 203's release review (new medium finding): cmd/jsonschema/peakmem_test.go calls syscall.Getrusage with no build constraint, so 'GOOS=windows go vet ./cmd/jsonschema' fails with 'undefined: syscall.Getrusage' and the cmd/jsonschema tests do not compile on Windows. Put the peak-memory measurement behind a unix build constraint (//go:build unix) and give non-unix builds a stub that skips TestCLIPeakMemory with a clear message; the flag definitions must exist on every platform. Prove it with 'GOOS=windows go vet ./...' and 'GOOS=linux go vet ./...' both passing. Add cmd/jsonschema/peakmem_test.go and the new stub file to the claims. Everything else in the plan stays.

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- The inventory is truncated, so the exact fixture file names in testdata/perf and the generator function names in internal/perfdata are not confirmed. They are claimed for reading only.
- The file holding story 202's vacuous F4 assertion is not named in the input. It is most likely compile_cache_limits_test.go or internal/compile/metacache_test.go, so both are claimed. The coder must confirm which one from story 202's release review and leave the other untouched.
- It is unknown which test flags peakmem_test.go defines and whether it already re-executes the CLI. The new peakmem_flags_test.go is claimed to hold them. If no flags exist, that file may be unnecessary.
- A raceEnabled helper may already exist in the root package, for example in concurrency_race_test.go. If so, reuse it and drop the two new race files, or name them so they do not clash.
- The limits are measured on the factory build machine. Its OS and Go version may differ from the operator's Mac (Go 1.27.2), and per the operator's answer a newer Go or another OS may need a deliberate re-measure.
- If test (a), (b) or (c) fails on the current code, the story cannot fix it, because only test files may change. It must be reported and routed to a follow-up story.
- The story is near the size limit: 13 written files and 6 tasks. If the routed items grow, split the story 202/203 follow-ups (T4 to T6) into a separate test-only story.

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
