package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// realFlags collects the flag names run defines: the literal names passed to
// flags.String/Bool/Var/Func in run.go, plus the limit flags from
// limitFlagNames. The flag set is local to run, so run.go is read as source.
func realFlags(t *testing.T) map[string]bool {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "run.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, n := range limitFlagNames {
		names[n] = true
	}
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if id, ok := sel.X.(*ast.Ident); !ok || id.Name != "flags" {
			return true
		}
		idx := map[string]int{"String": 0, "Bool": 0, "Var": 1, "Func": 0}
		i, ok := idx[sel.Sel.Name]
		if !ok || len(call.Args) <= i {
			return true
		}
		if lit, ok := call.Args[i].(*ast.BasicLit); ok && lit.Kind == token.STRING {
			s, _ := strconv.Unquote(lit.Value)
			names[s] = true
		}
		return true
	})
	return names
}

// docFlags reads the first column of the flag table in docs/cli.md.
func docFlags(t *testing.T) map[string]bool {
	t.Helper()
	data, err := os.ReadFile("../../docs/cli.md")
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile("^\\| `-{1,2}([A-Za-z0-9-]+)` \\|")
	names := map[string]bool{}
	for _, line := range strings.Split(string(data), "\n") {
		if m := re.FindStringSubmatch(line); m != nil {
			names[m[1]] = true
		}
	}
	return names
}

func keys(m map[string]bool) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func TestDocsFlagListMatchesTool(t *testing.T) {
	real, docs := realFlags(t), docFlags(t)
	if len(real) < 18 || len(docs) == 0 {
		t.Fatalf("flag collection failed: real=%v docs=%v", keys(real), keys(docs))
	}
	for _, n := range keys(real) {
		if !docs[n] {
			t.Errorf("flag -%s is defined by the tool but missing from docs/cli.md", n)
		}
	}
	for _, n := range keys(docs) {
		if !real[n] {
			t.Errorf("docs/cli.md lists -%s but the tool does not define it", n)
		}
	}
}

// Every collected flag must be accepted by the real flag set, so the list
// above cannot drift from what run parses.
func TestRealFlagsAreAccepted(t *testing.T) {
	for _, n := range keys(realFlags(t)) {
		var errb bytes.Buffer
		run([]string{"validate", "-" + n}, strings.NewReader(""), &bytes.Buffer{}, &errb)
		if strings.Contains(errb.String(), "unknown flag") {
			t.Errorf("flag -%s is not defined by run", n)
		}
	}
	var errb bytes.Buffer
	run([]string{"validate", "-no-such-flag"}, strings.NewReader(""), &bytes.Buffer{}, &errb)
	if !strings.Contains(errb.String(), "unknown flag") {
		t.Fatal("probe cannot detect an undefined flag")
	}
}
