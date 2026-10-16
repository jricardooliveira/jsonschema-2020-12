# Regular expressions: ECMA-262 versus RE2

`pattern` uses ECMA-262 regular expressions (Unicode mode, no flags). Go's `regexp` is RE2. The package `internal/ecmaregex` translates the pattern (`Translate`) and Go compiles the result. The translator fails closed: any construct it does not handle is a schema error (exit code 2), never passed to RE2 with its RE2 meaning. Matching is unanchored (`MatchString`). Patterns count characters as code points, like `maxLength`: `^.$` matches one emoji.

## Translated

| ECMA-262 | Meaning |
|---|---|
| `\d` `\D` | ASCII `0-9` / everything else |
| `\w` `\W` | ASCII `[0-9A-Za-z_]` / everything else |
| `\s` `\S` | ECMA WhiteSpace + LineTerminator (U+0009-000D, U+0020, U+00A0, U+1680, U+2000-200A, U+2028, U+2029, U+202F, U+205F, U+3000, U+FEFF) / everything else |
| `.` | any code point except `\n`, `\r`, U+2028, U+2029 |
| `\uXXXX` | that code point |
| `😀` | two `\u` escapes forming a surrogate pair become one code point |
| `\u{X...}` | code point up to U+10FFFF |
| `\xHH` | that code point |
| `\cX` | control character (X is an ASCII letter) |
| `\0` | NUL (not followed by a digit) |
| `\f \n \r \t \v` | the control characters |
| `\b` `\B` | word boundary (ASCII word characters); inside a class `\b` is backspace |
| `(?<name>...)` | named group; name is ASCII letters, digits, `_`, not starting with a digit |
| `(?:...)`, `(...)`, `* + ? {n} {n,} {n,m}`, `^ $ \|` | unchanged |
| `\p{...}` `\P{...}` | see property names below |
| `[]` / `[^]` | matches nothing / matches any code point |

The shorthands and the dot work inside character classes too. Negated shorthands inside a class (`[\D]`, `[^\S]`) are expanded to explicit ranges. In a class, a literal `[` is escaped, so `[[:alpha:]]` is the class of `[ : a l p h ]` followed by a literal `]`, never a POSIX class.

## Accepted as literals

- Identity escapes of ASCII punctuation: `\/ \_ \- \. \( \) \[ \] \{ \} \\ \^ \$ \* \+ \? \|` and similar.
- A lone `{` (not part of a valid quantifier such as `{2}`, `{2,}`, `{2,5}`), so `a{,5}` is the literal text `a{,5}`.
- A lone `}` or `]`.

## Schema errors

- Lookahead `(?=` `(?!` and lookbehind `(?<=` `(?<!`.
- Backreferences `\1`..`\9` and `\k<name>`.
- Inline flags and RE2 group syntax: `(?i)`, `(?i:...)`, `(?P<name>...)`; non-ASCII group names.
- RE2-only escapes `\z \A \Q...\E` and any unknown escape (`\a`, `\e`, escaped letters, escaped non-ASCII).
- Octal escapes (`\00`), malformed `\x`, `\u`, `\u{}`, `\c`, `\u{...}` above U+10FFFF.
- Lone surrogate escapes (`\ud800`, `\udc00`, `\ud83d` without a low half, `\u{D800}`).
- A class escape as a range endpoint (`[\d-z]`), an out-of-order range (`[z-a]`), `\B` in a class.
- Unterminated class, group name or property, a trailing backslash.
- Unsupported Unicode properties (below).
- Anything RE2 refuses after translation, for example a repeat count over 1000 (`a{2000}`).
- A pattern longer than the size limit.

Errors name the problem and its position in the pattern, and never include the instance.

## Unicode properties

`\p{...}` and `\P{...}` accept:

- General_Category as `\p{L}`, `\p{Lu}`, `\p{gc=Lu}`, `\p{General_Category=Uppercase_Letter}`. Short names are the Go tables (`L Lu Ll Lt Lm Lo M Mn Mc Me N Nd Nl No P Pc Pd Ps Pe Pi Pf Po S Sm Sc Sk So Z Zs Zl Zp C Cc Cf Cs Co`). Long names: Letter, Uppercase_Letter, Lowercase_Letter, Titlecase_Letter, Modifier_Letter, Other_Letter, Mark, Combining_Mark, Nonspacing_Mark, Spacing_Mark, Enclosing_Mark, Number, Decimal_Number, digit, Letter_Number, Other_Number, Punctuation, punct, Connector_Punctuation, Dash_Punctuation, Open_Punctuation, Close_Punctuation, Initial_Punctuation, Final_Punctuation, Other_Punctuation, Symbol, Math_Symbol, Currency_Symbol, Modifier_Symbol, Other_Symbol, Separator, Space_Separator, Line_Separator, Paragraph_Separator, Other, Control, cntrl, Format, Surrogate, Private_Use.
- Script as `\p{Script=Greek}` or `\p{sc=Greek}`, for every script in Go's `unicode.Scripts`.

Errors: `Script_Extensions` / `scx`, Cased_Letter (`LC`), Unassigned (`Cn`), binary properties such as `Alphabetic`, bare script names (`\p{Greek}`), and unknown names or values. Go has no table for these.

## Remaining differences

- No lookaround and no backreferences (RE2 has no such engine); such schemas are rejected, not approximated.
- No regex flags. Patterns behave as ECMA Unicode mode without `i`, `m`, `s`; `^` and `$` match only at the start and end of the string, and `.` never matches line terminators.
- `\b` and `\B` use ASCII word characters (as in ECMA without the `i` flag), matching `\w`.
- Unicode property coverage is limited to Go's tables, and Go's Unicode version may differ from the one ECMA-262 references.
- Case-insensitive matching, sticky/global behaviour and `lastIndex` do not exist here.
- Legacy (non-Unicode-mode) ECMA leniencies not listed under "Accepted as literals" are errors, for example octal escapes and identity escapes of letters.
- Matching runs in linear time; ECMA engines may be exponential on the same pattern.
- Group names are limited to ASCII letters, digits and `_`.

## Pattern size limit

A pattern may be at most `DefaultMaxPatternSize` bytes (64 KiB, counted in bytes of the schema's UTF-8 text). The check runs before translation and compilation. A longer pattern is a schema error (exit code 2). Currently the limit is a fixed constant in `internal/keywords/validation/pattern.go`; it is not yet configurable by a library option or CLI flag.
