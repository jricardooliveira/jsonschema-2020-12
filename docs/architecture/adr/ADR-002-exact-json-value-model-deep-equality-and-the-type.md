# ADR #2: Exact JSON value model, deep equality, and the type, enum and const keywords

- Date: 2026-10-16
- Story: #2
- Status: proposed (pending architecture sign-off)

## Decision
Revision 2 of the previous design. Everything the review accepted stays as it was: the file mapping (compiler.go is jsonschema.go, the flags live in cmd/jsonschema/run.go, main.go is not touched, the existing spec_test.go runner replaces internal/speccases), the `any` instance type with registry.go, eval.go and compile.go unchanged, the *SyntaxError Reasons "invalid UTF-8" and "unpaired surrogate escape", the *LimitError Limit names "number digits" and "number exponent", NumberLimits with DefaultMaxNumberDigits = DefaultMaxNumberExponent = 10000, Options in numberlimits.go (zero means the default, a negative value is an error, an explicit CLI value of 0 or less is a usage error), the BOM, trailing-data and cap-on-raw-text rules, the type, enum and const keywords, and the one allowed change to TestValidateSchemas (`{"type":"string"}` becomes `{"x-unknown":"string"}`). Only the two security fixes change.

Fix 1 (CPU amplification): the exact value is built once, at decode time. The new file internal/jsonvalue/number.go defines `type Number struct{ text string; rat *big.Rat; isInt bool }`. Its fields are unexported and it is used only as *Number, created only by the decoder and never changed afterwards. newNumber(text) runs only after the cap check has passed on the raw text. It is the only place in the package that calls big.Rat.SetString. It increments an unexported `var ratParses atomic.Int64`, which in-package tests read. The library decode path, DecodeLimited(data, NumberLimits), puts *Number leaves in the tree. Comparisons never parse again: `(*Number).IsInteger()` returns the cached isInt, `Cmp(a, b *Number) int` calls a.rat.Cmp(b.rat), and `(*Number).Rat() *big.Rat` returns new(big.Rat).Set(n.rat), a copy, so callers cannot change the shared value. Equal has a fast path: when a.text == b.text the numbers are equal and Cmp is skipped. Different texts with the same value (1 and 1.0) go to Cmp on the cached rats. The shared rats are only read (Cmp, IsInt, Sign, Set as a source), which is safe for concurrent use. Because const and enum values are decoded once with the schema, they hold their *Number for the life of the compiled Schema. Compatibility: TestDecodeValid (178, which must stay unchanged) expects Decode to return json.Number leaves. Decode and DecodeLimited therefore share one internal routine, `decode(data, lim, exact bool)`, with identical strict checks: BOM, pre-scan, duplicate names, depth, trailing data and the number cap. With exact=false (Decode, default limits) it keeps json.Number and builds no rat. With exact=true (DecodeLimited) it builds *Number. jsonschema.go uses only DecodeLimited. Equal, IsInteger and the keywords accept *Number. A json.Number is not a model value; Equal treats it as a different kind, and the doc comment says so.

Fix 2 (parser differential): the pre-scan moves into its own file, internal/jsonvalue/prescan.go, as `scanText(data []byte, onString func(start, end, fffd int)) *SyntaxError`. It is one linear pass. Outside strings it checks UTF-8 with utf8.DecodeRune. Inside a string it tracks escapes: `\\` consumes two bytes, `\"` does not end the string, and `\u` needs exactly 4 hex digits, else it leaves the rest to encoding/json's syntax error. A high surrogate must be followed directly by a `\u` low surrogate, and a lone low surrogate is an error. Invalid raw bytes inside a string are "invalid UTF-8". For every string literal it finds, including member names, it reports the byte span from the opening to the closing quote and fffd, the number of U+FFFD it legitimately holds (a literal U+FFFD rune or a \uFFFD escape). decode() calls scanText(data, nil). The new prescan_test.go holds a table test and a native Go fuzz target that check the same property. Whenever scanText accepts an input and json.Valid accepts it too: (1) the spans scanText reports, in order, decoded with json.Unmarshal(data[start:end], &s), equal the string tokens encoding/json's Decoder yields, in order and with the same count, so the two agree on every string boundary and escape; (2) each decoded string is valid UTF-8 and holds exactly fffd U+FFFD runes, so no invalid byte and no lone surrogate can come back as a silent replacement character. A third check: whenever Decode succeeds, the same FFFD invariant holds for every string in the tree. The seed corpus is added with f.Add only, so it stays text and plain `go test` runs it.

The rest is unchanged from the previous design. DecodeLimited enforces the cap on the raw token before newNumber: mantissa digits are counted, and the exponent has its leading zeros stripped and its length checked before Atoi. Equal does a type switch over nil, bool, string, *Number, []any (in order) and map[string]any (same size, every key equal). Values of different kinds are never equal. type maps *Number to number, and to integer as well when IsInteger() is true. enum and const use Equal. Compile errors name only the keyword or the offending type name from the schema, never instance content.

## Affected modules
- internal/jsonvalue/decode.go
- internal/jsonvalue/prescan.go
- internal/jsonvalue/limits.go
- internal/jsonvalue/number.go
- internal/jsonvalue/equal.go
- internal/jsonvalue/decode_test.go
- internal/jsonvalue/prescan_test.go
- internal/jsonvalue/number_test.go
- internal/jsonvalue/equal_test.go
- jsonschema.go
- jsonschema_test.go
- numberlimits.go
- numberlimits_test.go
- cmd/jsonschema/run.go
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

## Data / API impact
- DB impact: no
- API impact: yes
- Migration needed: no

## Constraints for implementation
- Standard library only. go.mod stays without require lines.
- Do not change internal/keywords/registry.go, internal/eval/eval.go, internal/compile/compile.go or cmd/jsonschema/main.go. The instance type stays `any`.
- Every existing 178 test passes unchanged. The single exception is the `{"type":"string"}` entry in TestValidateSchemas, which becomes `{"x-unknown":"string"}`. Decode keeps returning json.Number leaves; jsonschema.go calls only DecodeLimited.
- Decode and DecodeLimited share one internal decode(data, lim, exact) routine. Every strict check is identical in both paths; only the number leaf type differs.
- newNumber is the only caller of big.Rat.SetString in the package, and it increments ratParses (sync/atomic). Equal, Cmp and IsInteger never parse text. The shared *big.Rat is only read (Cmp, IsInt, Sign, and Set as a source); Rat() returns a copy. Number has no exported fields and no setters.
- Equal on two *Number returns true at once when their texts match, and otherwise uses Cmp on the cached rats. number_test.go decodes one number made of DefaultMaxNumberDigits digits and one with a DefaultMaxNumberExponent exponent. It records ratParses, then calls Equal, Cmp and IsInteger on them many times (for example 10000), and asserts that the counter has not moved. It also checks that Equal on a cap-sized number written two ways (different text, same value) still adds no parse. equal_test.go runs Equal from several goroutines on shared values, so `go test -race` covers concurrent reads.
- ADR-001 must not regress: the `len(stack)+1 > MaxDepth` check runs before descending into '[' or '{'; error text is offset plus line:column with a fixed phrase, never a JSON Pointer, member name or input bytes; readCapped stays the only input path in run.go.
- Check the number cap on the raw token text before newNumber. Count mantissa digits. For the exponent, strip leading zeros and compare the digit-text length with len(strconv.Itoa(max)) before calling Atoi. Test 1e999999999, 1e-999999999, an exponent of more than 20 digits, and leading-zero exponents that are within the cap.
- prescan.go: scanText(data, onString) is the only UTF-8 and surrogate check, and decode() calls it with nil. It must handle runs of backslashes (an even count before a quote ends the string, an odd count does not), \u escapes split at the end of the input, a lone high surrogate, a lone low surrogate, a high surrogate followed by a non-surrogate \u, a high surrogate followed by an escaped backslash and then `uDC00`, invalid bytes inside and outside strings, and UTF-8-encoded surrogates (ED A0 80), which are invalid UTF-8.
- prescan_test.go has a table test and `func FuzzScanTextAgreesWithEncodingJSON(f *testing.F)`. Its seeds are added with f.Add only (no binary corpus files) and cover every case in the previous constraint plus valid pairs, a literal U+FFFD and \uFFFD. The property, checked whenever scanText accepts and json.Valid accepts: the reported spans, decoded with json.Unmarshal, match the Decoder's string tokens in count, order and value; each one is valid UTF-8; and strings.Count(s, "\uFFFD") equals the reported fffd. A second assertion: when Decode succeeds, every string in its tree satisfies the same FFFD invariant. With plain `go test` the seeds run as ordinary tests, so there is no network and no extra flags.
- Tests name DefaultMaxNumberDigits, DefaultMaxNumberExponent and MaxDepth instead of writing the numbers: build inputs of length default and default+1.
- Strip at most one leading BOM. Error offsets refer to the original input bytes, BOM included.
- Options: 0 means the default, and a negative value makes CompileWithOptions return an error. CLI: an explicit flag value ≤ 0 is a usage error (one line, exit 2). The resolved limits sit in unexported Schema fields and never change after Compile returns.
- Each keyword lives in its own file and registers from init() with the validation vocabulary URI. Bad keyword values (an unknown type name, a duplicate or empty type array, a non-array enum) are compile errors. testdata cases cite validation §6.1.1, §6.1.2 or §6.1.3, or RFC 8259 §6 or §4.
- Keyword code never puts the instance into an error or message.
- parse_errors_test.go covers CLI exit 2 for: a broken schema, a broken document file and broken stdin; duplicate names at depth; invalid UTF-8; \uD800; a trailing second value; an over-cap number with default flags; acceptance of that same number once the flags are raised; and a usage error for a flag value of 0 or a negative value. Stderr never contains a planted secret.

## Risks
- Files outside the reservation: jsonschema.go is the real compiler.go, cmd/jsonschema/run.go is where the flags live, and jsonschema_test.go needs the one accepted entry change. internal/jsonvalue/prescan.go and prescan_test.go are new files in the reserved internal/jsonvalue component. They keep the pre-scan separate so the differential test can target it. No other story holds them.
- Two leaf shapes exist: Decode returns json.Number only to keep 178's TestDecodeValid unchanged, while the library uses DecodeLimited's *Number. Later stories (181-185) must decode through DecodeLimited and use Number.IsInteger, Cmp, Rat and Equal, never json.Number or strconv.ParseFloat. Removing Decode's json.Number shape is a follow-up once 178's test may change.
- Memory: each number keeps both its text and a big.Rat, bounded by the cap, so about 4 KB at worst for a 10^10000 value. Numbers are built eagerly even when no numeric keyword reads them; that is linear in input size and bounded.
- Differential test scope: the fuzz property uses encoding/json as the oracle only for inputs that both accept. Inputs scanText rejects but encoding/json accepts are by design: we are stricter.
- Options is defined in numberlimits.go while three other stories plan an options.go. They must add fields to this Options struct with the Max<Thing> / --max-<thing> naming and not declare a second type, so the scheduler should order them after this story.
- When flag parsing fails, run.go's existing message names the first argument starting with '-', which can be a valid flag (for example `--max-number-digits=abc`). The exit code is still 2, but the message may mislead. The coder may narrow it with flags.Lookup if the change stays small.
- Raising the cap a long way lets the caller opt into expensive big-number work. That is the caller's choice, and the default stays safe.

## Breaking changes
- The `type` keyword is now enforced, which changes 178's TestValidateSchemas expectation for {"type":"string"}; that entry becomes an unknown keyword.
- Input that 178 accepted is now rejected as broken JSON (exit 2): invalid UTF-8, unpaired \u surrogates, and numbers over the digit or exponent cap. A leading BOM is now accepted (178 rejected it).

## Sensitivity
- security
