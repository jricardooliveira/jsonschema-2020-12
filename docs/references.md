# References, base URIs and loops

Spec: 2020-12 core 8.2.1 `$id`, 8.2.2 `$anchor`, 8.2.3.1 `$ref`, 8.2.4 `$defs`, 8.3 `$comment`, 9.1, 9.2, 9.4.

## Base URI of the root schema
Used when the root has no `$id`:
1. `--base-uri` (CLI) / `Options.BaseURI` (library). It must be absolute and have no fragment (an empty `#` is dropped); otherwise it is an error (exit 2), never ignored.
2. The schema file's location as `file:///abs/path` (`Options.SchemaPath`; the CLI passes the schema argument). This is string work only: the path is not opened, stat'ed or resolved for symlinks.
3. Schema without a file: the fixed placeholder `https://stdin.invalid/schema.json` (`.invalid` never resolves; nothing is read).

## Resolution
- Every `$ref` is resolved once when the schema is loaded, so a broken reference is exit 2 even in a branch no document reaches.
- A reference to a resource not inside the loaded document is an unresolved-reference error. Nothing is fetched or read from disk.
- Pointer tokens are percent-decoded first, then `~1` and `~0` are unescaped. Invalid percent-encoding is an error.
- A plain-name fragment is percent-decoded before the anchor lookup, so `#%41` reaches anchor `A`.
- Every schema object is compiled once at load time and reused by every `$ref` that reaches it, so load time grows linearly with the schema, not with the number of `$ref` keywords.
- `$id` with a non-empty fragment, bad `$anchor`/`$dynamicAnchor` names (`^[A-Za-z_][-A-Za-z0-9._]*$`), and a repeated `$id` or anchor in one resource are schema errors (exit 2).

## Normalization
Declared `$id` values and resolved references get RFC 3986 6.2.2 syntax normalization before comparison: lowercase scheme and host, unescape unreserved characters only, uppercase escape digits, remove `.` and `..` segments. `https://Example.com/a%7Eb.json` equals `https://example.com/a~b.json`; path case, `%2F` vs `/`, or any other difference does not match.

## Non-schema targets
2020-12 core 9.4.2 leaves this undefined. A target that is an object or boolean is followed even under an unknown keyword (e.g. `#/definitions/x`) and compiled as a subschema with the base URI in effect there. Any other value (number, string, array, null) is an unresolved reference (exit 2).

## Loops
- Load time: schemas holding only `$ref` that point at each other are a schema error (exit 2).
- Evaluation: re-entering the same schema location at the same instance location on the current path is a loop error (exit 2). Recursion that descends into the document (trees, lists) is fine.
- Reference chain: at most `--max-ref-chain` / `Options.MaxRefChain` `$ref`/`$dynamicRef` jumps may be on one evaluation path. Past that, evaluation stops with a `reference chain` limit error (exit 2).
- Backstop: at most 4 × the nesting depth limit (`--max-depth` / `Options.MaxDepth`) schema nodes may be nested on one evaluation path. Past that, evaluation stops with a `nesting depth` limit error (exit 2), never a stack overflow.
- Steps: every schema-node entry (including through `$ref`) and every keyword evaluated counts on the per-call step counter (`--max-steps` / `Options.MaxSteps`). `$ref` branches that multiply at each level end with an `evaluation steps` limit error (exit 2), never a hang.

Defaults for all of these are in [limits.md](limits.md).

## Output locations
`keywordLocation` follows the evaluation path through `$ref` (core 12.3.1): a failure inside a referenced schema reached from `/properties/p/$ref` is reported at `/properties/p/$ref/minLength`, not at the target's own location such as `/$defs/a/minLength`. `absoluteKeywordLocation` gives the target's own absolute location where it is known; see [output-formats.md](output-formats.md).
