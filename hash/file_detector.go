package hash

import (
	"fmt"
	"os"

	"github.com/LarsArtmann/art-dupl/pkg/format"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/zeebo/xxh3"
)

// FileHash represents a hash of a complete file.
type FileHash struct {
	Hash     string
	Filename string
	Size     int
	Content  []byte
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

// FindFileDuplicates finds exact file duplicates by hashing file contents.
// This is a convenience function that takes file paths directly without requiring syntax nodes.
// Files smaller than the threshold (in bytes) are ignored.
func FindFileDuplicates(files []string, threshold int) []FileDuplicate {
	fd := NewFileDetector(threshold)

	// Hash all files
	fileHashes, _ := fd.hashFiles(files)

	// Group by hash
	hashGroups := fd.groupByHash(fileHashes)

	// Convert to FileDuplicate slice
	var duplicates []FileDuplicate

	for hash, group := range hashGroups {
		if len(group) >= 2 {
			// Filter by size threshold
			validFiles := make([]FileHash, 0)

			for _, fh := range group {
				if fh.Size >= threshold {
					validFiles = append(validFiles, fh)
				}
			}

			if len(validFiles) >= 2 {
				duplicates = append(duplicates, FileDuplicate{
					Hash:  hash,
					Files: validFiles,
				})
			}
		}
	}

	return duplicates
}

// FindDuplOver finds exact file duplicates using SHA-256 hashing.
func (f *FileDetector) FindDuplOver(data []*syntax.Node, threshold int) <-chan syntax.Match {
	resultChan := make(chan syntax.Match)

	go func() {
		defer close(resultChan)

		// Extract unique files from nodes
		fileList := f.extractUniqueFiles(data)

		// Read and hash all files
		fileHashes, err := f.hashFiles(fileList)
		if err != nil {
			return // Exit if can't read files
		}

		// Group files by identical hash
		hashGroups := f.groupByHash(fileHashes)

		// Convert groups to matches
		matches := f.convertToMatches(hashGroups, threshold)
		for _, match := range matches {
			resultChan <- match
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

// hashFiles calculates XXH3 hash for each file content.
//
// PERFORMANCE: XXH3 is ~20x faster than crypto/sha256 and includes
// native ARM64 NEON SIMD optimizations. DO NOT replace with cryptographic
// hash functions (SHA-256, etc.) - this hash is for deduplication only,
// not security. See: https://github.com/zeebo/xxh3
func (f *FileDetector) hashFiles(files []string) ([]FileHash, error) {
	var fileHashes []FileHash

	for _, filename := range files {
		// Read file content
		content, err := os.ReadFile(
			filename,
		) // #nosec G304 -- Filename comes from user-provided paths, verified by caller
		if err != nil {
			continue // Skip files that can't be read
		}

		// Calculate XXH3 hash
		//

		hash := xxh3.Hash(content)

		fileHash := FileHash{
			Hash:     format.Hash(hash),
			Filename: filename,
			Size:     len(content),
			Content:  content,
		}

		fileHashes = append(fileHashes, fileHash)
	}

	return fileHashes, nil
}

// groupByHash groups files by identical hash values.
func (f *FileDetector) groupByHash(fileHashes []FileHash) map[string][]FileHash {
	hashGroups := make(map[string][]FileHash)

	for _, fileHash := range fileHashes {
		hashGroups[fileHash.Hash] = append(hashGroups[fileHash.Hash], fileHash)
	}

	return hashGroups
}

// convertToMatches converts hash groups to syntax.Match format.
func (f *FileDetector) convertToMatches(
	hashGroups map[string][]FileHash,
	threshold int,
) []syntax.Match {
	var matches []syntax.Match

	for hash, group := range hashGroups {
		// Skip groups with only one file (no duplicate)
		if len(group) < 2 {
			continue
		}

		// Check if files meet minimum size threshold
		validFiles := make([]FileHash, 0)

		for _, fileHash := range group {
			if fileHash.Size >= threshold {
				validFiles = append(validFiles, fileHash)
			}
		}

		// Only create match if we have at least 2 substantial files
		if len(validFiles) >= 2 {
			// Create fragments for each valid file
			var fragments [][]*syntax.Node

			for _, fileHash := range validFiles {
				// Create a synthetic node representing the entire file
				node := &syntax.Node{
					Filename: fileHash.Filename,
					Pos:      0,
					End:      int32(fileHash.Size), //nolint:gosec // File sizes bounded by int32 in practice
					Type:     1,                    // Use a generic type
				}

				fragments = append(fragments, []*syntax.Node{node})
			}

			match := syntax.Match{
				Hash:  fmt.Sprintf("%x", hash),
				Frags: fragments,
			}
			matches = append(matches, match)
		}
	}

	return matches
}
