package keywords

// Recording reports whether the evaluated-property/item record of this call
// is kept. When it is false, keywords may skip building the keyword location
// for RecordProperties. Fail-safe: any Items that does not say otherwise (or
// none) counts as recording.
func (c *EvalContext) Recording() bool {
	if r, ok := c.Items.(interface{ Recording() bool }); ok {
		return r.Recording()
	}
	return true
}
