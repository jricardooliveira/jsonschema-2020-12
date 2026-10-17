# Scope

Only JSON Schema Draft 2020-12 is supported. The tests are written from the
2020-12 core and validation specification text.

## Mandatory 2020-12 behaviour

1. **Types and validation keywords, with exact JSON number semantics.** `1.0`
   is an integer, `multipleOf` is exact on decimals, big integers and decimals
   keep full precision (`math/big`). Keywords: `type`, `enum`, `const`,
   `multipleOf`, `maximum`, `exclusiveMaximum`, `minimum`, `exclusiveMinimum`,
   `maxLength`, `minLength` (counted in Unicode code points), `pattern`,
   `maxItems`, `minItems`, `uniqueItems` (deep JSON equality), `maxContains`,
   `minContains`, `maxProperties`, `minProperties`, `required`,
   `dependentRequired`. `enum` and `const` also use deep JSON equality.
2. **Applicators.** `properties`, `patternProperties`, `additionalProperties`,
   `propertyNames`, `dependentSchemas`, `prefixItems`, `items`, `contains`,
   `allOf`, `anyOf`, `oneOf`, `not`, `if`/`then`/`else`, and the boolean
   schemas `true` and `false`.
3. **Core.** `$schema`; `$vocabulary` (meta-schema validation of the schema,
   required and optional vocabularies; an unknown required vocabulary is an
   error); `$id` and base URI resolution per RFC 3986, including embedded
   resources; `$anchor`; `$ref` (local JSON Pointer with `~0`/`~1` escaping and
   percent-decoding, relative, and external through the registry);
   `$dynamicRef` and `$dynamicAnchor` with dynamic scope; `$defs`; `$comment`.
4. **Annotations and unevaluated keywords.** Annotation collection, and
   `unevaluatedProperties` and `unevaluatedItems` across every applicator,
   including through `$ref` and `$dynamicRef`.
5. **Unknown keywords** are ignored for validation and collected as
   annotations.
6. **Invalid schemas.** A schema that fails its meta-schema is an error (exit
   code 2), never "invalid".
7. **Output formats.** `flag` and `basic` are always available; see
   [output-formats.md](output-formats.md).
8. **Content keywords.** `contentEncoding`, `contentMediaType` and
   `contentSchema` are annotations only.
9. **`format` is annotation-only by default.** A format value never fails
   validation, and an unknown format is never an error.

Also mandatory: no network access (the validator never fetches a URI and reads
only the files and folders the caller lists), the embedded 2020-12
meta-schemas, the [limits](limits.md) against hostile input, and reports that
never contain instance values.

## Optional features

- **Format assertion.** `format` becomes an assertion in any of three ways:
  the CLI flag `--format-assertion`, the library option
  `Options.FormatAssertion`, or a meta-schema that lists the format-assertion
  vocabulary as required (an optional listing does not turn it on). Only string
  instances are checked. Meta-schema validation never asserts format. Which
  formats are checked, and how fully, is on [deviations.md](deviations.md).
- **Detailed and verbose output.** `-o detailed` and `-o verbose`. Verbose is
  only partly produced, see [output-formats.md](output-formats.md).
