# Release notes — #22: Build keyword locations only when a report or Result shows them

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #22
- Status: proposed (pending Checkpoint 3 sign-off)

**⚠️ the release-agent has concerns.**

## What changed

The change replaces eager keyword-location string building with lazy rendering, saving 24–40% of allocations and bytes on validation scenarios with no speed regression. All existing tests pass unchanged, including race tests, and all results stay identical. However, the implementation contains unused code that should be cleaned up before shipping.

- internal/eval/eval.go: replaced eager KeywordLocation and AbsoluteKeywordLocation string building with a lazyLoc type that renders on demand; the lazyLoc captures the pathFrame and renders only when a unit or annotation is created
- internal/eval/location.go: added nodeLoc and kwLoc types (dead code: never instantiated)
- internal/eval/location_test.go: added unit tests for nodeLoc/kwLoc (dead code: tests types not used in eval.go)
- internal/keywords/registry.go: changed EvalContext to hold a Location interface (Loc field) instead of eager KeywordLocation and AbsoluteKeywordLocation strings; added KeywordLoc() method to render on demand
- internal/keywords/annotate.go, applicator/{contains,if,items,properties,prefixitems}.go, unevaluated/{unevaluateditems,unevaluatedproperties}.go: updated to call ctx.KeywordLoc() instead of using ctx.KeywordLocation directly; if.go added branchLoc type to carry the base absolute location
- cmd/jsonschema/limits_edge_regression_test.go: new test checking every limit (nesting depth, ref chain, steps, output units, report bytes, annotations) at exactly the edge and one over
- docs/benchmarks.md: recorded before and after benchmark medians (10 runs per scenario) showing 24–40% lower allocs/op and B/op on validation scenarios, with no increase in compile or any other scenario; ns/op is 29–34% lower (speed gain is recorded but not required)

## How to verify

- Run go test ./... and confirm all tests pass, including existing output, verbose, annotation, unevaluated* and CLI contract tests
- Run go test -race ./... and confirm no race violations
- Run the benchmarks on the same machine: go test -run '^$' -bench . -benchmem -count=10 . (and repeat for ./internal/jsonvalue and ./cmd/jsonschema), and compare medians from docs/benchmarks.md to confirm allocs/op and B/op are lower on validation scenarios
- Manually validate a document with the tool and confirm the error messages, locations and results match the old output (no byte-for-byte assertion is required, but flag/basic/detailed/verbose reports must be logically identical)
- Run TestLimitsAtTheEdge and confirm each of the six limits still triggers at exactly the same point as before: one case at the edge gets a normal valid or invalid answer, one case over it gets exit code 2 with a limit-exceeded error

## Migration

none

## Rollback

Standard git revert: the change is self-contained within the evaluator and does not alter the public API or stored schema format

## Known limitations

- The implementation contains dead code: nodeLoc and kwLoc types in location.go are defined but never instantiated; the actual lazy location type (lazyLoc in eval.go) is simpler and uses the existing pathFrame
- Unit tests in location_test.go test the unused nodeLoc/kwLoc types, not the lazyLoc type actually deployed; the proof that lazyLoc renders correctly rests on the existing end-to-end test suites, not on dedicated unit tests of that type
- The limits regression test uses hard-coded edge values (depth 4, ref-chain 5, steps 186, output-units 4, report-bytes 414, annotations 37) instead of reading them from the code; if these defaults change, the test must be manually updated

## Concerns

- Dead code: location.go defines nodeLoc and kwLoc types that are never used. The actual code uses a simpler lazyLoc type in eval.go. Remove the unused types from location.go before shipping, or refactor to use the location.go types and remove the lazyLoc from eval.go.
- Test/code mismatch: location_test.go tests nodeLoc and kwLoc against an eager reimplementation, but these types are not deployed. The actual lazyLoc type in eval.go has no dedicated unit tests; its correctness is verified only by end-to-end tests. Add unit tests for lazyLoc or refactor to use the location.go types consistently.
- Hard-coded limit edge values in limits_edge_regression_test.go (depth 4, ref-chain 5, steps 186, output-units 4, report-bytes 414, annotations 37) are not derived from the code; the acceptance criterion says limits should be read from the code, not repeated as numbers. Consider reading these from options.go or limitflags.go instead.
- The change touches internal/keywords files in multiple vocabularies (applicator, unevaluated) to wire the lazy location interface. All touched files are in the granted scope, but the if.go modification introduces a branchLoc type to carry the base absolute location; confirm this works correctly in verbose's second evaluation pass.
