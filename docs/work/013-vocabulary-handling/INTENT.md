# INTENT #13

- Author: Joao Oliveira
- Date: 2026-10-17
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Add '$vocabulary' support as 2020-12 core describes. When a schema's meta-schema declares vocabularies, the validator processes them. A required vocabulary it does not know is an error with exit code 2. An optional unknown vocabulary is ignored. The format-assertion vocabulary is recognised so a later story can honour it. Use the embedded meta-schemas and custom meta-schemas supplied through the registry. Write spec-derived tests for required, optional and unknown vocabularies.

## Operator clarifications
- A custom meta-schema can list only some of the standard vocabularies. For example, it might list the 'core' and 'applicator' groups but leave out the 'validation' group, which holds rules like 'minimum' and 'maxLength'. When a schema uses such a meta-schema, what should happen to rules from a group that is not listed? — Follow the spec: treat them as unknown — Recommended. Rules from unlisted groups are ignored for validation and only kept as annotations. 'minimum: 5' would not reject the value 3 under that meta-schema. This matches 2020-12 exactly.
- Some meta-schemas say nothing about vocabularies at all, because they have no '$vocabulary' entry. Which rules should apply to a schema that uses one of them? — All standard 2020-12 groups — Recommended. The schema behaves exactly as if it used the official 2020-12 meta-schema. This is the safe, unsurprising choice.
- This story only recognises the format-checking vocabulary ('format-assertion'). A later story makes it actually check formats. Until that later story is done, what should happen when a meta-schema marks this vocabulary as required? — Accept it and record it; formats stay annotation-only for now — Recommended. There is no error and no exit code 2. The loaded schema notes that format checking was requested, so the later story only has to switch it on. Tests here check that it is accepted and recorded.
- The 2020-12 rules say the 'core' group (the basic rules such as $id, $ref and $defs) is always needed. What should happen when a custom meta-schema leaves 'core' out of its vocabulary list, or marks it as optional? — Treat it as a broken schema — Recommended. The schema is rejected with a clear message and exit code 2. This is the strict reading of the spec and catches a mistake in the meta-schema early.
- A schema can contain a separate schema inside it that has its own '$id' and its own '$schema'. That inner schema might use a different meta-schema with a different vocabulary list. Which vocabulary list should apply inside it? — The inner schema's own list — Recommended. Each part follows the meta-schema it declares, as 2020-12 allows. Example: the outer schema checks 'minimum', while an inner schema whose meta-schema leaves out the validation group does not.
- When a schema is refused because its meta-schema requires a vocabulary the validator does not know, what should the error message tell the person running the tool? — The unknown vocabulary's address and which meta-schema required it — Recommended. Example: 'meta-schema https://example.com/my-meta requires unknown vocabulary https://example.com/vocab/money'. It is written to the error output with exit code 2, and the library returns the same facts as a distinct error type that programs can check for.
- Older versions of JSON Schema are out of scope. Sometimes a schema still names an older version's meta-schema in '$schema', for example the Draft 7 address 'http://json-schema.org/draft-07/schema#'. What should happen then? — Refuse it as a broken schema — Recommended. Exit code 2 with a clear message such as 'unsupported $schema: draft-07; only 2020-12 is supported'. The library returns the same facts as an error. Nothing is checked with rules from the wrong version.
- A custom meta-schema could contain a badly written '$vocabulary' list. For example, an entry might be marked "yes" instead of true or false, or its address might not be a full web address. What should happen? — Refuse it as a broken schema — Recommended. Exit code 2 with a message that names the meta-schema and the bad entry. This matches the official rule that each entry must be a full address marked true or false.
- Operator note on refining again — Add to this story (Operator observation on story 189's build): $schema and $vocabulary are currently treated as unknown keywords and show up as annotations in the basic output of a valid document (keywordLocation '/$schema'). Register $schema, $vocabulary (and $comment if not yet) as core keywords that produce NO annotation and have no validation effect, with a test that a passing basic report for {"$schema":".../2020-12/schema","type":"string"} contains no $schema annotation. Keep everything else in the accepted plan.

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- The inventory is truncated. I could not confirm how compile.go dispatches keywords, nor whether internal/keywords/registry.go already exposes each keyword's vocabulary URI to another file in the package. If it does not, registry.go (a crowded file) needs a small read accessor, and its write claim would have to be added.
- Story 189 may already refuse non-2020-12 '$schema' in metaschema.go. If so, T2 only adjusts the message, and metaschema_cli_test.go may need updating. It is not claimed here.
- Whether cmd/jsonschema/run.go already maps every compile error to exit code 2 is assumed, not verified. If the new error type needs explicit mapping, run.go (crowded) would need a write claim.
- Existing tests (for example annotations_test.go and cmd/jsonschema/annotations_test.go) may expect a '/$schema' or '/$comment' annotation, and would break when these become silent core keywords. They are not claimed.
- Story 195 depends on this story's format-assertion record; the accessor name must be agreed so 195 can read it.

## Proposed outcome

See `SPEC.md` in this directory — the spec-agent's acceptance criteria are
the machine-readable statement of the outcome, derived from the problem above.

## Affected users and systems

See `PLAN.md` → *Files that change*, and the project's `project-spec.json`
for the systems in scope.

## Constraints

The project's `PROJECT_RULES.md` and `project-spec.json` (stack, conventions,
NFRs and the `forbidden` list) are injected into every agent and bind this work.

## Open questions

Recorded in `SPEC.md` → *Open questions*. A non-empty list parks the run at
Checkpoint 1 for the operator rather than failing it.
