package jsonschema

// FormatAssertionRequested reports whether any schema resource's meta-schema
// lists the format-assertion vocabulary as required. It is fixed at Compile;
// 'format' itself stays annotation-only.
func (s *Schema) FormatAssertionRequested() bool { return s.root.FormatAssertion }
