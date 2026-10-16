# INTENT #2

- Author: Joao Oliveira
- Date: 2026-10-16
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Read schemas and documents with an exact JSON value model built only on the standard library. Numbers are kept exactly as written using math/big. This means huge integers like 12345678901234567890 and decimals like 0.1 lose no precision, and 1.0 counts as an integer. Broken JSON in a schema or a document is an error with exit code 2, never 'invalid'. Add deep JSON equality: numbers compare by mathematical value, objects compare regardless of key order, and arrays compare element by element. Then add the 2020-12 validation keywords 'type' (string or array form, all seven JSON types), 'enum' and 'const', each in its own file registered through the keyword registry. Write positive and negative tests derived from the 2020-12 validation specification text.

## Operator clarifications
- A JSON object can list the same name twice, for example {"age": 5, "age": 9}. The JSON standard does not say which value counts, so different tools disagree. What should the validator do when a schema or a document has a repeated name inside the same object? — Treat it as broken JSON (exit code 2) — Recommended. The file is rejected with a clear error naming where the repeat is. There is no guessing about which value was checked, and a crafted file cannot slip one value past the check while another program reads the other.
- Exact numbers can be abused. A tiny number such as 1e999999999 means a 1 followed by almost a billion zeros, and working with it exactly could use huge memory or take a very long time. How should such extreme numbers be handled? — Generous built-in limit on number size, error when exceeded (exit code 2) — Recommended. This fits your earlier answer about hostile input. There is a default cap on how many digits a number may have and how large its exponent may be, set far above any real-world value. The caller can raise the cap with the same library option and command-line flag as the other limits. Going over the cap is a clear limit error, never a silent valid or invalid.
- Text in JSON is sometimes damaged: bytes that are not valid UTF-8, or a half character written as an escape such as "\uD800" with no partner. How should the validator treat text like this in a schema or a document? — Treat it as broken JSON (exit code 2) — Recommended. The file is rejected with a clear error naming where the bad text is. Length rules and equality checks then always work on real characters, so the answer is never wrong because of damaged text.
- Some files start with an invisible marker at the very beginning (a 'byte order mark', which some Windows editors add), and some have extra content after the JSON value, for example two objects one after the other. What should happen in each case? — Allow the invisible start marker; extra content after the value is broken JSON (exit code 2) — Recommended. Files saved by common editors still work. A file with something after its single JSON value is rejected, so a second hidden value is never silently ignored. Spaces and line breaks around the value are always fine.

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- The repository inventory is empty, so the skeleton's real file names are unknown. compiler.go (the root compile/load entry), internal/eval/eval.go and the CLI test file name are assumed and must be mapped to what story 178 actually created.
- Unknown whether story 178 typed instances as `any` or as a skeleton value type. If it already routes through a decoder hook, the edits to registry.go and eval.go may be unnecessary.
- Unknown whether all.go already lists the validation vocabulary package, and whether that package exists. The convention says it does. If not, a package doc file must be added and all.go edited.
- The operator's clarification requires the number cap to be raised 'with the same library option and command-line flag as the other limits', but the general limits come in story 193. This story adds the number-cap option and flag; story 193 should use the same naming pattern and must not duplicate them.
- The shared spec-case runner may already exist from story 178. If so, T6 reduces to reusing it, and internal/speccases is not created.
- Location reporting for parse errors (byte offset vs JSON Pointer) is not fixed by the brief. Both are proposed. The exact message text is left to the coder, as long as it never quotes values.

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
