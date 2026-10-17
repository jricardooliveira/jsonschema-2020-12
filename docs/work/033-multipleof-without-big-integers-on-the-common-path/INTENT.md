# INTENT #33

- Author: Joao Oliveira
- Date: 2026-10-17
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Remove allocations from multipleOf on ordinary numbers. Today evalMultipleOf allocates about 6 big.Ints per evaluation. Expose the number's inline mantissa and exponent without allocating. Keep the divisor's small form from compile time. Decide divisibility with int64 arithmetic: a remainder check, trailing zeros for the factor 2, and a bounded loop for the factor 5. The math/big path stays for big mantissas or huge exponents. Exactness (ADR-004) is kept: integer arithmetic only, never floats. Add a random differential test that compares the small path against the big path over at least 1 million inputs. All existing multipleOf tests must stay unchanged and pass. Validation behaviour, output, exit codes, limits, error messages and goroutine safety must not change at all. Use the standard library only. Do not remove or weaken any existing test. go test -race must pass. CLI stdout, stderr and exit code must stay byte-identical to the committed v1.1 golden files in flag, basic, detailed and verbose, for the perf fixtures, docs/examples and the testdata cases. Never regenerate or edit a golden file. If output differs, the story has failed and must stop. Targets, measured on an Apple M3 Max with Go 1.27.2 (on other hardware, compare relative to the before numbers): testing.AllocsPerRun of evalMultipleOf on small numbers equals 0; Validate5_5MB at most 6.25 M allocs/op for this change on its own. Record before and after numbers from the in-repo benchmarks on the same machine in docs/benchmarks.md. Re-measure the allocation limits in allocs_test.go and lower them on purpose, as docs/benchmarks.md describes. If a target is missed, record the numbers and hold the story for the operator's decision.

## Operator clarifications
- The new random test compares the fast way against the exact way on at least 1 million numbers. Should every normal test run, including the slower race-checking run, do all 1 million? — Always 1 million, same random seed every run — Recommended. Every run checks the same 1 million cases, so a failure can always be repeated. The seed is printed when a case fails. Costs a few seconds per run.
- The zero-allocation check needs a clear meaning of 'small numbers'. Which numbers must be checked without any memory allocation? — Digits fit a 64-bit integer, exponent within ±18 — Recommended. Covers ordinary prices, counts and decimals such as 0.01 or 12345.678. Anything larger may use the exact big-number way and may allocate.
- Stories 1 to 5 already lowered allocations, so Validate5_5MB will likely be far under 6.25 M even before this story. How should 'at most 6.25 M for this change on its own' be judged? — By the drop this story causes — Recommended. Measure before and after this story on the same machine. The drop must be at least as large as the drop from the v1.1 count down to 6.25 M. Both numbers are written in docs/benchmarks.md.
- Operator note on refining again — Operator note (parallel build): stories #206, #207, #210 and #211 will be built AT THE SAME TIME after #205 is integrated, so none of them may write a file another of them writes. Change only this:
1. Do NOT write allocs_test.go or docs/benchmarks.md (shared files; #208 and #209 stay serial and keep writing them). Instead:
   a. Put this story's re-measured, lowered allocation limits in a NEW root-package test file allocs_<topic>_test.go (for example allocs_childctx_test.go, allocs_objectkeys_test.go, allocs_multipleof_test.go, allocs_firstload_test.go), reusing the existing helpers and race gating of allocs_test.go / allocs_race_on/off_test.go read-only. Each new limit = this story's measured value + about 10% headroom, measured value recorded next to it. The existing limits in allocs_test.go stay untouched (they remain valid, looser upper bounds); that is the deliberate lowering for this story.
   b. Record this story's before/after medians (go test -count=10, before = main after #205, after = with this story) and the target verdicts in a NEW file docs/benchmarks/r3-<story number>-<topic>.md.
2. Do not write any file claimed by the other three stories (#206: internal/keywords/registry.go and childctx*; #207: internal/keywords/applicator/*, internal/keywords/annotate.go, internal/keywords/recording.go, internal/eval/evaluated.go; #210: internal/jsonvalue/number_inline*, internal/keywords/validation/*; #211: internal/registry/*, internal/compile/metacache.go). If you need one at the gate, the answer will be no.
3. Benchmark timings taken while other stories build in parallel are noisy: record them, but the operator re-measures interleaved at Checkpoint 3 and judges the 5% rule on those numbers. B/op and allocs/op are exact and judged as before.
Everything else in the accepted plan stays unchanged.
- The 'no memory use' check covers small numbers. When a number fails multipleOf, the validator must also write an error entry, and writing that entry uses memory today. Which case must use no memory at all? — Only the yes/no decision — Recommended. Deciding 'is it a multiple' uses no memory, whether the answer is yes or no. Writing the error entry for a failing number may use memory as it does today, so error messages and reports stay exactly the same.
- The 1 million random cases compare the fast way with the exact way. Which numbers should they include? — Random mix plus fixed edge cases as in option 1, and also: at least a third of the random cases built as exact multiples (k times the divisor, still in the small range) so both yes and no answers are well covered; divisors with high powers of 2 and 5 (for example 0.125, 0.0625, 1024, 3125, 0.00032); negative divisors' absolute value rules as today; and the int64 and plus/minus 18 exponent boundaries on both instance and divisor.

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- The inventory is truncated: the multipleOf keyword file is assumed to be internal/keywords/validation/multipleof.go. If the existing file has another name (for example multiple_of.go), the coder edits that file instead, and the new files keep the multipleof_ prefix.
- How story 203 stores small decimals is not known from the inventory. If decimals are held as float64 or big.Rat rather than as a decimal mantissa/exponent, number_inline.go alone cannot expose them without allocating. internal/jsonvalue/number.go (and possibly decode.go) would then have to change, and these files are not claimed here. That would need a send-back to add the write claims (no other parallel story claims them).
- The v1.1 Validate5_5MB allocs/op baseline must be read from docs/benchmarks.md to compute the required drop (v1.1 count minus 6.25 M). This story cannot assume that number.
- Running 1,000,000 differential cases under -race may take well over a few seconds on the build machine. The operator chose 'always 1 million'; if the -race run becomes impractical, that is an operator decision, not something this story changes.
- It is not yet confirmed which root allocation scenarios change measurably from this story (most likely Validate 0.5 MB / 5.5 MB). Only scenarios with a measured drop get a lowered limit in allocs_multipleof_test.go.
- Timings measured while #206, #207 and #211 build in parallel are noisy. They are recorded only; the operator judges the 5% rule at Checkpoint 3.

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
