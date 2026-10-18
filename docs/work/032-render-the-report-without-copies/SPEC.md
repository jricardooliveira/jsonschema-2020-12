# SPEC #32: Render the report without copies (round 3, story 5)

- Date: 2026-10-17
- Type: performance
- Verdict: pass
- Status: passed gate 1 with nothing for the operator to decide (run #169, 2026-10-17 23:21 UTC)

## Problem

With -o basic on a valid 5.5 MB document, the CLI's peak memory comes from report rendering. Every JSON Pointer segment goes through a strings.Replacer call, even when it has no '~'. The report buffer also grows by doubling, so peak RSS and BenchmarkCLI/5.5MB B/op and allocs/op stay high.

## Why this matters

Round 3 aims to roughly match the operator's external reference on memory. This story targets the CLI rendering path. It must leave every output byte, error, exit code, limit and goroutine-safety guarantee unchanged, as fixed by the v1.1 golden files.

## Acceptance criteria

- JSON Pointer segment decoding in the rendering code shared by all output shapes (flag, basic, detailed, verbose) returns the segment unchanged and allocates nothing when it contains no '~'. It decodes '~1' to '/' and '~0' to '~' exactly as before when it does contain '~', including '~01' becoming '~1'. A new test proves the decoding matches the previous strings.Replacer result on table and random inputs. A new testing.AllocsPerRun test proves 0 allocations for a segment without '~'.
- Only the 'basic' report buffer is pre-sized. The size hint is computed from the location lengths, capped at the remaining report-byte limit, plus 1 byte for the trailing newline. With no limit or a very high limit, the hint is never larger than that location-based size, and no further cap is added. Other shapes keep their current buffering.
- The report-bytes limit fails at exactly the same point as in v1.1, with the same error message and exit code 2. New boundary tests prove this, added only and with no existing test changed. A limit equal to the exact report size passes. A limit one byte smaller fails with the same error and exit code 2. A report whose pointers contain '~' behaves the same near the limit. The tests take the exact size and the limit values from the code or from an unlimited run, never from hard-coded numbers.
- CLI stdout, stderr and exit code are byte-identical to the committed v1.1 golden files in testdata/golden/ in flag, basic, detailed and verbose, for the perf fixtures, docs/examples and the testdata cases. No golden file is regenerated or edited. If any output differs, the story fails and stops. No existing test is removed or weakened. go test -race ./... passes. go.mod is unchanged (standard library only).
- docs/benchmarks.md records the before medians (code just before this story, after stories 205 to 208) and the after medians on the same machine, from go test -run '^$' -bench . -benchmem -count=10 for the in-repo benchmarks. BenchmarkCLI/5.5MB must drop by at least 49,252,608 B/op (499,252,608 - 450,000,000) and at least 1,902,785 allocs/op (9,202,785 - 7,300,000) versus before. It must also end at no more than 450,000,000 B/op and no more than 7,300,000 allocs/op. Each computation is written out with the v1.1 f35d9e4 source values and the exact measured values, with no placeholder digits or rounded guesses.
- CLI peak RSS on the 5.5 MB document with -o basic is measured 5 times before and 5 times after with /usr/bin/time -l on the built CLI. TestCLIPeakMemory is also run once before and once after with its -peakmem flag. All runs are recorded in docs/benchmarks.md, and the median after is at least 30 MiB lower than the median before. No default test run depends on peak memory or timings.
- The allocation limits in allocs_test.go are re-measured after the change. Every limit the change lowers is set on purpose to the new measured count plus about 10% headroom, as docs/benchmarks.md describes, with the measured value recorded next to the test. No limit is raised.
- If any target in criteria 5 or 6 is missed, the numbers are still recorded in docs/benchmarks.md and the story is held for the operator's decision. It is never marked done silently.

## Non-goals

- Changing verbose's second evaluation pass
- Pre-sizing the buffers of flag, detailed or verbose reports
- An extra absolute cap on the size hint beyond the remaining report-byte limit
- Any change to validation, annotations, limits, error messages or exit codes
- Regenerating or editing golden files
- New dependencies or a second implementation of pointer rendering

## Tasks

- **T1** — Measure the before numbers
  - purpose: On the code as it stands after stories 205 to 208, before any production change, run the in-repo benchmarks with -count=10, run the built CLI 5 times with /usr/bin/time -l on the 5.5 MB document with -o basic, and run TestCLIPeakMemory -peakmem once. Record every raw value in docs/benchmarks.md.
  - scope: docs/benchmarks.md
  - done when: A docs/benchmarks.md story-209 'before' section with exact medians and all 5 RSS runs plus the TestCLIPeakMemory value
  - delivers: AC5, AC6
- **T2** — Decode a pointer segment only when it contains '~'
  - purpose: In the shared segment-rendering function, return the segment unchanged when strings.IndexByte(seg, '~') < 0. Otherwise keep the existing decode. Add a differential test against the old strings.Replacer and an AllocsPerRun 0-allocation test in a new file.
  - scope: internal/output/pointer.go, internal/output/pointer_tilde_test.go
  - done when: New tests pass, and the golden tests pass unchanged in all four shapes
  - depends on: T1
  - delivers: AC1, AC4
- **T3** — Pre-size the basic report buffer
  - purpose: Add a size-hint helper in a new file: the location-based size, capped at the remaining report-byte limit, plus 1 byte for the trailing newline. Call it with one line where the basic report buffer is created. Byte counting and the limit check are not changed.
  - scope: internal/output/sizehint.go, internal/output/sizehint_test.go, internal/output/basic.go, cmd/jsonschema/run.go
  - done when: sizehint_test proves the cap and the +1 byte, and the golden tests pass unchanged
  - depends on: T1
  - delivers: AC2, AC4
- **T4** — Report-bytes limit boundary tests
  - purpose: Add CLI tests in a new file. One sets the limit to the exact basic report size taken from an unlimited run and passes. One sets it one byte smaller and fails with the identical error and exit code 2. One repeats both with a schema and document whose instance and keyword locations contain '~'.
  - scope: cmd/jsonschema/reportbytes_boundary_test.go, testdata/reportbytes/tilde_schema.json, testdata/reportbytes/tilde_document.json
  - done when: New boundary tests pass, and the existing limit tests pass unchanged
  - depends on: T2, T3
  - delivers: AC3, AC4
- **T5** — Measure after, lower allocation limits, record and judge
  - purpose: Run go test -race ./... and the golden tests. Re-run the benchmarks with -count=10, run the CLI 5 times with /usr/bin/time -l, and run TestCLIPeakMemory -peakmem once. Lower the affected limits in allocs_test.go to the measured count plus about 10%. Record the before/after medians and the target computations from the v1.1 f35d9e4 sources in docs/benchmarks.md, and hold the story if any target is missed.
  - scope: allocs_test.go, docs/benchmarks.md
  - done when: A docs/benchmarks.md story-209 'after' section with exact computations, lowered limits recorded with measured values, and -race and golden tests green
  - depends on: T4
  - delivers: AC4, AC5, AC6, AC7, AC8
