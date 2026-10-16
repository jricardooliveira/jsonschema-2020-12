# ADR #14: Annotation collection: metadata, format, content and unknown keywords, evaluated properties and items

- Date: 2026-10-16
- Story: #14
- Status: proposed (pending architecture sign-off)

## Decision
Annotation state is per call and lives in internal/eval. A new file, internal/eval/annotations.go, defines the Annotation record and the evaluation-tree node (keyword location, absolute keyword location, instance location, valid, errors, annotations, children). When a node fails, its annotations and those of its subtree are discarded. The same file holds the deterministic sort (instance location, then keyword location), the DefaultMaxAnnotations constant and the counter that returns the LimitError. eval.go is touched only to build the tree around each subschema evaluation, to call the counter and to collect unknown keywords. registry.go gains an annotate hook on EvalContext; any new methods go in eval-side files where the language allows. compile.go keeps unknown keywords and their raw schema values instead of discarding them. internal/keywords/annotate.go provides RegisterAnnotationOnly(name, vocabularyURI), which registers a compile function that keeps the raw value and an evaluate function that always passes and records the value. The metadata, formatannotation and content packages each call it from init() in one file, a deliberate exception to one-file-per-keyword chosen by the operator. Story 195 later extends format.go for assertion mode. The applicator keyword files built in stories 183 and 184 are extended to record their annotation results. The root package adds annotations.go with the public Annotation type and the Result.Annotations() accessor (a copy, so the Result stays immutable). jsonschema.go only stores the list on the Result, and options.go adds MaxAnnotations. internal/output/basic.go emits annotation units on a passing result. cmd/jsonschema/run.go adds the limit flag; main.go is untouched. The compiled Schema stays immutable, because all collection state lives in the per-call context.

## Affected modules
- internal/eval/annotations.go
- internal/eval/annotations_test.go
- internal/eval/eval.go
- internal/keywords/registry.go
- internal/compile/compile.go
- internal/keywords/annotate.go
- internal/keywords/annotate_test.go
- internal/keywords/metadata/annotations.go
- internal/keywords/metadata/annotations_test.go
- internal/keywords/formatannotation/format.go
- internal/keywords/formatannotation/format_test.go
- internal/keywords/content/content.go
- internal/keywords/content/content_test.go
- internal/keywords/applicator/properties.go
- internal/keywords/applicator/patternproperties.go
- internal/keywords/applicator/additionalproperties.go
- internal/keywords/applicator/prefixitems.go
- internal/keywords/applicator/items.go
- internal/keywords/applicator/contains.go
- annotations.go
- jsonschema.go
- options.go
- annotations_test.go
- internal/output/basic.go
- cmd/jsonschema/run.go
- cmd/jsonschema/annotations_test.go
- testdata/annotations/metadata.json
- testdata/annotations/format.json
- testdata/annotations/content.json
- testdata/annotations/unknown.json
- testdata/annotations/applicators.json

## Data / API impact
- DB impact: no
- API impact: yes: adds Result.Annotations(), the Annotation type, Options.MaxAnnotations and a CLI limit flag; passing basic reports gain annotation units
- Migration needed: no

## Constraints for implementation
- Only the standard library; go.mod stays without require lines.
- Annotation values only ever come from the schema. Instance values never appear in an annotation, a report or an error.
- Annotations from failing subschemas are discarded and appear nowhere, not even in the internal tree.
- The compiled Schema and everything reachable from it are immutable. Annotation state lives in the per-call evaluation context, and tests run under -race.
- Tests read DefaultMaxAnnotations instead of repeating 10000000.
- main.go is not edited, and no central keyword list or switch is edited: keywords register from init().
- The ordering is deterministic and covered by a test.

## Risks
- Recording every annotation on large documents costs time and memory. This is accepted by the operator (always collect) and bounded by the limit.
- eval.go and compile.go are crowded files, so changes there must stay minimal and only wire in the code in annotations.go.
- The applicator files from stories 183 and 184 may not exist yet, or may be named differently, when this story starts.

## Breaking changes
- A passing 'basic' report now lists annotation units, so it is longer than before (operator-approved).

## Sensitivity
- security
