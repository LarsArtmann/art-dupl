package hash

import (
	"crypto/sha1"
	"fmt"
	"strings"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// HashDetector implements hash-based code clone detection
type HashDetector struct {
	threshold int
}

// NewHashDetector creates a new hash-based detector
func NewHashDetector(threshold int) *HashDetector {
	return &HashDetector{
		threshold: threshold,
	}
}

// HashMatch represents a group of files with identical hash signatures
type HashMatch struct {
	Hash  string
	Frags [][]*syntax.Node
}

// FindDuplOver finds duplicates using hash-based comparison
func (h *HashDetector) FindDuplOver(data []*syntax.Node, threshold int) <-chan syntax.Match {
	resultChan := make(chan syntax.Match)
	
	go func() {
		defer close(resultChan)
		
		// Group nodes by file
		fileNodes := make(map[string][]*syntax.Node)
		for _, node := range data {
			if node.Filename != "" {
				fileNodes[node.Filename] = append(fileNodes[node.Filename], node)
			}
		}
		
		// Generate hash groups for each file
		hashGroups := make(map[string][]*syntax.Node)
		
		for _, nodes := range fileNodes {
			hashes := h.generateHashesForNodes(nodes)
			for hash, nodeList := range hashes {
				// Add filename context to each node
				for range nodeList {
					// filename already set from grouping
				}
				hashGroups[hash] = append(hashGroups[hash], nodeList...)
			}
		}
		
		// Filter by threshold and emit matches
		for hash, nodes := range hashGroups {
			if len(nodes) >= 2 {
				// Check if any sequence meets the threshold
				seqLength := h.calculateSequenceLength(nodes)
				if seqLength >= threshold {
					// Create positions for match
					var positions []int
					for i := range nodes {
						if i%seqLength == 0 { // Start of each sequence
							positions = append(positions, i)
						}
					}
					
					if len(positions) >= 2 {
						// Convert to syntax.Match format
						match := syntax.Match{
							Hash: hash,
							Frags: make([][]*syntax.Node, len(positions)),
						}
						
						for i, pos := range positions {
							end := pos + seqLength
							if end > len(nodes) {
								end = len(nodes)
							}
							match.Frags[i] = nodes[pos:end]
						}
						
						resultChan <- match
					}
				}
			}
		}
	}()
	
	return resultChan
}

// generateHashesForNodes generates sliding window hashes for nodes
func (h *HashDetector) generateHashesForNodes(nodes []*syntax.Node) map[string][]*syntax.Node {
	hashes := make(map[string][]*syntax.Node)
	
	// Generate sliding window hashes
	windowSize := h.threshold
	for i := 0; i <= len(nodes)-windowSize; i++ {
		window := nodes[i : i+windowSize]
		hash := h.computeHash(window)
		
		// Only keep if we have meaningful content
		if h.isSignificantHash(hash) {
			hashes[hash] = append(hashes[hash], nodes[i:i+windowSize]...)
		}
	}
	
	return hashes
}

// calculateSequenceLength estimates the length of repeated sequences
func (h *HashDetector) calculateSequenceLength(nodes []*syntax.Node) int {
	if len(nodes) == 0 {
		return 0
	}
	
	// Simple heuristic: look for repeating patterns
	// For now, just use the threshold as a minimum
	return h.threshold
}

// computeHash computes SHA1 hash for a sequence of nodes
func (h *HashDetector) computeHash(nodes []*syntax.Node) string {
	var content strings.Builder
	
	for _, node := range nodes {
		content.WriteString(fmt.Sprintf("%d:", node.Type))
		if node.Filename != "" {
			content.WriteString(node.Filename)
		}
		content.WriteString(";")
	}
	
	hash := sha1.Sum([]byte(content.String()))
	return fmt.Sprintf("%x", hash)
}

// isSignificantHash checks if a hash represents significant content
func (h *HashDetector) isSignificantHash(hash string) bool {
	// Skip very simple or repetitive patterns
	// This is a basic heuristic - could be enhanced
	return len(hash) > 10 && !strings.Contains(hash, "000000")
}