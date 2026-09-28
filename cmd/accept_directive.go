package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
)

// acceptDirectiveRe locates the //art-dupl:accept marker anywhere on a line,
// tolerating optional whitespace between "//" and "art-dupl". This recognizes
// the compact form (//art-dupl:accept), the gofmt-canonical form
// (// art-dupl:accept — gofmt enforces a space after "//" for comment lines),
// and trailing inline directives (code(); //art-dupl:accept). Without the
// whitespace tolerance, every standalone directive written in idiomatic Go
// style would be silently ignored.
var acceptDirectiveRe = regexp.MustCompile(`//\s*art-dupl:accept`)

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
	matched  map[directiveKey]bool          // directives that accepted ≥1 group this run
	readFile func(string) ([]byte, error)
}

// directiveKey identifies a directive by its file and line.
type directiveKey struct {
	filename string
	line     int
}

// DeadDirective is a hash-precision accept directive that matched zero clone
// groups during the run — almost always a stale hash left behind after the
// accepted code was edited (the group's content hash changed) or removed.
type DeadDirective struct {
	Filename string
	Line     int
	Hash     string
}

// NewAcceptedSet creates an AcceptedSet that uses the given file reader.
func NewAcceptedSet(readFile func(string) ([]byte, error)) *AcceptedSet {
	return &AcceptedSet{
		scanned:  make(map[string][]AcceptedDirective),
		matched:  make(map[directiveKey]bool),
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
				a.recordMatched(clone.Filename, d.Line)

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

		loc := acceptDirectiveRe.FindStringIndex(text)
		if loc == nil {
			continue
		}

		d := AcceptedDirective{Line: lineNum}

		rest := strings.TrimSpace(text[loc[1]:])

		// Grammar (suppression behavior is identical for every input — only
		// dead-directive reporting visibility changes at the edges):
		//   //art-dupl:accept                    → bare accept (hash-less)
		//   //art-dupl:accept <hex>              → precision hash (group
		//       hashes are XXH3 16-char hex, so a non-hex token could never
		//       match a group)
		//   //art-dupl:accept <multi-word text>  → description, bare accept
		//   //art-dupl:accept <single non-hex>   → NOT a directive. Prose
		//       merely mentioning the syntax ("...accepts it with
		//       //art-dupl:accept.") lands here; under the old single-token-
		//       is-a-hash rule it became a precision directive that could
		//       never fire — harmless before the dead-directive detector,
		//       nonsense-warning material after.
		switch {
		case rest == "":
			directives = append(directives, d)
		case isHexToken(rest):
			d.Hash = rest
			directives = append(directives, d)
		case strings.ContainsAny(rest, " \t"):
			directives = append(directives, d)
		}
	}

	return directives
}

// isHexToken reports whether s consists solely of hex digits — the only
// token shape that can ever equal a group content hash (XXH3, 16-char hex).
func isHexToken(s string) bool {
	for _, r := range s {
		isHex := (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
		if !isHex {
			return false
		}
	}

	return s != ""
}

// recordMatched marks the directive at (filename, line) as live: it accepted
// at least one group this run.
func (a *AcceptedSet) recordMatched(filename string, line int) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.matched == nil {
		a.matched = make(map[directiveKey]bool)
	}

	a.matched[directiveKey{filename: filename, line: line}] = true
}

// DeadDirectives returns the hash-precision directives seen this run that
// matched zero groups. Hash-less directives (//art-dupl:accept without a
// token) never go stale in this sense — they accept whatever sits nearby —
// so they are never reported. Only directives in files that the run actually
// consulted (files with at least one evaluated clone group) are visible here;
// the set scans lazily per group's clone files.
func (a *AcceptedSet) DeadDirectives() []DeadDirective {
	if a == nil {
		return nil
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	dead := make([]DeadDirective, 0)

	for filename, directives := range a.scanned {
		for _, d := range directives {
			if d.Hash == "" {
				continue
			}

			if a.matched[directiveKey{filename: filename, line: d.Line}] {
				continue
			}

			dead = append(dead, DeadDirective{Filename: filename, Line: d.Line, Hash: d.Hash})
		}
	}

	sort.Slice(dead, func(i, j int) bool {
		if dead[i].Filename != dead[j].Filename {
			return dead[i].Filename < dead[j].Filename
		}

		return dead[i].Line < dead[j].Line
	})

	return dead
}

// WarnDeadDirectives prints one stderr warning per dead directive. Diagnostics,
// not progress: it prints even under --quiet (same policy as the unmatched
// include/exclude pattern warnings). Nil-safe on receiver and writer.
func (a *AcceptedSet) WarnDeadDirectives(stderr io.Writer) {
	dead := a.DeadDirectives()
	if len(dead) == 0 || stderr == nil {
		return
	}

	for _, d := range dead {
		fmt.Fprintf(stderr,
			"warning: stale //art-dupl:accept %s at %s:%d — matched no clone groups this run; remove it or update the hash\n",
			d.Hash, d.Filename, d.Line)
	}
}
