# INTENT #19

- Author: Joao Oliveira
- Date: 2026-10-17
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

In the optional format-assertion mode, add the remaining 2020-12 format checks as new files registered through the format registration point: uri, uri-reference, iri, iri-reference, uri-template, json-pointer, relative-json-pointer and regex. The regex format uses the existing ECMA-262-to-RE2 translation step. Use only the standard library and check each format as far as practical. Any format that is only partly checked is recorded in the code. Outside assertion mode these formats stay annotations only. Write positive and negative tests for each format.

## Operator clarifications
- Some text patterns are valid JavaScript patterns but cannot be run by Go's built-in pattern engine. For example, "(?=abc)" means "followed by abc". When the regex format is checked, should a text like that count as a valid pattern? — Yes, count it as valid — Recommended. The text follows JavaScript pattern rules, so it is a real pattern, even though Go cannot run it. A small syntax check accepts the JavaScript-only parts (lookahead, back-references), and the code records this. A correct pattern is never reported as wrong.
- Go's built-in web address reader is lenient. For example, it accepts "http://exa mple.com" with a space in it. How strictly should the uri, uri-reference, iri and iri-reference formats be checked? — Strictly, against the official address rules — Recommended. The project writes its own check that follows the official address grammar (RFC 3986, and RFC 3987 for international addresses), built from the standard library only. Addresses with spaces, bad percent-codes or misplaced characters fail. A "uri" must also start with a scheme such as "http:".
- A uri-template is an address with fill-in slots, for example "/users/{id}{?page}". How deep should its check go? — Full check of the template's spelling — Recommended. Braces must match, slot names must use allowed characters, and only the official operator signs (+ # . / ; ? &) and modifiers (:3, *) are allowed, as defined in RFC 6570. The template is never filled in. No part is left unchecked.
- JavaScript patterns can be read in a strict mode or a relaxed mode. In the relaxed mode a text like "\a" (a backslash before an ordinary letter) is just the letter "a". In the strict mode it is a mistake. When the regex format is checked, which mode should decide whether a text is a valid pattern? — Strict mode — Recommended. Matches the strict JavaScript pattern rules (the 'u' flag). Useless or mistyped escapes such as "\a" or a lone "{" fail. It is closest to how most JSON Schema tools judge the regex format.
- A 'relative-json-pointer' is a short path that starts from the current spot in a document and moves up, for example "1/name" means "go up one level, then take name". The version of this rule used by 2020-12 also allows moving sideways in a list, for example "0+1" means "the next item in this list". Should the check accept these sideways moves? — Yes, accept sideways moves — Recommended. Follows the exact version of the rule that 2020-12 points to. "0+1" and "1-2/name" pass. A plain number with a leading zero such as "01", or a sign with no number after it such as "0+", fails.
- In strict JavaScript patterns, "\p{...}" means "any character of this kind", for example "\p{Letter}" or "\p{Script=Greek}". When the regex format is checked, how closely should the name inside the braces be checked? — Check the name against the official list — Recommended. A made-up name such as "\p{Lettter}" fails, the same way JavaScript rejects it. The list of names is written into the code from the JavaScript rules, using the standard library only. Any name the list does not yet cover is recorded in the code.

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- Story 195 (format registry and assertion mode) and story 182 (translation step) are not in the inventory yet. The exact signature of the format check function and the registration call are assumed from the conventions (internal/formats/registry.go, init() registration).
- Story 182 does not fix the file path or API of the ECMA-262-to-RE2 translation step; it is claimed only as a component read. If its API does not let the regex format tell 'unsupported by RE2' apart from 'syntax error', the regex format will rely on ecma_syntax.go for validity and record that in code.
- internal/uri (story 186) may already parse JSON Pointer and RFC 3986 references, but it is built for resolution and may be lenient. This story writes strict, separate checkers in internal/formats and does not depend on 186. A reviewer may prefer to share the JSON Pointer token check later.
- Story size: 22 files and 6 tasks fit the limits, but T4 (strict ECMA-262 u-flag grammar) is the largest piece. If it overruns, split it into two stories: (a) the address, template and pointer formats (T1-T3), and (b) the regex format with its syntax checker and property names (T4-T6).
- Go's unicode tables may not match the Unicode version ECMA-262 currently references, so some Script or binary property names may be missing from the list. Those gaps are recorded in code as the operator asked.

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
