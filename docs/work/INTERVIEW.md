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
