package formats

// RFC 3987 IRI: a scheme is required. See rfc3986.go for what is not checked.
func init() { Register("iri", func(s string) bool { return validRFC3986(s, true, true) }) }
