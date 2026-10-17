//go:build !race

package jsonschema

// raceEnabled is true when the test binary is built with -race.
const raceEnabled = false
