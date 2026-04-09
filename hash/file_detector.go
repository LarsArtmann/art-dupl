package hash

import (
	"io"
	"os"

	"github.com/LarsArtmann/art-dupl/pkg/format"
	"github.com/LarsArtmann/art-dupl/pkg/logger"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/zeebo/xxh3"
)

// FileHash represents a hash of a complete file.
type FileHash struct {
	Hash     string
	Filename string
	Size     int
}

// FileDetector implements exact file duplicate detection.
type FileDetector struct {
	threshold int
}

// NewFileDetector creates a new file-level hash detector.
func NewFileDetector(threshold int) *FileDetector {
	return &FileDetector{
		threshold: threshold,
	}
}

// FileDuplicate represents a group of files with identical content.
type FileDuplicate struct {
	Hash  string
	Files []FileHash
}

// hashEntry is a lightweight record for streaming deduplication.
// Only stores what's needed: hash value + file metadata. No file content.
type hashEntry struct {
	hash     string
	filename string
	size     int
}

// FindFileDuplicates finds exact file duplicates by hashing file contents.
// This is a convenience function that takes file paths directly without requiring syntax nodes.
// Files smaller than the threshold (in bytes) are ignored.
func FindFileDuplicates(files []string, threshold int) []FileDuplicate {
	fd := NewFileDetector(threshold)

	groups := make(map[string][]hashEntry)

	for _, filename := range files {
		entry, ok := fd.hashFile(filename)
		if !ok || entry.size < threshold {
			continue
		}

		groups[entry.hash] = append(groups[entry.hash], entry)
	}

	var duplicates []FileDuplicate

	for hash, group := range groups {
		if len(group) >= 2 {
			duplicates = append(duplicates, fd.convertGroup(hash, group))
		}
	}

	return duplicates
}

// FindDuplOver finds exact file duplicates using XXH3 hashing.
func (f *FileDetector) FindDuplOver(data []*syntax.Node, threshold int) <-chan syntax.Match {
	resultChan := make(chan syntax.Match)

	go func() {
		defer close(resultChan)

		fileList := f.extractUniqueFiles(data)

		groups := make(map[string][]hashEntry)

		for _, filename := range fileList {
			entry, ok := f.hashFile(filename)
			if !ok {
				continue
			}

			groups[entry.hash] = append(groups[entry.hash], entry)
		}

		for _, group := range groups {
			validFiles := make([]hashEntry, 0, len(group))

			for _, entry := range group {
				if entry.size >= threshold {
					validFiles = append(validFiles, entry)
				}
			}

			if len(validFiles) >= 2 {
				fragments := make([][]*syntax.Node, 0, len(validFiles))

				for _, entry := range validFiles {
					node := syntax.NewSyntheticFileNode(entry.filename, entry.size)
					fragments = append(fragments, []*syntax.Node{node})
				}

				resultChan <- syntax.Match{
					Hash:  validFiles[0].hash,
					Frags: fragments,
				}
			}
		}
	}()

	return resultChan
}

// extractUniqueFiles gets unique list of files from nodes.
func (f *FileDetector) extractUniqueFiles(data []*syntax.Node) []string {
	fileSet := make(map[string]bool)

	var files []string

	for _, node := range data {
		if node.Filename != "" && !fileSet[node.Filename] {
			fileSet[node.Filename] = true
			files = append(files, node.Filename)
		}
	}

	return files
}

// hashFile streams a file through XXH3 hasher and returns only the hash metadata.
// File content is never stored — the xxh3.Hasher reads via io.Copy in 32KB chunks.
//
// PERFORMANCE: XXH3 is ~20x faster than crypto/sha256 and includes
// native ARM64 NEON SIMD optimizations. DO NOT replace with cryptographic
// hash functions (SHA-256, etc.) - this hash is for deduplication only,
// not security. See: https://github.com/zeebo/xxh3
func (f *FileDetector) hashFile(filename string) (hashEntry, bool) {
	file, err := os.Open(
		filename,
	) // #nosec G304 -- Filename comes from user-provided paths, verified by caller
	if err != nil {
		logger.Default.Debug("skipping file that cannot be opened", "file", filename, "err", err)

		return hashEntry{}, false
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		logger.Default.Debug("skipping file that cannot be stat'd", "file", filename, "err", err)

		return hashEntry{}, false
	}

	size := int(fi.Size())

	hasher := xxh3.New()
	if _, err := io.Copy(hasher, file); err != nil {
		logger.Default.Debug("skipping file that cannot be read", "file", filename, "err", err)

		return hashEntry{}, false
	}

	return hashEntry{
		hash:     format.Hash(hasher.Sum64()),
		filename: filename,
		size:     size,
	}, true
}

// convertGroup converts a group of hash entries into a FileDuplicate.
func (f *FileDetector) convertGroup(hash string, group []hashEntry) FileDuplicate {
	files := make([]FileHash, len(group))

	for i, entry := range group {
		files[i] = FileHash{
			Hash:     entry.hash,
			Filename: entry.filename,
			Size:     entry.size,
		}
	}

	return FileDuplicate{
		Hash:  hash,
		Files: files,
	}
}
