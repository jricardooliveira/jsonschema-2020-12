// Package formats is the one registration point for format checks. A format
// file calls Register from init(); the format keyword finds checks only
// through Lookup. A check is a plain, linear-time function with no error path.
package formats

// registry is written only by init() functions and read-only afterwards.
var registry = map[string]func(string) bool{}

// Register adds the check for a format name. It panics on an empty name, a nil
// check or a duplicate name.
func Register(name string, check func(string) bool) {
	if name == "" || check == nil {
		panic("formats: incomplete registration")
	}
	if _, dup := registry[name]; dup {
		panic("formats: duplicate registration of " + name)
	}
	registry[name] = check
}

// Lookup returns the check registered under name.
func Lookup(name string) (func(string) bool, bool) {
	c, ok := registry[name]
	return c, ok
}
