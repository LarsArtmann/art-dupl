package hash

import (
	"fmt"
	"math"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// RollingHash implements a Rabin-Karp style rolling hash for node sequences
type RollingHash struct {
	base   uint64
	mod    uint64
	window []uint64
	hash   uint64
	power  uint64
}

// NewRollingHash creates a new rolling hash with specified window size
func NewRollingHash(windowSize int) *RollingHash {
	rh := &RollingHash{
		base:   257,            // Prime base
		mod:    math.MaxUint64, // Use full uint64 range to avoid expensive modulo
		window: make([]uint64, windowSize),
		power:  1,
	}

	// Pre-compute base^(windowSize-1) for efficient removal
	for i := 0; i < windowSize-1; i++ {
		rh.power *= rh.base
	}

	return rh
}

// Reset initializes hash with first window
func (rh *RollingHash) Reset(firstValue uint64) {
	rh.hash = 0
	for i := range rh.window {
		if i == 0 {
			rh.window[i] = firstValue
		} else {
			rh.window[i] = 0 // Will be filled by subsequent Roll calls
		}
		rh.hash = rh.hash*rh.base + rh.window[i]
	}
}

// Roll advances the window by one position
func (rh *RollingHash) Roll(newValue uint64) uint64 {
	// Remove oldest value: hash = hash - oldest * base^(windowSize-1)
	oldest := rh.window[0]
	rh.hash -= oldest * rh.power

	// Shift left by base position and add new value
	rh.hash = rh.hash*rh.base + newValue

	// Update window (circular buffer)
	copy(rh.window[:], rh.window[1:])
	rh.window[len(rh.window)-1] = newValue

	return rh.hash
}

// HashValue returns current hash
func (rh *RollingHash) HashValue() uint64 {
	return rh.hash
}

// WindowHash represents a hash value at a specific position
type WindowHash struct {
	Hash   uint64
	Pos    int    // Position in file's node sequence
	File   string // Source file
	Length int    // Window length
}

// HashDetector implements efficient hash-based code clone detection
type HashDetector struct {
	threshold int
}

// NewHashDetector creates a new hash detector
func NewHashDetector(threshold int) *HashDetector {
	return &HashDetector{
		threshold: threshold,
	}
}

// FindDuplOver finds duplicates using rolling hash-based comparison
func (h *HashDetector) FindDuplOver(data []*syntax.Node, threshold int) <-chan syntax.Match {
	resultChan := make(chan syntax.Match)

	go func() {
		defer close(resultChan)

		// Group nodes by file for proper processing
		fileNodes := h.groupNodesByFile(data)

		// Generate rolling hashes for each file
		allHashes := make([]WindowHash, 0)
		for filename, nodes := range fileNodes {
			fileHashes := h.generateRollingHashes(nodes, filename, threshold)
			allHashes = append(allHashes, fileHashes...)
		}

		// Find hash collisions (potential duplicates)
		hashGroups := h.groupHashesByValue(allHashes)

		// Validate and convert to syntax.Match
		matches := h.validateHashGroups(hashGroups, threshold, fileNodes)
		for _, match := range matches {
			resultChan <- match
		}
	}()

	return resultChan
}

// groupNodesByFile organizes nodes by their source file
func (h *HashDetector) groupNodesByFile(data []*syntax.Node) map[string][]*syntax.Node {
	fileNodes := make(map[string][]*syntax.Node)
	for _, node := range data {
		if node.Filename != "" {
			fileNodes[node.Filename] = append(fileNodes[node.Filename], node)
		}
	}
	return fileNodes
}

// generateRollingHashes creates rolling hashes for all sliding windows
func (h *HashDetector) generateRollingHashes(nodes []*syntax.Node, filename string, threshold int) []WindowHash {
	if len(nodes) < threshold {
		return nil
	}

	hashes := make([]WindowHash, 0)

	// Create rolling hasher
	roller := NewRollingHash(threshold)

	// Initialize with first window
	if len(nodes) >= threshold {
		var firstValues []uint64
		for i := range threshold {
			firstValues = append(firstValues, uint64(nodes[i].Type))
		}

		// Initialize hash
		roller.hash = 0
		for _, val := range firstValues {
			roller.hash = roller.hash*roller.base + val
		}

		// Fill window
		roller.window = make([]uint64, threshold)
		copy(roller.window, firstValues)

		// Record first hash
		hashes = append(hashes, WindowHash{
			Hash:   roller.HashValue(),
			Pos:    0,
			File:   filename,
			Length: threshold,
		})

		// Roll through remaining positions
		for i := threshold; i < len(nodes); i++ {
			newHash := roller.Roll(uint64(nodes[i].Type))
			hashes = append(hashes, WindowHash{
				Hash:   newHash,
				Pos:    i - threshold + 1,
				File:   filename,
				Length: threshold,
			})
		}
	}

	return hashes
}

// groupHashesByValue groups hashes by their value to find collisions
func (h *HashDetector) groupHashesByValue(hashes []WindowHash) map[uint64][]WindowHash {
	groups := make(map[uint64][]WindowHash)
	for _, wh := range hashes {
		groups[wh.Hash] = append(groups[wh.Hash], wh)
	}
	return groups
}

// validateHashGroups validates hash groups and converts to syntax.Match with improved logic
func (h *HashDetector) validateHashGroups(hashGroups map[uint64][]WindowHash, threshold int, fileNodes map[string][]*syntax.Node) []syntax.Match {
	var matches []syntax.Match

	for hash, group := range hashGroups {
		// Need at least 2 different files for a valid clone
		if len(group) < 2 || !h.hasMultipleFiles(group) {
			continue
		}

		// Group by file to avoid multiple fragments from same file
		fileGroups := make(map[string]WindowHash)
		for _, wh := range group {
			// Only keep first occurrence per file
			if _, exists := fileGroups[wh.File]; !exists {
				fileGroups[wh.File] = wh
			}
		}

		// Create fragments from unique file entries
		var fragments [][]*syntax.Node
		for file, wh := range fileGroups {
			nodes, exists := fileNodes[file]
			if !exists || wh.Pos+wh.Length > len(nodes) {
				continue
			}

			// Extract fragment at the correct position
			fragment := make([]*syntax.Node, wh.Length)
			copy(fragment, nodes[wh.Pos:wh.Pos+wh.Length])
			fragments = append(fragments, fragment)
		}

		// Only create match if we have at least 2 different files
		if len(fragments) >= 2 {
			match := syntax.Match{
				Hash:  fmt.Sprintf("%x", hash),
				Frags: fragments,
			}
			matches = append(matches, match)
		}
	}

	return matches
}

// hasMultipleFiles checks if hash group contains multiple files
func (h *HashDetector) hasMultipleFiles(group []WindowHash) bool {
	files := make(map[string]bool)
	for _, wh := range group {
		files[wh.File] = true
		if len(files) > 1 {
			return true
		}
	}
	return false
}
