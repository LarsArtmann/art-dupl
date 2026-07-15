package printer

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// TestSemanticPrecisionCorpus validates end-to-end semantic mode quality:
// true positives are found, false positives are suppressed.
func TestSemanticPrecisionCorpus(t *testing.T) {
	t.Parallel()

	t.Run("literal_normalization_enables_type2_detection", func(t *testing.T) {
		t.Parallel()

		code := `package test

import (
	"errors"
	"os"
)

func ReadConfigA(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.New("failed to read config")
	}
	if len(data) == 0 {
		return nil, errors.New("config is empty")
	}
	return data, nil
}

func ReadConfigB(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.New("failed to read template")
	}
	if len(content) == 0 {
		return nil, errors.New("template is empty")
	}
	return content, nil
}
`
		matches := findSemanticDupl(t, code, 3)
		if len(matches) == 0 {
			t.Error("Expected clone match for Type-2 clone with different literal values")
		}
	})

	t.Run("lock_defer_unlock_suppressed", func(t *testing.T) {
		t.Parallel()

		code := `package test

import "sync"

func ProcessA(m *sync.Mutex) {
	m.Lock()
	defer m.Unlock()
}

func ProcessB(m *sync.Mutex) {
	m.Lock()
	defer m.Unlock()
}
`
		seqs := findSemanticCloneNodes(t, code, 2)
		if len(seqs) == 0 {
			t.Skip("no matches found")
		}
		result := EvaluateActionability(seqs)
		if result != domain.NonActionable {
			t.Errorf("Lock+Defer Unlock should be non-actionable, got %s", result)
		}
	})

	t.Run("rlock_defer_runlock_suppressed", func(t *testing.T) {
		t.Parallel()

		code := `package test

import "sync"

func ReadA(m *sync.RWMutex) {
	m.RLock()
	defer m.RUnlock()
}

func ReadB(m *sync.RWMutex) {
	m.RLock()
	defer m.RUnlock()
}
`
		seqs := findSemanticCloneNodes(t, code, 2)
		if len(seqs) == 0 {
			t.Skip("no matches found")
		}
		result := EvaluateActionability(seqs)
		if result != domain.NonActionable {
			t.Errorf("RLock+Defer RUnlock should be non-actionable, got %s", result)
		}
	})

	t.Run("error_definitions_not_matched", func(t *testing.T) {
		t.Parallel()

		code := `package test

import "errors"

var (
	ErrNotFound     = errors.New("resource not found")
	ErrUnauthorized = errors.New("user is not authorized")
	ErrConflict     = errors.New("state conflict occurred")
)
`
		matches := findSemanticDupl(t, code, 2)
		if len(matches) > 0 {
			t.Errorf("Error definitions with different names should NOT match, got %d", len(matches))
		}
	})

	t.Run("real_business_logic_clone_detected", func(t *testing.T) {
		t.Parallel()

		code := `package test

import (
	"errors"
	"strings"
)

func CountWords(text string) (map[string]int, error) {
	counts := make(map[string]int)
	for _, word := range strings.Fields(text) {
		if word == "" {
			continue
		}
		c, ok := counts[word]
		if !ok {
			counts[word] = 1
		} else {
			counts[word] = c + 1
		}
	}
	if len(counts) == 0 {
		return nil, errors.New("no words found")
	}
	return counts, nil
}

func CountLines(text string) (map[string]int, error) {
	result := make(map[string]int)
	for _, line := range strings.Fields(text) {
		if line == "" {
			continue
		}
		n, found := result[line]
		if !found {
			result[line] = 1
		} else {
			result[line] = n + 1
		}
	}
	if len(result) == 0 {
		return nil, errors.New("no lines found")
	}
	return result, nil
}
`
		matches := findSemanticDupl(t, code, 3)
		if len(matches) == 0 {
			t.Error("Expected clone match for real business logic with renamed vars + different literals")
		}
	})

	t.Run("validation_chain_clone_detected", func(t *testing.T) {
		t.Parallel()

		code := `package test

import (
	"errors"
	"strings"
)

type UserValidator struct{}
type ProductValidator struct{}

func (v *UserValidator) ValidateUser(name, email string, age int) error {
	if name == "" {
		return errors.New("name is required")
	}
	if len(name) > 100 {
		return errors.New("name too long")
	}
	if email == "" {
		return errors.New("email is required")
	}
	if !strings.Contains(email, "@") {
		return errors.New("invalid email")
	}
	if age < 0 || age > 150 {
		return errors.New("invalid age")
	}
	return nil
}

func (v *ProductValidator) ValidateProduct(title, sku string, price int) error {
	if title == "" {
		return errors.New("title is required")
	}
	if len(title) > 200 {
		return errors.New("title too long")
	}
	if sku == "" {
		return errors.New("sku is required")
	}
	if !strings.Contains(sku, "-") {
		return errors.New("invalid sku")
	}
	if price < 0 || price > 1000000 {
		return errors.New("invalid price")
	}
	return nil
}
`
		matches := findSemanticDupl(t, code, 5)
		if len(matches) == 0 {
			t.Error("Expected clone match for validation chain with different field names and literals")
		}
	})
}

func findSemanticDupl(t *testing.T, code string, threshold int) []syntax.Match {
	t.Helper()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	node, err := golang.ParseWithConfig(tmpFile, golang.MustParseConfig(golang.DetectionModeSemantic))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	stream := syntax.Serialize(node)
	// Append a unique sentinel so the suffix tree creates proper leaf nodes
	// for all suffixes (standard Ukkonen requirement).
	stream = append(stream, &syntax.Node{Type: math.MinInt32})

	tree := suffixtree.New()
	for _, n := range stream {
		if err := tree.Update(n); err != nil {
			t.Fatalf("tree.Update() error = %v", err)
		}
	}

	ctx := context.Background()
	var matches []syntax.Match

	for m := range tree.FindDuplOver(ctx, threshold) {
		unit := syntax.FindSyntaxUnits(stream, m, threshold)
		if len(unit.Frags) > 0 {
			matches = append(matches, unit)
		}
	}

	return matches
}

func findSemanticCloneNodes(t *testing.T, code string, threshold int) [][]*domain.CloneNode {
	t.Helper()

	matches := findSemanticDupl(t, code, threshold)
	if len(matches) == 0 {
		return nil
	}

	var allSeqs [][]*domain.CloneNode

	for _, m := range matches {
		allSeqs = append(allSeqs, ToCloneNodeSeqs(m.Frags)...)
	}

	return allSeqs
}
