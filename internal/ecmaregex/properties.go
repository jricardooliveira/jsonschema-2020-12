package ecmaregex

import (
	"fmt"
	"strings"
	"unicode"
)

// longCategory maps ECMA-262 long General_Category value names to the short
// names Go's regexp understands. Cased_Letter (LC) and Unassigned (Cn) have no
// Go table and are therefore not listed: they are schema errors.
var longCategory = map[string]string{
	"Letter": "L", "Uppercase_Letter": "Lu", "Lowercase_Letter": "Ll",
	"Titlecase_Letter": "Lt", "Modifier_Letter": "Lm", "Other_Letter": "Lo",
	"Mark": "M", "Combining_Mark": "M", "Nonspacing_Mark": "Mn",
	"Spacing_Mark": "Mc", "Enclosing_Mark": "Me",
	"Number": "N", "Decimal_Number": "Nd", "digit": "Nd",
	"Letter_Number": "Nl", "Other_Number": "No",
	"Punctuation": "P", "punct": "P", "Connector_Punctuation": "Pc",
	"Dash_Punctuation": "Pd", "Open_Punctuation": "Ps", "Close_Punctuation": "Pe",
	"Initial_Punctuation": "Pi", "Final_Punctuation": "Pf", "Other_Punctuation": "Po",
	"Symbol": "S", "Math_Symbol": "Sm", "Currency_Symbol": "Sc",
	"Modifier_Symbol": "Sk", "Other_Symbol": "So",
	"Separator": "Z", "Space_Separator": "Zs", "Line_Separator": "Zl",
	"Paragraph_Separator": "Zp",
	"Other":               "C", "Control": "Cc", "cntrl": "Cc", "Format": "Cf",
	"Surrogate": "Cs", "Private_Use": "Co",
}

// resolveProperty turns the text between the braces of \p{...} into the name
// Go's regexp accepts.
func resolveProperty(spec string) (string, error) {
	name, value, hasEq := strings.Cut(spec, "=")
	if !hasEq {
		return category(spec)
	}
	switch name {
	case "General_Category", "gc":
		return category(value)
	case "Script", "sc":
		if _, ok := unicode.Scripts[value]; ok {
			return value, nil
		}
		return "", fmt.Errorf("unsupported Unicode script %q", value)
	case "Script_Extensions", "scx":
		return "", fmt.Errorf("Script_Extensions is not supported (%q)", spec)
	}
	return "", fmt.Errorf("unsupported Unicode property %q", name)
}

func category(v string) (string, error) {
	if short, ok := longCategory[v]; ok {
		v = short
	}
	if _, ok := unicode.Categories[v]; ok {
		return v, nil
	}
	return "", fmt.Errorf("unsupported Unicode property %q", v)
}
