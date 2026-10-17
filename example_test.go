package jsonschema_test

import (
	"fmt"

	"github.com/neoface/jsonschema"
)

// ExampleCompile compiles a schema once and checks several documents with it.
func ExampleCompile() {
	schema, err := jsonschema.Compile([]byte(`{
		"type": "object",
		"properties": {"age": {"type": "integer", "minimum": 0}},
		"required": ["age"]
	}`))
	if err != nil {
		fmt.Println("schema error:", err)
		return
	}
	for _, doc := range []string{`{"age": 36}`, `{"age": -1}`} {
		res, err := schema.Validate([]byte(doc))
		if err != nil {
			fmt.Println("error:", err)
			continue
		}
		out, _ := res.Output("flag")
		fmt.Println(res.Valid(), string(out))
	}
	// Output:
	// true {"valid":true}
	// false {"valid":false}
}
