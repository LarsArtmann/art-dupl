package cmd

import (
	"os"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/internal/accept"
)

// The //art-dupl:accept directive machinery is canonical in internal/accept
// (extracted 2026-10-05, same pattern as internal/gitignore) so the toolsdk
// provider can share the exact suppression semantics; these aliases keep the
// cmd-level API — tests and run_output call sites — unchanged.
type (
	// AcceptedSet tracks //art-dupl:accept directives across source files.
	AcceptedSet = accept.AcceptedSet
	// AcceptedDirective is a single //art-dupl:accept comment in source code.
	AcceptedDirective = accept.AcceptedDirective
	// DeadDirective is a hash-precision directive that matched zero groups.
	DeadDirective = accept.DeadDirective
)

// NewAcceptedSet creates an AcceptedSet that uses the given file reader.
var NewAcceptedSet = accept.NewAcceptedSet

// newAcceptSet returns an AcceptedSet from config, or nil when disabled.
// Returns nil so that shouldSuppressGroup skips the accept-directive check
// entirely (nil-safe via AcceptedSet.IsAccepted).
func newAcceptSet(cfg *config.Config) *AcceptedSet {
	if cfg.NoAcceptDirectives {
		return nil
	}

	return NewAcceptedSet(os.ReadFile)
}
