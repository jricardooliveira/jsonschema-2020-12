package formats

// RFC 3987 IRI-reference. See rfc3986.go for what is not checked.
func init() {
	Register("iri-reference", func(s string) bool { return validRFC3986(s, true, false) })
}
