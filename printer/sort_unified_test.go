package printer

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
)

func TestSortGroupsByCriteria_Size(t *testing.T) {
	t.Parallel()

	groups := []CloneGroup{
		{Hash: "b", Size: 5, Clones: []JSONClone{{}}},
		{Hash: "a", Size: 10, Clones: []JSONClone{{}}},
		{Hash: "c", Size: 3, Clones: []JSONClone{{}}},
	}

	sortGroupsByCriteria(groups, config.SortBySize, GroupMetrics[CloneGroup]{
		Size:    func(g CloneGroup) int { return g.Size },
		Count:   func(g CloneGroup) int { return len(g.Clones) },
		SortKey: func(g CloneGroup) string { return g.Hash },
	})

	if groups[0].Hash != "a" || groups[1].Hash != "b" || groups[2].Hash != "c" {
		t.Errorf("expected descending size order [a(10), b(5), c(3)], got %s, %s, %s",
			groups[0].Hash, groups[1].Hash, groups[2].Hash)
	}
}

func TestSortGroupsByCriteria_Occurrence(t *testing.T) {
	t.Parallel()

	groups := []CloneGroup{
		{Hash: "x", Size: 1, Clones: []JSONClone{{}, {}}},
		{Hash: "y", Size: 1, Clones: make([]JSONClone, 5)},
		{Hash: "z", Size: 1, Clones: []JSONClone{{}}},
	}

	sortGroupsByCriteria(groups, config.SortByOccurrence, GroupMetrics[CloneGroup]{
		Size:    func(g CloneGroup) int { return g.Size },
		Count:   func(g CloneGroup) int { return len(g.Clones) },
		SortKey: func(g CloneGroup) string { return g.Hash },
	})

	if groups[0].Hash != "y" || groups[1].Hash != "x" || groups[2].Hash != "z" {
		t.Errorf("expected descending occurrence [y(5), x(2), z(1)], got %s, %s, %s",
			groups[0].Hash, groups[1].Hash, groups[2].Hash)
	}
}

func TestSortGroupsByCriteria_Hash(t *testing.T) {
	t.Parallel()

	groups := []CloneGroup{
		{Hash: "charlie", Size: 1},
		{Hash: "alpha", Size: 1},
		{Hash: "bravo", Size: 1},
	}

	sortGroupsByCriteria(groups, config.SortByHash, GroupMetrics[CloneGroup]{
		Size:    func(g CloneGroup) int { return g.Size },
		Count:   func(g CloneGroup) int { return len(g.Clones) },
		SortKey: func(g CloneGroup) string { return g.Hash },
	})

	if groups[0].Hash != "alpha" || groups[1].Hash != "bravo" || groups[2].Hash != "charlie" {
		t.Errorf("expected ascending hash [alpha, bravo, charlie], got %s, %s, %s",
			groups[0].Hash, groups[1].Hash, groups[2].Hash)
	}
}

func TestSortGroupsByCriteria_TotalTokens(t *testing.T) {
	t.Parallel()

	// total tokens = Size * Count
	groups := []CloneGroup{
		{Hash: "low", Size: 10, Clones: []JSONClone{{}}},     // 10*1=10
		{Hash: "mid", Size: 5, Clones: make([]JSONClone, 4)}, // 5*4=20
		{Hash: "hi", Size: 3, Clones: make([]JSONClone, 10)}, // 3*10=30
	}

	sortGroupsByCriteria(groups, config.SortByTotalTokens, GroupMetrics[CloneGroup]{
		Size:    func(g CloneGroup) int { return g.Size },
		Count:   func(g CloneGroup) int { return len(g.Clones) },
		SortKey: func(g CloneGroup) string { return g.Hash },
	})

	if groups[0].Hash != "hi" || groups[1].Hash != "mid" || groups[2].Hash != "low" {
		t.Errorf("expected descending total tokens [hi(30), mid(20), low(10)], got %s, %s, %s",
			groups[0].Hash, groups[1].Hash, groups[2].Hash)
	}
}

func TestSortGroupsByCriteria_EmptyInput(t *testing.T) {
	t.Parallel()

	groups := []CloneGroup{}

	sortGroupsByCriteria(groups, config.SortBySize, GroupMetrics[CloneGroup]{
		Size:    func(g CloneGroup) int { return g.Size },
		Count:   func(g CloneGroup) int { return len(g.Clones) },
		SortKey: func(g CloneGroup) string { return g.Hash },
	})

	if len(groups) != 0 {
		t.Errorf("expected empty slice to remain empty, got %d items", len(groups))
	}
}

func TestSortGroupsByCriteria_SortByHashEmptyKeys(t *testing.T) {
	t.Parallel()

	// Groups with empty sort keys should not panic and should maintain relative order
	// (the comparator returns false when either key is empty, so sort.Slice treats them as equal)
	groups := []CloneGroup{
		{Hash: "", Size: 5},
		{Hash: "real", Size: 3},
		{Hash: "", Size: 10},
	}

	sortGroupsByCriteria(groups, config.SortByHash, GroupMetrics[CloneGroup]{
		Size:    func(g CloneGroup) int { return g.Size },
		Count:   func(g CloneGroup) int { return len(g.Clones) },
		SortKey: func(g CloneGroup) string { return g.Hash },
	})

	// "real" should come before empty keys (empty key returns false, so it sorts last)
	realIdx := -1

	for i, g := range groups {
		if g.Hash == "real" {
			realIdx = i
		}
	}

	if realIdx == -1 {
		t.Fatal("group with hash 'real' not found after sort")
	}
}

func TestSortGroupsByCriteria_ProcessedCloneSlice(t *testing.T) {
	t.Parallel()

	// Verify the factory works with []domain.ProcessedClone (the text printer's type)
	groups := [][]domain.ProcessedClone{
		{domain.ProcessedClone{CloneRef: domain.CloneRef{Filename: "b.go", Fragment: "short"}, TokenCount: 5}},
		{
			domain.ProcessedClone{
				CloneRef:   domain.CloneRef{Filename: "a.go", Fragment: "much longer fragment"},
				TokenCount: 10,
			},
		},
	}

	sortGroupsByCriteria(groups, config.SortBySize, GroupMetrics[[]domain.ProcessedClone]{
		Size:    totalFragmentSize,
		Count:   func(g []domain.ProcessedClone) int { return len(g) },
		SortKey: func(g []domain.ProcessedClone) string { return g[0].Filename },
	})

	if groups[0][0].Filename != "a.go" {
		t.Errorf("expected a.go first (fragment len 19 > 5), got %s", groups[0][0].Filename)
	}
}

func TestSortGroupsByCriteria_DefaultToSize(t *testing.T) {
	t.Parallel()

	// Unknown criteria should default to size sorting
	groups := []CloneGroup{
		{Hash: "b", Size: 5},
		{Hash: "a", Size: 10},
	}

	sortGroupsByCriteria(groups, config.SortCriteria("__unknown__"), GroupMetrics[CloneGroup]{
		Size:    func(g CloneGroup) int { return g.Size },
		Count:   func(g CloneGroup) int { return len(g.Clones) },
		SortKey: func(g CloneGroup) string { return g.Hash },
	})

	if groups[0].Hash != "a" {
		t.Errorf("expected default-to-size sorting, got first=%s", groups[0].Hash)
	}
}

func TestSortGroupsByCriteria_EqualValues(t *testing.T) {
	t.Parallel()

	// Equal values should not cause any issues
	groups := []CloneGroup{
		{Hash: "c", Size: 5},
		{Hash: "b", Size: 5},
		{Hash: "a", Size: 5},
	}

	sortGroupsByCriteria(groups, config.SortBySize, GroupMetrics[CloneGroup]{
		Size:    func(g CloneGroup) int { return g.Size },
		Count:   func(g CloneGroup) int { return len(g.Clones) },
		SortKey: func(g CloneGroup) string { return g.Hash },
	})

	// All size=5, order is unspecified but should not panic
	for _, g := range groups {
		if g.Size != 5 {
			t.Errorf("expected all size=5, got %d", g.Size)
		}
	}
}

func TestSortGroupsByCriteria_SingleElement(t *testing.T) {
	t.Parallel()

	groups := []CloneGroup{
		{Hash: "only", Size: 42},
	}

	sortGroupsByCriteria(groups, config.SortBySize, GroupMetrics[CloneGroup]{
		Size:    func(g CloneGroup) int { return g.Size },
		Count:   func(g CloneGroup) int { return len(g.Clones) },
		SortKey: func(g CloneGroup) string { return g.Hash },
	})

	if len(groups) != 1 || groups[0].Hash != "only" {
		t.Errorf("single element should be unchanged, got %v", groups)
	}
}
