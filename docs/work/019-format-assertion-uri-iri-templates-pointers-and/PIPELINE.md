# Pipeline record — #19: Format assertion: URI, IRI, templates, pointers and regex

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #154 · **Status:** waiting_human · **Stage:** gate-2-human
- **Started:** 2026-10-17 03:17:13 UTC · **Stopped:** 2026-10-17 03:17:36 UTC
- **Story spend:** ~$0.17 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-2-architect: `factory approve 154` to accept, or `factory reject 154 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-2-architect**:
  > 🧱 BOUNDARY REVIEW WARNINGS — tenant / authorization / API contract / security concerns found before any code is written:
  >   • security: Hostile-input depth bound is ambiguous. The design says ECMA parser recursion is 'bounded by pattern length or nesting'. A pattern of millions of '(' or '[' characters, which fits within the default input-size limit, would recurse once per character. Go stack exhaustion is a fatal, non-recoverable crash, not an error. A depth cap is also not a clean fix, because a check function returning bool cannot report a limit error. Returning false would turn a limit hit into a silent 'invalid', which the project forbids. A human should settle the approach; the safe choice is a non-recursive parser (explicit heap stack or iteration) whose work is linear in input length, so no cap is needed.
  >   • security: The same concern applies to the RFC 3986/3987 and RFC 6570 checkers and to the JSON pointer checkers. The design does not say they are iterative or linear. These formats are simple enough that they must be single-pass with no recursion over input length.
  >   • security: The ECMA-262-to-RE2 translation step is reused on instance-supplied strings, which are hostile. The design does not say that a panic or error from it is contained. It must be treated as non-authoritative: any error, or a panic from the step, must not change the syntax checker's verdict and must not crash the process.
  >   → Approve the design with these boundary risks, or reject with the fix?

## Trail

Every boss authorization (`boss`), agent verdict (`agent`) and gate verdict (`gate`), in the order they happened.

| When (UTC) | Kind | Step | Outcome |
|---|---|---|---|
| 2026-10-17 03:17:13 | run | run started | • |
| 2026-10-17 03:17:13 | agent | spec-agent | ✅ pass |
| 2026-10-17 03:17:14 | gate | gate-1-spec | ✅ Story 'Format assertion: URI, IRI, templates, pointers and regex' has 11 AC and 6 tasks |
| 2026-10-17 03:17:14 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-17 03:17:14 | agent | architect-agent | ✅ pass |
| 2026-10-17 03:17:14 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it touches sensitive areas (security) |
| 2026-10-17 03:17:36 | agent | boundary-agent | ✅ warn · 21.4s · $0.1737 |
| 2026-10-17 03:17:36 | gate | gate-2-architect | ⏸️ awaiting human sign-off |
| 2026-10-17 03:17:36 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
