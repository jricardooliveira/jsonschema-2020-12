# INTENT #14

- Author: Joao Oliveira
- Date: 2026-10-16
- Project: PROJ-010
- Status: captured (the request as received, before interpretation)

## Problem

Collect 2020-12 annotations during evaluation, with their instance and keyword locations. Annotations are dropped from any subschema that fails. Collect these: annotation keywords such as title, description, default, deprecated, readOnly, writeOnly and examples; 'format' values, which by default are annotations only and never fail validation; contentEncoding, contentMediaType and contentSchema, which are annotations only; and unknown keywords, which are ignored for validation and collected as annotations. Unknown formats are never an error. Applicators record which properties and items they evaluated. Annotations appear in the output formats wherever the 2020-12 shapes include them. Write spec-derived tests.

## Operator clarifications
- When a document passes, the 2020-12 report shapes can also list the annotations that were collected, such as a title, a default value or a format name. Should the default 'basic' report from the command-line tool include them? — Yes, as the 2020-12 shape allows — A passing 'basic' report lists each annotation with its document location and rule location. 'flag' never lists them. The report gets longer, but it follows the standard shape exactly.
- An annotation's value comes from the schema, not from the checked document. Examples are a 'default' value, an 'examples' list, or the value of an unknown keyword. Earlier you said reports must never quote the document's data. May reports show these schema values? — Yes, schema values may be shown — The value is written by the schema author, not taken from the checked document, so the no-document-data rule still holds. This matches the standard shape, where an annotation carries its value.
- Collecting every annotation takes time and memory on large documents. Should the validator always collect them, or only when something needs them? — Always collect — This is simpler and gives one code path. Every check pays the cost, even a plain yes/no check.
- Besides the JSON report, should programmers using the Go library be able to read the collected annotations directly? For example, a program could ask for the 'default' value at a given document location so it can fill in a missing setting. — Yes, a simple list in the result — Recommended. Each passing result carries a list of annotations. Each entry has the document location, the rule location, the keyword and the value. This is the same information as the report, but programs do not need to parse JSON to get it.
- Because annotations are always collected, a crafted schema and document could create millions of them and use up memory. For example, a 'title' on every item of a list with a million items. Earlier you asked for built-in limits that stop hostile files. Should the number of collected annotations have a limit too? — Yes, a limit like the others (library option + CLI flag, raise or lower, never off; exceeding = exit 2, never a silent valid/invalid), but generous enough for any document within the 64 MiB input cap: default 10 million collected annotations per document. It must be consistent with story 193's work limit, which uses the same option and flag style.
- When a report lists several annotations, should they always come out in the same order? Scripts and evaluators that compare reports can treat a changed order as a difference. — Always the same order — Recommended. Annotations are ordered by document location, then by rule location, so the same input always gives exactly the same report. This is easy to test and easy to compare.
- When a document fails, some parts of the schema may still have passed and collected annotations, such as a 'title' on a field that was fine. Should a failing report also list those annotations? — Only in 'verbose' — Recommended. A failing 'basic' or 'detailed' report lists only the errors, so it stays short and easy for scripts to read. 'verbose', the full report, also shows the annotations from the parts that passed. 'flag' never lists them.
- The 2020-12 rules allow the full 'verbose' report to also show annotations that were thrown away because their part of the schema failed, marked as dropped. This can help someone find out why a schema did not behave as expected. Should 'verbose' show them? — No, leave them out — Recommended. The rules make this optional. Leaving it out keeps one simple rule: annotations from a failing part are discarded and appear nowhere.
- Operator note on refining again — Do not split. (1) The metadata keywords (title, description, default, deprecated, readOnly, writeOnly, examples), 'format' in annotation mode and contentEncoding/contentMediaType/contentSchema all behave as annotation-only keywords whose annotation is the keyword's value, so register them through ONE shared annotate-only helper with one file per vocabulary: internal/keywords/metadata/annotations.go, internal/keywords/formatannotation/format.go, internal/keywords/content/content.go (plus one test file each and testdata). This deliberate exception to one-file-per-keyword keeps the story within the file limit; story 195 later extends format.go for assertion mode. (2) Remove the dependency on story 193: use the limit style that already exists from stories 178/179 (a named default in the numberlimits.go pattern, a LimitError with a new Limit name, an Options field and a CLI flag); story 193 follows the same style. (3) Expose the per-subschema evaluation tree (each subschema's result, errors and annotations, including passing parts inside a failing evaluation) through an internal API that story 194 uses for detailed and verbose output; Result.Annotations() stays empty on failure.

## Open points the operator accepted
Each point below is decided: take the standard safe choice, record it as an assumption, and do not ask about it again.
- The applicator keyword files from stories 183 and 184 are not in the inventory yet. Their exact file names (for example patternproperties.go vs pattern_properties.go) and their vocabulary folder are assumed to be internal/keywords/applicator/; the claims must be adjusted to the real names.
- options.go, the numberlimits.go limit pattern and the LimitError limit-name field are expected from story 179 (started) or story 180. The inventory shows LimitError in internal/jsonvalue/decode.go but not its fields. If adding the new limit name needs an edit to decode.go, that file becomes a write claim.
- internal/output/basic.go comes from story 180, which is not built yet. The annotation-unit shape is assumed to fit its registered renderer without changes to the output registry.
- Whether the annotation limit counts every annotation recorded (including ones later dropped) or only those kept at the end is not settled by the operator. This proposal counts every recorded annotation, which bounds both work and memory.
- Whether EvalContext's annotate hook can be added without changing the struct in registry.go depends on its current fields. registry.go is claimed as a write.
- The proposal is 31 files and 6 tasks, close to the 32-file cap. If the applicator files turn out to need their own tests or the existing spec runner needs changes, split T3 (applicator annotations) into its own story that story 192 depends on.

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
