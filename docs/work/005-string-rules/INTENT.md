# INTENT #5

- Author: Joao Oliveira
- Date: 2026-10-16
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Add the 2020-12 string keywords maxLength, minLength and pattern. Each keyword lives in its own file and registers through the keyword registry. Lengths count Unicode code points, not bytes. The 'pattern' keyword uses Go's standard regexp package through a small translation step that handles common ECMA-262 differences, for example \d meaning ASCII digits only. A pattern the translation step or RE2 cannot handle, such as lookahead or a backreference, is a schema error with exit code 2. Put the translation step in its own reusable file so patternProperties and the regex format can use it later. The keywords apply only to string instances. Write spec-derived positive and negative tests, including multi-byte characters and translated patterns.

## Operator clarifications
- JavaScript and Go disagree on a few pattern shortcuts. Besides \d (digits 0-9 only), which differences should the translation step handle in this story? — The common shortcuts (\d \D \w \W \s \S with ECMA's Unicode whitespace set, and '.' excluding \n \r U+2028 U+2029), plus the ECMA escapes RE2 does not accept: \uXXXX (and surrogate pairs written as two \u escapes), \u{...}, \cX, \0, \/ and named groups (?<name>...), where they translate safely. Anything that cannot be translated (lookaround, backreferences) is a schema error, and every remaining difference is listed in the docs.
- A JSON text can contain a broken character, for example half of an emoji written as \ud83d with nothing after it. How should maxLength and minLength count it? — Reject the document as unreadable — Treated as broken JSON, so the check stops with exit code 2.
- Go's pattern engine can't hang, but a very long pattern in a hostile schema can still use a lot of memory when it is prepared. Should 'pattern' have a size limit? — Yes, a generous limit the caller can raise — Recommended. Fits the existing limits (nesting depth, input size). A pattern over the limit is a schema error with exit code 2. Ordinary patterns never come close.
- When 'pattern' fails, reports never show the document's text. May the report show the pattern itself? It comes from the schema, not the document. — Yes, show the pattern — Recommended. Example: 'does not match pattern ^[a-z]+$'. Easier to fix, and it never shows the checked data.
- A pattern can treat an emoji or another character outside the basic range either as one character or as two halves. Take the pattern ^.$, meaning 'exactly one character', checked against a single emoji. Which result do you want? — One character: it matches — Recommended. This is JavaScript's Unicode mode. Patterns count characters the same way maxLength and minLength do, so a schema never says 'one character' in one rule and 'two' in another.
- JavaScript patterns can name groups of characters by Unicode property, for example \p{L} or \p{Letter} for 'any letter in any alphabet'. Go accepts the short names like \p{L} but not the long names like \p{Letter}. What should happen with these? — Accept both, translating long names to short ones — Recommended. Schemas written for JavaScript keep working. Any property name that Go has no table for is a schema error with exit code 2, and the docs list which names are supported.
- JavaScript has an older, looser way of reading patterns and a newer, stricter one. You chose the newer one for counting characters (clarification 5). The two disagree on a few odd spellings that the older way quietly accepts, for example \_ or \- (a backslash before a character that needs none) or a lone { or ]. What should happen when a schema's pattern uses one of these? — Accept them where the meaning is clear — Recommended. \_ means '_' and a lone { means '{', as most schemas written for JavaScript expect, so these schemas keep working. A spelling with no clear meaning, such as \a, is a schema error with exit code 2. The docs list what is accepted.
- Operator note on refining again — Do not change internal/jsonvalue/decode.go in this story. Rejecting invalid UTF-8 and lone surrogate escapes is already story 179's job (its strict decoder, accepted plan); this story depends on 179 and relies on strings being valid Unicode, so maxLength/minLength count code points of already-valid strings. Keep the pattern-size limit here, named in the same style as story 179's number-size option and flag. All other open points are accepted as stated (Script_Extensions: schema error, documented).

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- The exact names and files for the library options and the compile context come from stories 178 and 179, which are not yet in the repository (no inventory was supplied). options.go and internal/compile/limits.go are best guesses for where story 179's number-size option lives. The coder must follow 179's actual files and naming style (for example MaxNumberSize to MaxPatternSize, -max-number-size to -max-pattern-size).
- It is unknown which file holds the shared table-driven testdata runner and the typed schema-error type that maps to exit code 2. Both are assumed to exist from stories 178 and 179, and are only read here.
- The default pattern-size value is not fixed by the operator. It must be generous (for example tens of kilobytes), and tests must read it from code.
- Whether Go's unicode tables cover every ECMA-262 property value name exactly is not verified. Names without a table are schema errors, as agreed, and the docs list the supported set.
- The CLI flag requires a one-line edit in the crowded file cmd/jsonschema/main.go. This cannot be avoided, because the flags live there by convention.

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
