package detection

// Config configures the MultiDetector's behavior.
type Config struct {
	// Methods specifies which detection algorithms to run.
	// Accepted values: MethodArtDupl, MethodHash.
	// Empty means use the default (art-dupl only).
	Methods []string

	// Verbose enables detailed logging during detection.
	Verbose bool
}

// Detection method names for Config.Methods.
const (
	MethodArtDupl = "art-dupl"
	MethodHash    = "hash"
)
