package artdupl

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// combinedFamilyA/B: structurally identical functions on different concrete
// types (the generics-extraction shape — matched only by the type-erased pass).
const combinedFamilyA = `package main

type First struct{ Value int }

func SumFirst(f First, g First, times int) int {
	total := f.Value + g.Value
	for i := 0; i < times; i++ {
		total += i
	}
	return total
}
`

const combinedFamilyB = `package main

type Second struct{ Value int }

func SumSecond(s Second, t Second, times int) int {
	total := s.Value + t.Value
	for i := 0; i < times; i++ {
		total += i
	}
	return total
}
`

// combinedPairA/B: structurally identical with identical local types —
// matched by the type-aware pass as a regular clone group.
const combinedPairA = `package main

func pairSumInts(a int, b int) int {
	total := a + b
	steps := 3
	for i := 0; i < steps; i++ {
		total += i
	}
	return total
}
`

const combinedPairB = `package main

func pairSumIntsAgain(x int, y int) int {
	total := x + y
	steps := 3
	for i := 0; i < steps; i++ {
		total += i
	}
	return total
}
`

// writeCombinedFixtures writes the four combined-mode fixtures and returns
// their paths.
func writeCombinedFixtures(t *testing.T) []string {
	t.Helper()

	dir := t.TempDir()
	fixtures := map[string]string{
		"family_a.go": combinedFamilyA,
		"family_b.go": combinedFamilyB,
		"pair_a.go":   combinedPairA,
		"pair_b.go":   combinedPairB,
	}

	files := make([]string, 0, len(fixtures))

	for name, content := range fixtures {
		path := filepath.Join(dir, name)

		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}

		files = append(files, path)
	}

	return files
}

// TestDetector_CombinedTypeAwareAndSuggestGenerics pins the combined-mode
// contract at the SDK layer: one shared type-check, two passes, both outputs —
// the identical-type pair group (type-aware pass) AND the cross-type family
// group (type-erased pass).
func TestDetector_CombinedTypeAwareAndSuggestGenerics(t *testing.T) {
	t.Parallel()

	files := writeCombinedFixtures(t)

	opts := DefaultOptions()
	opts.TypeAware = true
	opts.SuggestGenerics = true
	opts.Threshold = 1

	detector, err := NewDetector(opts)
	if err != nil {
		t.Fatalf("NewDetector: %v", err)
	}

	t.Cleanup(func() { cleanupDetector(t, detector) })

	result, err := detector.FindClones(t.Context(), files)
	if err != nil {
		t.Fatalf("FindClones (combined): %v", err)
	}

	var hasPairGroup, hasFamilyGroup bool

	for _, group := range result.CloneGroups {
		var pairFiles, familyFiles int

		for _, clone := range group.Clones {
			switch filepath.Base(clone.Filename) {
			case "pair_a.go", "pair_b.go":
				pairFiles++
			case "family_a.go", "family_b.go":
				familyFiles++
			}
		}

		if pairFiles == 2 {
			hasPairGroup = true
		}

		if familyFiles == 2 {
			hasFamilyGroup = true
		}
	}

	if !hasPairGroup {
		t.Error("combined mode must report the identical-type pair group (type-aware pass)")
	}

	if !hasFamilyGroup {
		t.Error("combined mode must report the cross-type family group (type-erased pass)")
	}
}

// TestDetector_CombinedSuppressesReceiverNoise pins the false-positive
// elimination at the SDK layer: a cross-type family with a single divergent
// position (receiver-only noise) is dropped in combined mode while
// suggest-generics alone still reports it.
func TestDetector_CombinedSuppressesReceiverNoise(t *testing.T) {
	t.Parallel()

	const noiseA = `package main

type Clock struct{ Name string }

func renderClock(c Clock) string {
	parts := c.Name
	report := "ts:" + parts
	width := len(report) + 1
	return report[:width-1]
}
`

	const noiseB = `package main

type Meter struct{ Name string }

func renderMeter(m Meter) string {
	parts := m.Name
	report := "ts:" + parts
	width := len(report) + 1
	return report[:width-1]
}
`

	runMode := func(t *testing.T, combined bool) (groupCount int, divergentStmtFound bool) {
		t.Helper()

		dir := t.TempDir()

		files := []string{
			filepath.Join(dir, "noise_a.go"),
			filepath.Join(dir, "noise_b.go"),
		}

		if err := os.WriteFile(files[0], []byte(noiseA), 0o644); err != nil {
			t.Fatalf("write noise_a.go: %v", err)
		}

		if err := os.WriteFile(files[1], []byte(noiseB), 0o644); err != nil {
			t.Fatalf("write noise_b.go: %v", err)
		}

		opts := DefaultOptions()
		opts.Threshold = 1
		opts.SuggestGenerics = true
		opts.TypeAware = combined

		detector, err := NewDetector(opts)
		if err != nil {
			t.Fatalf("NewDetector: %v", err)
		}

		t.Cleanup(func() { cleanupDetector(t, detector) })

		result, err := detector.FindClones(t.Context(), files)
		if err != nil {
			if errors.Is(err, ErrNoDuplicatesFound) {
				return 0, false
			}

			t.Fatalf("FindClones: %v", err)
		}

		for _, group := range result.CloneGroups {
			groupCount++

			for _, clone := range group.Clones {
				// Line 6 is the divergent receiver statement (`parts := c.Name`).
				// A clone starting there means the full cross-type body leaked
				// through; the same-type core starts at line 7.
				if clone.LineStart <= 6 {
					divergentStmtFound = true
				}
			}
		}

		return groupCount, divergentStmtFound
	}

	groups, leaked := runMode(t, true)
	if leaked {
		t.Error("combined mode must drop the full-body cross-type noise group (line 6 divergent statement)")
	}

	if groups == 0 {
		t.Error("combined mode should still report the same-type statement core from the type-aware pass")
	}

	sgGroups, sgLeaked := runMode(t, false)
	if sgGroups == 0 || !sgLeaked {
		t.Error("suggest-generics alone must report the full cross-type noise family including the divergent statement")
	}
}

// TestDetector_CombinedStreamResults smoke-tests the streaming path through
// the combined mode.
func TestDetector_CombinedStreamResults(t *testing.T) {
	t.Parallel()

	files := writeCombinedFixtures(t)

	opts := DefaultOptions()
	opts.TypeAware = true
	opts.SuggestGenerics = true
	opts.Threshold = 1

	detector, err := NewDetector(opts)
	if err != nil {
		t.Fatalf("NewDetector: %v", err)
	}

	t.Cleanup(func() { cleanupDetector(t, detector) })

	resultChan, err := detector.FindClonesStreamResult(t.Context(), files)
	if err != nil {
		t.Fatalf("FindClonesStreamResult: %v", err)
	}

	groupCount := 0

	for streamResult := range resultChan {
		if streamResult.Err != nil {
			t.Fatalf("stream error: %v", streamResult.Err)
		}

		if streamResult.Group != nil {
			groupCount++
		}
	}

	if groupCount == 0 {
		t.Error("combined streaming must emit clone groups")
	}
}
