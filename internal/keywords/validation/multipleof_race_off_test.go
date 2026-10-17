//go:build !race

package validation

// raceEnabled is true when the test binary is built with -race.
const raceEnabled = false
