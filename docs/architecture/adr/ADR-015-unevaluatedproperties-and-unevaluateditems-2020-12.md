# ADR #15: unevaluatedProperties and unevaluatedItems (2020-12 unevaluated vocabulary)

- Date: 2026-10-17
- Story: #15
- Status: proposed (pending architecture sign-off)

## Decision
Both keywords are plug-ins in the existing internal/keywords/unevaluated package. all.go already blank-imports that package, so it is only read. They need two things from the engine. (1) Ordering: 2020-12 core 11 requires unevaluated* to see every sibling's annotations. A boolean on keywords.Keyword (for example AfterSiblings) lets eval.go run flagged keywords last without a central name list. This is the only edit to the crowded internal/eval/eval.go, and it is kept to the dispatch loop. (2) Evaluated-location annotations: story 191 adds them to the per-call EvalContext and drops them from failing subschemas. These keywords read them for the current instance location and, on success, add every property or item they examined. 'not' contributes nothing by construction, because its subschema must fail. anyOf and oneOf must evaluate every branch, not stop at the first match, for annotations to be complete. Errors are one per uncovered child, at that child's instance location, through the error/output-unit model from story 180, and never carry values.

## Affected modules
- internal/keywords/registry.go
- internal/eval/eval.go
- internal/eval/order_test.go
- internal/keywords/unevaluated/unevaluatedproperties.go
- internal/keywords/unevaluated/unevaluatedproperties_test.go
- internal/keywords/unevaluated/unevaluateditems.go
- internal/keywords/unevaluated/unevaluateditems_test.go
- testdata/unevaluated/unevaluatedProperties.json
- testdata/unevaluated/unevaluatedItems.json

## Data / API impact
- DB impact: no
- API impact: no
- Migration needed: no

## Constraints for implementation
- No central list or switch of keyword names; ordering comes from the registration flag.
- Compiled Schema stays immutable; all evaluated-location state lives in the per-call evaluation context (tests run with -race).
- Errors never include instance values, only instance location, keyword location and absolute keyword location.
- Only annotations from successful subschemas count.
- Standard library only; go.mod unchanged.
- Tests cite 2020-12 core sections; nothing from the official JSON-Schema-Test-Suite or Bowtie is copied or referenced.

## Risks
- If story 191 does not track evaluated items/properties per instance location, or drops annotations at the wrong granularity, this story cannot be correct without changing 191's files.
- anyOf/oneOf short-circuiting in story 185 would hide annotations from later passing branches.
- $dynamicRef extension cases depend on story 188 propagating annotations from the resolved target.

## Breaking changes
- none

## Sensitivity
- security
