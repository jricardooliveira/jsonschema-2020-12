# Formats

`format` is an annotation by default. It only fails validation in assertion
mode, which is on when any of these holds:

- the library option `Options.FormatAssertion` is true;
- the CLI flag `--format-assertion` is given;
- the schema's meta-schema lists the format-assertion vocabulary as required
  (an optional listing does not turn it on).

In assertion mode, checks apply only to string instances; every other JSON
type passes. Meta-schema validation of a schema never asserts format.
Failure reports give the instance and keyword locations, never the value.
Checks are hand-written, linear-time and use the standard library only.

## Checked formats

| Format | Rule |
|---|---|
| `date-time` | RFC 3339 5.6; `t`/`z` lowercase allowed; offset required; second 60 only at 23:59:60 UTC after the offset |
| `date` | RFC 3339 5.6 full-date, real calendar days incl. leap years |
| `time` | RFC 3339 5.6 full-time, offset required, same leap-second rule |
| `duration` | RFC 3339 appendix A ABNF; `P2W` alone; units chain only to the next one (`P1Y2M3D`, `PT1H2M` pass; `P1Y1D`, `PT1H1S`, `P1W3D`, `P`, `PT` fail) |
| `email` | RFC 5321 Mailbox: dot-atom or quoted local part (max 64), domain or `[ipv4]` / `[IPv6:...]` / general `[tag:content]` literal (tag not checked against the IANA registry), max 254 bytes |
| `idn-email` | partly checked, see below |
| `hostname` | RFC 1123: labels 1-63 of letters, digits, hyphen, no hyphen at an edge, max 253, trailing dot fails |
| `idn-hostname` | partly checked, see below |
| `ipv4` | RFC 2673 strict dotted-quad, no leading zeros |
| `ipv6` | RFC 4291 text forms, `::` compression, embedded IPv4 tail; zone suffix (`%`) fails |
| `uuid` | shape only: 8-4-4-4-12 hex digits, either case; version and variant digits are not examined |

## Partly checked: idn-hostname and idn-email

No Unicode/IDNA tables are available (standard library only; `golang.org/x/net/idna`
was declined). Plain-ASCII input follows `hostname` / `email` exactly. Input
with non-ASCII characters gets basic checks only:

- `idn-hostname`: valid UTF-8, at most 253 code points, labels of 1-63 code
  points, no trailing dot, no hyphen at a label edge, no space or control
  characters, ASCII characters limited to letters, digits and `-`.
- `idn-email`: valid UTF-8, at most 254 bytes, an unquoted local part of 1-64
  bytes whose ASCII characters are atext or `.` (so no stray punctuation such
  as `,` `(` `)` `<` `>` `;` `:` `[` `]` `\` `"`), with no spaces, control
  characters or empty dot segments, and a domain that is an address literal or
  passes `idn-hostname`. A quoted local part with non-ASCII characters fails.

Not checked: IDNA2008 code point categories (PVALID/CONTEXTJ/CONTEXTO),
normalization, bidi rules, `--` in label positions 3-4, Unicode punctuation,
ideographic full stops, RFC 6531 non-ASCII character categories.

## Not yet checked (always pass)

These formats are not built yet. In assertion mode they pass like unknown
formats, until a later story adds them:

`uri`, `uri-reference`, `iri`, `iri-reference`, `uri-template`,
`json-pointer`, `relative-json-pointer`, `regex`.

Unknown format names always pass and are never an error.
