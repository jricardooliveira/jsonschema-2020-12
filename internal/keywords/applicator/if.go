package applicator

import (
	"fmt"
	"strings"

	"github.com/neoface/jsonschema/internal/keywords"
)

func init() {
	keywords.Register(keywords.Keyword{Name: "if", Vocabulary: VocabularyURI, Compile: compileIf, Evaluate: evalIf})
}

// branch is a compiled sibling 'then' or 'else' and its keyword location.
type branch struct {
	name, loc string
	node      any
}

// conditional is the compiled 'if': its subschema plus the sibling branches
// (nil when absent). Each subschema is compiled exactly once, here.
type conditional struct {
	cond      any
	then, els *branch
}

func compileIf(ctx *keywords.CompileContext, value any) (any, error) {
	n, err := ctx.Sub(value, "")
	if err != nil {
		return nil, fmt.Errorf("if: %w", err)
	}
	c := &conditional{cond: n}
	parent := strings.TrimSuffix(ctx.Location, "/if")
	for _, name := range []string{"then", "else"} {
		v, ok := ctx.Siblings[name]
		if !ok {
			continue
		}
		b := &branch{name: name, loc: parent + "/" + name}
		if b.node, err = ctx.CompileAt(v, b.loc); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		if name == "then" {
			c.then = b
		} else {
			c.els = b
		}
	}
	return c, nil
}

// evalIf (core 10.2.2) evaluates 'if' once, dropping its units (still
// counted), then applies 'then' when it passed or 'else' when it failed. A
// failing branch keeps its units, then adds one unit at the branch location.
func evalIf(ctx *keywords.EvalContext, compiled any, instance any) (bool, error) {
	c := compiled.(*conditional)
	ok, _, err := ctx.EvalChild(c.cond, instance, ctx.InstanceLocation)
	if err != nil {
		return false, err
	}
	b := c.els
	if ok {
		b = c.then
	}
	if b == nil {
		return true, nil
	}
	ok, units, err := ctx.EvalChild(b.node, instance, ctx.InstanceLocation)
	if err != nil || ok {
		return ok, err
	}
	ctx.Keep(units)
	ifLoc := ctx.KeywordLocation
	ctx.KeywordLocation = b.loc
	err = ctx.Report(b.name + ": subschema did not match")
	ctx.KeywordLocation = ifLoc
	return false, err
}

// compileBranch is the compile step of 'then' and 'else'. With a sibling
// 'if', that keyword compiles the branch, so only the shape is checked here;
// alone, the branch is compiled once so a malformed schema is still an error.
// Either way the branch evaluates nothing itself.
func compileBranch(name string, ctx *keywords.CompileContext, value any) (any, error) {
	if _, ok := ctx.Siblings["if"]; !ok {
		if _, err := ctx.Sub(value, ""); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		return nil, nil
	}
	switch value.(type) {
	case bool, map[string]any:
		return nil, nil
	}
	return nil, fmt.Errorf("%s: must be a schema", name)
}

// evalBranch: 'then' and 'else' have no evaluation of their own ('if' runs them).
func evalBranch(*keywords.EvalContext, any, any) (bool, error) { return true, nil }
