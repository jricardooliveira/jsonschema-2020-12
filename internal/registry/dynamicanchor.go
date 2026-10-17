package registry

// HasDynamicAnchor reports whether the resource holds a subschema carrying
// "$dynamicAnchor": name (2020-12 core 8.2.2). Plain-name anchors share one
// namespace, so the answer comes from the same Anchors table as $anchor.
func (r *Resource) HasDynamicAnchor(name string) bool {
	ptr, ok := r.Anchors[name]
	if !ok {
		return false
	}
	v, ok := navigate(r.Doc, ptr)
	if !ok {
		return false
	}
	m, ok := v.(map[string]any)
	if !ok {
		return false
	}
	got, ok := m["$dynamicAnchor"].(string)
	return ok && got == name
}
