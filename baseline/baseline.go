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
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"time"
)

// DefaultBaselinePath is the conventional location for the baseline file.
const DefaultBaselinePath = ".art-dupl-baseline.json"

// baselineVersion is the file-format version this binary reads and writes.
const baselineVersion = "1.0"

// ErrUnsupportedBaselineVersion is returned by Load when the on-disk baseline
// uses a file-format version this binary cannot interpret. CI must fail loudly
// rather than silently treating an incompatible file as an empty baseline.
var ErrUnsupportedBaselineVersion = errors.New("unsupported baseline version")

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
// The files slice is copied so later caller mutation cannot corrupt the entry.
func (bf *File) Add(hash string, files []string, tokens int) {
	if hash == "" || bf.hashes[hash] {
		return
	}

	filesCopy := slices.Clone(files)

	bf.Entries = append(bf.Entries, Entry{Hash: hash, Files: filesCopy, Tokens: tokens})
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

	if loaded.Version != baselineVersion {
		return nil, fmt.Errorf("load baseline %s: %w: %q", path, ErrUnsupportedBaselineVersion, loaded.Version)
	}

	loaded.ensureIndex()

	return &loaded, nil
}

// Save writes the baseline to disk as formatted JSON. Entries are sorted by
// hash for deterministic, reviewable diffs. The write is atomic: data is
// written to a temp file in the same directory and then renamed, so a crash
// mid-write never leaves a truncated baseline that would silently empty the
// CI accepted-set.
func (bf *File) Save(path string) error {
	bf.ensureIndex()

	sort.Slice(bf.Entries, func(i, j int) bool {
		return bf.Entries[i].Hash < bf.Entries[j].Hash
	})

	data, err := json.Marshal(bf, jsontext.WithIndentPrefix(""), jsontext.WithIndent("  "))
	if err != nil {
		return fmt.Errorf("marshal baseline: %w", err)
	}

	data = append(data, '\n')

	tmp, err := os.CreateTemp(filepath.Dir(path), ".art-dupl-baseline-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp baseline file: %w", err)
	}

	tmpName := tmp.Name()
	removeTemp := func() { _ = os.Remove(tmpName) }

	_, err = tmp.Write(data)
	if err != nil {
		_ = tmp.Close()

		removeTemp()

		return fmt.Errorf("write baseline %s: %w", path, err)
	}

	err = tmp.Close()
	if err != nil {
		removeTemp()

		return fmt.Errorf("close baseline %s: %w", path, err)
	}

	err = os.Chmod(tmpName, baselineFilePerm)
	if err != nil {
		removeTemp()

		return fmt.Errorf("chmod baseline %s: %w", path, err)
	}

	err = os.Rename(tmpName, path)
	if err != nil {
		removeTemp()

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
			if e.Hash != "" {
				bf.hashes[e.Hash] = true
			}
		}
	}
}
