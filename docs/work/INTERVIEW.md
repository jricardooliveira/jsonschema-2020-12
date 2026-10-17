# Intake interview — Jsonschema 2020 12

1. **[goal] In your own words, what should this product do for you, and what problem does it solve? For example: check that files follow a set of rules, help people write those rules, or something else.**
   - Answer: A JSON Schema Draft 2020-12 validator written from scratch in Go: a Go library (programmatic API) plus a non-interactive command-line tool. Given a schema and a JSON instance it returns an unambiguous valid/invalid result with machine-readable error output, so that people and automated evaluators can check JSON documents against 2020-12 schemas fully offline. The problem it solves: a dependable, spec-faithful, self-contained validator whose validation engine is our own code (no existing JSON Schema library as a dependency), covering the mandatory behaviour of the 2020-12 core and validation specifications.

2. **[users] Who will mainly use this validator?**
   - Options offered: Developers and automated checkers / Developers only / Also non-programmers
   - Answer: Developers and automated checkers — Programmers calling the Go library, and scripts or test harnesses running the command-line tool

3. **[must_do] How should the validator treat "format" rules (for example checking that a text is a real email address or date)?**
   - Options offered: Only a note by default, checking can be switched on / Always check formats / Never check formats
   - Answer: Annotation only by default, exactly as 2020-12 requires: format values are collected as annotations and never fail validation. Format assertion is an explicit OPTIONAL mode: an opt-in library option and CLI flag, and also honoured when a meta-schema declares the format-assertion vocabulary as required. In assertion mode, check the formats defined in the 2020-12 validation spec (date-time, date, time, duration, email, idn-email, hostname, idn-hostname, ipv4, ipv6, uri, uri-reference, iri, iri-reference, uuid, uri-template, json-pointer, relative-json-pointer, regex) as far as practical, and document any format that is only partly checked. Unknown formats are never an error.

4. **[must_do] The 2020-12 rules allow a schema to point to other schemas by web address. Since it must work fully offline, how should those references be found?**
   - Options offered: Only local files or schemas given in advance / Allow downloading when asked / Same file only
   - Answer: No network access ever: the validator never fetches a remote URI on its own. A controlled schema registry: the caller supplies external schemas explicitly (individual files, or directories, each mapped to a base URI; via library API and CLI flags), schemas are found by their $id or the mapped URI after RFC 3986 resolution, and the official 2020-12 meta-schema and all its vocabulary meta-schemas (base https://json-schema.org/draft/2020-12/schema) are embedded in the binary so they work offline. A reference that cannot be resolved from the registry is a clear error (exit code for error, not invalid).

5. **[must_not_do] Should it support older versions of JSON Schema (like Draft 7 or 2019-09), or only 2020-12?**
   - Options offered: Only 2020-12 / 2020-12 plus 2019-09 / Several older drafts
   - Answer: Only 2020-12 — Keeps the work focused; older versions are out of scope

6. **[data] How should the command-line tool receive the schema and the JSON documents to check?**
   - Options offered: Files, plus standard input for the document / Files only / Files, standard input, and many documents in one run
   - Answer: Files, plus standard input for the document — Schema from a file path; one or more documents from file paths, or piped in through standard input. Fits scripts and test harnesses.

7. **[screens] What should the result look like when a document is checked? The 2020-12 rules describe standard report shapes: a simple yes/no ('flag'), a flat list of errors ('basic'), and nested reports ('detailed', 'verbose').**
   - Options offered: Standard report shapes, chosen with a flag / Only yes/no and a flat error list / Our own JSON error format
   - Answer: Standard report shapes, chosen with a flag — Output follows the 2020-12 report shapes as JSON. 'flag' and 'basic' always available, 'detailed' and 'verbose' as far as practical. One default, with a switch to pick another.

8. **[errors] Scripts tell the outcome apart by the tool's exit code. Which codes should it use?**
   - Options offered: 0 valid, 1 invalid, 2 tool or usage error / 0 valid, 1 invalid, separate codes per kind of failure / 0 valid, anything else not valid
   - Answer: 0 valid, 1 invalid, 2 tool or usage error — Usage error, unreadable file, broken JSON, bad schema or unresolved reference all give 2, so 'invalid document' is never confused with 'could not check'.

9. **[success] How will you know the validator is correct? The JSON Schema project publishes an official shared test suite that most validators are measured against.**
   - Options offered: Pass the official test suite for 2020-12 / Pass the official suite, including all optional tests / Our own tests only
   - Answer: Our own tests only, derived directly from the 2020-12 core and validation specification text: every mandatory keyword and behaviour gets spec-derived positive and negative tests written by this project, plus CLI contract tests (exit codes, output formats, stdin/file input, registry). Do NOT copy, vendor, download or reference the official JSON-Schema-Test-Suite or Bowtie or any copy of them in this project: correctness against those is measured later by an independent external evaluator, outside this project. No network at build or test time.

10. **[security] Could the schemas or JSON documents ever come from people you don't trust, for example files uploaded by strangers and checked by an automated service?**
   - Options offered: Yes, treat them as possibly hostile / Limits only when switched on / No, inputs are always trusted
   - Answer: Yes, treat them as possibly hostile: built-in limits (nesting depth, $ref/$dynamicRef recursion that never terminates, input size) stop a crafted file from hanging or crashing the tool; exceeding a limit is a clear error with exit code 2, never a silent valid/invalid. The defaults must be generous enough that ordinary, spec-conformant schemas and documents are never affected, and the caller can raise them via a library option and CLI flag.

11. **[security] When a schema points to another schema by a file-style address (for example file:///etc/...), may the validator read that file from disk on its own?**
   - Options offered: Never, only schemas supplied up front / Only inside the folders supplied / Yes, any local file
   - Answer: Never, only schemas supplied up front — It reads only the files and folders the caller listed. Anything else is an unresolved reference (exit code 2). Nothing on the machine is read by surprise.

12. **[security] Error reports can quote the actual piece of the document that failed, for example a wrong password value or a personal email. Should they?**
   - Options offered: Show only where it failed, not the value / Include the value, can be switched off / Always include the value
   - Answer: Show only where it failed, not the value — Reports give the location in the document and the rule that failed, never the data itself. Safe to save in logs.

13. **[technical] The validator has many separate parts: about 60 schema rules (keywords such as type, properties, $ref, unevaluatedProperties), 19 format checks, 4 report shapes, the schema registry, the limits and the command-line tool. How should the Go code be organised so these parts can be built side by side? In Go, the files every part would otherwise have to edit are a central list or switch of keywords, `main.go` and `go.mod`.**
   - Options offered: One package, one file per rule, a single registration table / Sub-packages per vocabulary / One central evaluator with a keyword switch / You decide
   - Answer: You decide — Uses the first option.

14. **[technical] The 2020-12 spec says the `pattern` and `patternProperties` keywords and the `regex` format use JavaScript (ECMA-262) regular expressions. Go's built-in `regexp` package uses a different syntax (RE2). RE2 is fast and safe against malicious input, but it does not support lookahead or backreferences, and some character classes behave differently. How should regular expressions be handled?**
   - Options offered: Go's standard regexp, with differences documented / Add a third-party ECMA-compatible regex library / Write our own ECMA-262 regex engine / You decide
   - Answer: Go's standard regexp, with differences documented — Recommended. It needs no dependency and cannot hang on a hostile pattern, which fits the untrusted-input answer. Patterns that RE2 cannot compile are reported as a schema error (exit code 2). Common ECMA syntax differences, such as \d meaning ASCII digits only, are handled by a small translation step, and the remaining gaps are listed in the docs. Cost: a few rare patterns, such as lookahead, are not supported.

15. **[technical] JSON documents can hold numbers that are too big or too precise for ordinary computer numbers, for example 12345678901234567890 or 0.1. This affects rules like 'multipleOf', 'maximum' and whether 1.0 counts as a whole number. A common shortcut rounds such numbers, so a document can get the wrong valid/invalid answer. How exact should number handling be?**
   - Options offered: Exact numbers using Go's built-in big-number support / Ordinary 64-bit numbers, with the limits documented / Ordinary numbers, falling back to exact ones only when needed
   - Answer: Exact numbers using Go's built-in big-number support — Recommended. Numbers are read exactly as written using the standard library (math/big), with no extra dependency. 'multipleOf 0.1', huge integers and 1.0 as an integer all give the correct answer. Cost: a little slower and more code in the number-related rules.

16. **[technical] Some format checks need large Unicode tables that Go's standard library does not include: 'idn-hostname' and 'idn-email' (international web and email addresses). The official extension package golang.org/x/net/idna has these tables. It is not a JSON Schema library, but it is an extra dependency in go.mod, downloaded once at build setup. Should we allow it?**
   - Options offered: Standard library only, international formats partly checked and documented / Allow golang.org/x/net/idna for international formats / Allow small official golang.org/x packages wherever they help
   - Answer: Standard library only, international formats partly checked and documented — Recommended. No dependency at all, so the build is fully self-contained and go.mod never changes. idn-hostname and idn-email get basic checks only, and the docs list what they miss. This affects only the optional format-checking mode.

17. **[technical] Answer 216 said the command-line tool has one default report shape, with a switch to pick another, but did not say which shape is the default. The default is what scripts and evaluators get when they pass no switch. Which should it be?**
   - Options offered: 'basic' (flat list of errors) / 'flag' (yes/no only) / 'detailed' (nested report)
   - Answer: 'basic' (flat list of errors) — Recommended. Always fully supported. It gives a yes/no plus each failing location and rule, without the data itself (per answer 221). Easy for scripts to read, and short enough for logs.

18. **[technical] The tool can check several documents in one run (answer 215). How should the results and the exit code look when there is more than one document, for example when two documents pass and one fails?**
   - Options offered: One JSON report per line, worst result sets the exit code / One combined JSON list, worst result sets the exit code / Stop at the first document that fails or errors
   - Answer: The first option (one JSON report per line, worst result sets the exit code: 2 if any document could not be checked, else 1 if any invalid, else 0), with one rule: when exactly one document is checked, stdout is exactly one 2020-12 output-format object for the chosen shape and nothing else, so an evaluator can parse it directly. The per-line label for multiple documents must not change the shape of the standard output unit inside it.

19. **[technical] How should programmers use the Go library? The command-line tool already checks several documents against one schema in a single run, so the library needs some way to load a schema and then check documents with it.**
   - Options offered: Load once, check many, safe to share / Load once, check many, one at a time / One call per check
   - Answer: Load once, check many, safe to share — Recommended. The caller loads and checks the schema once (registry, limits and format mode are fixed at that point), then uses it to check any number of documents. It can also be used from several goroutines at the same time. The tool works the same way. Cost: the loaded schema must never be changed after loading, which the design has to guarantee and test. (assumed — the operator said "you decide")

20. **[technical] Which Go version should the project require? This is the minimum version set in go.mod, and the version people need installed to build the library and the tool.**
   - Options offered: Go 1.26 or newer / Go 1.27 or newer / An older version, such as Go 1.22
   - Answer: Go 1.26 or newer — Recommended. Still supported by the Go team and widely installed, with all the standard-library features this project needs (math/big, regexp, Unicode tables). Builds on the current release too. (assumed — the operator said "you decide")

21. **[execution] May routine planning, coding, checks and bounded repairs proceed automatically within the agreed requirements? Batch launches, material changes, budget increases and releases will still require your decision.**
   - Options offered: Yes / Pause
   - Answer: Yes — Automate routine work within the agreed scope

22. **[correction] Correction to the proposed brief and technical approach**
   - Answer: Keep every answer, and add an explicit section 'Mandatory scope (2020-12 core + validation)' that the backlog must cover, nothing silently dropped: (1) all types and validation keywords with JSON number semantics (1.0 is an integer, exact multipleOf on decimals, big numbers): type, enum, const, multipleOf, maximum, exclusiveMaximum, minimum, exclusiveMinimum, maxLength/minLength (Unicode code points), pattern, maxItems, minItems, uniqueItems (deep JSON equality), maxContains, minContains, maxProperties, minProperties, required, dependentRequired; (2) applicators: properties, patternProperties, additionalProperties, propertyNames, dependentSchemas, prefixItems, items, contains, allOf, anyOf, oneOf, not, if/then/else; boolean schemas true/false; (3) core: $schema, $vocabulary (meta-schema validation of the schema, required vs optional vocabularies, unknown required vocabulary is an error), $id and base URI resolution per RFC 3986 including embedded resources, $anchor, $ref (local JSON Pointer with ~0/~1 escaping and percent-decoding, relative, external through the registry), $dynamicRef/$dynamicAnchor with dynamic scope, $defs, $comment; (4) annotation collection and unevaluatedProperties / unevaluatedItems across every applicator including through $ref and $dynamicRef; (5) unknown keywords ignored for validation and collected as annotations; (6) invalid-schema handling: a schema that fails the meta-schema is an error (exit 2), never 'invalid'; (7) output formats flag/basic (mandatory), detailed/verbose; (8) content* keywords (contentEncoding, contentMediaType, contentSchema) as annotations only; (9) format annotation-only by default. Add a section 'Optional features': format-assertion mode and vocabulary, detailed/verbose output. Also state as a hard constraint: no existing JSON Schema validation library may be a dependency (e.g. santhosh-tekuri/jsonschema, xeipuuv/gojsonschema, qri-io/jsonschema, google/jsonschema-go, kaptinlin/jsonschema); auxiliary libraries only if documented, prefer the standard library; document regex deviations from ECMA-262.

23. **[success] What result would make this performance work a success? Your measurements show about 2 times slower than tekuri, 3 times the memory and 10 times slower loading of the schema.**
   - Options offered: Clear targets against today's numbers / Roughly match tekuri on your comparison / Best effort, no fixed target
   - Answer: Roughly match tekuri on your comparison — Same speed and memory as tekuri, give or take. This is checked only with your separate comparison outside the project, because tekuri may never be part of it.

24. **[must_do] Your list has two kinds of fixes. Fixes 1 to 4 skip work nobody needs: building locations up front, recording evaluated properties, collecting annotations, and reloading the meta-schema. Fix 5 changes how numbers are stored, which is the bigger change and riskier for the exact-number rules (1.0 is a whole number, exact multipleOf). Which should this round include?**
   - Options offered: Fixes 1–4 now, numbers unchanged / Fixes 1–5, with a fast path for small numbers / Fixes 1–4, plus the verbose report without a second pass
   - Answer: Fixes 1–5, with a fast path for small numbers — Small numbers use ordinary computer numbers, and exact math is used whenever needed. Saves more memory on large documents. Costs extra tests in every number rule to prove the answers stay exact.

25. **[success] Must every result stay exactly the same as today? This means the same valid or invalid answer, the same reports in every shape, and the same Result from the library. Only the speed would change.**
   - Options offered: Yes, identical results everywhere / Library may skip annotations unless asked
   - Answer: Yes, identical results everywhere — Safest choice. Work is skipped only when nothing would ever show it, for example when the schema has no unevaluated* rule. All existing tests must pass unchanged.

26. **[success] How should the project itself guard the speed gains so they don't slowly get lost later? Your comparison with tekuri stays outside the project.**
   - Options offered: Benchmarks plus allocation limits in tests / Benchmarks only / Nothing in the project
   - Answer: Benchmarks plus allocation limits in tests — Go benchmarks in the project, using your order schema and documents (standard library only). Ordinary tests also fail if checking or loading a schema uses more memory allocations than an agreed limit. Allocation counts are stable, unlike timings.

27. **[must_do] The 'verbose' report checks the whole document a second time, so it takes about twice as long (140 ms instead of 70 ms on the 0.5 MB file). Your earlier choice of fixes 1–5 did not include changing this. Should this round also remove that second check?**
   - Options offered: Leave verbose as it is this round / Also build the verbose report in one pass
   - Answer: Leave verbose as it is this round — Keeps this round to the five fixes you chose. Verbose stays correct but about 2× slower; it can be improved in a later round.

28. **[success] Tests will fail if checking a document or loading a schema uses more memory allocations than an agreed limit. How should those limits be set?**
   - Options offered: Measured count after the fixes, plus about 10% headroom / Exactly the measured count, no headroom / Fixed targets based on tekuri's numbers
   - Answer: Measured count after the fixes, plus about 10% headroom — Each limit is set from what the improved code actually uses, with a little room. Catches real slowdowns without failing over tiny harmless changes.

29. **[success] The speed tests will use your order schema and documents (answer 26), but those files now sit in a temporary folder outside the project that can be wiped. The large ones are big: 0.5 MB, 0.33 MB with 1,000 errors, and 5.5 MB. How should the project keep them?**
   - Options offered: Copy the schema and small document in; generate the large ones / Copy every file into the project as it is / Copy the schema and small document only, with no large-document tests
   - Answer: Copy the schema and small document in; generate the large ones — The schema and the 0.5 KB document are saved in the project's testdata. The large documents are built by test code from a fixed recipe, so the results are the same on every run and the repository stays small.

30. **[correction] Operator note on a retried step**
   - Answer: Record of the user's round-2 amendment (the user's own measurements, so the brief keeps the baseline): one Mac, 16 cores, Go 1.27.2, one order schema with nested items using $ref, pattern, multipleOf, uniqueItems and additionalProperties; timings vary about +/-5%, allocation counts are reliable. Library baseline: small doc (~0.5 KB) per validation 56 us and 1,582 allocations; 500 KB valid doc 48 ms and 74 MB allocated; 330 KB doc with 1,000 errors 33 ms and 54 MB allocated; compiling the schema 1.1 ms and 28k allocations; throughput about 10 MB/s (parsing the 500 KB doc alone with encoding/json takes 7.7 ms). CLI baseline: 0.5 MB doc 70 ms, 39 MB peak; 5.5 MB doc 580 ms, 330 MB peak (about 60x the file size); 0.33 MB doc with 1,000 errors 50 ms, 32 MB peak; 0.5 MB doc with -o verbose 140 ms, 65 MB peak. Profiled causes: (1) keyword and absolute locations built for every keyword, even passing ones (about 30% of allocated memory); (2) evaluated properties always recorded (about 15%); (3) annotations always collected and sorted, splitting JSON Pointers on every comparison; (4) one heap node per document value with exact big-number arithmetic for every number (the 60x memory growth); Compile rebuilds the registry index and built-in meta-schemas, recompiles the 2020-12 meta-schema and collects and sorts annotations during the meta-schema check. The in-project benchmarks cover these same scenarios so before and after can be compared; the comparison with the other validator stays outside the project.

31. **[correction] Operator amendment**
   - Answer: Round 3: cut memory and time further, with the same answers. A profiling study (outside the factory, on v1.1) found where the allocations are and prototyped each change below. Every prototype kept all tests green, -race green and CLI output byte-identical in all 4 output formats. Build these, one story each, in this order. Validation behaviour, output, exit codes, limits, error messages and goroutine safety must not change at all. Standard library only, no new dependency, no other validator anywhere in the repo.

1. Reuse evaluation path frames.
- Today enterNode (internal/eval/refguard.go) allocates a pathFrame and returns a leave closure at every schema node, evalNode allocates a lazyLoc, and DynamicScope.push returns a closure. That is about 27% of what Validate allocates.
- Keep a per-call stack of pathFrame indexed by depth (the path is strictly last-in first-out), embed lazyLoc in the frame, and replace the closures with an explicit exit call. A frame is only valid while its node is being evaluated: document that invariant and add a nested $ref / if / allOf verbose-output test.
- Targets: BenchmarkValidate5_5MB B/op at most 280 MB and allocs/op at most 4.8 M; ValidateErrors B/op at most 22.5 MB; CompileOrder allocs/op at most 1,600; no scenario more than 5% slower.

2. One child evaluation context per depth.
- EvalChild (internal/keywords/registry.go) allocates a new EvalContext for every child: about 19% of the bytes.
- The per-call collector keeps one context per depth and resets it for each child. The caller still owns the returned units slice. Document that no keyword may keep its child context after EvalChild returns.
- Targets: Validate5_5MB B/op at most 305 MB on its own (at most 200 MB together with story 1); ValidateErrors at most 24.5 MB; ValidateSmall at most 35,500 B/op.

3. Object keywords without wasted copies.
- properties, additionalProperties and patternProperties sort a heap copy of the member names twice per object, grow the applied slice from 1, copy the names again for the annotation, and render the keyword location for the evaluated record even when that record is switched off.
- Sort into a small stack buffer (same byte order as today; heap only above 8 members), pre-size, annotate without a copy, and render the location only when the record is kept (a Recording() check). unevaluated* keywords must keep recording.
- Targets: Validate5_5MB B/op at most 345 MB and allocs/op at most 6.0 M on its own; annotation values and order unchanged.

4. Annotations collected in growing chunks.
- Annotate appends to one growing slice, the biggest single allocations in evaluation.
- Collect annotations in chunks whose sizes grow 16, 32, … up to 4,096, with rollback by mark and truncate. Flatten and sort only on the first Annotations() or Output() call, under sync.Once. Fixed 512-entry chunks were measured to make small documents allocate 114% more, so the growing sizes are required.
- Add a randomised test of add / truncate / flatten against a plain slice.
- Targets: Validate5_5MB B/op at most 342 MB on its own; ValidateErrors at most 27.4 MB; ValidateSmall no worse than 42,000 B/op; Annotations() and Output() byte-identical.

5. Render the report without copies.
- With -o basic on a valid 5.5 MB document, the CLI's peak memory comes from rendering the report: a strings.Replacer call for every pointer segment, plus a report buffer that grows by doubling.
- Decode a segment only when it contains "~". Give the report buffer a size hint from the location lengths, capped at the remaining report-byte limit, plus 1 byte for the trailing newline. The report-bytes limit must still fail at exactly the same point.
- Targets: BenchmarkCLI/5.5MB allocs/op at most 7.3 M and B/op at most 450 MB; CLI peak RSS on the 5.5 MB document with -o basic at least 30 MiB lower than today.

6. multipleOf without big integers on the common path.
- evalMultipleOf allocates about 6 big.Ints per evaluation.
- Expose the number's inline mantissa and exponent without allocating, keep the divisor's small form from compile time, and decide divisibility with int64 arithmetic: remainder, trailing zeros for the factor 2, a bounded loop for the factor 5. The big path stays for big mantissas or huge exponents. Exactness (ADR-004) is kept: integers only, never floats.
- Add a random differential test, small path against big path, over at least 1 M inputs.
- Targets: testing.AllocsPerRun of evalMultipleOf on small numbers = 0; Validate5_5MB allocs/op at most 6.25 M on its own; all existing multipleOf tests unchanged.

7. Build the built-in meta-schema index once.
- compile/metacache.go calls registry.BuildBuiltinIndex() although registry.AddBuiltins already built and shared the index (registry.go). That second build costs about 165 µs and 118 KB per process.
- Use one shared builder. Add a counter test showing exactly one built-in index build per process. FirstLoadOrder rebuilds on purpose, so add a separate first-compile measure.

For every story:
- no existing test removed or weakened;
- go test -race green;
- the CLI's output (stdout, stderr, exit code) byte-identical to v1.1 on the perf fixtures, docs/examples and the testdata cases, in flag, basic, detailed and verbose;
- benchmark numbers before and after from the in-repo benchmarks on the same machine;
- allocation limits in allocs_test.go re-measured and lowered on purpose, as docs/benchmarks.md says.

The thresholds come from one Apple M3 Max with Go 1.27.2; on other hardware compare relative to the before numbers.

32. **[success] Each of the seven stories has number targets, for example 'Validate5_5MB at most 280 MB'. These numbers come from your M3 Max. If a story keeps every result identical and clearly cuts memory, but misses its target on the build machine, what should happen?**
   - Options offered: Report it and let you decide / Hard gate: the story is not done until every target is met / Targets are guidance only
   - Answer: Report it and let you decide — The story records its before and after numbers and is held for your decision when a target is missed; nothing is merged as done silently

33. **[success] Every story must keep the CLI output (stdout, stderr, exit code) byte-identical to v1.1. What should each story compare against?**
   - Options offered: Golden files recorded once from v1.1 / Build v1.1 from its git tag during the check / Rely on the existing tests only
   - Answer: Golden files, as in option 1, but captured from the unchanged v1.1 code (f35d9e4) in the first task of story 1, before any production code changes, and committed with that story. Generated large documents stored as size plus SHA-256. No later story may regenerate or edit a golden file; a story whose output differs from a golden file has failed and must stop, never update the golden. The existing tests stay as they are.

34. **[success] Story 5 asks for CLI peak memory on the 5.5 MB document with -o basic to be at least 30 MiB lower. Peak memory is measured by the operating system, not by Go tests. How should it be checked?**
   - Options offered: Documented manual measurement / An automated test that runs the CLI and reads peak memory / Leave it to your outside comparison
   - Answer: Documented measurement as in option 1, using the existing opt-in peak-memory test in cmd/jsonschema (TestCLIPeakMemory with its -peakmem flag, unix only, skipped by default) and /usr/bin/time -l on the built CLI; record the before and after values in docs/benchmarks.md. No default test run may depend on peak memory. The operator re-measures it before release.

35. **[success] Story 1 says 'no scenario more than 5% slower', but your timings vary by about ±5% from run to run. How should 'slower' be judged?**
   - Options offered: Repeated runs, compare the medians / Use the benchstat tool / Single run, rerun on a miss
   - Answer: Repeated runs, compare the medians — Run each benchmark about 10 times before and after with go test -count=10 and compare medians by hand or with a small script; only the standard toolchain is used, and the result is reported, not asserted in a test
