# Release notes — #17: Detailed and verbose output formats, with URI-safe absoluteKeywordLocation

> Written by the release-agent from the story, the design, the real diff and the review. It does not decide readiness: the release gate does, and only the operator releases (Checkpoint 3).

- Story: #17
- Status: proposed (pending Checkpoint 3 sign-off)

**⚠️ the release-agent has concerns.**

## What changed

The validator now renders 'detailed' and 'verbose' output shapes per 2020-12 spec, with nested error and annotation trees. Both library and CLI can select them. absoluteKeywordLocation is percent-encoded in all shapes to comply with RFC 3986. The report byte cap applies to both new shapes and is enforced during rendering.

- internal/compile/fragment.go and fragment_test.go: percent-encoding helper for URI-safe absoluteKeywordLocation fragments
- internal/compile/compile.go: one-line call to encodeFragment in Resource.AbsoluteLocation
- internal/output/tree.go and nested_limit_test.go: shared nested-report builder, enforces DefaultMaxReportBytes per unit
- internal/output/detailed.go and detailed_test.go: new 'detailed' shape with 2020-12 condensing (drops passing nodes without annotations, collapses single-child wrappers)
- internal/output/verbose.go and verbose_test.go: new 'verbose' shape that includes every evaluated keyword, passing or failing
- internal/eval/units_verbose_test.go: proves passing tree nodes do not count against MaxOutputUnits
- internal/output/registry.go: Tree field on Report, RegisterTreeReport and NeedsTree functions
- jsonschema.go: Result gains private fields (root, inst, max, maxAnn) for per-call tree rebuilding
- output.go: Output calls EvaluateTree if shape needs tree
- output_unchanged_test.go: regression test pinning flag and basic output SHA-256 hashes to base commit c941ecb
- output_nested_test.go and cmd/jsonschema/output_nested_test.go: library and CLI contract tests for detailed, verbose, cap enforcement and exit codes
- cmd/jsonschema/abslocation_escape_test.go: CLI test for percent-encoded absoluteKeywordLocation

## How to verify

- Library: res.Output('detailed') and res.Output('verbose') render valid JSON with nested errors and annotations on any Result
- Library: passing tree nodes do not count against MaxOutputUnits; set a cap equal to the failure count to verify both shapes succeed
- Library: output contains no instance values; test with a document whose property values are secret strings, verify they never appear
- Library: property-name annotations show matched names as an array, never property values or indices
- CLI: jsonschema validate -output detailed schema.json doc.json produces exactly one JSON output object on stdout
- CLI: jsonschema validate -output detailed schema.json doc1.json doc2.json produces two JSON lines, worst exit code wins
- CLI: absoluteKeywordLocation is valid URI syntax; parse with net/url and verify no error for schema keys with space, '%' and non-ASCII
- CLI: exit codes unchanged (0 valid, 1 invalid, 2 error) for detailed and verbose
- CLI: a report exceeding DefaultMaxReportBytes exits 2, outputs empty stdout and error message mentioning 'report bytes'

## Migration

No data or schema changes required. Code parsing absoluteKeywordLocation must now handle percent-encoding: space becomes %20, '%' becomes %25, non-ASCII becomes UTF-8 bytes in hex (%C3%A9 for é). The encoding is applied once in Resource.AbsoluteLocation, consistent across all output shapes. The existing DefaultMaxReportBytes cap now applies to detailed and verbose; no separate limit added. Raising the cap requires Options.MaxReportBytes in code; CLI flag -report-bytes is planned for story 193. Passing tree nodes no longer consume MaxOutputUnits; only failures count, same as flag and basic.

## Rollback

Revert the commit. The 'detailed' and 'verbose' output shapes become unrecognized (exit code 2 error). All flag and basic output is byte-identical except for absoluteKeywordLocation fragment encoding. Reverting fully restores pre-encoding behavior; however, schema keys with spaces, '%' or non-ASCII will again produce invalid URI fragments.

## Known limitations

- Passing leaf keywords with no subschema and no annotation (e.g., passing minLength, maxLength) have no tree node and never appear in verbose. Recording passing leaf units on every Validate call was ruled out to avoid per-call overhead. Only passing subschema nodes are shown.
- absoluteKeywordLocation appears only where already known from the tree node or unit; no absolute location is synthesized.
- Annotation values come from the schema (title, format, etc.) or are lists of matched property and array indices, never instance values. Property-name annotations list matched keys as an array.
- CLI has no -report-bytes flag to raise DefaultMaxReportBytes for detailed and verbose. Library callers raise it with Options{MaxReportBytes: N}. The flag is story 193.
- Result now holds the compiled root node and decoded instance, keeping them alive for as long as the caller holds the Result. Validate does no extra work; tree rebuilding only happens on Output calls for shapes that need it.

## Concerns

- Breaking change in all output shapes: absoluteKeywordLocation now percent-encodes schema-key fragments per RFC 3986. Keys with spaces, '%', or non-ASCII characters (like 'a b', '50%', 'é') are now encoded as '%20', '%25', '%C3%A9'. Code parsing this field must handle URI percent-encoding. Existing tests asserting unencoded locations need updating. The operator accepted this fix for spec compliance (story 188).
- AC6 partially deferred: the story proves the cap is enforced in library and CLI for detailed and verbose; it shows error exit codes and error messages. However, the CLI flag to raise the cap does not exist yet (story 193), so only library callers can raise it via Options.MaxReportBytes. The test for raising the cap is library-level, not CLI-level.
- Hash verification gap: output_unchanged_test.go pins flag and basic output SHA-256 values from base commit c941ecb to prove they are unchanged. The diff shows the claim but not the verification that the hashes were taken before the change; an operator should spot-check that the hashes match the base commit.
- Performance trade-off: detailed and verbose each rebuild the tree on every Output call, doubling CPU cost for those shapes. Flag and basic are unaffected. The tree is never cached, so Result stays thread-safe, but long-lived Results or many calls to Output will incur repeated evaluation.
- Memory trade-off: Result now keeps the root compile.Node and decoded instance, increasing size for each Result held by the caller. Validate does not rebuild the tree, so the fast path (flag and basic) has no extra cost.
