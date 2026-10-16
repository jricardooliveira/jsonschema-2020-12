# ADR #14: Annotation collection: metadata, format, content and unknown keywords, evaluated properties and items

- Date: 2026-10-16
- Story: #14
- Status: proposed (pending architecture sign-off)

## Decision
This revision settles the four boundary points the reviewer raised. Every file, name, constraint and test from the reviewed design stays the same, apart from the one file noted at the end.

The design itself is unchanged:
- Annotation state is per call and lives in internal/eval. A new file, internal/eval/annotations.go, defines the Annotation record and the evaluation-tree node: keyword location, absolute keyword location, instance location, valid, errors, annotations and children.
- When a node fails, its annotations and those of its whole subtree are discarded. The same file holds the deterministic sort, the DefaultMaxAnnotations constant and the counter that returns the LimitError.
- eval.go is touched only to build the tree around each subschema evaluation, to call the counter and to collect unknown keywords.
- registry.go gains an annotate hook on EvalContext. Any other new methods go in eval-side files where the language allows.
- compile.go keeps unknown keywords and their raw schema values instead of discarding them.
- internal/keywords/annotate.go provides RegisterAnnotationOnly(name, vocabularyURI). It registers a compile function that keeps the raw value and an evaluate function that always passes and records that value.
- The metadata, formatannotation and content packages each call it from init() in one file. This is a deliberate exception to one-file-per-keyword, chosen by the operator. Story 195 later extends format.go for assertion mode.
- The applicator files from stories 183 and 184 are extended to record their annotation results.
- The root package adds annotations.go, with the public Annotation type and Result.Annotations(). The accessor returns a copy, so the Result stays immutable.
- jsonschema.go only stores the list on the Result. internal/output/basic.go emits annotation units on a passing result. cmd/jsonschema/run.go adds the limit flag, and main.go is not touched.
- The compiled Schema stays immutable, because all collection state lives in the per-call context.

(1) The limit follows the MaxOutputUnits style exactly. I checked the existing code:
- MaxOutputUnits is a field of the Options struct in numberlimits.go. Options.maxUnits() turns 0 into DefaultMaxOutputUnits and rejects a negative value with a compile error.
- The root output.go re-exports eval.DefaultMaxOutputUnits.
- run.go treats an explicitly set non-positive limit flag as a usage error, inside its flags.Visit check.

MaxAnnotations mirrors this:
- eval.DefaultMaxAnnotations = 10_000_000 lives in internal/eval/annotations.go. The root annotations.go re-exports it as jsonschema.DefaultMaxAnnotations.
- The field MaxAnnotations int is added to the Options struct in numberlimits.go, next to MaxOutputUnits. The struct is defined there, so the field cannot go anywhere else.
- Options.maxAnnotations() goes in options.go, next to baseURI(). Unset or 0 means the default. A negative value is a compile error ('options: MaxAnnotations must not be negative').
- The CLI flag is -max-annotations, read through the existing intFlag helper. An explicitly set value of 0 or less is a usage error (exit 2), checked in the same Visit check.
- Exceeding the limit returns &jsonvalue.LimitError{Limit: "annotations", Max: n}, which gives exit 2. No value ever means 'collection off'.
- The counter increments on every recorded annotation, including ones later dropped, so both work and memory are bounded. This matches how the units collector counts.
- decode.go is only read, because LimitError already has the Limit and Max fields.

(2) Order. The sort is a stable sort with this comparator:
1. Instance location, compared token by token. Each JSON Pointer is split and each token decoded (~1 to /, ~0 to ~). Tokens are compared in byte order, and a shorter prefix sorts first.
2. Keyword location, compared the same way.
3. Absolute keyword location, as a byte compare, with empty sorting first.
4. Keyword name.
5. Evaluation order (the record sequence number).

The pointer decoding uses the existing internal/uri/pointer.go helpers. Tests cover two kinds of ties:
- the same instance and keyword location reached through different dynamic paths, for example two $ref paths to one $defs subschema, which differ only in absolute location or sequence;
- '/a~1b' against '/a/b', which must sort differently from a raw byte compare.

(3) The NDJSON wrapper ({document, output|error}) and the shape of a failing basic report do not change. Only a passing basic report gains annotation units.

(4) Values. This follows the operator decision already on record from stories 180 and 194. Annotation values come from the schema only. The one exception is the evaluated-location results of properties, patternProperties and additionalProperties: they list evaluated property NAMES. These names already appear in instanceLocation, so they disclose nothing new. Instance VALUES never appear. This rule is written in the godoc of the root Annotation type and in the internal/output/basic.go comment, and the ADR records it. A CLI test asserts that a secret instance value is absent from both passing and failing output.

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
- numberlimits.go
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
- API impact: yes
- Migration needed: no

## Constraints for implementation
- Only the standard library; go.mod stays without require lines.
- Annotation values come only from the schema. Evaluated property names may appear, because they are already in instanceLocation. Instance VALUES never appear in an annotation, a report or an error.
- Annotations from failing subschemas are discarded and appear nowhere, not even in the internal tree.
- The compiled Schema and everything reachable from it are immutable. Annotation state lives in the per-call evaluation context, and tests run under -race.
- MaxAnnotations copies MaxOutputUnits exactly. Unset or 0 means DefaultMaxAnnotations, a negative value is a compile error, and an explicit CLI value of 0 or less is a usage error. There is never a 'collection off' mode.
- The counter increments on every recorded annotation, including ones later dropped. Going over the limit gives LimitError{Limit: "annotations"}.
- The sort is stable. It compares instance location, then keyword location, both as decoded JSON Pointer tokens (byte order, prefix first), then absolute keyword location, then keyword name, then evaluation sequence. Tests include ties from different dynamic paths and a ~1-escaped token.
- The NDJSON wrapper and the failing-report shape are unchanged.
- Tests read DefaultMaxAnnotations instead of repeating 10000000.
- main.go is not edited, and no central keyword list or switch is edited: keywords register from init().

## Risks
- numberlimits.go is outside the reservation. The Options struct, and its MaxOutputUnits field, are defined there, so the MaxAnnotations field must be added there too (one line plus a comment). options.go keeps the maxAnnotations() method.
- The plan now has 32 files, which is the cap.
- Recording every annotation on large documents costs time and memory. The operator accepted this ('always collect'), and the limit bounds it.
- eval.go and compile.go are crowded files. Changes there must stay minimal and only wire in the code in annotations.go.
- There is no docs/ file in the plan for the 'names, not values' rule, to stay within the cap. The rule is recorded in the Annotation godoc and in the ADR.

## Breaking changes
- A passing 'basic' report now lists annotation units, so it is longer than before. The operator approved this.

## Sensitivity
- security
