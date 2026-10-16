# Pipeline record — #2: Exact JSON value model, deep equality, and the type, enum and const keywords

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #135 · **Status:** waiting_human · **Stage:** gate-2-human
- **Started:** 2026-10-16 16:50:29 UTC · **Stopped:** 2026-10-16 16:51:03 UTC
- **Story spend:** ~$0.16 of $26.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-2-architect: `factory approve 135` to accept, or `factory reject 135 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-2-architect**:
  > ⚠️ BREAKING CHANGES detected — these will affect existing API consumers:
  >   • If the skeleton's keyword evaluate signature used a different instance type, it changes to jsonvalue.Value.
  >   • Keyword evaluate signature may change to take jsonvalue.Value. This is an internal-only API (internal/keywords) and changes no public library or CLI contract. It must land before stories 181-185.
  >   → Approve these breaking changes?
  >
  > 🧱 BOUNDARY REVIEW WARNINGS — tenant / authorization / API contract / security concerns found before any code is written:
  >   • api_contract: The number-cap option is only loosely specified ("an Options field or option function"). Story 178 and the project spec fix limits at Compile time. The design does not say whether the cap is stored in the immutable Schema and applied to both the schema and every Validate document.
  >   • api_contract: The design lists two limits (MaxNumberDigits, MaxNumberExponent) but a single CLI flag. A human should settle whether one flag raises both or each gets its own flag. Story 193 must reuse the same naming pattern and not duplicate it.
  >   • api_contract: The design names ParseError, but ADR #1 already defines SyntaxError and LimitError in internal/jsonvalue. Two parallel error families would break the typed errors.As mapping to exit code 2.
  >   • security: ParseError is designed to carry a JSON Pointer to the error location. Pointer segments are document member names, which is content. ADR #1, approved at Checkpoint 2, deliberately reports offset plus line:column only for this reason. The design contradicts that decision, so a human should confirm that parse and limit error text stays free of member names.
  >   • security: The caller-supplied cap is a hostile-input control. The design does not say what a zero, negative or absent cap value means. An unvalidated value could turn the cap off (fail-open) or reject everything.
  >   • security: The cap is only safe if the exponent text itself is parsed without overflow or allocation. An exponent with thousands of digits must be rejected by its text length, not converted with Atoi into a wrapped value. The cap must also cover negative exponents (1e-999999999), which expand into huge denominators.
  >   • security: The new strict decoder replaces ADR #1's decode.go. The depth limit (checked before descending), the location-only error text, and the input-size cap in run.go must not regress when the decoder is rewritten.
  >   → Approve the design with these boundary risks, or reject with the fix?

## Trail

Every boss authorization (`boss`), agent verdict (`agent`) and gate verdict (`gate`), in the order they happened.

| When (UTC) | Kind | Step | Outcome |
|---|---|---|---|
| 2026-10-16 16:50:29 | run | run started | • |
| 2026-10-16 16:50:29 | agent | spec-agent | ✅ pass |
| 2026-10-16 16:50:29 | gate | gate-1-spec | ✅ Story 'Exact JSON value model, deep equality, and the type, enum and const keywords' has 11 AC and 6 tasks |
| 2026-10-16 16:50:30 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-16 16:50:30 | agent | architect-agent | ✅ pass |
| 2026-10-16 16:50:30 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it changes an API; it declares breaking changes; it touches sensitive areas (security) |
| 2026-10-16 16:51:03 | agent | boundary-agent | ✅ warn · 33.1s · $0.1646 |
| 2026-10-16 16:51:03 | gate | gate-2-architect | ⏸️ awaiting human sign-off |
| 2026-10-16 16:51:03 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
