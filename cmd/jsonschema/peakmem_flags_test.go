package main

import "flag"

// Flags of the CLI peak-memory test. Defined on every platform so the test
// binary accepts them everywhere; only the unix build measures anything.
var (
	peakMem      = flag.Bool("peakmem", false, "measure CLI peak memory (TestCLIPeakMemory)")
	peakMemChild = flag.String("peakmem-child", "", "internal: schema|doc|exit for one child run")
)
