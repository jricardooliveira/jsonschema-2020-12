# Pipeline record — #14: Annotation collection: metadata, format, content and unknown keywords, evaluated properties and items

> Written by the factory's orchestrator (the boss) each time this story's run stops. The factory database is the live record; this is its committed snapshot. Do not hand-edit it.

- **Project:** jsonschema-2020-12
- **Run:** #144 · **Status:** waiting_human · **Stage:** gate-question
- **Started:** 2026-10-16 21:28:36 UTC · **Stopped:** 2026-10-16 21:43:46 UTC
- **Story spend:** ~$5.26 of $35.00 (estimated at API list prices)

## Next authorized step

The operator decides at gate-question: `factory approve 144` to accept, or `factory reject 144 "<your answers>"` to send it back.

## Blockers

- Waiting for the operator at **gate-question**:
  > The coder asks for files the approved design does not name: internal/compile/refs.go (resolver.compile, around line 157) does `if !ok { continue }` for unregistered keywords. Replace the `continue` with `n.Items = append(n.Items, Item{keywords.AnnotateOnly(name, ""), s[name], loc + "/" + escape.Replace(name)})` before `continue`, as compile.go already does. Then unknown keywords are collected as annotations. Not granted automatically: internal/compile/refs.go (not in the approved design and not a test). An answer that starts with "yes" (a note may follow) approves these files for this run: the same pass runs again and may change them. An answer that starts with "no" (or a rejection) refuses them: the pass runs again without them, with your words as a finding.
  > paths: internal/compile/refs.go

## Trail

Every boss authorization (`boss`), agent verdict (`agent`) and gate verdict (`gate`), in the order they happened.

| When (UTC) | Kind | Step | Outcome |
|---|---|---|---|
| 2026-10-16 21:28:36 | run | run started | • |
| 2026-10-16 21:28:37 | agent | spec-agent | ✅ pass |
| 2026-10-16 21:28:37 | gate | gate-1-spec | ✅ Story 'Annotation collection: metadata, format, content and unknown keywords, evaluated properties and items' has 11 AC and 6 tasks |
| 2026-10-16 21:28:37 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-16 21:28:37 | agent | architect-agent | ✅ pass |
| 2026-10-16 21:28:38 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it changes an API; it declares breaking changes; it touches sensitive areas (security) |
| 2026-10-16 21:29:11 | agent | boundary-agent | ✅ warn · 33.6s · $0.2020 |
| 2026-10-16 21:29:11 | gate | gate-2-architect | ❌ REJECTED: Keep the design and settle the boundary points: (1) MaxAnnotations follows the existing limit style exactly like MaxOutputUnits: unset/0 in the library = DefaultMaxAnnotations, negative = compile error, expl… |
| 2026-10-16 21:29:38 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-16 21:30:21 | agent | architect-agent | ✅ pass · 42.5s · $0.5745 |
| 2026-10-16 21:30:21 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it changes an API; it declares breaking changes; it touches sensitive areas (security) |
| 2026-10-16 21:30:40 | agent | boundary-agent | ✅ warn · 18.5s · $0.2005 |
| 2026-10-16 21:30:40 | gate | gate-2-architect | ❌ REJECTED: Keep this design and add one bound: a cap on the total bytes of a rendered report, DefaultMaxReportBytes = 256 MiB, in the existing limit style (library option, 0 = default, negative = error; CLI flag left t… |
| 2026-10-16 21:32:12 | boss | authorize architect-agent | ✅ gate-1-spec passed |
| 2026-10-16 21:34:10 | agent | architect-agent | ✅ pass · 117.7s · $0.8801 |
| 2026-10-16 21:34:10 | boss | authorize boundary-agent | ✅ the design requires a boundary review: it changes an API; it declares breaking changes; it touches sensitive areas (security) |
| 2026-10-16 21:34:27 | agent | boundary-agent | ✅ pass · 16.4s · $0.2078 |
| 2026-10-16 21:34:27 | gate | gate-2-architect | ✅ APPROVED: Approved by human reviewer |
| 2026-10-16 21:34:40 | boss | authorize coder-agent:T1 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-16 21:36:50 | agent | coder-agent | ✅ complete · T1 · 129.2s · $0.5768 |
| 2026-10-16 21:36:57 | gate | gate-build | ✅ [T1] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-16 21:36:57 | boss | authorize coder-agent:T2 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-16 21:37:44 | agent | coder-agent | ✅ complete · T2 · 46.9s · $0.4180 |
| 2026-10-16 21:37:59 | gate | gate-build | ✅ [T2] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-16 21:37:59 | boss | authorize coder-agent:T3 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-16 21:38:57 | agent | coder-agent | ✅ complete · T3 · 58.1s · $0.4537 |
| 2026-10-16 21:39:11 | gate | gate-build | ✅ [T3] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-16 21:39:12 | boss | authorize coder-agent:T4 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-16 21:40:49 | agent | coder-agent | ✅ complete · T4 · 97.0s · $0.6153 |
| 2026-10-16 21:40:50 | gate | gate-build | ✅ [T4] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-16 21:40:50 | boss | authorize coder-agent:T5 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-16 21:42:09 | agent | coder-agent | ✅ complete · T5 · 78.0s · $0.5266 |
| 2026-10-16 21:42:16 | gate | gate-build | ✅ [T5] go_build:pass, go_vet:pass, go_test:pass, go_race:pass |
| 2026-10-16 21:42:16 | boss | authorize coder-agent:T6 | ✅ gate-2-architect, approved by the operator at Checkpoint 2 |
| 2026-10-16 21:43:46 | agent | coder-agent | ✅ scope-request · T6 · 89.7s · $0.6040 |
| 2026-10-16 21:43:46 | gate | gate-question | ⏸️ awaiting human sign-off |
| 2026-10-16 21:43:46 | run | run waiting_human | ⏸️ |

## Warnings (recorded, not blocking)

- none
