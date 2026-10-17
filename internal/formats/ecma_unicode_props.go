package formats

import (
	"strings"
	"unicode"
)

// ECMA-262 (u flag) \p{...} / \P{...} property names, checked by name only.
// The table is built once at init and is read-only afterwards.
//
// ponytail: partly covered. Not covered, so these fail closed:
//   - script aliases missing from scriptAliases below (the long Script name
//     still works for every script in Go's unicode.Scripts);
//   - scripts newer than Go's Unicode tables (not in unicode.Scripts);
//   - the ECMA-262 properties of strings (RGI_Emoji etc.), which need the v flag.
//
// Binary properties are accepted by name only; whether Go's tables hold the
// property is not checked because only the spelling matters here.

var (
	ecmaGeneralCategory = map[string]bool{}
	ecmaScript          = map[string]bool{}
	ecmaBinaryProp      = map[string]bool{}
)

// gcAliases lists the General_Category spellings: short code, long name and
// further aliases (unicode.Categories lacks LC and Cn and all long names).
var gcAliases = []string{
	"L Letter", "LC Cased_Letter", "Lu Uppercase_Letter", "Ll Lowercase_Letter",
	"Lt Titlecase_Letter", "Lm Modifier_Letter", "Lo Other_Letter",
	"M Mark Combining_Mark", "Mn Nonspacing_Mark", "Mc Spacing_Mark", "Me Enclosing_Mark",
	"N Number", "Nd Decimal_Number digit", "Nl Letter_Number", "No Other_Number",
	"P Punctuation punct", "Pc Connector_Punctuation", "Pd Dash_Punctuation",
	"Ps Open_Punctuation", "Pe Close_Punctuation", "Pi Initial_Punctuation",
	"Pf Final_Punctuation", "Po Other_Punctuation",
	"S Symbol", "Sm Math_Symbol", "Sc Currency_Symbol", "Sk Modifier_Symbol", "So Other_Symbol",
	"Z Separator", "Zs Space_Separator", "Zl Line_Separator", "Zp Paragraph_Separator",
	"C Other", "Cc Control cntrl", "Cf Format", "Cs Surrogate", "Co Private_Use", "Cn Unassigned",
}

// scriptAliases maps ISO 15924 codes to Script long names (as in unicode.Scripts).
var scriptAliases = map[string]string{
	"Adlm": "Adlam", "Aghb": "Caucasian_Albanian", "Ahom": "Ahom", "Arab": "Arabic",
	"Armi": "Imperial_Aramaic", "Armn": "Armenian", "Avst": "Avestan", "Bali": "Balinese",
	"Bamu": "Bamum", "Bass": "Bassa_Vah", "Batk": "Batak", "Beng": "Bengali",
	"Bhks": "Bhaiksuki", "Bopo": "Bopomofo", "Brah": "Brahmi", "Brai": "Braille",
	"Bugi": "Buginese", "Buhd": "Buhid", "Cakm": "Chakma", "Cans": "Canadian_Aboriginal",
	"Cari": "Carian", "Cham": "Cham", "Cher": "Cherokee", "Chrs": "Chorasmian",
	"Copt": "Coptic", "Qaac": "Coptic", "Cprt": "Cypriot", "Cyrl": "Cyrillic",
	"Deva": "Devanagari", "Diak": "Dives_Akuru", "Dogr": "Dogra", "Dsrt": "Deseret",
	"Dupl": "Duployan", "Egyp": "Egyptian_Hieroglyphs", "Elba": "Elbasan", "Elym": "Elymaic",
	"Ethi": "Ethiopic", "Geor": "Georgian", "Glag": "Glagolitic", "Gong": "Gunjala_Gondi",
	"Gonm": "Masaram_Gondi", "Goth": "Gothic", "Gran": "Grantha", "Grek": "Greek",
	"Gujr": "Gujarati", "Guru": "Gurmukhi", "Hang": "Hangul", "Hani": "Han",
	"Hano": "Hanunoo", "Hatr": "Hatran", "Hebr": "Hebrew", "Hira": "Hiragana",
	"Hluw": "Anatolian_Hieroglyphs", "Hmng": "Pahawh_Hmong", "Hmnp": "Nyiakeng_Puachue_Hmong",
	"Hung": "Old_Hungarian", "Ital": "Old_Italic", "Java": "Javanese", "Kali": "Kayah_Li",
	"Kana": "Katakana", "Khar": "Kharoshthi", "Khmr": "Khmer", "Khoj": "Khojki",
	"Kits": "Khitan_Small_Script", "Knda": "Kannada", "Kthi": "Kaithi", "Lana": "Tai_Tham",
	"Laoo": "Lao", "Latn": "Latin", "Lepc": "Lepcha", "Limb": "Limbu", "Lina": "Linear_A",
	"Linb": "Linear_B", "Lisu": "Lisu", "Lyci": "Lycian", "Lydi": "Lydian",
	"Mahj": "Mahajani", "Maka": "Makasar", "Mand": "Mandaic", "Mani": "Manichaean",
	"Marc": "Marchen", "Medf": "Medefaidrin", "Mend": "Mende_Kikakui", "Merc": "Meroitic_Cursive",
	"Mero": "Meroitic_Hieroglyphs", "Mlym": "Malayalam", "Modi": "Modi", "Mong": "Mongolian",
	"Mroo": "Mro", "Mtei": "Meetei_Mayek", "Mult": "Multani", "Mymr": "Myanmar",
	"Nand": "Nandinagari", "Narb": "Old_North_Arabian", "Nbat": "Nabataean", "Newa": "Newa",
	"Nkoo": "Nko", "Nshu": "Nushu", "Ogam": "Ogham", "Olck": "Ol_Chiki",
	"Orkh": "Old_Turkic", "Orya": "Oriya", "Osge": "Osage", "Osma": "Osmanya",
	"Palm": "Palmyrene", "Pauc": "Pau_Cin_Hau", "Perm": "Old_Permic", "Phag": "Phags_Pa",
	"Phli": "Inscriptional_Pahlavi", "Phlp": "Psalter_Pahlavi", "Phnx": "Phoenician",
	"Plrd": "Miao", "Prti": "Inscriptional_Parthian", "Rjng": "Rejang", "Rohg": "Hanifi_Rohingya",
	"Runr": "Runic", "Samr": "Samaritan", "Sarb": "Old_South_Arabian", "Saur": "Saurashtra",
	"Sgnw": "SignWriting", "Shaw": "Shavian", "Shrd": "Sharada", "Sidd": "Siddham",
	"Sind": "Khudawadi", "Sinh": "Sinhala", "Sogd": "Sogdian", "Sogo": "Old_Sogdian",
	"Sora": "Sora_Sompeng", "Soyo": "Soyombo", "Sund": "Sundanese", "Sylo": "Syloti_Nagri",
	"Syrc": "Syriac", "Tagb": "Tagbanwa", "Takr": "Takri", "Tale": "Tai_Le",
	"Talu": "New_Tai_Lue", "Taml": "Tamil", "Tang": "Tangut", "Tavt": "Tai_Viet",
	"Telu": "Telugu", "Tfng": "Tifinagh", "Tglg": "Tagalog", "Thaa": "Thaana",
	"Thai": "Thai", "Tibt": "Tibetan", "Tirh": "Tirhuta", "Ugar": "Ugaritic",
	"Vaii": "Vai", "Wara": "Warang_Citi", "Wcho": "Wancho", "Xpeo": "Old_Persian",
	"Xsux": "Cuneiform", "Yezi": "Yezidi", "Yiii": "Yi", "Zanb": "Zanabazar_Square",
	"Zinh": "Inherited", "Qaai": "Inherited", "Zyyy": "Common", "Zzzz": "Unknown",
}

// binaryProps lists the ECMA-262 binary property spellings: long name, then
// short alias if any.
var binaryProps = []string{
	"ASCII", "ASCII_Hex_Digit AHex", "Alphabetic Alpha", "Any", "Assigned",
	"Bidi_Control Bidi_C", "Bidi_Mirrored Bidi_M", "Case_Ignorable CI", "Cased",
	"Changes_When_Casefolded CWCF", "Changes_When_Casemapped CWCM",
	"Changes_When_Lowercased CWL", "Changes_When_NFKC_Casefolded CWKCF",
	"Changes_When_Titlecased CWT", "Changes_When_Uppercased CWU", "Dash",
	"Default_Ignorable_Code_Point DI", "Deprecated Dep", "Diacritic Dia", "Emoji",
	"Emoji_Component EComp", "Emoji_Modifier EMod", "Emoji_Modifier_Base EBase",
	"Emoji_Presentation EPres", "Extended_Pictographic ExtPict", "Extender Ext",
	"Grapheme_Base Gr_Base", "Grapheme_Extend Gr_Ext", "Hex_Digit Hex",
	"IDS_Binary_Operator IDSB", "IDS_Trinary_Operator IDST", "ID_Continue IDC",
	"ID_Start IDS", "Ideographic Ideo", "Join_Control Join_C",
	"Logical_Order_Exception LOE", "Lowercase Lower", "Math",
	"Noncharacter_Code_Point NChar", "Pattern_Syntax Pat_Syn",
	"Pattern_White_Space Pat_WS", "Quotation_Mark QMark", "Radical",
	"Regional_Indicator RI", "Sentence_Terminal STerm", "Soft_Dotted SD",
	"Terminal_Punctuation Term", "Unified_Ideograph UIdeo", "Uppercase Upper",
	"Variation_Selector VS", "White_Space space", "XID_Continue XIDC", "XID_Start XIDS",
}

func init() {
	for k := range unicode.Categories {
		ecmaGeneralCategory[k] = true
	}
	for _, line := range gcAliases {
		for _, n := range strings.Fields(line) {
			ecmaGeneralCategory[n] = true
		}
	}
	for k := range unicode.Scripts {
		ecmaScript[k] = true
	}
	for short, long := range scriptAliases {
		if ecmaScript[long] { // only aliases of scripts Go's tables know
			ecmaScript[short] = true
		}
	}
	for _, line := range binaryProps {
		for _, n := range strings.Fields(line) {
			ecmaBinaryProp[n] = true
		}
	}
}

// validECMAUnicodeProperty reports whether expr, the text between the braces
// of \p{...} or \P{...}, is an ECMA-262 property: a General_Category value
// (alone or as gc=/General_Category=), a Script=/sc=/Script_Extensions=/scx=
// value, or a binary property name. Matching is case-sensitive, as in ECMA-262.
func validECMAUnicodeProperty(expr string) bool {
	name, value, hasValue := strings.Cut(expr, "=")
	if !hasValue {
		return ecmaGeneralCategory[expr] || ecmaBinaryProp[expr]
	}
	switch name {
	case "General_Category", "gc":
		return ecmaGeneralCategory[value]
	case "Script", "sc", "Script_Extensions", "scx":
		return ecmaScript[value]
	}
	return false
}
