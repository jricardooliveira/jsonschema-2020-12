# Deviations

Everything here is a place where the validator is not exactly 2020-12 or not
exactly the brief. This page also holds the contents of the earlier pages on
regular expressions and formats.

## Regular expressions: ECMA-262 versus RE2

`pattern` and `patternProperties` use ECMA-262 regular expressions (Unicode
mode, no flags). Go's `regexp` is RE2. `internal/ecmaregex` translates the
pattern (`Translate`) and Go compiles the result. The translator fails closed:
any construct it does not handle is a schema error (exit code 2), never passed
to RE2 with its RE2 meaning. Matching is unanchored. Characters count as code
points, like `maxLength`: `^.$` matches one emoji.

### Translated

| ECMA-262 | Meaning |
|---|---|
| `\d` `\D` | ASCII `0-9` / everything else |
| `\w` `\W` | ASCII `[0-9A-Za-z_]` / everything else |
| `\s` `\S` | ECMA WhiteSpace + LineTerminator (U+0009-000D, U+0020, U+00A0, U+1680, U+2000-200A, U+2028, U+2029, U+202F, U+205F, U+3000, U+FEFF) / everything else |
| `.` | any code point except `\n`, `\r`, U+2028, U+2029 |
| `\uXXXX` | that code point; two escapes forming a surrogate pair become one code point |
| `\u{X...}` | code point up to U+10FFFF |
| `\xHH` | that code point |
| `\cX` | control character (X is an ASCII letter) |
| `\0` | NUL (not followed by a digit) |
| `\f \n \r \t \v` | the control characters |
| `\b` `\B` | word boundary (ASCII word characters); inside a class `\b` is backspace |
| `(?<name>...)` | named group; name is ASCII letters, digits, `_`, not starting with a digit |
| `(?:...)`, `(...)`, `* + ? {n} {n,} {n,m}`, `^ $ \|` | unchanged |
| `\p{...}` `\P{...}` | see Unicode properties below |
| `[]` / `[^]` | matches nothing / matches any code point |

The shorthands and the dot work inside classes too. Negated shorthands inside a
class (`[\D]`, `[^\S]`) are expanded to explicit ranges. A literal `[` inside a
class is escaped, so `[[:alpha:]]` is never a POSIX class.

### Accepted as literals

- Identity escapes of ASCII punctuation: `\/ \_ \- \. \( \) \[ \] \{ \} \\ \^ \$ \* \+ \? \|` and similar.
- A lone `{` that is not a valid quantifier (`a{,5}` is the literal text).
- A lone `}` or `]`.

### Rejected as a schema error (exit code 2)

- Lookahead `(?=` `(?!` and lookbehind `(?<=` `(?<!`.
- Backreferences `\1`..`\9` and `\k<name>`.
- Inline flags and RE2 group syntax: `(?i)`, `(?i:...)`, `(?P<name>...)`; non-ASCII group names.
- RE2-only escapes `\z \A \Q...\E` and any unknown escape (`\a`, `\e`, escaped letters, escaped non-ASCII).
- Octal escapes, malformed `\x`, `\u`, `\u{}`, `\c`, `\u{...}` above U+10FFFF.
- Lone surrogate escapes.
- A class escape as a range endpoint (`[\d-z]`), an out-of-order range (`[z-a]`), `\B` in a class.
- Unterminated class, group name or property; a trailing backslash.
- Unsupported Unicode properties (below).
- Anything RE2 refuses after translation, for example a repeat count over 1000 (`a{2000}`).
- A pattern longer than the pattern size limit (`--max-pattern-size`, see [limits.md](limits.md)); that case is a limit error, also exit code 2.

Error messages name the keyword, the pattern text, the problem and its
position. They never include the instance.

### Unicode properties

`\p{...}` and `\P{...}` accept General_Category (`\p{L}`, `\p{Lu}`,
`\p{gc=Lu}`, `\p{General_Category=Uppercase_Letter}`, short names from Go's
tables and the long names) and Script (`\p{Script=Greek}`, `\p{sc=Greek}`) for
every script in Go's `unicode.Scripts`. These are errors: `Script_Extensions` /
`scx`, `Cased_Letter` (`LC`), `Unassigned` (`Cn`), binary properties such as
`Alphabetic`, bare script names (`\p{Greek}`), and unknown names or values.

### Remaining differences

- No lookaround and no backreferences. Such schemas are rejected, not approximated.
- No regex flags. Patterns behave as ECMA Unicode mode without `i`, `m`, `s`; `^` and `$` match only at the start and end of the string, and `.` never matches line terminators.
- `\b` and `\B` use ASCII word characters.
- Unicode property coverage is limited to Go's tables; Go's Unicode version may differ from the one ECMA-262 references.
- Legacy (non-Unicode-mode) ECMA leniencies not listed under "Accepted as literals" are errors.
- Matching runs in linear time; ECMA engines may be exponential on the same pattern.

## Formats

`format` is an annotation by default; see [scope.md](scope.md) for how to turn
assertion on. In assertion mode only string instances are checked, every other
type passes. Unknown format names always pass and are never an error. Failure
reports give locations, never the value. All checks are hand-written,
linear-time and standard-library only.

| Format | Checked | What it misses |
|---|---|---|
| `date-time` | RFC 3339 5.6; lowercase `t`/`z` allowed; offset required; second 60 only at 23:59:60 UTC | none known |
| `date` | RFC 3339 full-date, real calendar days, leap years | none known |
| `time` | RFC 3339 full-time, offset required, same leap-second rule | none known |
| `duration` | RFC 3339 appendix A ABNF (`P2W` alone; units chain to the next one) | none known |
| `email` | RFC 5321 Mailbox: dot-atom or quoted local part (max 64), domain or address literal, max 254 bytes | the tag of a general address literal is not checked against the IANA registry |
| `idn-email` | **partly checked**, see below | see below |
| `hostname` | RFC 1123: labels 1-63 of letters, digits, hyphen; no hyphen at an edge; max 253; trailing dot fails | none known |
| `idn-hostname` | **partly checked**, see below | see below |
| `ipv4` | strict dotted-quad, no leading zeros | none known |
| `ipv6` | RFC 4291 text forms, `::`, embedded IPv4 tail; a zone suffix (`%`) fails | none known |
| `uuid` | shape only: 8-4-4-4-12 hex digits, either case | version and variant digits are not examined |
| `uri`, `uri-reference`, `iri`, `iri-reference` | **partly checked**: RFC 3986 / 3987 grammar in one forward scan (scheme, authority, path, query, fragment, percent-escapes, IP-literals; IRIs also admit ucschar and iprivate) | the host is checked only as reg-name or IP-literal syntax (no dotted-quad rule, no IDN or bidi rules of RFC 3987 4.2); port digits are not range-limited |
| `uri-template` | **partly checked**: RFC 6570 level 4 grammar, never expanded | operators reserved for the future (`= , ! @ \|`) are rejected; nothing beyond the grammar is judged |
| `json-pointer` | RFC 6901 | none known |
| `relative-json-pointer` | relative-json-pointer draft grammar | none known |
| `regex` | **partly checked**: ECMA-262 Unicode-mode (u flag) syntax in one forward scan | see below |

### idn-hostname and idn-email

No Unicode or IDNA tables are available (standard library only;
`golang.org/x/net/idna` was declined). Plain ASCII input follows `hostname`
and `email` exactly. Input with non-ASCII characters gets basic checks only:

- `idn-hostname`: valid UTF-8, at most 253 code points, labels of 1-63 code points, no trailing dot, no hyphen at a label edge, no space or control characters, ASCII characters limited to letters, digits and `-`.
- `idn-email`: valid UTF-8, at most 254 bytes, an unquoted local part of 1-64 bytes whose ASCII characters are atext or `.`, no spaces, control characters or empty dot segments, and a domain that is an address literal or passes `idn-hostname`. A quoted local part with non-ASCII characters fails.

Not checked: IDNA2008 code point categories (PVALID, CONTEXTJ, CONTEXTO),
normalization, bidi rules, `--` in label positions 3-4, Unicode punctuation,
ideographic full stops, RFC 6531 non-ASCII character categories.

### regex format

The `regex` format judges syntax only, under the strict ECMA-262 `u` grammar:
lookahead, lookbehind, backreferences and named groups are **valid** here, even
though `pattern` rejects the ones RE2 cannot run. Group names use Go's
`unicode` tables as an approximation of ID_Start/ID_Continue; the `(?i:...)`
modifier groups are rejected; `\p{...}` names come from the same table as
`pattern`.

## Known differences

Places where the finished tool differs from the brief, or leaves work behind:

1. **Regex that RE2 cannot run.** The brief asks for ECMA-262 patterns; lookahead, lookbehind and backreferences are rejected as schema errors in `pattern` and `patternProperties`, as the brief accepts. The `regex` format still reports them as valid, so a pattern can pass `format: regex` and fail to compile as a `pattern`.
2. **Partly checked formats.** `idn-hostname`, `idn-email`, the four URI/IRI formats, `uri-template`, `regex` and `uuid` (shape only) do not check everything the brief's "as far as practical" could cover; see the table above.
3. **Verbose output is only partly produced.** A passing leaf assertion without annotation has no unit, so verbose lists evaluated subschemas and annotations only. For a failing document the detailed and verbose output can be the same. See [output-formats.md](output-formats.md).
4. **Superseded pages still on disk.** `docs/regex-deviations.md` and `docs/formats.md` hold the same subjects as this page, and `docs/cli-contract.md` describes an early CLI (one document, valid/invalid only; its current content is [cli.md](cli.md)). This story could not delete them; they are stale and should be removed in a separate story. `docs/regex-deviations.md` also says the pattern size limit is a fixed constant, but it is now configurable. `docs/references.md` (reference resolution notes) is also partly stale: it says `absoluteKeywordLocation` is not emitted and that the evaluation-depth backstop is fixed, but both changed (see [output-formats.md](output-formats.md) and [limits.md](limits.md)).
5. **Stray debug test files** are still present: `dbg_test.go` and `internal/registry/dbg_tmp_test.go`.
6. **Duplicated spec runners** are still present: `internal/speccases` and `spec_applicator_test.go`.

No other difference from the brief is known.
