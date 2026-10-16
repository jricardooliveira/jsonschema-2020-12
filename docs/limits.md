# Default limits

Limits stop hostile input from hanging or exhausting the validator. Exceeding
one is an error (CLI exit code 2; in the library, `Compile`, `CompileWithOptions`
or `Validate` return an error wrapping the module's internal
`*jsonvalue.LimitError`), never a valid or invalid result. The error names the limit and
its value, never the offending input.

| Limit | Default | Raise it with | Applies to |
|-------|---------|---------------|------------|
| Number digits | 10000 (`DefaultMaxNumberDigits`) | `Options.MaxNumberDigits`, `--max-number-digits` | every number in the schema and the document |
| Number exponent | 10000 (`DefaultMaxNumberExponent`) | `Options.MaxNumberExponent`, `--max-number-exponent` | every number in the schema and the document |
| Nesting depth | 10000 (`jsonvalue.MaxDepth`) | fixed for now | arrays and objects |
| Input size | 64 MiB per input | fixed for now (CLI only) | schema file, document file, stdin |

## Number caps

- The digit cap counts every digit written in the integer and fraction parts,
  leading and trailing zeros included, so `0.000…01` and `1000…0` cannot slip
  past it.
- The exponent cap is the largest absolute written exponent (`1e10000` and
  `1e-10000` pass by default, `1e10001` does not). Leading zeros in the
  exponent do not count.
- Both are checked on the raw number text before any big-number work. Numbers
  within the caps are kept exactly (`math/big`); comparison and equality never
  expand a power of ten from the exponent.
- In the library, `CompileWithOptions(schema, Options{...})` fixes the caps for
  the schema and every later `Validate` call. A zero field means the default;
  a negative field is an error. In the CLI, an explicit value of 0 or less is a
  usage error.
- Raising the caps far above the defaults lets a caller opt into more expensive
  number work; the defaults stay safe. With absurdly raised caps an exponent
  that no longer fits in 64 bits is still a "number exponent" limit error.

## Input changes in this release

- A leading UTF-8 byte order mark is now accepted.
- Invalid UTF-8, unpaired `\u` surrogate escapes and numbers over the caps are
  now rejected as errors (exit 2). See [cli-contract.md](cli-contract.md).
