# Library

Package `github.com/jricardooliveira/jsonschema-2020-12`. Load a schema once, then check any
number of documents with it. A compiled `*Schema` is immutable and safe for
concurrent use by several goroutines.

## API

- `Compile(schema []byte) (*Schema, error)`: compile with default options.
- `CompileWithOptions(schema []byte, opts Options) (*Schema, error)`: limits
  (`MaxDepth`, `MaxSteps`, ...), `BaseURI` and `FormatAssertion` are fixed here.
- `CompileWithRegistry(schema []byte, opts Options, reg Registry) (*Schema, error)`:
  also supplies the registry files and folders (`Registry.Files`,
  `Registry.Dirs`) that `$ref` may resolve to. Nothing else is ever read.
- `(*Schema).Validate(document []byte) (Result, error)`: an immutable `Result`.
  An error (broken JSON, exceeded limit) is never a valid or invalid result.
- `Result.Valid()`, `Result.Annotations()` and `Result.Output(shape)` where
  shape is `flag`, `basic`, `detailed` or `verbose`.

A bad schema, an unresolved reference or an exceeded limit makes `Compile`
fail. A limit error is a `*jsonschema.LimitError`; see [limits.md](limits.md).
Reports never contain instance values.

## Example

The same code runs as `ExampleCompile` in `example_test.go`, and `go test`
checks its output.

```go
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
```
