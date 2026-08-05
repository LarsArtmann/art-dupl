package detection

import "github.com/LarsArtmann/art-dupl/domain"

// Config configures the MultiDetector's behavior.
type Config struct {
	// Methods specifies which detection algorithms to run.
	// Accepted values: MethodArtDupl, MethodHash.
	// Empty means use the default (art-dupl only).
	Methods []domain.DetectionMethod

	// Verbose enables detailed logging during detection.
	Verbose bool

	// SearchWorkers controls suffix tree search parallelism.
	// 0 or 1 = sequential (default), >1 = parallel DFS with N workers.
	// Parallel search dispatches root-level subtrees to goroutines.
	SearchWorkers int
}

// Detection method name constants — aliases for domain.DetectionMethod
// to prevent drift across packages.
//
// art-dupl:accept architectural alias: detection and SDK independently re-export domain constants (arch-lint boundary)
const (
	MethodArtDupl = domain.MethodArtDupl
	MethodHash    = domain.MethodHash
)
