# Backlog — Jsonschema 2020 12

| # | Story | Status | Evidence | Runs |
|---|---|---|---|---|
| 1 | Walking skeleton | released | [001-walking-skeleton/](001-walking-skeleton/) | 134 |
| 2 | Exact JSON values, type, enum and const | released | [002-exact-json-values-type-enum-and-const/](002-exact-json-values-type-enum-and-const/) | 135 |
| 3 | Error reports: flag and basic output, multiple documents | released | [003-error-reports-flag-and-basic-output-multiple/](003-error-reports-flag-and-basic-output-multiple/) | 136, 137 |
| 4 | Number rules | released | [004-number-rules/](004-number-rules/) | 139 |
| 5 | String rules | released | [005-string-rules/](005-string-rules/) | 138 |
| 6 | Object rules | released | [006-object-rules/](006-object-rules/) | 141 |
| 7 | Array rules | released | [007-array-rules/](007-array-rules/) | 140 |
| 8 | Combining schemas | released | [008-combining-schemas/](008-combining-schemas/) | 142 |
| 9 | Base URIs, anchors and local references | released | [009-base-uris-anchors-and-local-references/](009-base-uris-anchors-and-local-references/) | 143 |
| 10 | Schema registry for external references | released | [010-schema-registry-for-external-references/](010-schema-registry-for-external-references/) | 145 |
| 11 | Dynamic references | released | [011-dynamic-references/](011-dynamic-references/) | 146, 147 |
| 12 | Embedded meta-schemas and schema validation | released | [012-embedded-meta-schemas-and-schema-validation/](012-embedded-meta-schemas-and-schema-validation/) | 148 |
| 13 | $vocabulary handling | to do |  |  |
| 14 | Annotation collection | released | [014-annotation-collection/](014-annotation-collection/) | 144 |
| 15 | unevaluatedProperties and unevaluatedItems | to do |  |  |
| 16 | Limits against hostile input | to do |  |  |
| 17 | Detailed and verbose output | to do |  |  |
| 18 | Format assertion: mode, dates, email, hostnames, IP and UUID | to do |  |  |
| 19 | Format assertion: URI, IRI, templates, pointers and regex | to do |  |  |
| 20 | User documentation: CLI contract and scope | to do |  |  |

1. **Walking skeleton** — released
   Create the smallest end-to-end version of the JSON Schema Draft 2020-12 validator in Go. Set up a single Go module requiring Go 1.26 or newer, with no dependencies beyond the standard library. No existing JSON Schema validation library may ever be a dependency. Go packages pick up new files by themselves, so adding a file never requires editing a shared manifest. Provide a library package with a load-once, check-many API: the caller loads a schema once, then checks any number of JSON documents with it. At this stage the API accepts only the boolean schemas true and false and the empty object schema {}. Give the library ONE registration point for keywords: each schema keyword lives in its own file and registers itself with a central keyword registry, so later keywords add files instead of editing a central list or switch. Provide a non-interactive command-line tool whose main.go only wires flags to the library. It takes a schema from a file path and one document from a file path or from standard input. It prints a minimal JSON result saying valid or invalid. Exit codes are 0 for valid, 1 for invalid, and 2 for any usage error or unreadable file. Include tests for the library and for the CLI exit codes. The tests use no network.
   Why: Proves the module, library, CLI and keyword registration point all work before any keyword depends on them.

2. **Exact JSON values, type, enum and const** — released
   Read schemas and documents with an exact JSON value model built only on the standard library. Numbers are kept exactly as written using math/big. This means huge integers like 12345678901234567890 and decimals like 0.1 lose no precision, and 1.0 counts as an integer. Broken JSON in a schema or a document is an error with exit code 2, never 'invalid'. Add deep JSON equality: numbers compare by mathematical value, objects compare regardless of key order, and arrays compare element by element. Then add the 2020-12 validation keywords 'type' (string or array form, all seven JSON types), 'enum' and 'const', each in its own file registered through the keyword registry. Write positive and negative tests derived from the 2020-12 validation specification text.
   Why: Every later rule needs exact numbers and equality, and type/enum/const are the simplest rules that exercise them.

3. **Error reports: flag and basic output, multiple documents** — released
   Make the validator produce the standard 2020-12 output formats 'flag' and 'basic' as JSON, in both the library and the CLI. The CLI picks the shape with an output-format option, and 'basic' is the default. Every reported error gives the instance location, the keyword location, the absolute keyword location where one is known, and which rule failed. A report never quotes the document's values, so it is safe to save in logs. Give output formats one registration point so later shapes can be added as new files. The CLI accepts several document paths in one run. When exactly one document is checked, stdout is exactly one 2020-12 output object for the chosen shape and nothing else. When several documents are checked, the CLI prints one JSON report per line. Each line carries a label for its document, and the label does not change the standard output object inside the line. The worst result sets the exit code: 2 if any document could not be checked, otherwise 1 if any document is invalid, otherwise 0. Add CLI contract tests for both shapes, stdin input, multiple documents and exit codes.
   Why: Fixes the machine-readable result contract early so every keyword story reports errors the same way.

4. **Number rules** — released
   Add the 2020-12 numeric validation keywords multipleOf, maximum, exclusiveMaximum, minimum and exclusiveMinimum. Each keyword lives in its own file and registers through the keyword registry. All comparisons use the exact number model. For example, 0.3 is a multiple of 0.1, comparisons on huge integers are exact, and 1.0 behaves as the integer 1. The keywords apply only to number instances and ignore every other type. Write spec-derived positive and negative tests, including decimal and big-number cases.
   Why: Isolated numeric rules that only need the value model and can be built alongside the other assertion stories.

5. **String rules** — released
   Add the 2020-12 string keywords maxLength, minLength and pattern. Each keyword lives in its own file and registers through the keyword registry. Lengths count Unicode code points, not bytes. The 'pattern' keyword uses Go's standard regexp package through a small translation step that handles common ECMA-262 differences, for example \d meaning ASCII digits only. A pattern the translation step or RE2 cannot handle, such as lookahead or a backreference, is a schema error with exit code 2. Put the translation step in its own reusable file so patternProperties and the regex format can use it later. The keywords apply only to string instances. Write spec-derived positive and negative tests, including multi-byte characters and translated patterns.
   Why: Delivers the string rules and the shared regex translation that object rules and formats later reuse.

6. **Object rules** — released
   Add the 2020-12 object keywords properties, patternProperties, additionalProperties, propertyNames, dependentSchemas, maxProperties, minProperties, required and dependentRequired. Each keyword lives in its own file and registers through the keyword registry. patternProperties uses the existing ECMA-262-to-RE2 translation step. additionalProperties applies only to properties not matched by properties or patternProperties in the same schema object. The keywords apply only to object instances. Errors report the failing instance location without quoting values. Write spec-derived positive and negative tests for every keyword.
   Why: Groups the object keywords, which depend on each other, so they can be built in parallel with the array and combining rules.

7. **Array rules** — released
   Add the 2020-12 array keywords prefixItems, items, contains, maxContains, minContains, maxItems, minItems and uniqueItems. Each keyword lives in its own file and registers through the keyword registry. 'items' applies only to elements after those covered by prefixItems. maxContains and minContains work with contains, and minContains 0 makes contains always pass. uniqueItems uses deep JSON equality, so 1 and 1.0 are duplicates and objects compare regardless of key order. The keywords apply only to array instances. Write spec-derived positive and negative tests for every keyword.
   Why: Groups the array keywords, which depend on each other, as a separate part that can be built at the same time as the object rules.

8. **Combining schemas** — released
   Add the 2020-12 applicator keywords allOf, anyOf, oneOf, not, and if/then/else. Each keyword lives in its own file and registers through the keyword registry. 'then' and 'else' have no effect without 'if', and a failure inside 'if' alone never makes the document invalid. Boolean schemas true and false work as subschemas everywhere. Write spec-derived positive and negative tests, including nested combinations and oneOf with more than one match.
   Why: Completes the in-place applicators, which annotation and unevaluated handling later rely on.

9. **Base URIs, anchors and local references** — released
   Add 2020-12 core identification and referencing within one schema document. '$id' sets the base URI using RFC 3986 resolution, including for embedded schema resources with their own $id. '$anchor' defines plain-name fragments. '$defs' holds reusable subschemas. '$ref' resolves local JSON Pointer fragments, which includes ~0/~1 escaping and percent-decoding. It also resolves anchor fragments and relative references against the current base URI. '$ref' works alongside sibling keywords. '$comment' is accepted and has no effect on validation. A reference that cannot be resolved is an error with exit code 2, never 'invalid'. The validator never fetches anything from the network or from disk on its own. Write spec-derived positive and negative tests.
   Why: Referencing inside one document is the base that the registry, dynamic references and meta-schemas all build on.

10. **Schema registry for external references** — released
   Add a controlled schema registry. Through a library option and CLI flags, the caller supplies extra schemas as individual files or as whole directories, each mapped to a base URI. '$ref' to another resource is found by that resource's $id or by its mapped URI, after RFC 3986 resolution. The validator never accesses the network and never reads any file or folder the caller did not list, including file:// addresses. Anything not found in the registry is an unresolved-reference error with exit code 2. The registry is fixed when the schema is loaded. Add library tests and CLI contract tests for file mapping, directory mapping and unresolved references.
   Why: Delivers offline external references as soon as local referencing exists, and is independent of the keyword stories.

11. **Dynamic references** — released
   Add 2020-12 '$dynamicRef' and '$dynamicAnchor' with dynamic scope, as the core specification describes. A $dynamicRef whose initial target holds a matching $dynamicAnchor resolves to the outermost matching $dynamicAnchor in the dynamic scope. Otherwise it behaves like $ref. This works across embedded resources and across schemas supplied through the registry. Write spec-derived positive and negative tests, including recursive extension of a schema.
   Why: Must come before meta-schema validation, because the official meta-schemas reach every subschema through "$dynamicRef": "#meta".

12. **Embedded meta-schemas and schema validation** — released
   Embed the official 2020-12 meta-schema and all of its vocabulary meta-schemas, under base https://json-schema.org/draft/2020-12/schema, in the binary so they work offline. Honour '$schema'. Before checking any document, validate every loaded schema against its meta-schema. A schema that fails is an invalid-schema error with exit code 2, never 'invalid'. This story depends on dynamic references: the official meta-schemas reach every subschema through "$dynamicRef": "#meta", so validation against them is incomplete without $dynamicRef. Only 2020-12 is supported. Write tests that use valid and invalid schemas, run through both the library and the CLI.
   Why: Placed after dynamic references so meta-schema validation is complete, and before $vocabulary, which reads these meta-schemas.

13. **$vocabulary handling** — to do
   Add '$vocabulary' support as 2020-12 core describes. When a schema's meta-schema declares vocabularies, the validator processes them. A required vocabulary it does not know is an error with exit code 2. An optional unknown vocabulary is ignored. The format-assertion vocabulary is recognised so a later story can honour it. Use the embedded meta-schemas and custom meta-schemas supplied through the registry. Write spec-derived tests for required, optional and unknown vocabularies.
   Why: Builds directly on the embedded meta-schemas and prepares the hook that format assertion needs.

14. **Annotation collection** — released
   Collect 2020-12 annotations during evaluation, with their instance and keyword locations. Annotations are dropped from any subschema that fails. Collect these: annotation keywords such as title, description, default, deprecated, readOnly, writeOnly and examples; 'format' values, which by default are annotations only and never fail validation; contentEncoding, contentMediaType and contentSchema, which are annotations only; and unknown keywords, which are ignored for validation and collected as annotations. Unknown formats are never an error. Applicators record which properties and items they evaluated. Annotations appear in the output formats wherever the 2020-12 shapes include them. Write spec-derived tests.
   Why: Annotation tracking is what unevaluatedProperties and unevaluatedItems depend on, so it comes first.

15. **unevaluatedProperties and unevaluatedItems** — to do
   Add the 2020-12 'unevaluatedProperties' and 'unevaluatedItems' keywords, each in its own file and registered through the keyword registry. They use the collected annotations from every adjacent applicator and from nested ones: properties, patternProperties, additionalProperties, prefixItems, items, contains, allOf, anyOf, oneOf, not, if/then/else, dependentSchemas, $ref, $dynamicRef, and nested unevaluated keywords. Only successful subschemas count. Write spec-derived positive and negative tests, including cases that go through $ref and $dynamicRef.
   Why: Needs every applicator, references and annotation collection in place, so it comes after all of them.

16. **Limits against hostile input** — to do
   Add built-in limits on nesting depth, input size, and $ref/$dynamicRef recursion that never terminates, so a crafted schema or document cannot hang or crash the tool. Exceeding a limit is a clear error with exit code 2, never a silent valid or invalid result. The defaults are generous enough that ordinary, spec-conformant schemas and documents are never affected. The caller can raise each limit through a library option and a CLI flag. Also prove the load-once guarantee: a loaded schema is never changed after loading and is safe to use from several goroutines at the same time. Cover this with tests run under the race detector. Add tests for each limit and CLI contract tests for the limit flags.
   Why: Hardens the finished engine against untrusted input once all recursive features exist.

17. **Detailed and verbose output** — to do
   Add the 2020-12 'detailed' and 'verbose' output formats, as far as practical, as new formats registered through the output-format registration point. Both the library and the CLI output-format option can select them. The nested reports carry the instance location, keyword location and absolute keyword location, and they include annotations where the shapes call for them. They never quote document values. When one document is checked, stdout is exactly one output object. With several documents, each line holds the unchanged output object. Document in the code any part of the shapes that is only partly produced. Write tests for both shapes.
   Why: An optional feature that plugs into the existing output registration point without touching the keywords.

18. **Format assertion: mode, dates, email, hostnames, IP and UUID** — to do
   Add the optional format-assertion mode. It is enabled by an opt-in library option, by a CLI flag, or when a meta-schema declares the format-assertion vocabulary as required. In this mode, 'format' fails validation when the value does not match. Default behaviour stays annotation-only, and unknown formats are never an error. Implement these format checks from the 2020-12 validation spec, as far as practical, using only the standard library: date-time, date, time, duration, email, idn-email, hostname, idn-hostname, ipv4, ipv6 and uuid. Give format checks one registration point so each check lives in its own file. idn-hostname and idn-email get basic checks only, and the code records what they miss. Write positive and negative tests for each format and for each way of turning the mode on.
   Why: Adds the optional assertion mode with the first half of the formats, sized to fit alongside the shared format registration point.

19. **Format assertion: URI, IRI, templates, pointers and regex** — to do
   In the optional format-assertion mode, add the remaining 2020-12 format checks as new files registered through the format registration point: uri, uri-reference, iri, iri-reference, uri-template, json-pointer, relative-json-pointer and regex. The regex format uses the existing ECMA-262-to-RE2 translation step. Use only the standard library and check each format as far as practical. Any format that is only partly checked is recorded in the code. Outside assertion mode these formats stay annotations only. Write positive and negative tests for each format.
   Why: Completes the formats in separate files that plug into the existing registration point.

20. **User documentation: CLI contract and scope** — to do
   Write README.md and a docs/ folder with no code changes. Cover how to build, test and run the project, using go build and go test ./.... Describe the complete CLI contract: every flag; schema and document input from files and from standard input; the registry flags that map files and directories to base URIs; the output-format flag and the exact JSON shapes for flag, basic, detailed and verbose; and the line format for multiple documents. State the exit codes: 0 valid, 1 invalid, 2 error, where invalid schemas, unresolved references, bad input and exceeded limits are errors, never invalid. List the limit flags and their defaults, and the format-assertion flag. Document the library API with a short example. Include a section that clearly separates MANDATORY 2020-12 behaviour from OPTIONAL features, which are format-assertion mode and vocabulary and detailed/verbose output. List the documented deviations: ECMA-262 regex versus RE2, and partly checked formats. Add the dependency statement: standard library only, no JSON Schema library. This story depends on these stories: Walking skeleton; Error reports: flag and basic output, multiple documents; Schema registry for external references; Embedded meta-schemas and schema validation; Limits against hostile input; Detailed and verbose output; Format assertion: mode, dates, email, hostnames, IP and UUID; Format assertion: URI, IRI, templates, pointers and regex.
   Why: Comes last because it documents the finished CLI contract, the scope and the deviations.
