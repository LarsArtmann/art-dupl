package syntax

import (
	"testing"
	"time"
)

// TestPerfRegressionSerialize guards against gross performance regressions
// in the core Serialize path. Thresholds are intentionally generous (10x the
// observed baseline on a 2024 laptop) to avoid CI flakiness while still
// catching catastrophic regressions.
func TestPerfRegressionSerialize(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping perf regression test in short mode")
	}

	root := genDeepTree(100)

	const threshold = 50 * time.Millisecond

	start := time.Now()
	_ = Serialize(root)
	elapsed := time.Since(start)

	if elapsed > threshold {
		t.Errorf("Serialize(100-depth tree) took %v (threshold %v)", elapsed, threshold)
	}
}

// TestPerfRegressionHashSeq guards against regressions in the grouping-hash
// path used by the suffix tree.
func TestPerfRegressionHashSeq(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping perf regression test in short mode")
	}

	nodes := make([]*Node, 0, 10000)
	for i := range 10000 {
		nodes = append(nodes, &Node{Type: int32(i % 128), Name: "var"})
	}

	const threshold = 50 * time.Millisecond

	start := time.Now()
	_ = hashSeq(nodes)
	elapsed := time.Since(start)

	if elapsed > threshold {
		t.Errorf("hashSeq(10k nodes) took %v (threshold %v)", elapsed, threshold)
	}
}
