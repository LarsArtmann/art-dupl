package cmd

import (
	"bufio"
	"bytes"
	"os"
	"strings"
	"sync"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
)

const acceptDirectivePrefix = "//art-dupl:accept"

// acceptDirectiveScanAbove is the number of lines above clone.LineStart to
// also check for directives. Users naturally place directives on the line
// above the code they want to accept, matching every other linter convention
// (golangci-lint, eslint, revive). The window is capped to avoid accepting
// unrelated groups that happen to be nearby.
const acceptDirectiveScanAbove = 5

const (
	scannerInitBufSize = 65536   // 64 KB
	scannerMaxBufSize  = 1048576 // 1 MB
)

// AcceptedDirective represents a single //art-dupl:accept comment in source code.
type AcceptedDirective struct {
	Line int    // 1-indexed source line number where the directive appears
	Hash string // optional group hash for precision matching (empty = match any group)
}

// AcceptedSet tracks //art-dupl:accept directives across source files.
// It lazily scans files on first access and caches the result so each file is
// read at most once per run.
type AcceptedSet struct {
	mu       sync.RWMutex
	scanned  map[string][]AcceptedDirective // filename → directives found
	readFile func(string) ([]byte, error)
}

// NewAcceptedSet creates an AcceptedSet that uses the given file reader.
func NewAcceptedSet(readFile func(string) ([]byte, error)) *AcceptedSet {
	return &AcceptedSet{
		scanned:  make(map[string][]AcceptedDirective),
		readFile: readFile,
	}
}

// newAcceptSet returns an AcceptedSet from config, or nil when disabled.
// Returns nil so that shouldSuppressGroup skips the accept-directive check
// entirely (nil-safe via AcceptedSet.IsAccepted).
func newAcceptSet(cfg *config.Config) *AcceptedSet {
	if cfg.NoAcceptDirectives {
		return nil
	}

	return NewAcceptedSet(os.ReadFile)
}

// IsAccepted checks if any clone in the group has an accept directive
// within its line range or up to acceptDirectiveScanAbove lines above
// LineStart. A directive on line N accepts clones where
// (LineStart - acceptDirectiveScanAbove) <= N <= LineEnd. This matches user
// expectations: directives placed on the line above the clone (like every
// other linter) are honored. When the directive includes a hash
// (//art-dupl:accept <hash>), it only matches groups with that exact hash.
func (a *AcceptedSet) IsAccepted(group domain.ProcessedCloneGroup) bool {
	if a == nil {
		return false
	}

	for _, clone := range group.Clones {
		directives := a.getDirectives(clone.Filename)

		scanFrom := max(1, clone.LineStart-acceptDirectiveScanAbove)

		for _, d := range directives {
			if d.Line < scanFrom || d.Line > clone.LineEnd {
				continue
			}

			if d.Hash == "" || d.Hash == group.Hash {
				return true
			}
		}
	}

	return false
}

// getDirectives returns the accept directives for the given file, scanning it
// on first access. Thread-safe via double-checked locking.
func (a *AcceptedSet) getDirectives(filename string) []AcceptedDirective {
	a.mu.RLock()
	directives, ok := a.scanned[filename]
	a.mu.RUnlock()

	if ok {
		return directives
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if directives, ok := a.scanned[filename]; ok {
		return directives
	}

	directives = a.scanFile(filename)
	a.scanned[filename] = directives

	return directives
}

// scanFile reads a file and extracts all //art-dupl:accept directives.
func (a *AcceptedSet) scanFile(filename string) []AcceptedDirective {
	data, err := a.readFile(filename)
	if err != nil {
		return nil
	}

	var directives []AcceptedDirective

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, scannerInitBufSize), scannerMaxBufSize)

	lineNum := 0

	for scanner.Scan() {
		lineNum++

		text := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(text, acceptDirectivePrefix) {
			continue
		}

		d := AcceptedDirective{Line: lineNum}

		rest := strings.TrimPrefix(text, acceptDirectivePrefix)
		rest = strings.TrimSpace(rest)

		// Only treat the text after the prefix as a hash if it is a single
		// token (no spaces). This distinguishes precision-hash directives
		// (//art-dupl:accept a1b2c3d4) from human-readable descriptions
		// (//art-dupl:accept idiomatic test helper boilerplate). Without this,
		// any text after the prefix is stored as the hash and never matches a
		// real group hash, silently disabling the directive.
		if rest != "" && !strings.ContainsAny(rest, " \t") {
			d.Hash = rest
		}

		directives = append(directives, d)
	}

	return directives
}
