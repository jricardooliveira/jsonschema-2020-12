package eval

import (
	"strings"

	"github.com/neoface/jsonschema/internal/compile"
)

// nodeLoc is a cheap reference to one schema node on the evaluation path. It
// holds only immutable data (the parent keyword reference, the node's static
// location and its resource), never per-call state, so rendering later gives
// the same string as rendering at once. Strings are built only on demand.
type nodeLoc struct {
	parent *kwLoc // keyword of the parent node that led here; nil at the root
	static string // JSON Pointer of the node from the schema root
	res    *compile.Resource
	viaRef bool // the path to this node crossed $ref or $dynamicRef
}

// kwLoc is a cheap reference to one keyword being evaluated in a node.
type kwLoc struct {
	node   *nodeLoc
	name   string // keyword name
	static string // JSON Pointer of the keyword from the schema root
}

// isRef reports whether the keyword is a reference that jumps to another node.
func (k *kwLoc) isRef() bool { return k.name == "$ref" || k.name == "$dynamicRef" }

// child returns the reference for the node a keyword evaluates into.
func (k *kwLoc) child(static string, res *compile.Resource) *nodeLoc {
	return &nodeLoc{parent: k, static: static, res: res, viaRef: k.node.viaRef || k.isRef()}
}

// Dynamic renders the node's location along the evaluation path (core
// 12.3.1): a reference target starts at the $ref keyword itself, any other
// child keeps its path below the parent node.
func (n *nodeLoc) Dynamic() string {
	p := n.parent
	if p == nil {
		return n.static
	}
	if !p.isRef() {
		if rel, ok := strings.CutPrefix(n.static, p.node.static); ok && (rel == "" || rel[0] == '/') {
			return p.node.Dynamic() + rel
		}
	}
	return p.String()
}

// String renders the keyword location along the evaluation path.
func (k *kwLoc) String() string {
	return k.node.Dynamic() + strings.TrimPrefix(k.static, k.node.static)
}

// Absolute renders the absolute keyword location (core 12.3.2); it is empty
// until a reference was crossed.
func (k *kwLoc) Absolute() string {
	if !k.node.viaRef {
		return ""
	}
	return k.node.res.AbsoluteLocation(k.static)
}
