# Command-line tool

```
jsonschema validate [flags] <schema-file> [<document-file>|-]...
```

The schema comes from a file path. Each document comes from a file path, or
from standard input when no document is given or the name is `-`. Flags go
before the schema file. Single and double dashes are both accepted.

## Examples

One valid document (the default output format is `basic`):

```sh
jsonschema validate docs/examples/schema.json docs/examples/valid.json
```

```json
{"valid":true,"annotations":[{"valid":true,"instanceLocation":"","keywordLocation":"/properties","annotation":["age","name"]}]}
```

Exit code: 0

One invalid document, as a flag report:

```sh
jsonschema validate -o flag docs/examples/schema.json docs/examples/invalid.json
```

```json
{"valid":false}
```

Exit code: 1

## Flags

| Flag | Default | Meaning |
|------|---------|---------|
| `-o` | `basic` | Output format: `flag`, `basic`, `detailed` or `verbose`. |
| `--output` | empty | Same as `-o`; when set it wins over `-o`. |
| `--base-uri` | empty | Base URI for a schema without `$id`. Absolute, no fragment. Empty means the schema file's `file:` URI (string work only, never read). |
| `--format-assertion` | off | Make `format` an assertion. Off means format is an annotation only. |
| `--ref` | none | Repeatable. A registry file: `<uri>=<path>` maps it to a base URI, a plain path uses the file's `$id`. |
| `--ref-dir` | none | Repeatable. A registry folder: `<uri>=<folder>`; the mapping is required. |
| `--max-depth` | 10000 | Deepest array/object nesting. |
| `--max-input-bytes` | 67108864 | Largest single input (schema, document, stdin). |
| `--max-ref-chain` | 20000 | `$ref`/`$dynamicRef` jumps on one evaluation path. |
| `--max-steps` | 100000000 | Work steps for one document. |
| `--max-pattern-size` | 65536 | Longest regular expression, in bytes. |
| `--max-output-units` | 10000000 | Failure units one document may produce. |
| `--max-report-bytes` | 268435456 | Size of one rendered report. |
| `--max-number-digits` | 10000 | Digits a number may write. |
| `--max-number-exponent` | 10000 | Largest absolute written exponent. |
| `--max-annotations` | 10000000 | Annotations one document may record. |
| `--max-ref-file-bytes` | 67108864 | Size of each registry file. |
| `--max-ref-total-bytes` | 268435456 | Size of all registry files together. |

A `--ref` or `--ref-dir` value is split at the first `=`. It is a mapping only
when the text before it is an absolute URI (scheme of 2 or more characters, no
whitespace); otherwise the whole value is a path (prefix `./` to force a
path). Only the files and folders you list are ever read; a reference to
anything else is an unresolved reference.

Every `--max-*` value must be a whole number greater than 0. `0`, a negative
value or a non-number is a usage error. The limits are explained in
[limits.md](limits.md).

`-h` prints the usage line and exits with 2.

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Every document is valid. |
| 1 | At least one document is invalid and none could not be checked. |
| 2 | Error: the check could not be done. |

These are always errors (2), never invalid: a usage error, an unreadable file,
broken JSON, a schema that fails its meta-schema, an unknown required
vocabulary, an unresolved reference, and an exceeded limit. On an error the
tool prints one line on stderr, for example:

```text
schema docs/examples/bad-schema.json: invalid against its meta-schema at "/type": rule "enum" failed
```

Error lines never contain the content of the schema or the document.

## Several documents

With exactly one document, stdout is exactly one output object of the chosen
format and nothing else.

With several documents, stdout has one line per document, in the order given.
Each line is `{"document":<path>,"output":<the unchanged output object>}`; a
document that could not be checked gives `{"document":<path>,"error":<text>}`
instead. The worst result sets the exit code: 2 if any document could not be
checked, else 1 if any is invalid, else 0.

```sh
jsonschema validate -o flag docs/examples/schema.json docs/examples/valid.json docs/examples/invalid.json
```

```json
{"document":"docs/examples/valid.json","output":{"valid":true}}
{"document":"docs/examples/invalid.json","output":{"valid":false}}
```

Exit code: 1
