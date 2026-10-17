# Pipeline record — #18: Format assertion: mode, dates, email, hostnames, IP and UUID

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #153 · **Status:** waiting_human · **Stage:** gate-question
- **Started:** 2026-10-17 02:32:34 UTC · **Stopped:** 2026-10-17 02:33:26 UTC
- **Story spend:** ~$0.52 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-question: `factory approve 153` to accept, or `factory reject 153 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-question**:
  > The coder asks for files the approved design does not name: The agreed design says to carry the assert boolean into the format keyword's Compile function. The only available route is a new field on keywords.CompileContext. resolver.compile in refs.go must set that field per resource. The library option has to be added to the Options struct in numberlimits.go and passed through jsonschema.go and CompileReachedWith. None of these four files is in T1's scope. Not granted automatically: internal/keywords/registry.go (not in the approved design and not a test); internal/compile/refs.go (not in the approved design and not a test); jsonschema.go (not in the approved design and not a test); numberlimits.go (not in the approved design and not a test). An answer that starts with "yes" (a note may follow) approves these files for this run: the same pass runs again and may change them. An answer that starts with "no" (or a rejection) refuses them: the pass runs again without them, with your words as a finding.
  > paths: internal/keywords/registry.go, internal/compile/refs.go, jsonschema.go, numberlimits.go

## Trail

Every boss authorization (`boss`), agent verdict (`agent`) and gate verdict (`gate`), in the order they happened.

| When (UTC) | Kind | Step | Outcome |
|---|---|---|---|
| 2026-10-17 02:32:34 | run | run started | • |
| 2026-10-17 02:32:34 | agent | spec-agent | ✅ pass |
| 2026-10-17 02:32:34 | gate | gate-1-spec | ✅ Story 'Format assertion: mode, dates, email, hostnames, IP and UUID' has 17 AC and 6 tasks |
| 2026-10-17 02:32:35 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-17 02:32:35 | agent | architect-agent | ✅ pass |
| 2026-10-17 02:32:35 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it changes an API; it touches sensitive areas (security) |
| 2026-10-17 02:32:53 | agent | boundary-agent | ✅ pass · 18.2s · $0.1732 |
| 2026-10-17 02:32:53 | gate | gate-2-architect | ✅ Architecture approved. 32 modules, 3 risks flagged |
| 2026-10-17 02:32:53 | boss | authorize coder-agent:T1 | ✅ gate-2-architect passed |
| 2026-10-17 02:33:26 | agent | coder-agent | ✅ scope-request · T1 · 32.1s · $0.3477 |
| 2026-10-17 02:33:26 | gate | gate-question | ⏸️ awaiting human sign-off |
| 2026-10-17 02:33:26 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
