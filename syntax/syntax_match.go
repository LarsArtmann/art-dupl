package syntax

import (
	"github.com/LarsArtmann/art-dupl/suffixtree"
)

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
	match := Match{
		Frags: make([][]*Node, len(m.Ps)),
	}
	for i, pos := range m.Ps {
		match.Frags[i] = make([]*Node, len(indexes))
		for j, index := range indexes {
			match.Frags[i][j] = data[int(pos)+index]
		}
	}

	lastIndex := indexes[len(indexes)-1]
	lastNode := firstSeq[lastIndex]
	endIdx := lastIndex + int(lastNode.Owns)
	if lastNode.Statement || lastNode.Owns == 0 {
		endIdx = lastIndex + 1
	}

	match.Hash = hashSeq(firstSeq[indexes[0]:endIdx])

	return match
}

// fileContainsStatements reports whether any node in the token stream for the
// given filename is marked as a statement.
func fileContainsStatements(data []*Node, filename string) bool {
	for i := range data {
		if data[i].Filename == filename && data[i].Statement {
			return true
		}
	}

	return false
}

func getUnitsIndexes(nodeSeq []*Node, threshold int) []int {
	var (
		indexes []int
		split   bool
	)

	hasStatements := false

	for i := range nodeSeq {
		if nodeSeq[i].Statement {
			hasStatements = true

			break
		}
	}

	for i := 0; i < len(nodeSeq); {
		n := nodeSeq[i]

		if n.Statement {
			if split {
				indexes = indexes[:0]
				split = false
			}

			indexes = append(indexes, i)
			i++

			continue
		}

		if hasStatements {
			i++

			continue
		}

		switch {
		case int(n.Owns) > len(nodeSeq)-i:
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

// isCyclic finds out whether there is a repetive pattern in the found clone.
func isCyclic(indexes []int, nodes []*Node) bool {
	count := len(indexes)
	if count <= 1 {
		return false
	}

	alts := findDivisors(count)
	if len(alts) == 0 {
		return false
	}

	for i := range indexes[count/2] {
		checkPatternCycle(i, indexes, nodes, alts, count)

		if len(alts) == 0 {
			return false
		}
	}

	return true
}

func findDivisors(count int) map[int]bool {
	alts := make(map[int]bool)

	for i := 1; i <= count/2; i++ {
		if count%i == 0 {
			alts[i] = true
		}
	}

	return alts
}

func checkPatternCycle(startIdx int, indexes []int, nodes []*Node, alts map[int]bool, count int) {
	if startIdx+indexes[0] >= len(nodes) {
		return
	}

	startNode := nodes[startIdx+indexes[0]]

	for alt := range alts {
		if !isPatternRepeating(startIdx, alt, indexes, nodes, startNode, count) {
			delete(alts, alt)
		}
	}
}

func isPatternRepeating(
	startIdx, alt int,
	indexes []int,
	nodes []*Node,
	startNode *Node,
	count int,
) bool {
	for j := alt; j < count; j += alt {
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
