# ADR #31: Annotations collected in growing chunks, flattened lazily under sync.Once

- Date: 2026-10-17
- Story: #31
- Status: proposed (pending architecture sign-off)

## Decision
New file internal/eval/annotation_chunks.go holds a chunk store: a slice of chunks, where chunk k has capacity min(16<<k, 4096), plus a (chunk, offset) cursor. Mark returns the total length. Truncate moves the cursor back and keeps emptied chunks, so later Adds reuse them. Flatten allocates exactly Len entries once. The annotation sink keeps its existing API and only swaps its backing storage, so eval.go, the keyword files and the child-context code from 206 need no edits. The store lives in the per-call evaluation context: fresh per check, never shared. The Result holds the store plus a sync.Once. The Once flattens, applies the existing order (annotation_order.go, precomputed segments), stores the sorted slice and nils the chunk references. Annotations() and Output() both go through it. Measurement compares the code just before this story with the code after it, against the operator's v1.1 f35d9e4 values.

## Affected modules
- internal/eval/annotation_chunks.go
- internal/eval/annotation_chunks_test.go
- internal/eval/annotation_chunks_random_test.go
- internal/eval/annotation_sink.go
- internal/eval/annotations.go
- annotations.go
- jsonschema.go
- annotations_once_race_test.go
- allocs_test.go
- docs/benchmarks.md
- cmd/jsonschema/golden_dbg_test.go

## Data / API impact
- DB impact: no
- API impact: no
- Migration needed: no

## Constraints for implementation
- Chunk capacities grow 16, 32, … and cap at 4,096; fixed-size chunks are forbidden.
- Fresh chunks per check; no sync.Pool or cross-call reuse.
- Flatten, sort and chunk release all happen inside the same sync.Once.
- The sort order must be exactly the current order: reuse annotation_order.go, do not reimplement it.
- Do not edit internal/eval/eval.go, internal/keywords/registry.go, testdata/golden/*, allocs_<topic>_test.go or docs/benchmarks/r3-*.md.
- The random test uses math/rand with a fixed seed printed on failure. It goes in its own new _test.go file.
- Standard library only; go.mod unchanged.
- Delete only cmd/jsonschema/golden_dbg_test.go, using git rm.

## Risks
- Rollback across a chunk boundary could leave stale entries visible if the cursor arithmetic is off; the random test targets this case.
- ValidateSmall: more than 16 annotations means several small blocks, so allocs/op could rise. The operator rule makes that a hold, not a raised limit.
- Releasing chunks outside the Once would race with concurrent readers; the race test covers this.
- If 206's per-depth child contexts each own a sink, every depth gets its own chunk store. Byte totals could differ from the prototype, so this must be measured.

## Breaking changes
- none

## Sensitivity
- none
