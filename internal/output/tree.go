package output

import (
	"encoding/json"

	"github.com/neoface/jsonschema/internal/keywords"
)

// nestedHead is the common start of every nested unit.
type nestedHead struct {
	Valid                   bool   `json:"valid"`
	KeywordLocation         string `json:"keywordLocation"`
	AbsoluteKeywordLocation string `json:"absoluteKeywordLocation,omitempty"`
	InstanceLocation        string `json:"instanceLocation"`
}

type nestedError struct {
	nestedHead
	Error string `json:"error"`
}

type nestedAnnotation struct {
	nestedHead
	Annotation any `json:"annotation"`
}

// hasInfo reports whether a node carries information for 'detailed': it
// failed, or it or a descendant holds an annotation.
func hasInfo(n *keywords.TreeNode) bool {
	if !n.Valid || len(n.Annotations) > 0 {
		return true
	}
	for _, c := range n.Children {
		if hasInfo(c) {
			return true
		}
	}
	return false
}

// kept returns the children rendered under n. Verbose keeps all; detailed
// drops passing children without annotations (core 12.4.2).
func kept(n *keywords.TreeNode, verbose bool) []*keywords.TreeNode {
	if verbose {
		return n.Children
	}
	var out []*keywords.TreeNode
	for _, c := range n.Children {
		if hasInfo(c) {
			out = append(out, c)
		}
	}
	return out
}

// renderTree writes the nested report for r.Tree unit by unit through w, so
// the byte cap applies while rendering. verbose keeps passing nodes; otherwise
// the detailed condensing applies. Instance values are never written: node
// fields are locations and messages, annotation values come from the schema
// (or are matched property names, treated like instanceLocation).
// A nil Tree renders one root unit with Report.Valid.
func renderTree(r Report, w *Budget, verbose bool) error {
	if r.Tree == nil {
		return put(w, nestedHead{Valid: r.Valid})
	}
	return writeNode(w, r.Tree, verbose, true)
}

func writeNode(w *Budget, n *keywords.TreeNode, verbose, root bool) error {
	kids := kept(n, verbose)
	leaves := len(n.Annotations)
	if !n.Valid {
		leaves = len(n.Errors)
	}
	// Detailed collapses a node that only wraps a single child into it.
	if !verbose && !root && leaves == 0 && len(kids) == 1 {
		return writeNode(w, kids[0], verbose, false)
	}
	head, err := json.Marshal(nestedHead{n.Valid, n.KeywordLocation, n.AbsoluteKeywordLocation, n.InstanceLocation})
	if err != nil {
		return err
	}
	if leaves == 0 && len(kids) == 0 {
		_, err = w.Write(head)
		return err
	}
	key := "annotations"
	if !n.Valid {
		key = "errors"
	}
	if _, err := w.Write(head[:len(head)-1]); err != nil {
		return err
	}
	if err := puts(w, `,"`+key+`":[`); err != nil {
		return err
	}
	first := true
	sep := func() error {
		if !first {
			return puts(w, ",")
		}
		first = false
		return nil
	}
	if n.Valid {
		for _, a := range n.Annotations {
			if err := sep(); err != nil {
				return err
			}
			if err := put(w, nestedAnnotation{nestedHead{true, a.KeywordLocation, a.AbsoluteKeywordLocation, a.InstanceLocation}, jsonValue(a.Value)}); err != nil {
				return err
			}
		}
	} else {
		for _, u := range n.Errors {
			if err := sep(); err != nil {
				return err
			}
			if err := put(w, nestedError{nestedHead{false, u.KeywordLocation, u.AbsoluteKeywordLocation, u.InstanceLocation}, u.Error}); err != nil {
				return err
			}
		}
	}
	for _, c := range kids {
		if err := sep(); err != nil {
			return err
		}
		if err := writeNode(w, c, verbose, false); err != nil {
			return err
		}
	}
	return puts(w, "]}")
}
