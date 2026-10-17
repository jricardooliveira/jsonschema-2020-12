//go:build !unix

package main

import "testing"

// TestCLIPeakMemory is skipped where getrusage is not available.
func TestCLIPeakMemory(t *testing.T) {
	t.Skip("CLI peak-memory measurement needs getrusage (unix only)")
}
