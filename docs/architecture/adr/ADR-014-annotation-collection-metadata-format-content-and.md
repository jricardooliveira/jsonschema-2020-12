# ADR #14: Annotation collection: metadata, format, content and unknown keywords, evaluated properties and items

- Date: 2026-10-16
- Story: #14
- Status: approved by the operator at Checkpoint 2 (run #144, 2026-10-16 21:34 UTC)

## Decision
This revision adds the one bound the reviewer asked for: a cap on the total bytes of a rendered report. Everything else from the reviewed design stays, except three file-list fixes that reading the code made necessary. They are listed at the end.

The design itself is unchanged:
- Annotation state is per call and lives in internal/eval. A new file, internal/eval/annotations.go, defines the Annotation record and the evaluation-tree node. A node holds keyword location, absolute keyword location, instance location, valid, errors, annotations and children.
- When a node fails, its annotations and those of its whole subtree are thrown away. The same file holds the deterministic sort, DefaultMaxAnnotations and the counter that returns the LimitError.
- eval.go is touched only to build the tree around each subschema evaluation, to call the counter and to collect unknown keywords.
- registry.go gains an annotate hook on EvalContext.
- compile.go keeps unknown keywords and their raw schema values instead of dropping them.
- internal/keywords/annotate.go provides RegisterAnnotationOnly(name, vocabularyURI). It registers a compile function that keeps the raw value and an evaluate function that always passes and records that value. The metadata, formatannotation and content packages each call it from init() in one file. This is the operator's chosen exception to one-file-per-keyword. Story 195 later extends format.go.
- The applicator files are extended to record their annotation results.
- The root annotations.go holds the public Annotation type and Result.Annotations(). The accessor returns a copy.
- jsonschema.go only stores the list (and now the byte cap) on the Result.
- internal/output/basic.go emits annotation units on a passing result.
- cmd/jsonschema/run.go adds -max-annotations. main.go is not touched.
- The compiled Schema stays immutable.

(1) MaxAnnotations copies the MaxOutputUnits style:
- eval.DefaultMaxAnnotations is 10_000_000. The root package re-exports it as jsonschema.DefaultMaxAnnotations.
- The field Options.MaxAnnotations goes in numberlimits.go next to MaxOutputUnits.
- Options.maxAnnotations() works like maxUnits(): 0 means the default, and a negative value is a compile error ('options: MaxAnnotations must not be negative').
- The CLI flag -max-annotations is read with the existing intFlag helper. An explicit value of 0 or less is a usage error in the same flags.Visit check.
- Going over the limit returns &jsonvalue.LimitError{Limit: "annotations", Max: n}, which gives exit 2.
- The counter counts every recorded annotation, including ones later dropped.

(2) The sort is stable. It compares instance location, then keyword location, both as decoded JSON Pointer tokens (byte order, shorter prefix first, using internal/uri/pointer.go). Ties break on absolute keyword location (byte compare, empty first), then keyword name, then the evaluation sequence number. Tests cover ties from different dynamic paths and '/a~1b' against '/a/b'.

(3) The NDJSON wrapper and the failing basic shape do not change. Only a passing basic report gains annotation units.

(4) Annotation values come from the schema only. The exception is the evaluated property NAMES from properties, patternProperties and additionalProperties, which already appear in instanceLocation. This rule is written in the Annotation godoc and in the basic.go comment.

(5) NEW, the report-bytes cap:
- DefaultMaxReportBytes = 256 << 20 lives in internal/output/registry.go. The root output.go re-exports it as jsonschema.DefaultMaxReportBytes, the same way it re-exports DefaultMaxOutputUnits.
- Options.MaxReportBytes int goes in numberlimits.go, with maxReportBytes() next to maxUnits(): 0 means the default, and a negative value is a compile error ('options: MaxReportBytes must not be negative').
- The Schema stores the resolved cap, and each Result carries it.
- No CLI flag in this story. Story 193 adds it in the same intFlag/Visit style.
- The cap is checked while rendering, not after:
  - registry.go gains a small capped writer, Budget, built on bytes.Buffer. Its Write returns &jsonvalue.LimitError{Limit: "report bytes", Max: max} as soon as the running length would go over the cap. It never returns partial output.
  - registry.go also gains Report{Valid, Units, Annotations []eval.Annotation}, RegisterReport(name, func(Report, *Budget) error) and RenderReport(name, Report, maxBytes) ([]byte, error).
  - The existing Register(name, Renderer) and Render(name, valid, units) keep their signatures. Register wraps an old-style renderer so its bytes go through the same Budget. Render calls RenderReport with DefaultMaxReportBytes. So flag.go, run.go's shape check and the existing output tests (registry_test, basic_test, flag_test, units_test) are not edited.
  - There is one map, and the duplicate-name check covers both entry points.
- basic.go switches to RegisterReport. It writes the opening bytes, then marshals each annotation or error unit on its own and writes it to the Budget, then writes the closing bytes. Memory is bounded by the cap plus one unit, and a unit's size is bounded by the 64 MiB input cap.
- Story 194 registers detailed and verbose with RegisterReport and gets the same enforcement.
- Result.Output(shape) in output.go calls output.RenderReport with the result's valid flag, units, annotations and cap.
- The CLI already gives exit 2 when rendering fails, but run.go currently replaces the error with 'cannot render result'. run.go now passes the LimitError text through. With one document, stdout stays empty and stderr gets one line. With several, that document gets the existing ADR-003 {document, error} line and the worst result gives exit 2. No report is ever truncated.

File-list fixes found by reading the code:
- output.go and internal/output/registry.go were missing from the reviewed design. Without them, basic.go could not receive the annotations, because the current Renderer only gets (valid, units). They are now listed.
- To stay within 32 files, maxAnnotations() moves from options.go into numberlimits.go beside maxUnits(), where the matching limit helper already lives. options.go, a crowded file, is no longer touched.
- internal/keywords/annotate_test.go is dropped. The helper is tested through the metadata, format and content test files, which call it for every keyword it registers.

## Affected modules
- internal/eval/annotations.go
- internal/eval/annotations_test.go
- internal/eval/eval.go
- internal/keywords/registry.go
- internal/compile/compile.go
- internal/keywords/annotate.go
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
- numberlimits.go
- output.go
- annotations_test.go
- internal/output/registry.go
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
- Annotation values come only from the schema. Evaluated property names may appear because they are already in instanceLocation. Instance VALUES never appear in an annotation, a report or an error.
- Annotations from failing subschemas are discarded and appear nowhere, not even in the internal tree.
- The compiled Schema and everything reachable from it are immutable. Annotation state and the render Budget are per call. Tests run under -race.
- MaxAnnotations and MaxReportBytes copy MaxOutputUnits exactly: the field is in numberlimits.go, its resolver method sits beside maxUnits(), 0 means the default and a negative value is a compile error. There is never an 'off' mode.
- The annotation counter counts every recorded annotation, including ones later dropped. Going over gives LimitError{Limit: "annotations"}.
- The report-bytes cap is enforced inside rendering through output.Budget, unit by unit. Going over returns LimitError{Limit: "report bytes", Max: cap} and nil bytes, never a truncated report. Every shape goes through the Budget: old-style renderers are wrapped by Register, and new shapes use RegisterReport.
- Register and Render keep their current signatures, so flag.go and the existing output tests are not edited.
- run.go passes a render LimitError through as the stderr message (exit 2, empty stdout for one document). Other render failures keep 'cannot render result'.
- The sort is stable and compares decoded JSON Pointer tokens, with the documented tie-breaks. Tests include dynamic-path ties and a ~1-escaped token.
- The NDJSON wrapper and the failing-report shape are unchanged.
- Tests read DefaultMaxAnnotations and DefaultMaxReportBytes instead of repeating the numbers. The report-bytes test uses a large 'default' value under 'items', applied to many array items, with a lowered Options.MaxReportBytes, and checks for *LimitError 'report bytes'. The same input renders fine under the default cap. A negative MaxReportBytes is a compile error.
- main.go is not edited, and no central keyword list or switch is edited: keywords register from init().

## Risks
- output.go and internal/output/registry.go are outside the reservation. The current Renderer only receives (valid, units), so annotations and the byte cap cannot reach basic.go without these edits. Neither file is crowded, and the existing Register/Render signatures are kept so the other output callers are not touched.
- numberlimits.go now holds both new Options fields and their resolver methods, next to maxUnits(). options.go is no longer touched, which frees a crowded file.
- internal/keywords/annotate_test.go was dropped to stay within 32 files. RegisterAnnotationOnly is tested only through the three vocabulary test files.
- The CLI path for an exceeded report-bytes cap cannot be triggered from the command line until story 193 adds the flag. In this story it is covered at the library level, and run.go's pass-through of the error text is a small reviewed change.
- Wrapped old-style renderers (flag) are checked against the cap after json.Marshal. Their output has a fixed small size, so this is safe. New shapes must use RegisterReport to keep memory bounded.
- The plan has 32 files, which is the cap.
- Recording every annotation costs time and memory. The operator accepted this, and the annotation limit and the report-bytes limit bound it.
- eval.go and compile.go are crowded files. Changes there must stay minimal and only wire in the code in annotations.go.

## Breaking changes
- A passing 'basic' report now lists annotation units, so it is longer than before. The operator approved this.
- Result.Output can now return a 'report bytes' LimitError, but only above 256 MiB under the default.

## Sensitivity
- security
