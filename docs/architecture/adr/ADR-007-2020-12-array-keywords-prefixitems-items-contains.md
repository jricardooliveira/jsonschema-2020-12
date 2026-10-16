# ADR #7: 2020-12 array keywords: prefixItems, items, contains, maxContains, minContains, maxItems, minItems, uniqueItems

- Date: 2026-10-16
- Story: #7
- Status: proposed (pending architecture sign-off)

## Decision
The applicator keywords (prefixItems, items, contains) go in the applicator vocabulary package. The validation keywords (maxContains, minContains, maxItems, minItems, uniqueItems) go in the validation vocabulary package. Each is one file plus its own test, and each registers itself from init(). Neither all.go nor registry.go is edited. items reads its sibling prefixItems length, and contains reads its sibling minContains and maxContains values, through whatever sibling-keyword access the skeleton's compile step provides. minContains and maxContains register as compile-only keywords that check their own value and do nothing on their own. Item annotations go into a new per-evaluation record, internal/eval/evaluated_items.go. Per the operator note, that record extends internal/eval/evaluated.go when story 183 has created it, using the same style. The compiled schema holds no per-evaluation state. The only edit to the crowded internal/eval/eval.go is at most one line attaching the record to the evaluation context. The uniqueItems duplicate finder is an unexported function in uniqueitems.go that returns the groups and the number of full deep-equality comparisons. The fingerprint must give equal values the same key (numbers by their normalised exact value, objects independent of key order), and it is built on internal/jsonvalue deep equality from story 179.

## Affected modules
- internal/eval/evaluated_items.go
- internal/eval/evaluated_items_test.go
- internal/eval/evaluated.go
- internal/eval/eval.go
- internal/keywords/applicator/prefixitems.go
- internal/keywords/applicator/prefixitems_test.go
- internal/keywords/applicator/items.go
- internal/keywords/applicator/items_test.go
- internal/keywords/applicator/contains.go
- internal/keywords/applicator/contains_test.go
- internal/keywords/validation/mincontains.go
- internal/keywords/validation/mincontains_test.go
- internal/keywords/validation/maxcontains.go
- internal/keywords/validation/maxcontains_test.go
- internal/keywords/validation/maxitems.go
- internal/keywords/validation/maxitems_test.go
- internal/keywords/validation/minitems.go
- internal/keywords/validation/minitems_test.go
- internal/keywords/validation/uniqueitems.go
- internal/keywords/validation/uniqueitems_test.go
- testdata/applicator/prefixItems.json
- testdata/applicator/items.json
- testdata/applicator/contains.json
- testdata/validation/minContains.json
- testdata/validation/maxContains.json
- testdata/validation/maxItems.json
- testdata/validation/minItems.json
- testdata/validation/uniqueItems.json

## Data / API impact
- DB impact: no
- API impact: no
- Migration needed: no

## Constraints for implementation
- No instance values in any error message. Messages give positions and counts only.
- No short-circuit: prefixItems, items, contains and uniqueItems examine the whole array.
- No global mutable state. The comparison counter is returned per call.
- Use the exact number model and deep equality from internal/jsonvalue.
- No wall-clock assertions in tests.
- At most one wiring line in internal/eval/eval.go.
- Standard library only. go.mod unchanged.

## Risks
- The fingerprint for numbers must normalise exact values (1, 1.0, 1e0) or real duplicates fall into different groups. Tests must cover this.
- Evaluating every element for contains and items increases work on huge arrays. The evaluation-step limits arrive in story 193.

## Breaking changes
- none

## Sensitivity
- security
