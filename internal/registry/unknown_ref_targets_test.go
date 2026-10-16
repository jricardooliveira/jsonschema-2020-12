package registry

import "testing"

func TestIDUnderUnknownKeywordRefTargetStartsResource(t *testing.T) {
	doc := mustDoc(t, `{
		"$id":"https://e.example/root.json",
		"$ref":"#/definitions/x",
		"definitions":{"x":{"$id":"https://e.example/sub/x.json",
			"properties":{"a":{"$ref":"y.json"}},
			"definitions":{"y":{"$id":"y.json"}}}}}`)
	idx, err := Build(doc, "https://unused.example/")
	if err != nil {
		t.Fatal(err)
	}
	if idx.Resources["https://e.example/sub/x.json"] == nil {
		t.Fatal("$id inside the $ref target must start a resource")
	}
	if got := idx.Bases["/definitions/x/properties/a"]; got != "https://e.example/sub/x.json" {
		t.Fatalf("base of relative $ref inside target = %q", got)
	}
}

func TestUnknownKeywordNotReferencedIsIgnored(t *testing.T) {
	doc := mustDoc(t, `{"$id":"https://e.example/r.json","definitions":{"x":{"$id":"https://e.example/x.json"}}}`)
	idx, err := Build(doc, "https://unused.example/")
	if err != nil {
		t.Fatal(err)
	}
	if idx.Resources["https://e.example/x.json"] != nil {
		t.Fatal("unreferenced unknown keyword must not be walked")
	}
}

func TestCyclicRefsUnderUnknownKeywordTerminate(t *testing.T) {
	doc := mustDoc(t, `{"$ref":"#/definitions/a","definitions":{"a":{"$ref":"#/definitions/b"},"b":{"$ref":"#/definitions/a"}}}`)
	if _, err := Build(doc, "https://e.example/c.json"); err != nil {
		t.Fatal(err)
	}
}
