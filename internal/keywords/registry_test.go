package keywords

import "testing"

func testKeyword(name string) Keyword {
	return Keyword{
		Name:       name,
		Vocabulary: "urn:test",
		Compile:    func(*CompileContext, any) (any, error) { return 1, nil },
		Evaluate:   func(*EvalContext, any, any) (bool, error) { return true, nil },
	}
}

func TestRegisterAndLookup(t *testing.T) {
	const name = "x-registry-test-lookup"
	Register(testKeyword(name))
	t.Cleanup(func() { delete(registry, name) })
	k, ok := Lookup(name)
	if !ok || k.Name != name || k.Vocabulary != "urn:test" {
		t.Fatalf("lookup failed: %v %v", k, ok)
	}
	if _, ok := Lookup(name + "-missing"); ok {
		t.Fatal("unknown name found")
	}
}

func TestDuplicateRegistrationPanics(t *testing.T) {
	const name = "x-registry-test-dup"
	Register(testKeyword(name))
	t.Cleanup(func() { delete(registry, name) })
	defer func() {
		if recover() == nil {
			t.Fatal("want panic on duplicate")
		}
	}()
	Register(testKeyword(name))
}

func TestIncompleteRegistrationPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("want panic on incomplete keyword")
		}
	}()
	Register(Keyword{Name: "x-registry-test-incomplete"})
}
