# SPEC #19: Format assertion: URI, IRI, templates, pointers and regex

- Date: 2026-10-17
- Type: feature
- Verdict: pass
- Status: proposed (pending Checkpoint 1 sign-off)

## Problem

In format-assertion mode the validator checks only date-time, date, time, duration, email, idn-email, hostname, idn-hostname, ipv4, ipv6 and uuid. The other 2020-12 formats are never checked: uri, uri-reference, iri, iri-reference, uri-template, json-pointer, relative-json-pointer and regex.

## Why this matters

The brief makes format assertion an optional mode that checks every format defined in the 2020-12 validation spec as far as practical, using the standard library only, and records any format that is only partly checked. Go's net/url accepts addresses the official grammars reject, so the operator chose strict, project-written grammar checks.

## Acceptance criteria

- In assertion mode, 'uri' accepts only strings that match the RFC 3986 URI grammar, including the required scheme. It rejects relative references, spaces, bad percent-encodings (for example '%zz' or '%4') and misplaced characters. The tests are positive and negative and live in internal/formats/uri_test.go.
- In assertion mode, 'uri-reference' accepts any RFC 3986 URI-reference, absolute or relative (for example '/a/b?c#d', '#frag', ''). It rejects strings with spaces, bad percent-encodings or characters the grammar does not allow. The tests are positive and negative.
- In assertion mode, 'iri' and 'iri-reference' follow RFC 3987. They accept the same strings as uri and uri-reference, plus non-ASCII characters in the allowed ucschar and iprivate ranges, for example 'http://ƒøø.example/ü'. They reject spaces, bad percent-encodings and disallowed characters. 'iri' requires a scheme. The tests are positive and negative.
- In assertion mode, 'uri-template' accepts a template only when its spelling follows RFC 6570. Braces must match and not nest. Variable names may use only the allowed characters and pct-encoded triplets. Only the operators + # . / ; ? & are allowed, and only the modifiers :N (1 to 9999) and *. The template is never expanded. Examples: '/users/{id}{?page}' passes; '{', '{a', '{a}}', '{@a}', '{a:0}' and '{a b}' fail. No part of the template is left unchecked.
- In assertion mode, 'json-pointer' accepts '' and strings made of '/'-prefixed reference tokens in which '~' is followed only by '0' or '1' (RFC 6901). Examples: '/a~1b' and '/' pass; 'a', '/~2' and '/~' fail.
- In assertion mode, 'relative-json-pointer' follows the draft-bhutton-relative-json-pointer-00 grammar that 2020-12 references. That grammar is: a non-negative integer with no leading zero, an optional index manipulation '+N' or '-N', then either '#' or a JSON pointer. Examples: '0', '1/name', '0#', '0+1' and '1-2/name' pass; '01', '0+', '-1', '0##' and '/a' fail.
- In assertion mode, 'regex' accepts a string only when it is a valid ECMA-262 pattern under the strict (u-flag) grammar. Lookahead, lookbehind, back-references and named groups are accepted as valid even though RE2 cannot run them, and the code records this. Invalid patterns fail, for example '\a', a lone '{', '(', '[b-a]' and '(?<n>a)(?<n>b)'. Patterns that the existing ECMA-262-to-RE2 translation step can handle are also passed through it, and the step's result never turns a syntactically valid pattern into a failure.
- For the regex format, '\p{...}' and '\P{...}' names are checked against a list of ECMA-262 property names and values written into the code. This covers General_Category and its aliases, Script and Script_Extensions with their aliases, and the binary properties. The list is built using the standard library only (for example unicode.Scripts and unicode.Categories plus hand-written aliases). Names such as '\p{Letter}', '\p{L}' and '\p{Script=Greek}' pass; '\p{Lettter}' and '\p{Script=Foo}' fail. Any names the list does not yet cover are recorded in the code.
- Each of the eight formats lives in its own file internal/formats/<format>.go and registers itself from init() through the format registration point (internal/formats/registry.go). No existing shared file is edited to add them.
- Outside assertion mode, all eight formats stay annotations only: a value that breaks the format still validates as valid. Each format has a test for this.
- Every format here that is only partly checked carries a code comment that says what it misses. go.mod gains no require line, and no test uses the network.

## Non-goals

- Expanding uri-template templates
- Resolving or dereferencing any URI, IRI or JSON pointer against a document
- Running ECMA-262-only regex features (lookahead, back-references) at validation time; the 'pattern' keyword keeps its schema-error behaviour from story 182
- Writing docs/ pages (story 197 documents the partly checked formats)
- Any new dependency, including golang.org/x packages

## Tasks

- **T1** — Strict RFC 3986/3987 grammar and the four address formats
  - purpose: Write one shared, stdlib-only recursive-descent checker for URI and URI-reference, with an IRI switch that admits the RFC 3987 ucschar and iprivate ranges. Register uri, uri-reference, iri and iri-reference as thin files on top of it.
  - scope: internal/formats/rfc3986.go, internal/formats/rfc3986_test.go, internal/formats/uri.go, internal/formats/uri_test.go, internal/formats/uri_reference.go, internal/formats/uri_reference_test.go, internal/formats/iri.go, internal/formats/iri_test.go, internal/formats/iri_reference.go, internal/formats/iri_reference_test.go
  - done when: go test -race ./internal/formats/... passes with positive, negative and annotation-only cases for the four formats.
  - delivers: AC1, AC2, AC3, AC9, AC10, AC11
- **T2** — uri-template spelling check
  - purpose: Check the RFC 6570 grammar (literals, expressions, operators, varspecs, prefix and explode modifiers) without expanding the template.
  - scope: internal/formats/uri_template.go, internal/formats/uri_template_test.go
  - done when: Tests cover matching braces, operators, modifier ranges, varname characters, pct-encoding and annotation-only mode.
  - delivers: AC4, AC9, AC10, AC11
- **T3** — json-pointer and relative-json-pointer
  - purpose: Check RFC 6901 pointers, and relative pointers per draft-bhutton-relative-json-pointer-00, including the index manipulation.
  - scope: internal/formats/json_pointer.go, internal/formats/json_pointer_test.go, internal/formats/relative_json_pointer.go, internal/formats/relative_json_pointer_test.go
  - done when: Tests cover escapes, leading zeros, '+N'/'-N', '#', and annotation-only mode.
  - delivers: AC5, AC6, AC9, AC10, AC11
- **T4** — Strict ECMA-262 (u-flag) pattern syntax checker
  - purpose: Write a stdlib-only parser that only answers whether a pattern is valid ECMA-262 with the u flag. It covers alternatives, quantifiers and their bounds, groups, named groups (no duplicates in one alternative), lookaround, back-references (numbered refs within the group count, named refs to existing names), class ranges (in order), identity escapes limited to syntax characters and '/', \u{...} and surrogate escapes, and \p/\P delegated to the name list.
  - scope: internal/formats/ecma_syntax.go, internal/formats/ecma_syntax_test.go
  - done when: Table tests accept lookaround, back-references and named groups, and reject '\a', a lone '{', unbalanced parentheses, reversed ranges, out-of-range back-references and duplicate group names.
  - depends on: T5
  - delivers: AC7, AC11
- **T5** — Unicode property name list
  - purpose: Build the ECMA-262 property name and value list for \p{...}. Use unicode.Categories and unicode.Scripts from the standard library, plus hand-written long-name aliases, script aliases and the binary property names. Record in code any names not yet covered.
  - scope: internal/formats/ecma_unicode_props.go, internal/formats/ecma_unicode_props_test.go
  - done when: Tests accept 'Letter', 'L', 'Lu', 'Script=Greek', 'sc=Grek', 'Script_Extensions=Latin', 'ASCII' and 'Any', and reject misspellings and unknown script values.
  - delivers: AC8, AC11
- **T6** — regex format wiring
  - purpose: Register the 'regex' format. A value is valid when the strict ECMA syntax check accepts it. Patterns without ECMA-only constructs are also run through the existing ECMA-262-to-RE2 translation step. Record in code that ECMA-only constructs are judged by syntax only.
  - scope: internal/formats/regex.go, internal/formats/regex_test.go
  - done when: Tests cover valid, invalid and JS-only patterns in assertion mode, and annotation-only behaviour outside it.
  - depends on: T4, T5
  - delivers: AC7, AC8, AC9, AC10, AC11
