# Default limits

Limits stop hostile input from hanging or exhausting the validator. Exceeding
one is an error, never a valid or invalid result: the CLI exits with code 2,
and the library returns a `*jsonschema.LimitError` from `Compile`,
`CompileWithOptions`, `Validate` or `Result.Output`. The error names the limit
and its value, never the offending input.

A zero `Options` field means the default; a negative field is an error. A CLI
value of 0 or less is a usage error. A limit can be raised or lowered, never
switched off.

| Limit | Default | CLI flag | Library option | Applies to |
|-------|---------|----------|----------------|------------|
| Nesting depth | 10000 (`DefaultMaxDepth`) | `--max-depth` | `Options.MaxDepth` | arrays and objects in schemas, registry files and documents |
| Input size | 64 MiB (`DefaultMaxInputBytes`) | `--max-input-bytes` | `Options.MaxInputBytes` | each schema, registry file and document, and stdin |
| Reference chain | 20000 (`DefaultMaxRefChain`) | `--max-ref-chain` | `Options.MaxRefChain` | `$ref`/`$dynamicRef` jumps on one evaluation path |
| Evaluation steps | 100000000 (`DefaultMaxSteps`) | `--max-steps` | `Options.MaxSteps` | work steps for one document |
| Pattern size | 65536 bytes (`DefaultMaxPatternSize`) | `--max-pattern-size` | `Options.MaxPatternSize` | each regular expression |
| Output units | 10000000 (`DefaultMaxOutputUnits`) | `--max-output-units` | `Options.MaxOutputUnits` | failure units for one document |
| Annotations | 10000000 (`DefaultMaxAnnotations`) | `--max-annotations` | `Options.MaxAnnotations` | annotations for one document |
| Report bytes | 256 MiB (`DefaultMaxReportBytes`) | `--max-report-bytes` | `Options.MaxReportBytes` | one rendered report |
| Number digits | 10000 (`DefaultMaxNumberDigits`) | `--max-number-digits` | `Options.MaxNumberDigits` | every number in the schema and the document |
| Number exponent | 10000 (`DefaultMaxNumberExponent`) | `--max-number-exponent` | `Options.MaxNumberExponent` | every number in the schema and the document |
| Registry file size | 64 MiB (`DefaultMaxRegistryFileBytes`) | `--max-ref-file-bytes` | `Registry.MaxFileBytes` | each file supplied to the registry |
| Registry total size | 256 MiB (`DefaultMaxRegistryTotalBytes`) | `--max-ref-total-bytes` | `Registry.MaxTotalBytes` | all registry files together |

## Number caps

- The digit cap counts every digit written in the integer and fraction parts,
  leading and trailing zeros included, so `0.000…01` and `1000…0` cannot slip
  past it.
- The exponent cap is the largest absolute written exponent (`1e10000` and
  `1e-10000` pass by default, `1e10001` does not). Leading zeros in the
  exponent do not count.
- Both are checked on the raw number text before any big-number work. Numbers
  within the caps are kept exactly (`math/big`).

## Fixing and rendering

In the library, the limits are fixed when the schema is compiled and apply to
every later `Validate` call. The step, output-unit and annotation budgets are
fresh for each document. A limit error on stderr ends with
`(raise with --<flag>)` naming the flag.

See [cli.md](cli.md) for the flags and [library.md](library.md) for the options.
