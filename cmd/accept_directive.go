package cmd

import (
	"bufio"
	"bytes"
	"strings"
	"sync"

	"github.com/LarsArtmann/art-dupl/domain"
)

const acceptDirectivePrefix = "//art-dupl:accept"

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

// IsAccepted checks if any clone in the group has an accept directive within
// its line range. A directive on line N accepts clones where LineStart <= N <=
// LineEnd. When the directive includes a hash (//art-dupl:accept <hash>), it
// only matches groups with that exact hash.
func (a *AcceptedSet) IsAccepted(group domain.ProcessedCloneGroup) bool {
	if a == nil {
		return false
	}

	for _, clone := range group.Clones {
		directives := a.getDirectives(clone.Filename)

		for _, d := range directives {
			if d.Line < clone.LineStart || d.Line > clone.LineEnd {
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

		if rest != "" {
			d.Hash = rest
		}

		directives = append(directives, d)
	}

	return directives
}
