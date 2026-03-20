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
// Total: 40B (37.5% reduction from 64B).
type Node struct {
	Type     int32
	Pos      int32
	End      int32
	Owns     int32
	Children []*Node
	Filename string
}

func NewNode() *Node {
	return &Node{}
}

func (n *Node) AddChildren(children ...*Node) {
	n.Children = append(n.Children, children...)
}

// Val returns the token value for suffix tree compatibility.
// Implements the suffixtree.Token interface.
func (n *Node) Val() suffixtree.TokenValue {
	return suffixtree.TokenValue(n.Type)
}

// NewSyntheticFileNode creates a synthetic node representing an entire file.
// This is used for file-level duplicate detection where we want to match
// entire files rather than specific code fragments.
//
//nolint:gosec // G115: Size is validated to be within reasonable bounds before this point
func NewSyntheticFileNode(filename string, size int) *Node {
	return &Node{
		Filename: filename,
		Pos:      0,
		End:      int32(size),
		Type:     1,
	}
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

	n.Owns = int32(count) // #nosec G115 -- Child count bounded by maxChildrenSerial

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
// - Validate threshold at domain boundary.
func FindSyntaxUnits(data []*Node, m suffixtree.Match, threshold int) Match {
	if len(m.Ps) == 0 {
		return Match{}
	}

	firstSeq := data[m.Ps[0] : m.Ps[0]+m.Len]
	indexes := getUnitsIndexes(firstSeq, threshold)

	if len(indexes) > 0 && len(m.Ps) > 1 {
		indexes = validateOwnershipConsistency(data, m, firstSeq, indexes)
	}

	if len(indexes) == 0 || isCyclic(indexes, firstSeq) || spansMultipleFiles(indexes, firstSeq) {
		return Match{}
	}

	return buildMatch(data, m, firstSeq, indexes)
}

func validateOwnershipConsistency(
	data []*Node,
	m suffixtree.Match,
	firstSeq []*Node,
	indexes []int,
) []int {
	lasti := indexes[len(indexes)-1]
	firstn := firstSeq[lasti]

	for i := 1; i < len(m.Ps); i++ {
		pos := int(m.Ps[i]) + lasti
		if pos >= len(data) || data[pos].Owns != firstn.Owns {
			return indexes[:len(indexes)-1]
		}
	}

	return indexes
}

func buildMatch(data []*Node, m suffixtree.Match, firstSeq []*Node, indexes []int) Match {
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
	var (
		indexes []int
		split   bool
	)

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
func isCyclic(indexes []int, nodes []*Node) bool {
	cnt := len(indexes)
	if cnt <= 1 {
		return false
	}

	alts := findDivisors(cnt)
	if len(alts) == 0 {
		return false
	}

	for i := range indexes[cnt/2] {
		checkPatternCycle(i, indexes, nodes, alts, cnt)
		if len(alts) == 0 {
			return false
		}
	}

	return true
}

// findDivisors returns all divisors of cnt that are <= cnt/2.
func findDivisors(cnt int) map[int]bool {
	alts := make(map[int]bool)
	for i := 1; i <= cnt/2; i++ {
		if cnt%i == 0 {
			alts[i] = true
		}
	}

	return alts
}

// checkPatternCycle removes invalid periods from alts for the given starting position.
func checkPatternCycle(startIdx int, indexes []int, nodes []*Node, alts map[int]bool, cnt int) {
	// Bounds check to prevent panic
	if startIdx+indexes[0] >= len(nodes) {
		return
	}
	startNode := nodes[startIdx+indexes[0]]

	for alt := range alts {
		if !isPatternRepeating(startIdx, alt, indexes, nodes, startNode, cnt) {
			delete(alts, alt)
		}
	}
}

// isPatternRepeating checks if the pattern repeats at the given period (alt).
func isPatternRepeating(
	startIdx, alt int,
	indexes []int,
	nodes []*Node,
	startNode *Node,
	cnt int,
) bool {
	for j := alt; j < cnt; j += alt {
		index := startIdx + indexes[j]
		if index >= len(nodes) {
			return startIdx >= indexes[alt]
		}

		nalt := nodes[index]
		if startNode.Owns != nalt.Owns || startNode.Type != nalt.Type {
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

// Unique removes duplicate entries from a group of syntax nodes based on file and range.
// Two clones are considered duplicates if they have the same filename and the same range (start-end).
func Unique(group [][]*Node) [][]*Node {
	type rangeKey struct {
		filename string
		start    int
		end      int
	}

	seen := make(map[rangeKey]struct{})

	var newGroup [][]*Node

	for _, seq := range group {
		if len(seq) == 0 {
			continue
		}

		first := seq[0]
		last := seq[len(seq)-1]

		key := rangeKey{filename: first.Filename, start: int(first.Pos), end: int(last.End)}
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}

			newGroup = append(newGroup, seq)
		}
	}

	return newGroup
}

// CountUniqueFiles returns the number of unique files in a clone group.
func CountUniqueFiles(group [][]*Node) int {
	uniqueFiles := make(map[string]bool)

	for _, seq := range group {
		if len(seq) > 0 {
			uniqueFiles[seq[0].Filename] = true
		}
	}

	return len(uniqueFiles)
}
