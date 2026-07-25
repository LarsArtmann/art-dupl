package hash

import (
	"context"
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
type FileDetector struct{}

// NewFileDetector creates a new file-level hash detector.
func NewFileDetector() *FileDetector {
	return &FileDetector{}
}

// FileDuplicate represents a group of files with identical content.
type FileDuplicate struct {
	Hash  string
	Files []FileHash
}

// groupByHash hashes each file and groups results by content hash.
// Files that cannot be hashed are silently skipped (errors logged at Debug).
// Respects context cancellation between files.
func (f *FileDetector) groupByHash(ctx context.Context, files []string) map[string][]FileHash {
	groups := make(map[string][]FileHash)

	for _, filename := range files {
		if ctx.Err() != nil {
			return groups
		}

		fh, ok := f.hashFile(filename)
		if !ok {
			continue
		}

		groups[fh.Hash] = append(groups[fh.Hash], fh)
	}

	return groups
}

// filterDuplicateGroups returns groups containing at least 2 files that meet
// the size threshold. Because the hash is over file CONTENTS, every file in a
// group shares the same size, so checking group[0].Size is sufficient.
func filterDuplicateGroups(groups map[string][]FileHash, threshold int) []FileDuplicate {
	var dups []FileDuplicate

	for hash, group := range groups {
		if len(group) >= 2 && group[0].Size >= threshold {
			dups = append(dups, FileDuplicate{Hash: hash, Files: group})
		}
	}

	return dups
}

// FindFileDuplicates finds exact file duplicates by hashing file contents.
// Files are hashed one at a time via io.Copy through XXH3 — file content is
// never held in memory. Only (hash, filename, size) is retained per file.
func FindFileDuplicates(ctx context.Context, files []string, threshold int) []FileDuplicate {
	fd := NewFileDetector()
	groups := fd.groupByHash(ctx, files)

	return filterDuplicateGroups(groups, threshold)
}

// fileHashToFragment converts a FileHash into a synthetic syntax node fragment.
func fileHashToFragment(fh FileHash) []*syntax.Node {
	return []*syntax.Node{syntax.NewSyntheticFileNode(fh.Filename, fh.Size)}
}

// FindDuplOver finds exact file duplicates using XXH3 hashing.
func (f *FileDetector) FindDuplOver(ctx context.Context, data []*syntax.Node, threshold int) <-chan syntax.Match {
	resultChan := make(chan syntax.Match)

	go func() {
		defer close(resultChan)

		fileList := f.extractUniqueFiles(data)
		groups := f.groupByHash(ctx, fileList)
		dups := filterDuplicateGroups(groups, threshold)

		for _, dup := range dups {
			fragments := make([][]*syntax.Node, 0, len(dup.Files))
			for _, fh := range dup.Files {
				fragments = append(fragments, fileHashToFragment(fh))
			}

			select {
			case resultChan <- syntax.Match{
				Hash:  dup.Hash,
				Frags: fragments,
			}:
			case <-ctx.Done():
				return
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
func (f *FileDetector) hashFile(filename string) (FileHash, bool) {
	file, err := os.Open(
		filename,
	) // #nosec G304 -- Filename comes from user-provided paths, verified by caller
	if err != nil {
		return FileHash{}, f.fileError(filename, err, "skipping file that cannot be opened")
	}

	//art-dupl:accept standard defer-close idiom (different types: os.File vs io.Closer)
	defer func() { _ = file.Close() }()

	fi, err := file.Stat()
	if err != nil {
		return FileHash{}, f.fileError(filename, err, "skipping file that cannot be stat'd")
	}

	size := int(fi.Size())

	hasher := xxh3.New()

	_, err = io.Copy(hasher, file)
	if err != nil {
		return FileHash{}, f.fileError(filename, err, "skipping file that cannot be read")
	}

	return FileHash{
		Hash:     format.Hash(hasher.Sum64()),
		Filename: filename,
		Size:     size,
	}, true
}

// fileError logs a file operation error and returns failure state.
func (f *FileDetector) fileError(filename string, err error, reason string) bool {
	logger.Default.Debug(reason, "file", filename, "err", err)

	return false
}
