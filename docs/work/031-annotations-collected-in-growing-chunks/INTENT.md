# INTENT #31

- Author: Joao Oliveira
- Date: 2026-10-17
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Replace the single growing annotation slice with chunked collection. Today Annotate appends to one slice that keeps growing, and these are the biggest single allocations during evaluation. Collect annotations in chunks whose sizes grow 16, 32, and so on, up to 4,096. Support rollback by mark and truncate. Flatten and sort only on the first call to Annotations() or Output(), guarded by sync.Once. The growing chunk sizes are required: fixed 512-entry chunks were measured to make small documents allocate 114% more. Add a randomised test that runs add, truncate and flatten against a plain slice and checks they agree. Annotations() and Output() must stay byte-identical. Validation behaviour, output, exit codes, limits, error messages and goroutine safety must not change at all. Use the standard library only. Do not remove or weaken any existing test. go test -race must pass. CLI stdout, stderr and exit code must stay byte-identical to the committed v1.1 golden files in flag, basic, detailed and verbose, for the perf fixtures, docs/examples and the testdata cases. Never regenerate or edit a golden file. If output differs, the story has failed and must stop. Targets, measured on an Apple M3 Max with Go 1.27.2 (on other hardware, compare relative to the before numbers): Validate5_5MB at most 342 MB B/op for this change on its own; ValidateErrors at most 27.4 MB B/op; ValidateSmall no worse than 42,000 B/op. Record before and after numbers from the in-repo benchmarks on the same machine in docs/benchmarks.md. Re-measure the allocation limits in allocs_test.go and lower them on purpose, as docs/benchmarks.md describes. If a target is missed, record the numbers and hold the story for the operator's decision.

## Operator clarifications
- The new test will try many random sequences of adding, undoing and reading back notes, then compare the results with a simple list. When it finds a mismatch, how should the random choices be made so the failure can be investigated? — Fixed starting seed, printed on failure — Recommended. Every run tries the same sequences, so a failure can always be repeated exactly. The seed is shown in the failure message.
- How much should the random test try on each ordinary test run? Larger runs find more problems but make 'go test' slower, including with -race. — About 10,000 random steps per run as in option 1, but the test must assert that it actually reached an undo across a chunk boundary and a chunk at the 4,096-entry cap at least once (about 8,200 entries are needed before the first capped chunk), adding a fixed growth sequence past the cap if the random walk alone does not get there. Must stay well under a second without -race.
- When an undo removes notes back past the start of a chunk, those chunks become empty. What should happen to them during the same check? — Keep them and reuse them for the next notes — Recommended. Avoids allocating the same memory again when a failed branch (for example inside anyOf or oneOf) is undone and evaluation continues. Most likely to meet the memory targets.
- Stories 1 to 3 will already be in the code when this story starts, so memory will already be lower before this change. How should the 'on its own' targets (342 MB, 27.4 MB, 42,000 B) be checked? — Compare with the code just before this story — Recommended. Measure the benchmarks right before and right after this change on the same machine, and check that the saving is at least what the target implies compared with v1.1 (for example, 5.5 MB at least as far below its before number as 342 MB is below the v1.1 number). Record both sets of numbers in docs/benchmarks.md.
- After the notes are sorted into one list the first time someone asks for them, the original chunks are no longer needed. What should happen to them? — Let them go after the first sort, as in option 1, but only inside the same sync.Once that builds the sorted list, so concurrent Annotations()/Output() callers never see a half-released state; covered by a concurrent -race test on one shared Result. Answers stay exactly the same.
- Splitting notes into chunks means a document that collects more than 16 notes now gets several smaller memory blocks instead of one growing block. On small documents, the number of separate memory requests (allocs/op) could go up slightly, even while total bytes go down. Your rule says the limits in allocs_test.go are re-measured and lowered. What should happen if a limit would need to go UP for this story? — Treat it as a missed target and hold the story — Recommended. The limit is never raised silently. The story records the before and after counts in docs/benchmarks.md and waits for your decision, like any other missed target.
- The note chunks could be kept and reused by the next document check, instead of being created fresh for each check. Reuse saves more memory, but chunks would then pass between checks that may run at the same time. Which should this story do? — Fresh chunks for every check — Recommended. Each check creates its own chunks and nothing passes from one check to the next, so running checks at the same time stays exactly as safe as today. The memory targets are expected to be met without reuse.
- Operator note on refining again — Operator note (one addition; everything else in the accepted plan stays):
Remove the stray file cmd/jsonschema/golden_dbg_test.go with `git rm`. Story #205 left it by mistake: it contains only the line "package main", holds no test and is referenced by nothing. Add it to this story's claims as the one file deleted, delete nothing else, and name it in the release notes. All other existing tests stay unchanged. Note: #206, #210 and #211 recorded their numbers in docs/benchmarks/r3-<n>-*.md and their lowered limits in allocs_<topic>_test.go; leave those files as they are. This story keeps writing allocs_test.go and docs/benchmarks.md as planned.
- Operator note on refining again — Operator note (numbers only; everything else in the accepted plan stays): "v1.1" in every target means the v1.1 release f35d9e4 (round 2 final), NOT the round-1 code (the round-1 tables in docs/benchmarks.md are history; do not use them). v1.1 values on the M3 Max: Validate5_5MB 368,958,349 B/op and 6,600,096 allocs/op; ValidateErrors 29,282,834 B/op; ValidateSmall 41,682 B/op and 785 allocs/op; BenchmarkCLI/5.5MB 499,252,608 B/op and 9,202,785 allocs/op; CLI peak RSS 5.5 MB -o basic about 195 MiB. Required drops versus the code just before the story therefore are:
- #208: Validate5_5MB at least 26,958,349 B/op (368,958,349 - 342,000,000); ValidateErrors at least 1,882,834 B/op (29,282,834 - 27,400,000); ValidateSmall may rise by at most 318 B/op (41,682 to 42,000) and allocs/op may not rise.
- #209: BenchmarkCLI/5.5MB at least 49,252,608 B/op (499,252,608 - 450,000,000) and 1,902,785 allocs/op (9,202,785 - 7,300,000); peak RSS at least 30 MiB lower than the code just before the story.
Record each computation with these sources and your exact measured values; never write placeholder digits or rounded guesses.

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- The inventory is truncated, so it is not confirmed where Result's fields and Output() live (jsonschema.go is assumed, alongside annotations.go). If Output() is in another root file, that file must be added to the claims via send-back.
- Current mark/rollback is assumed to be exposed by annotation_sink.go with signatures the chunk store can keep. If eval.go calls slice operations directly, eval.go (crowded) needs a write claim for a one-line wiring change.
- How 206's per-depth child contexts hand annotations to the parent is unverified. If they copy child slices into the parent sink, internal/keywords/registry.go may need a change, which this proposal does not claim.
- The ValidateSmall allocs/op 'may not rise' rule may fail if the small document collects more than 16 annotations; in that case the story is held per the operator rule.
- Release notes are under factory-owned docs/releases/. This story cannot write them; the factory must name the golden_dbg_test.go removal there.
- Story 207 is still 'started'. The 'before' baseline must be measured after 207 lands, and 207 also writes allocs_test.go and docs/benchmarks.md.

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
