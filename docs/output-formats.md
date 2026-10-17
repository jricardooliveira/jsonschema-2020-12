# Output formats

Choose a format with `-o flag|basic|detailed|verbose` (default `basic`). With
one document, stdout is exactly one object of the chosen format. With several
documents each object is wrapped in a label, see [cli.md](cli.md).

Reports carry the instance location (`instanceLocation`), the keyword location
(`keywordLocation`) and, where known, the absolute keyword location
(`absoluteKeywordLocation`). They never carry instance values: only where a
check failed and which rule failed. They are safe to store in logs.

`absoluteKeywordLocation` is left out when it is unknown, or when the
evaluation path crossed no reference. The examples below use no `$ref`, so none
shows it.

All examples use the schema `docs/examples/schema.json`.

## flag

Only the verdict.

```sh
jsonschema validate -o flag docs/examples/schema.json docs/examples/invalid.json
```

```json
{"valid":false}
```

Exit code: 1

## basic (default)

A flat list. An invalid document gets `errors`, each with the location of the
failing value, the keyword location and a message. A valid document gets
`annotations`, one unit per collected annotation.

```sh
jsonschema validate docs/examples/schema.json docs/examples/invalid.json
```

```json
{"valid":false,"errors":[{"instanceLocation":"/age","keywordLocation":"/properties/age/minimum","error":"value is less than minimum"},{"instanceLocation":"/name","keywordLocation":"/properties/name/minLength","error":"string is shorter than minLength"}]}
```

Exit code: 1

```sh
jsonschema validate docs/examples/schema.json docs/examples/valid.json
```

```json
{"valid":true,"annotations":[{"valid":true,"instanceLocation":"","keywordLocation":"/properties","annotation":["age","name"]}]}
```

Exit code: 0

## detailed

A hierarchy that follows the schema structure, condensed as 2020-12 core
12.4.2 describes: passing nodes without annotations are dropped, a node with
one child is replaced by that child, and annotations under a failing node are
not kept. A failing document has `errors`, a passing one has `annotations`.

```sh
jsonschema validate -o detailed docs/examples/schema.json docs/examples/invalid.json
```

```json
{"valid":false,"keywordLocation":"","instanceLocation":"","errors":[{"valid":false,"keywordLocation":"/properties/age/minimum","instanceLocation":"/age","error":"value is less than minimum"},{"valid":false,"keywordLocation":"/properties/name/minLength","instanceLocation":"/name","error":"string is shorter than minLength"},{"valid":false,"keywordLocation":"/properties/age","instanceLocation":"/age","errors":[{"valid":false,"keywordLocation":"/properties/age/minimum","instanceLocation":"/age","error":"value is less than minimum"}]},{"valid":false,"keywordLocation":"/properties/name","instanceLocation":"/name","errors":[{"valid":false,"keywordLocation":"/properties/name/minLength","instanceLocation":"/name","error":"string is shorter than minLength"}]}]}
```

Exit code: 1

```sh
jsonschema validate -o detailed docs/examples/schema.json docs/examples/valid.json
```

```json
{"valid":true,"keywordLocation":"","instanceLocation":"","annotations":[{"valid":true,"keywordLocation":"/properties","instanceLocation":"","annotation":["age","name"]}]}
```

Exit code: 0

## verbose

The same nested shape, with every evaluated subschema node and every annotation
kept, passing or failing.

```sh
jsonschema validate -o verbose docs/examples/schema.json docs/examples/valid.json
```

```json
{"valid":true,"keywordLocation":"","instanceLocation":"","annotations":[{"valid":true,"keywordLocation":"/properties","instanceLocation":"","annotation":["age","name"]},{"valid":true,"keywordLocation":"/properties/age","instanceLocation":"/age"},{"valid":true,"keywordLocation":"/properties/name","instanceLocation":"/name"}]}
```

Exit code: 0

For a failing document, the verbose output of this schema equals the detailed
output, because nothing was condensed away.

## Only partly produced

- A passing keyword that has no subschema and records no annotation (for
  example a passing `minLength`) has no node in the evaluation tree. Verbose
  therefore lists every evaluated subschema and every annotation, but not a
  separate unit for each passing leaf assertion. Detailed would drop such a
  unit anyway.
- `absoluteKeywordLocation` appears only where the node or unit already carries
  one.
- Both nested formats are rendered on demand from a second evaluation of the
  document. A limit hit during that step (see [limits.md](limits.md)) is an
  error with exit code 2.
