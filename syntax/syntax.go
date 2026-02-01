// Package syntax provides unified AST representation for code duplication detection.
//
// This package bridges the gap between language-specific AST parsers
// (golang/ast for Go code) and the language-agnostic suffix tree
// used by the detection algorithm.
//
// Core Types:
// - Node: Unified syntax tree node representing any language construct
// - Match: Represents a clone match with fragments (group of nodes)
// - Frags: Slice of node sequences (each fragment is a sequence of nodes)
//
// Design:
// - Language-agnostic: Works with any language that provides a parser
// - Type-safe: Uses int32 for types (see golang/constants for mapping)
// - Memory-optimized: Careful field ordering for cache efficiency
// - Position-aware: Tracks byte positions and line numbers for all nodes
//
// Usage Flow:
// 1. Parse source files -> language-specific AST (go/ast, etc.)
// 2. Transform AST -> unified syntax.Node tree (see syntax/golang/)
// 3. Build suffix tree from Node sequence (suffixtree.Update())
// 4. Find duplicates using suffix tree (FindDuplOver())
// 5. Convert matches to complete syntax units (FindSyntaxUnits())
//
// Key Functions:
// - FindSyntaxUnits(): Converts suffix tree matches to complete syntax units
// - hashSeq(): Creates hash of node sequence for duplicate detection
// - isCyclic/spansMultipleFiles(): Validation helpers
//
// Performance:
// - maxChildrenSerial constant prevents goroutine stack overflow
// - Node struct is 40B (37.5% reduction from 64B) via int32 fields
// - See MEMORY_LAYOUT_OPTIMIZATION_PLAN.md for details
//
package syntax

import (
	"github.com/LarsArtmann/art-dupl/suffixtree"
)

// To avoid "goroutine stack exceeds" with gigantic slices (Composite Literals).
// 10_000 => 0.89s
// 20_000 => 1.53s
// 30_000 => 2.57s
// 40_000 => 3.89s
// 50_000 => 5.58s
// 60_000 => 7.95s
// 70_000 => 10.15s
// 80_000 => 13.11s
// 90_000 => 16.62s
// 100_000 => 21.42s.
const maxChildrenSerial = 10_000

// Node represents a syntax tree node.
//
// Memory Layout Optimized with int32 fields:
// - int32 fields grouped for cache efficiency (4B each, 16B total)
// - pointer field (8B)
// - string header at end (16B)
// Total: 40B (37.5% reduction from 64B)
type Node struct {
	Type int32
	Pos  int32
	End  int32
	Owns int32
	Children []*Node
	Filename string
}

func NewNode() *Node {
	return &Node{}
}

func (n *Node) AddChildren(children ...*Node) {
	n.Children = append(n.Children, children...)
}

func (n *Node) Val() int {
	return int(n.Type)
}

type Match struct {
	Hash  string
	Frags [][]*Node
}

func Serialize(n *Node) []*Node {
	stream := make([]*Node, 0, 10)
	serial(n, &stream)
	return stream
}

func serial(n *Node, stream *[]*Node) int {
	*stream = append(*stream, n)
	var count int
	for i, child := range n.Children {
		// To avoid "goroutine stack exceeds" with gigantic slices (Composite Literals).
		if i > maxChildrenSerial {
			break
		}

		count += serial(child, stream)
	}
	n.Owns = int32(count)
	return int(n.Owns) + 1
}

// FindSyntaxUnits finds all complete syntax units in the match group and returns them
// with the corresponding hash.
//
// TODO: TYPE SAFETY ISSUE - This function uses int for positions and thresholds
// but the domain package has strongly-typed LineNumber, BytePosition, TokenCount, Threshold.
// Consider:
// - Accept domain.Threshold instead of int
// - Return domain types instead of primitive types
// - Validate threshold at domain boundary
//
// Also: This function is complex (cyclop lint suppression). Consider breaking into smaller
// functions for better testability and readability.
func FindSyntaxUnits(data []*Node, m suffixtree.Match, threshold int) Match { //nolint:cyclop // Syntax unit matching with multiple validation paths
	if len(m.Ps) == 0 {
		return Match{}
	}
	firstSeq := data[m.Ps[0] : m.Ps[0]+m.Len]
	indexes := getUnitsIndexes(firstSeq, threshold)

	// Validate that syntax units have consistent ownership across all positions
	// This ensures we're matching complete syntactic structures with identical tree shapes
	if len(indexes) > 0 && len(m.Ps) > 1 {
		lasti := indexes[len(indexes)-1]
		firstn := firstSeq[lasti]

		// Check each occurrence of the pattern
		for i := 1; i < len(m.Ps); i++ {
			// Ensure we don't go out of bounds
			pos := int(m.Ps[i]) + lasti
			if pos >= len(data) {
				// Position out of bounds, remove this index
				indexes = indexes[:len(indexes)-1]
				break
			}

			n := data[pos]
			if firstn.Owns != n.Owns {
				// Different ownership structure means different tree shapes
				// Remove the problematic index to ensure only complete matches
				indexes = indexes[:len(indexes)-1]
				break
			}
		}
	}
	if len(indexes) == 0 || isCyclic(indexes, firstSeq) || spansMultipleFiles(indexes, firstSeq) {
		return Match{}
	}

	match := Match{Frags: make([][]*Node, len(m.Ps))}
	for i, pos := range m.Ps {
		match.Frags[i] = make([]*Node, len(indexes))
		for j, index := range indexes {
			match.Frags[i][j] = data[int(pos)+index]
		}
	}

	lastIndex := indexes[len(indexes)-1]
	match.Hash = hashSeq(firstSeq[indexes[0] : lastIndex+int(firstSeq[lastIndex].Owns)])
	return match
}

func getUnitsIndexes(nodeSeq []*Node, threshold int) []int {
	var indexes []int
	var split bool
	for i := 0; i < len(nodeSeq); {
		n := nodeSeq[i]
		switch {
		case int(n.Owns) > len(nodeSeq)-i:
			// not complete syntax unit
			i++
			split = true
			continue
		case int(n.Owns)+1 < threshold:
			split = true
		default:
			if split {
				indexes = indexes[:0]
				split = false
			}
			indexes = append(indexes, i)
		}
		i += int(n.Owns) + 1
	}
	return indexes
}

// isCyclic finds out whether there is a repetive pattern in the found clone. If positive,
// it return false to point out that the clone would be redundant.
func isCyclic(indexes []int, nodes []*Node) bool { //nolint:cyclop // Cyclic pattern detection with multiple iteration paths
	cnt := len(indexes)
	if cnt <= 1 {
		return false
	}

	alts := make(map[int]bool)
	for i := 1; i <= cnt/2; i++ {
		if cnt%i == 0 {
			alts[i] = true
		}
	}

	for i := range indexes[cnt/2] {
		nstart := nodes[i+indexes[0]]
	AltLoop:
		for alt := range alts {
			for j := alt; j < cnt; j += alt {
				index := i + indexes[j]
				if index < len(nodes) {
					nalt := nodes[index]
					if nstart.Owns == nalt.Owns && nstart.Type == nalt.Type {
						continue
					}
				} else if i >= indexes[alt] {
					return true
				}
				delete(alts, alt)
				continue AltLoop
			}
		}
		if len(alts) == 0 {
			return false
		}
	}
	return true
}

func spansMultipleFiles(indexes []int, nodes []*Node) bool {
	if len(indexes) < 2 {
		return false
	}
	f := nodes[indexes[0]].Filename
	for i := 1; i < len(indexes); i++ {
		if nodes[indexes[i]].Filename != f {
			return true
		}
	}
	return false
}
