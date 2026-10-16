# SPEC #11: $dynamicRef and $dynamicAnchor with dynamic scope

- Date: 2026-10-16
- Type: feature
- Verdict: pass
- Status: passed gate 1 with nothing for the operator to decide (run #147, 2026-10-16 23:53 UTC)

## Problem

The validator has no 2020-12 dynamic referencing. Schemas that use $dynamicRef/$dynamicAnchor cannot be checked correctly, and that includes recursive extension of a schema and the official meta-schemas reached through "$dynamicRef": "#meta".

## Why this matters

$dynamicRef/$dynamicAnchor with dynamic scope is in the mandatory 2020-12 core scope of the brief. Story 189 (meta-schema validation) and story 192 (unevaluated* through $dynamicRef) need it.

## Acceptance criteria

- $dynamicAnchor defines a plain-name fragment in the same namespace as $anchor (2020-12 core 8.2.2). A $ref to '#name' resolves to a subschema that carries "$dynamicAnchor": "name".
- A $dynamicRef is first resolved like $ref against the current base URI. When that initially resolved target holds a $dynamicAnchor whose name matches the plain-name fragment, the reference resolves to the outermost schema resource in the dynamic scope that has a matching $dynamicAnchor (2020-12 core 8.2.3.2).
- When the initially resolved target has no matching $dynamicAnchor (bookending), or the fragment is not a plain name, $dynamicRef behaves exactly like $ref.
- Dynamic resolution works across embedded schema resources with their own $id. Entering an embedded resource adds it to the dynamic scope.
- Dynamic resolution works across schemas supplied through the schema registry (story 187), with the dynamic scope spanning resources from different documents.
- A duplicate plain name within one schema resource is a broken schema, with exit code 2 in the CLI and a schema error from the library, never 'invalid'. This covers two $dynamicAnchor with the same name, or a $dynamicAnchor and an $anchor with the same name. The rule is consistent with story 186.
- A $dynamicRef whose initial target cannot be resolved locally or from the registry is an unresolved-reference error: exit code 2, never 'invalid', with no network or disk access.
- Errors reached through $dynamicRef carry a keywordLocation that contains '/$dynamicRef' at the point of the reference, and the absoluteKeywordLocation of the dynamically resolved target. No instance value is ever quoted.
- $dynamicRef evaluation goes through story 186's evaluation-time loop check: re-entering the same schema location at the same instance location on the current evaluation path is a loop error. A test with a non-terminating $dynamicRef cycle ends with exit code 2 and does not hang.
- A spec-derived recursive-extension test passes: a base tree schema whose recursion uses $dynamicRef to '#node' is extended by a schema that $refs it and declares its own "$dynamicAnchor": "node". The extension's constraints then apply at every level of nesting, with matching positive and negative cases, including the same schema used without extension.
- The dynamic scope lives only in the per-call evaluation context. The compiled Schema is not modified during evaluation, and concurrent validation with the same compiled schema passes under go test -race.
- Positive and negative cases live as JSON in testdata/core/dynamicRef.json and testdata/core/dynamicAnchor.json, run by the shared table-driven runner, and each case cites its 2020-12 core section.

## Non-goals

- The configurable recursion limit with its library option and CLI flag (story 193); this story only uses story 186's loop check.
- unevaluatedProperties/unevaluatedItems through $dynamicRef (story 192).
- Meta-schema validation and the embedded meta-schemas (story 189).
- $recursiveRef/$recursiveAnchor from 2019-09 (older drafts are out of scope).

## Tasks

- **T1** — Index $dynamicAnchor names in the resource registry
  - purpose: Register each $dynamicAnchor as a plain-name fragment in the same namespace as $anchor, mark it as dynamic, and reject duplicate plain names within one resource as a schema error.
  - scope: internal/registry/dynamicanchor.go, internal/registry/dynamicanchor_test.go, internal/registry/anchor.go
  - done when: Registry tests show '#name' resolving to a $dynamicAnchor, a lookup reporting whether a resource has a matching dynamic anchor, and duplicate names (dynamic+dynamic, dynamic+static) failing compilation.
  - delivers: AC1, AC6
- **T2** — Dynamic scope in the per-call evaluation context
  - purpose: Keep the stack of entered schema resources in the evaluation context. Push the stack when evaluation enters a resource (root, embedded $id, $ref/$dynamicRef target) and pop it when evaluation leaves. Provide an outermost-match lookup by anchor name.
  - scope: internal/eval/dynamicscope.go, internal/eval/dynamicscope_test.go, internal/eval/eval.go
  - done when: Unit tests for push/pop and outermost lookup pass. eval.go changes only in the one place that enters and leaves a schema resource.
  - delivers: AC4, AC11
- **T3** — $dynamicAnchor and $dynamicRef keywords
  - purpose: Add keyword files that register through the keyword registry. $dynamicRef resolves its initial target at compile time like $ref, applies bookending, and at evaluation time picks the outermost matching resource from the dynamic scope. It evaluates through the existing loop check and reports '/$dynamicRef' keywordLocation with the resolved target's absoluteKeywordLocation.
  - scope: internal/keywords/core/dynamicanchor.go, internal/keywords/core/dynamicref.go, internal/keywords/core/dynamicref_test.go
  - done when: Keyword unit tests pass for dynamic resolution, bookending, an unresolved reference (schema error), error locations and the loop error.
  - depends on: T1, T2
  - delivers: AC2, AC3, AC5, AC7, AC8, AC9
- **T4** — Spec-derived dynamic reference test cases
  - purpose: Write JSON test cases citing 2020-12 core 8.2.2/8.2.3.2. They cover: $ref to a dynamic anchor; outermost resolution; bookending; embedded resources; registry-supplied resources; duplicate names; unresolved targets; error locations; the non-terminating cycle; and recursive extension with and without extension.
  - scope: testdata/core/dynamicRef.json, testdata/core/dynamicAnchor.json, internal/keywords/core/dynamicref_spec_test.go
  - done when: The shared runner runs both files green under go test -race ./...
  - depends on: T3
  - delivers: AC1, AC2, AC3, AC4, AC5, AC6, AC7, AC8, AC9, AC10, AC12
- **T5** — CLI contract tests for dynamic references
  - purpose: Prove the exit codes at the CLI boundary. A recursive-extension document gives 0 or 1. A duplicate plain name, an unresolved $dynamicRef and a non-terminating $dynamicRef cycle each give exit 2 without hanging. The basic output of an invalid document shows '/$dynamicRef' in keywordLocation.
  - scope: cmd/jsonschema/dynamicref_cli_test.go, cmd/jsonschema/testdata/dynamicref/
  - done when: The CLI tests pass, with a timeout guarding the cycle case.
  - depends on: T3
  - delivers: AC5, AC6, AC7, AC8, AC9, AC10
- **T6** — Concurrency check on dynamic scope
  - purpose: Validate many documents in parallel goroutines with one compiled schema that uses $dynamicRef, to show the dynamic scope never leaks into the compiled Schema.
  - scope: internal/eval/dynamicscope_race_test.go
  - done when: go test -race passes with parallel recursive-extension validations giving the expected results.
  - depends on: T2, T3
  - delivers: AC11
