package domain_test

import (
	"fmt"
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
)

// TestStringID_IntegrationMinimal tests minimal StringID integration.
func TestStringID_IntegrationMinimal(t *testing.T) {
	// Test basic accessor methods work
	clone := domain.Clone{}

	// Set fields using StringID API
	clone.SetFilename("main.go")
	clone.SetFragment("func main() {}")
	clone.SetHash("abc123")

	// Verify accessors
	if clone.FilenameString() != "main.go" {
		t.Errorf("FilenameString() = %s, want main.go", clone.FilenameString())
	}

	if clone.FragmentString() != "func main() {}" {
		t.Errorf("FragmentString() = %s, want func main() {}", clone.FragmentString())
	}

	if clone.HashString() != "abc123" {
		t.Errorf("HashString() = %s, want abc123", clone.HashString())
	}

	//nolint:forbidigo // debug output in test
	fmt.Println("✅ StringID integration works!")
	//nolint:forbidigo // debug output in test
	fmt.Printf("   Filename: %s (ID: %d)\n", clone.FilenameString(), clone.Filename)
	//nolint:forbidigo // debug output in test
	fmt.Printf("   Fragment: %s (ID: %d)\n", clone.FragmentString(), clone.Fragment)
	//nolint:forbidigo // debug output in test
	fmt.Printf("   Hash: %s (ID: %d)\n", clone.HashString(), clone.Hash)
}
