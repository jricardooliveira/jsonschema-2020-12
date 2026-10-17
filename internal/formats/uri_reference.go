package formats

// RFC 3986 URI-reference. See rfc3986.go for what is not checked.
func init() {
	Register("uri-reference", func(s string) bool { return validRFC3986(s, false, false) })
}
