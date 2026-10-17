# ADR #19: Format assertion: URI, IRI, templates, pointers and regex

- Date: 2026-10-17
- Story: #19
- Status: proposed (pending architecture sign-off)

## Decision
All work goes in new files inside internal/formats, the package created by story 195. Each format registers itself from init() into internal/formats/registry.go, which this story reads and never edits. The four address formats share one strict grammar file (rfc3986.go) with an IRI switch; this is a real second use, not a speculative seam. The regex format combines a new strict ECMA-262 u-flag syntax checker (ecma_syntax.go) with a property-name table (ecma_unicode_props.go), and reuses the translation step from story 182 without changing it. No crowded file is touched. Assertion and annotation mode switching already lives in story 195's mechanism.

## Affected modules
- internal/formats/rfc3986.go
- internal/formats/rfc3986_test.go
- internal/formats/uri.go
- internal/formats/uri_test.go
- internal/formats/uri_reference.go
- internal/formats/uri_reference_test.go
- internal/formats/iri.go
- internal/formats/iri_test.go
- internal/formats/iri_reference.go
- internal/formats/iri_reference_test.go
- internal/formats/uri_template.go
- internal/formats/uri_template_test.go
- internal/formats/json_pointer.go
- internal/formats/json_pointer_test.go
- internal/formats/relative_json_pointer.go
- internal/formats/relative_json_pointer_test.go
- internal/formats/ecma_syntax.go
- internal/formats/ecma_syntax_test.go
- internal/formats/ecma_unicode_props.go
- internal/formats/ecma_unicode_props_test.go
- internal/formats/regex.go
- internal/formats/regex_test.go

## Data / API impact
- DB impact: no
- API impact: no
- Migration needed: no

## Constraints for implementation
- Standard library only; go.mod stays without require lines.
- Do not use net/url to judge validity: it is lenient (it accepts 'http://exa mple.com').
- Never edit internal/formats/registry.go, internal/keywords/registry.go, internal/eval/eval.go, jsonschema.go, options.go or cmd/jsonschema/*.
- Checkers must run in linear or bounded time on hostile strings: no backtracking regexes over the input, and recursion in the ECMA parser is bounded by pattern length or nesting.
- Check functions are pure and keep no shared mutable state.
- Each partly checked format has a comment that says what it misses.
- Tests ask the code for any defined limit (for example the max prefix length 9999) instead of repeating the number.
- Format errors never quote the instance value.

## Risks
- The strict ECMA-262 u-flag grammar is large. Annex B leniencies must stay disabled, and named-group and back-reference rules are subtle.
- The Unicode property alias list is hand-written. Go's unicode tables may lag behind the Unicode version ECMA-262 references, so some script names may be missing. This is recorded in code.
- Story 182's translation step may reject constructs that the syntax checker accepts. The regex format must not let that turn a valid pattern into a failure.

## Breaking changes
- none

## Sensitivity
- security
