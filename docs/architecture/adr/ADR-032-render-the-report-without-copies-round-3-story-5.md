# ADR #32: Render the report without copies (round 3, story 5)

- Date: 2026-10-17
- Story: #32
- Status: proposed (pending architecture sign-off)

## Decision
Both changes stay inside internal/output, the package that renders flag, basic, detailed and verbose reports. Change 1 is a fast path at the top of the existing segment-decoding function. It sits in code all shapes share, so every shape benefits and no call site changes. Change 2 puts the size-hint arithmetic in a new file, internal/output/sizehint.go, so the crowded files are touched by one line at most. The hint is applied only where the basic report buffer is created. Report-byte counting and the limit check are not touched, so the limit fails at the same byte with the same error. The hint only sets capacity and never changes content. Measurements and judgement follow docs/benchmarks.md. 'Before' is the code just before this story, on the same machine.

## Affected modules
- internal/output/pointer.go
- internal/output/pointer_tilde_test.go
- internal/output/sizehint.go
- internal/output/sizehint_test.go
- internal/output/basic.go
- cmd/jsonschema/run.go
- cmd/jsonschema/reportbytes_boundary_test.go
- testdata/reportbytes/tilde_schema.json
- testdata/reportbytes/tilde_document.json
- allocs_test.go
- docs/benchmarks.md

## Data / API impact
- DB impact: no
- API impact: no
- Migration needed: no

## Constraints for implementation
- Standard library only; go.mod unchanged
- Never regenerate or edit testdata/golden/*; a golden mismatch fails the story
- The '~' fast path must produce identical output to the previous strings.Replacer decode, including the order of '~1' before '~0'
- Size hint = min(location-based size, remaining report-byte limit) + 1; basic shape only; no further cap
- Do not move or alter the report-byte counting or the limit check
- New tests go in new _test.go files; existing tests unchanged
- Tests derive limit and size values from the code or an unlimited run, never hard-coded
- No peak-memory or timing assertions in default test runs; TestCLIPeakMemory stays opt-in
- Record exact measured values and computations against the v1.1 f35d9e4 sources; no rounded or placeholder digits
- Touch crowded files (cmd/jsonschema/run.go, docs/benchmarks.md) minimally: one wiring line in run.go if the buffer is created there

## Risks
- The hint may over-reserve if the location-based estimate exceeds the real report size; the cap keeps it at or below the remaining limit, so the measured B/op decides
- The 30 MiB RSS target may be missed because stories 205 to 208 already lowered peak memory; recorded and held per criterion 8
- The BenchmarkCLI/5.5MB required drop (49,252,608 B/op, 1,902,785 allocs/op) comes from story 209 alone and may not be reachable from rendering changes; held for the operator if missed
- Golden output drift if the '~' fast path mishandles an edge case; the differential test and golden tests guard this

## Breaking changes
- none

## Sensitivity
- none
