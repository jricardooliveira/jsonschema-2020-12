package eval

// RecordProperties records the member names a keyword applied a subschema to
// for the object at instLoc. kwLoc is the keyword location. The slice is
// copied. It shares the item log, so a failed branch's Rollback drops it too.
func (r *EvaluatedItems) RecordProperties(instLoc, kwLoc string, names []string) {
	r.entries = append(r.entries, itemEntry{
		key:   itemKey{instLoc, kwLoc},
		props: true,
		names: append([]string(nil), names...),
	})
}

// PropertiesAt returns the recorded names for the object at instLoc, keyed by
// keyword location. Slices are copies.
func (r *EvaluatedItems) PropertiesAt(instLoc string) map[string][]string {
	out := map[string][]string{}
	for _, e := range r.entries {
		if e.props && e.key.instance == instLoc {
			out[e.key.keyword] = append([]string(nil), e.names...)
		}
	}
	return out
}
