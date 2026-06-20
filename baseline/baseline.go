// Package baseline records and compares sets of accepted clone hashes so
// art-dupl can run as a CI gate: a baseline captures the clones a codebase
// currently has, and a check reports only clones that are new relative to the
// baseline.
//
// File format (JSON):
//
//	{
//	  "version": "1.0",
//	  "threshold": 15,
//	  "recorded_at": "2026-06-20T18:30:00Z",
//	  "entries": [
//	    {"hash": "abc123", "files": ["a.go", "b.go"], "tokens": 42}
//	  ]
//	}
//
// Matching is by clone-group hash (content fingerprint), so renamed files or
// moved code that keeps the same structure still matches the baseline.
package baseline

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"
)

// DefaultBaselinePath is the conventional location for the baseline file.
const DefaultBaselinePath = ".art-dupl-baseline.json"

// baselineFilePerm is the file mode for written baseline files.
const baselineFilePerm = 0o600

// Entry records a single accepted clone group.
type Entry struct {
	Hash   string   `json:"hash"`
	Files  []string `json:"files"`
	Tokens int      `json:"tokens"`
}

// File is the on-disk baseline representation.
type File struct {
	Version    string    `json:"version"`
	Threshold  int       `json:"threshold"`
	RecordedAt time.Time `json:"recordedAt"`
	Entries    []Entry   `json:"entries"`

	// hashes is a lookup index built by indexHashes for O(1) Has() calls.
	hashes map[string]bool
}

// NewFile creates an empty baseline ready to accept entries.
func NewFile(threshold int) *File {
	return &File{
		Version:    "1.0",
		Threshold:  threshold,
		RecordedAt: time.Now().UTC(),
		Entries:    []Entry{},
		hashes:     make(map[string]bool),
	}
}

// Add records a clone group in the baseline. Duplicate hashes are ignored.
func (bf *File) Add(hash string, files []string, tokens int) {
	if hash == "" || bf.hashes[hash] {
		return
	}

	bf.Entries = append(bf.Entries, Entry{Hash: hash, Files: files, Tokens: tokens})
	bf.hashes[hash] = true
}

// Has reports whether the baseline already accepts the given clone hash.
func (bf *File) Has(hash string) bool {
	bf.ensureIndex()

	return bf.hashes[hash]
}

// Len returns the number of accepted clone groups.
func (bf *File) Len() int { return len(bf.Entries) }

// Load reads a baseline from disk. Returns a wrapped error if the file cannot
// be read or parsed.
func Load(path string) (*File, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- path is user-provided, validated by caller
	if err != nil {
		return nil, fmt.Errorf("read baseline %s: %w", path, err)
	}

	var loaded File

	err = json.Unmarshal(data, &loaded)
	if err != nil {
		return nil, fmt.Errorf("parse baseline %s: %w", path, err)
	}

	loaded.ensureIndex()

	return &loaded, nil
}

// Save writes the baseline to disk as formatted JSON. Entries are sorted by
// hash for deterministic, reviewable diffs.
func (bf *File) Save(path string) error {
	bf.ensureIndex()

	sort.Slice(bf.Entries, func(i, j int) bool {
		return bf.Entries[i].Hash < bf.Entries[j].Hash
	})

	data, err := json.MarshalIndent(bf, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal baseline: %w", err)
	}

	data = append(data, '\n')

	err = os.WriteFile(path, data, baselineFilePerm) // #nosec G304 -- path is user-provided
	if err != nil {
		return fmt.Errorf("write baseline %s: %w", path, err)
	}

	return nil
}

// Exists reports whether a baseline file is present at the given path.
func Exists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}

// ensureIndex builds the hash lookup map if it hasn't been built yet.
func (bf *File) ensureIndex() {
	if bf.hashes == nil {
		bf.hashes = make(map[string]bool, len(bf.Entries))
		for _, e := range bf.Entries {
			bf.hashes[e.Hash] = true
		}
	}
}
