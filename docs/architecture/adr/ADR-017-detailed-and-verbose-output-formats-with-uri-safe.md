# ADR #17: Detailed and verbose output formats, with URI-safe absoluteKeywordLocation

- Date: 2026-10-17
- Story: #17
- Status: proposed (pending architecture sign-off)

## Decision
I changed only the point the reviewer settled: where detailed and verbose get their data. Everything else is the same as the reviewed design (percent-encoded fragment, shapes in their own files, the report byte cap, no new CLI flag, run.go and main.go not edited).

What I found in the code. Story 191 already keeps the per-subschema evaluation tree. It is keywords.TreeNode, returned by eval.EvaluateTree. Each node holds its result, errors and kept annotations, and passing nodes inside a failing evaluation stay in it. Tree nodes never call EvalContext.Report, so passing nodes never count against MaxOutputUnits. Schema.Validate calls eval.EvaluateFull, which builds no tree. The CLI calls s.Validate and then res.Output(shape) (run.go lines 165-169). It checks the shape name with output.Render(shape, true, nil) (line 119).

Decision: Validate keeps calling EvaluateFull, so it builds no tree and records nothing extra. Flag and basic results, units, memory and exit codes stay byte-for-byte the same. The tree is built only when a shape that needs it is rendered:
- Result also keeps the compiled root node, the decoded instance and the unit and annotation caps it was evaluated with. These are a few read-only references; the Result is still immutable.
- Result.Output(shape) (output.go) asks output.NeedsTree(shape). If the shape needs the tree, it calls eval.EvaluateTree with the same caps and passes the tree in Report.Tree. The tree is built per Output call and never cached, so a Result stays safe to share across goroutines.
- Evaluation is deterministic, so the tree has the same valid flag and units as the Validate call. A limit error during this second evaluation is returned from Output, and the CLI already turns that into exit 2.

Registry seam: internal/output/registry.go gets a Tree *keywords.TreeNode field on Report, a RegisterTreeReport(name, fn) that registers the shape and marks it as needing the tree, and NeedsTree(name). It keeps one map, one duplicate check and one Budget. detailed.go and verbose.go call RegisterTreeReport from init(). No central list of shape names exists anywhere.

tree.go walks the TreeNode and writes each unit through the Budget as it goes, so the cap is enforced while rendering. A nil Tree (the run.go name probe, or output.Render) renders a single root unit {valid, keywordLocation:"", instanceLocation:""}, so run.go's shape check accepts the new names without any edit.
- Verbose writes every evaluated subschema node, passing or failing, with its errors as leaf units and its kept annotations as annotation units.
- Detailed applies the 2020-12 condensing rules: it drops passing nodes that carry no annotations, collapses a node with one child into that child, and drops annotations under a failing node.

Part only partly produced, to be recorded in comments in detailed.go and verbose.go: a passing keyword that has no subschema and records no annotation (for example a passing minLength) has no tree node. Verbose therefore lists every evaluated subschema and every annotation, but not a separate unit for each passing leaf assertion. Producing those would mean recording passing units on every Validate call, which the reviewer ruled out. absoluteKeywordLocation on tree nodes is set only where the tree or its units already carry one.

Crowded files: jsonschema.go gets three Result fields and one line in Validate that fills them. compile.go keeps the one-line call to the fragment helper. eval.go, internal/eval/units.go and internal/keywords/registry.go are not edited. I removed internal/eval/units.go and units_verbose_test.go from the plan because they are no longer needed.

New regression test (output_unchanged_test.go):
- It builds a large document in code, with thousands of items and many failures. It compares the SHA-256 of its flag and basic output with hashes taken at base commit c941ecb before any other change, and checks the CLI exit codes for the same document.
- It also checks that with MaxOutputUnits set to exactly the number of failure units, Validate and the detailed and verbose renders all succeed. This proves passing nodes are not counted.

## Affected modules
- internal/compile/compile.go
- internal/compile/fragment.go
- internal/compile/fragment_test.go
- internal/output/registry.go
- internal/output/tree.go
- internal/output/nested_limit_test.go
- internal/output/detailed.go
- internal/output/detailed_test.go
- internal/output/verbose.go
- internal/output/verbose_test.go
- jsonschema.go
- output.go
- output_nested_test.go
- output_unchanged_test.go
- cmd/jsonschema/output_nested_test.go
- cmd/jsonschema/abslocation_escape_test.go

## Data / API impact
- DB impact: no
- API impact: yes
- Migration needed: no

## Constraints for implementation
- Standard library only; go.mod unchanged
- Each shape is its own file, registered from init() with output.RegisterTreeReport; no edit to a central list or switch
- Schema.Validate keeps calling eval.EvaluateFull: it builds no tree and records no passing units. Only Result.Output, for a shape where NeedsTree is true, calls eval.EvaluateTree, with the same MaxOutputUnits and MaxAnnotations
- Passing tree nodes never count against MaxOutputUnits; do not add Report calls or passing units anywhere
- Do not edit internal/eval/eval.go, internal/eval/units.go or internal/keywords/registry.go
- Touch internal/compile/compile.go only for the one-line fragment call, and jsonschema.go only for the Result fields and the line in Validate that fills them
- Do not edit cmd/jsonschema/run.go and do not add a report-bytes CLI flag (story 193 owns it). A nil Report.Tree must render a minimal valid root unit, so run.go's output.Render name probe accepts the new shapes
- Enforce the cap while rendering, by writing each unit through the existing output.Budget, not after building the whole report in memory
- Never emit instance values; property names appear only where an annotation lists matched names
- Write output_unchanged_test.go first, at the base commit, so its flag/basic SHA-256 values and CLI exit codes come from the code before this story
- Cap tests use DefaultMaxReportBytes and Options.MaxReportBytes from the code instead of repeating the number; the 'raise the cap' test is library-level
- The compiled Schema and Result stay immutable; the tree is per Output call and never cached (go test -race)
- Comments in detailed.go and verbose.go record the parts produced only partly: no unit for a passing leaf assertion that has no subschema and no annotation, and absoluteKeywordLocation only where it is known

## Risks
- internal/output/registry.go, jsonschema.go and output.go were read-only in the reservation and are now write claims. The Tree field on Report and RegisterTreeReport/NeedsTree are the smallest seam that lets a shape receive the tree, and Result has to keep the root node and instance so it can rebuild the tree. jsonschema.go is crowded; the change is three fields and one line.
- A Result now keeps the decoded instance alive for as long as the caller holds it. Validate does no extra work and records nothing extra, but memory is held longer when callers keep many Results.
- Detailed and verbose evaluate the document a second time inside Output. That doubles the CPU cost only for those shapes. Flag and basic are unaffected.
- Verbose does not list passing leaf assertions that have no subschema and no annotation, because the reviewer ruled out recording passing units. This is documented in the shape files.
- Encoding the fragment changes absoluteKeywordLocation for schemas with special characters in their keys. Existing tests that assert the unencoded form will need updating.
- Verbose output for large documents can hit the report cap. That is intended. Until story 193 adds the CLI flag, only library callers can raise the cap, through Options.MaxReportBytes.
- output_unchanged_test.go relies on hashes taken before the change. If the coder takes them after changing flag or basic, the test proves nothing. The new fragment encoding does not affect it because the test schema's keys need no encoding.

## Breaking changes
- absoluteKeywordLocation for schema keys that contain spaces, '%' or non-ASCII characters is now percent-encoded, in flag and basic output too (accepted by the reviewer)

## Sensitivity
- security
- pii
