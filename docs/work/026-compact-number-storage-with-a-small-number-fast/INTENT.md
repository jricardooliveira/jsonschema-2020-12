# INTENT #26

- Author: Joao Oliveira
- Date: 2026-10-17
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Reduce the memory used by decoded JSON documents. Today every document value gets its own heap node, and every number is stored with exact math/big arithmetic, which causes about 60 times memory growth over file size. Store values more compactly. Small numbers that an ordinary machine integer or float represents exactly, and that compare exactly, use those ordinary numbers. Exact math/big arithmetic is still used whenever needed: huge integers, decimals such as 0.1, and any multipleOf or comparison where an ordinary number could round. Number answers must stay exact. 1.0 is an integer, 0.3 is a multiple of 0.1, comparisons on huge integers are exact, and uniqueItems, enum and const deep equality treat 1 and 1.0 as equal. Add tests to every number rule that prove the answers stay exact on both the fast path and the exact path, and at the boundary between them. The rules are type integer, multipleOf, maximum, exclusiveMaximum, minimum, exclusiveMinimum, enum, const and uniqueItems. Results must stay identical everywhere: the same valid or invalid answer, the same flag, basic, detailed and verbose reports, and the same library Result. All existing tests must pass unchanged. Every hostile-input limit must keep working exactly as before. That covers nesting depth, input size, ref chain, evaluation-step charging, output units, report bytes and annotation count. Each must trigger at exactly the same inputs, with the same error and exit code 2. Add one regression test that proves this. The compiled Schema must stay immutable and safe for concurrent use, checked with go test -race. Run the project's benchmarks on the same machine before and after the change, and add both sets of numbers to docs/benchmarks.md, including the retained-heap metric. The change must show a reduction in allocs/op or B/op and in retained heap on the document scenarios. No scenario's allocs/op or B/op may increase.

## Operator clarifications
- Some tests already fail if checking a document uses more memory allocations than a set limit. After this change, the document checks will use fewer allocations, so the current limits will have spare room. Should this story lower those limits to match the new numbers? — No allocation-limit tests exist yet: story 'Allocation limits in tests' adds every limit once, after all five fixes, from the measured counts plus about 10% (the user's answer). This story adds or changes no allocation-limit test; it records its before and after benchmark numbers in docs/benchmarks.md.
- How big must the memory saving be for this story to count as done? Today a 5.5 MB document needs about 330 MB at peak, which is about 60 times its size. — Any clear reduction, no fixed target — Recommended. The story is accepted when the benchmarks show less memory and fewer allocations on the document checks, and nothing gets worse. Whether you now roughly match tekuri is still judged by your own comparison outside the project.
- Storing document values more compactly changes how the library holds a document inside. Programmers using the Go library could see this if they hand the library a document they have already read themselves, or read values back out of a result. May this story change what programmers see in the library? — No visible change for programmers — Recommended. Every public function and type stays exactly as it is today. The compact storage is hidden inside the library, so existing programs keep working without any change. This matches 'identical results everywhere'.
- This story must record a 'retained heap' number before and after the change: the memory still held while a decoded document is kept. If the project's benchmarks do not report that number yet, what should this story do? — Add it to the existing document benchmarks — Recommended. Each document benchmark also reports the memory still held after decoding (standard library only), measured the same way before and after. The before numbers come from running the new measurement on the old code.
- Your own baseline also measured the command-line tool's peak memory, for example 330 MB on the 5.5 MB document. Should docs/benchmarks.md also show the tool's peak memory before and after this change? — Yes, record the tool's peak memory too, for information only — Recommended. The 0.5 MB, 5.5 MB and 0.33 MB documents are run through the tool before and after, and peak memory is written in the docs. It does not decide whether the story passes, because it varies between runs.
- This story is judged on memory: fewer allocations and less memory used, and nothing may get worse on those. It does not say anything about speed. Storing numbers more compactly can make checking faster, but some paths could get a little slower. How should time per check count for this story? — No clear slowdown allowed — Recommended. Time per check is recorded before and after in docs/benchmarks.md. The story fails only if a scenario gets clearly slower, by more than 10%, which is about twice the normal ±5% variation in your timings. This protects your goal of matching tekuri on speed without failing over noise.
- Operator note on refining again — Narrow and correct the design before it is built. (1) Keep the decoded document's Go shape exactly as today: objects map[string]any, arrays []any, string, bool, nil and *jsonvalue.Number. Do not introduce a new Value struct or ordered member slices: every keyword, the evaluator, output and the library's caller-decoded path switch on these types, so that change would reach far beyond the claimed files and could add linear lookups (for example required or properties on huge objects). Drop the 'no heap node per value' part of AC1; this story is the user's fix 5, number storage. (2) Change only Number's internal representation. Exact fast path: a small exact decimal held inline (int64 mantissa without trailing zeros plus the int64 exponent), so 0.99, 0.01, 1.0 and 1e2 are native and exact; a big.Int mantissa only when the digits do not fit int64. Never use float64 to decide an answer. Cut per-number allocations (no big.Int allocation on the fast path), keep the canonical and original number text byte-identical in every report. (3) Keep every guarantee from story 179: never raise 10 to a large power (Cmp aligns mantissas only within the digit cap, Rat() stays lazy and counted, multipleOf keeps the exact decimal/modular method), the number digit and exponent caps and their LimitErrors stay identical, and the existing amplification and allocation-free-equality tests in internal/jsonvalue/number_test.go pass unchanged. Mixed fast/big operands compare, test equal and hash (uniqueItems) exactly; 1, 1.0 and 1e0 stay equal. (4) Keep the public methods of Number (Decimal, IsInteger, Rat, Canonical, String, Cmp, equality) with the same behaviour; edit keyword files only where a faster exact helper is used, and say which. Keep all other acceptance criteria: per-rule fast/big/boundary tests, the limits regression test, the race test, before/after benchmarks with the retained-heap metric, CLI peak memory for information, no public API change, no allocation-limit test, no slowdown over 10%.

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- The inventory is truncated. The validation keyword package path (assumed internal/keywords/validation/) and the exact keyword file names must be confirmed before coding.
- It is unknown whether the exact multipleOf method lives in jsonvalue or in multipleof.go. multipleof.go is claimed for writing in case it does; if not, it stays untouched and no keyword file is edited.
- It is unknown whether uniqueItems hashes numbers in equal.go or in its own keyword file. If the hash is in uniqueitems.go, that file also needs a write claim.
- It is unknown whether Number currently stores its original text or recomputes it. Byte-identical text may require keeping it, which bounds the saving.
- Story 198 (benchmarks and retained-heap metric) is still 'started'. This story needs its benchmarks and metric to record before/after numbers. If 198 does not add the retained-heap metric, this story would have to edit the benchmark files, which are not claimed here.
- The paths of the benchmark files from story 198 (for example bench_test.go at the root and cmd/jsonschema/bench_test.go) are not in the visible inventory.
- Stories 199 to 202 also write docs/benchmarks.md. Their order changes which 'before' numbers this story measures against.
- CLI peak memory depends on an external tool (for example /usr/bin/time -l on macOS) run by hand; it is information only.

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
