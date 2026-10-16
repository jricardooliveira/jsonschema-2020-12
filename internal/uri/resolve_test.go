package uri

import "testing"

func TestResolve(t *testing.T) {
	base := "http://a/b/c/d;p?q"
	for ref, want := range map[string]string{ // RFC 3986 5.4.1/5.4.2
		"g":           "http://a/b/c/g",
		"./g":         "http://a/b/c/g",
		"g/":          "http://a/b/c/g/",
		"/g":          "http://a/g",
		"//g":         "http://g",
		"?y":          "http://a/b/c/d;p?y",
		"#s":          "http://a/b/c/d;p?q#s",
		"":            "http://a/b/c/d;p?q",
		"../g":        "http://a/b/g",
		"../../g":     "http://a/g",
		"../../../g":  "http://a/g",
		"g;x?y#s":     "http://a/b/c/g;x?y#s",
		"/./g":        "http://a/g",
		"g.":          "http://a/b/c/g.",
		"g/../h":      "http://a/b/c/h",
		"https://x/y": "https://x/y",
	} {
		got, err := Resolve(base, ref)
		if err != nil || got != want {
			t.Errorf("Resolve(%q)=%q,%v want %q", ref, got, err, want)
		}
	}
	if _, err := Resolve(base, "a%zz"); err == nil {
		t.Error("bad percent-encoding accepted")
	}
}

func TestNormalize(t *testing.T) {
	same := [][2]string{
		{"https://Example.com/a%7Eb.json", "https://example.com/a~b.json"},
		{"HTTP://x/a%2fb", "http://x/a%2Fb"},
		{"http://x/a/./b/../c", "http://x/a/c"},
	}
	for _, c := range same {
		a, _ := Normalize(c[0])
		b, _ := Normalize(c[1])
		if a != b {
			t.Errorf("%q (%q) != %q (%q)", c[0], a, c[1], b)
		}
	}
	diff := [][2]string{
		{"http://x/A", "http://x/a"},
		{"http://x/a%2Fb", "http://x/a/b"},
		{"http://x/a", "http://x/b"},
	}
	for _, c := range diff {
		a, _ := Normalize(c[0])
		b, _ := Normalize(c[1])
		if a == b {
			t.Errorf("%q and %q must differ", c[0], c[1])
		}
	}
	if _, err := Normalize("http://x/%zz"); err == nil {
		t.Error("bad percent-encoding accepted")
	}
}
