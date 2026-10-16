# CLI contract

```
jsonschema validate [--max-number-digits N] [--max-number-exponent N] <schema-file> [<document-file>|-]
```

The document is read from `<document-file>`, or from standard input when it is
omitted or given as `-`. Only one document is supported for now.

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | The document is valid. Stdout is `{"valid":true}`. |
| 1 | The document is invalid. Stdout is `{"valid":false}`. |
| 2 | The document could not be checked: usage error, unreadable file, broken JSON, bad schema or an exceeded limit. Stdout is empty and stderr holds one line. |

Exit code 2 never means "invalid". Error lines give a location (byte offset,
line and column) and a fixed reason, never the content of the schema or the
document. A usage error may echo only the flag name as typed.

## Flags

| Flag | Default | Meaning |
|------|---------|---------|
| `--max-number-digits N` | 10000 (`jsonschema.DefaultMaxNumberDigits`) | Most digits a number may write in its integer and fraction parts, leading and trailing zeros included. |
| `--max-number-exponent N` | 10000 (`jsonschema.DefaultMaxNumberExponent`) | Largest absolute value of a number's written exponent. |

Both caps apply to the schema and to the document. `N` must be a whole number
greater than 0: `0`, a negative value or a non-number is a usage error (exit 2,
one stderr line, `invalid value for flag -<name>` or
`flag -<name> must be greater than 0`), naming the flag whose value failed.
An unknown flag is reported as `unknown flag -<name>`, without any `=value`. `-h` / `-help` print the usage line and exit 2.

Defaults and the other built-in limits are listed in [limits.md](limits.md).

## Accepted input

Schemas and documents must be one JSON value (RFC 8259) in UTF-8:

- A single leading UTF-8 byte order mark is accepted; a second one is broken JSON.
- Whitespace around the value is accepted. Anything else after the value,
  including a second value, is broken JSON.
- Bytes that are not valid UTF-8 are broken JSON.
- A `\u` escape that leaves a surrogate without its partner (for example
  `"\uD800"`) is broken JSON. A valid surrogate pair decodes to one code point.
- An object that repeats a member name, at any depth, is broken JSON. Names are
  compared after unescaping, so `"a"` and `"a"` are the same name. The
  error gives the location of the repeat.
- A number over the digit or exponent cap is a limit error.

All of these exit with code 2 and an empty stdout.
