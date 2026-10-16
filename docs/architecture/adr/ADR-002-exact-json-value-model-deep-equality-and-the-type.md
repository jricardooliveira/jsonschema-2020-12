# ADR #2: Exact JSON value model, deep equality, and the type, enum and const keywords

- Date: 2026-10-16
- Story: #2
- Status: proposed (pending architecture sign-off)

## Decision
New package internal/jsonvalue holds the whole value model. The strict decoder uses encoding/json's Decoder.Token stream with UseNumber to read numbers as written, but adds its own checks. It tracks member names per object to reject duplicates. It validates raw input with utf8.Valid before decoding and checks surrogate escapes, because encoding/json silently replaces bad text with U+FFFD. It strips one leading BOM and requires io.EOF after the value. It checks digit count and exponent length on the number text before calling big.Rat.SetString, so a huge exponent cannot allocate. Number holds a *big.Rat (an exact decimal from mantissa and exponent) and exposes IsInteger, Cmp and the original text. Equal is in its own file. The type, enum and const keywords go in the validation vocabulary package, one file each, registered via init(). all.go already imports the vocabulary, so no hub is edited for the keywords. The library wiring touches the root compile entry and the CLI main.go only to route input through the decoder and to add the number-cap option and flag. The option lives in a new root file, numberlimits.go, so later limits from story 193 can sit beside it without editing it.

## Affected modules
- internal/jsonvalue/value.go
- internal/jsonvalue/number.go
- internal/jsonvalue/decode.go
- internal/jsonvalue/limits.go
- internal/jsonvalue/equal.go
- internal/jsonvalue/number_test.go
- internal/jsonvalue/decode_test.go
- internal/jsonvalue/equal_test.go
- compiler.go
- numberlimits.go
- numberlimits_test.go
- internal/keywords/registry.go
- internal/eval/eval.go
- cmd/jsonschema/main.go
- cmd/jsonschema/parse_errors_test.go
- internal/keywords/validation/type.go
- internal/keywords/validation/enum.go
- internal/keywords/validation/const.go
- internal/keywords/validation/type_test.go
- internal/keywords/validation/enum_test.go
- internal/keywords/validation/const_test.go
- testdata/validation/type.json
- testdata/validation/enum.json
- testdata/validation/const.json
- internal/speccases/runner.go
- internal/speccases/runner_test.go

## Data / API impact
- DB impact: no
- API impact: yes: the library gains a number-cap option and returns parse/limit errors for broken input; the CLI gains a number-cap flag; the keyword evaluate signature may change to jsonvalue.Value
- Migration needed: no

## Constraints for implementation
- Standard library only. go.mod stays without require lines.
- Check digit count and exponent size on the raw number text before any math/big conversion.
- Do not rely on encoding/json for UTF-8 or surrogate validation; it replaces bad text silently.
- Errors and keyword messages give locations only, never instance values.
- Tests read defaults from the exported constants instead of repeating numbers.
- Each keyword in its own file, registered via init(). No central keyword list or switch is edited.

## Risks
- The skeleton (story 178) may have defined the instance type in the evaluate signature as `any`. Changing it to jsonvalue.Value touches registry.go and eval.go, which every later keyword story depends on, so this must land before stories 181-185.
- encoding/json's Decoder.Token cannot see raw escapes, so lone-surrogate detection may need a small pre-scan or a hand-written string scanner.
- The number-cap flag overlaps with story 193's limit flags. Their names and option shape should line up.

## Breaking changes
- If the skeleton's keyword evaluate signature used a different instance type, it changes to jsonvalue.Value.

## Sensitivity
- security
