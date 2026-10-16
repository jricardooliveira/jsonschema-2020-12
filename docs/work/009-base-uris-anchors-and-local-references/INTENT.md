# INTENT #9

- Author: Joao Oliveira
- Date: 2026-10-16
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Add 2020-12 core identification and referencing within one schema document. '$id' sets the base URI using RFC 3986 resolution, including for embedded schema resources with their own $id. '$anchor' defines plain-name fragments. '$defs' holds reusable subschemas. '$ref' resolves local JSON Pointer fragments, which includes ~0/~1 escaping and percent-decoding. It also resolves anchor fragments and relative references against the current base URI. '$ref' works alongside sibling keywords. '$comment' is accepted and has no effect on validation. A reference that cannot be resolved is an error with exit code 2, never 'invalid'. The validator never fetches anything from the network or from disk on its own. Write spec-derived positive and negative tests.

## Operator clarifications
- When should a broken reference be caught? Say a schema has a reference that points nowhere, but it sits in a branch that a given document never reaches. — When the schema is loaded — Recommended. Every reference is checked once, before any document is looked at. A broken schema always gives exit code 2, whatever the document is, so results never depend on which branch a document happens to take.
- The 2020-12 rules say each '$id' and each '$anchor' name must appear only once within a schema resource, but they do not say what to do when one is repeated. What should happen? — Treat it as a broken schema (exit 2) — Recommended. The error names the repeated '$id' or anchor. The validator never has to guess which of the two copies a reference means.
- A main schema without its own '$id' still needs a starting address, so that relative references inside it (like 'other.json#/x') can be worked out. The rules leave that address up to us. What should it be? — Its file location, with an optional override — Recommended. A schema read from a file gets that file's own address, for example file:///home/me/schema.json. The caller can set another base address through a library option or CLI flag. Schemas piped in without a file get a fixed placeholder address. Nothing extra is ever read from disk because of this.
- A reference can point somewhere that is not a schema at all, for example into a list of allowed values inside 'enum', or into an unknown keyword. The rules leave this case open. How should it be handled? — Follow it when the target value is an object or a boolean and use it as a schema (2020-12 core 9.4.2 leaves this undefined; real-world schemas often use "$ref": "#/definitions/x", and 'definitions' is an unknown keyword in 2020-12). Such a target is compiled as a subschema with the base URI in effect at that location. A target that is any other value (number, string, array, null) is an unresolvable reference: broken schema, exit 2. Document this choice.
- A schema can contain references that only point at each other and never reach a real rule. For example, rule A says 'same as B' and rule B says 'same as A', so checking any document would loop forever. How should the validator handle this? — Never reject a legitimate recursive schema: references that descend into the document (trees, linked lists, the official meta-schemas) are normal and must work. At load time, reject only pure reference chains that loop without any other keyword (A is only {"$ref": B} and B is only {"$ref": A}): broken schema, exit 2. Everything else is caught during evaluation, deterministically: re-entering the same schema location at the same instance location on the current evaluation path is a loop error (exit 2, never a hang), with the recursion limit as the backstop.
- Some '$id' and '$anchor' values break the 2020-12 writing rules. For example, an '$id' with a '#part' on the end, an anchor name that starts with a digit, or a reference with a broken percent code such as '%zz'. Should this story reject these itself? — Yes, reject them as a broken schema (exit 2), exactly per the 2020-12 rules: an $id with a NON-empty fragment is an error, but an $id ending in an empty fragment ('#') is accepted and the empty fragment is ignored; $anchor and $dynamicAnchor must match ^[A-Za-z_][-A-Za-z0-9._]*$; a reference with invalid percent-encoding is an error. The error names the location in the schema.
- A reference has to match an address declared with '$id'. Sometimes the two are written slightly differently but mean the same place. For example, the schema declares 'https://Example.com/a%7Eb.json' and a reference says 'https://example.com/a~b.json'. The difference here is a capital letter in the site name and an escaped character. Should these count as the same address? — Yes, apply the standard safe clean-up first — Recommended. Before comparing, both addresses get the safe clean-up the web address standard (RFC 3986) describes. The scheme and site name become lowercase, escaped characters that need no escaping are unescaped, escape codes use capital letters, and '.' and '..' segments are removed. Equivalent addresses then match, and nothing else does. This choice gets its own positive and negative tests and is documented.

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- The inventory is empty, so the exact skeleton layout from story 178 (still in progress) is unknown: the name of the Options type, whether options.go already exists, and how compile and eval expose hooks. If options.go already exists and other stories write to it, it becomes a shared file; the option could instead go in a new file.
- Dependency 183 (properties) is assumed so that recursive tree tests can descend into the document. If 184 (items) lands first, a linked-list test could use it instead. Sibling-keyword tests rely on 'type' and 'const' from 179.
- The recursion-limit backstop is assumed to exist minimally in the skeleton or eval. Story 193 makes it configurable; until then a fixed internal bound is used.
- The exact placeholder base URI for stdin schemas is not fixed by the operator. A documented fixed value is chosen here.
- The core keyword package may need to be listed in internal/keywords/all/all.go. The skeleton is expected to list every vocabulary already; if it does not, a one-line edit there would be needed.
- At 30 files the story fits the limit. If the skeleton requires extra wiring files, consider splitting the base URI option and CLI flag into a separate story.

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
