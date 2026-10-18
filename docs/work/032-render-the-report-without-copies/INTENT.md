# INTENT #32

- Author: Joao Oliveira
- Date: 2026-10-17
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Lower the CLI's memory use when rendering reports. With -o basic on a valid 5.5 MB document, peak memory comes from report rendering: a strings.Replacer call for every JSON Pointer segment, and a report buffer that grows by doubling. Decode a segment only when it contains '~'. Give the report buffer a size hint computed from the location lengths, capped at the remaining report-byte limit, plus 1 byte for the trailing newline. The report-bytes limit must still fail at exactly the same point, with the same error and exit code 2. Validation behaviour, output, exit codes, limits, error messages and goroutine safety must not change at all. Use the standard library only. Do not remove or weaken any existing test. go test -race must pass. CLI stdout, stderr and exit code must stay byte-identical to the committed v1.1 golden files in flag, basic, detailed and verbose, for the perf fixtures, docs/examples and the testdata cases. Never regenerate or edit a golden file. If output differs, the story has failed and must stop. Targets, measured on an Apple M3 Max with Go 1.27.2 (on other hardware, compare relative to the before numbers): BenchmarkCLI/5.5MB at most 7.3 M allocs/op and at most 450 MB B/op; CLI peak RSS on the 5.5 MB document with -o basic at least 30 MiB lower than before. Measure peak RSS with the existing opt-in TestCLIPeakMemory in cmd/jsonschema (run with its -peakmem flag; unix only; skipped by default) and with /usr/bin/time -l on the built CLI. Record the before and after values in docs/benchmarks.md. No default test run may depend on peak memory. Also record before and after numbers from the in-repo benchmarks on the same machine. Re-measure the allocation limits in allocs_test.go and lower them on purpose, as docs/benchmarks.md describes. If a target is missed, record the numbers and hold the story for the operator's decision.

## Operator clarifications
- The peak-memory target says 'at least 30 MiB lower than before'. Stories 1 to 4 will already have lowered memory by the time this story starts. What should 'before' mean for this story? — The code just before this story — Recommended. Measure peak memory on the code as it stands after stories 1–4, then again after this story, on the same machine. The 30 MiB must come from this story's own change.
- There are two changes in this story. The first decodes a location only when it contains '~'. The second sizes the report up front. Which report shapes should get them? Output stays byte-identical either way. — '~' check in all shapes, sizing only for 'basic' — Recommended. The '~' check is in code all shapes share, so every shape benefits. Up-front sizing goes only where the target is measured ('basic'), which keeps the change small and low-risk.
- Peak memory changes a little from one run to the next, so a single run before and after could pass or miss the 30 MiB target by chance. How should this story decide if peak memory dropped by at least 30 MiB? — Median of 5 runs each — Recommended. Run the built CLI 5 times before and 5 times after with /usr/bin/time -l, and once with TestCLIPeakMemory -peakmem. Compare the middle values, record all runs in docs/benchmarks.md, and hold the story if the drop is under 30 MiB.
- The allocation and memory targets for the 5.5 MB benchmark (at most 7.3 M allocations and 450 MB) were set on v1.1. By the time this story starts, stories 1 to 4 will already have lowered both numbers, so they may already be met before any change. How should this story judge its benchmark targets? — Judged by the drop this story causes, as for stories 3, 4 and 6: before = the code just before this story, after = with it, same machine. BenchmarkCLI/5.5MB must drop by at least the v1.1 gap to its targets (v1.1 about 499 MB and 9.20 M allocs/op, so at least about 49 MB and 1.9 M allocs/op less), and the absolute 7.3 M / 450 MB must also hold. A smaller drop is recorded and the story is held for my decision.
- The report-size limit must still stop at exactly the same point as v1.1. How should this story prove that, beyond the golden files? — New boundary tests at the exact limit — Recommended. Add tests where the limit equals the report's exact size (passes), one byte less (fails with the same error and exit code 2), and a pointer containing '~' near the limit. These are added only; no existing test changes.
- When the caller raises the report-size limit very high, or there is no limit, the size hint would rely only on the location lengths. Should the up-front size be capped as well, so a huge report never reserves a large block of memory at once? — Use the location-based size, capped at the remaining limit only — Recommended. This is exactly what you asked for. The size is never larger than the report will really be, so nothing is wasted. No extra cap is needed.
- Operator note on refining again — Operator note (numbers only; everything else in the accepted plan stays): "v1.1" in every target means the v1.1 release f35d9e4 (round 2 final), NOT the round-1 code (the round-1 tables in docs/benchmarks.md are history; do not use them). v1.1 values on the M3 Max: Validate5_5MB 368,958,349 B/op and 6,600,096 allocs/op; ValidateErrors 29,282,834 B/op; ValidateSmall 41,682 B/op and 785 allocs/op; BenchmarkCLI/5.5MB 499,252,608 B/op and 9,202,785 allocs/op; CLI peak RSS 5.5 MB -o basic about 195 MiB. Required drops versus the code just before the story therefore are:
- #208: Validate5_5MB at least 26,958,349 B/op (368,958,349 - 342,000,000); ValidateErrors at least 1,882,834 B/op (29,282,834 - 27,400,000); ValidateSmall may rise by at most 318 B/op (41,682 to 42,000) and allocs/op may not rise.
- #209: BenchmarkCLI/5.5MB at least 49,252,608 B/op (499,252,608 - 450,000,000) and 1,902,785 allocs/op (9,202,785 - 7,300,000); peak RSS at least 30 MiB lower than the code just before the story.
Record each computation with these sources and your exact measured values; never write placeholder digits or rounded guesses.

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- The inventory is truncated: internal/output file names are not visible. The strings.Replacer may live in internal/output/pointer.go, internal/uri or internal/eval/location.go. The coder must confirm the real paths and adjust modules_affected before writing.
- It is unclear whether the basic report buffer and the trailing newline are created in internal/output/basic.go or in cmd/jsonschema/run.go. run.go is claimed for one wiring line and may turn out to be read-only.
- If the shared segment decoder lives outside internal/output (for example internal/uri, also used by $ref resolution), the '~' fast path also affects reference decoding. Output must stay identical, but the claim then moves to that file and the scope widens.
- The inventory shows no cmd/jsonschema/allocs_test.go. CLI allocation limits, if any, may live elsewhere. Only the root allocs_test.go is claimed.
- The 'before' baseline needs stories 207 and 208 merged. Story 207 is still in progress, so scheduling must wait for both.
- Whether the numeric targets are reachable from rendering alone is unknown until measured. A miss is recorded and the story is held for the operator.

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
