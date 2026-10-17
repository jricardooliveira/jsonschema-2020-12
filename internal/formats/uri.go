package formats

// RFC 3986 URI: a scheme is required. See rfc3986.go for what is not checked.
func init() { Register("uri", func(s string) bool { return validRFC3986(s, false, true) }) }
