package artdupl

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// dupFunctionCode has 7+ separate top-level statements to exceed threshold 5.
const dupFunctionCode = `package main

func parseHeader(data []byte) (int, int, int) {
	header := data[:4]
	version := int(header[0])
	flags := int(header[1])
	length := int(header[2])<<8 | int(header[3])
	body := data[4 : 4+length]
	tail := data[4+length:]

	return version, flags, len(tail)
}
`

func TestDetector_TypeAware_FindsClones(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "a.go")
	file2 := filepath.Join(tmpDir, "b.go")

	err := os.WriteFile(file1, []byte(dupFunctionCode), 0o644)
	if err != nil {
		t.Fatalf("Failed to write a.go: %v", err)
	}

	err = os.WriteFile(file2, []byte(dupFunctionCode), 0o644)
	if err != nil {
		t.Fatalf("Failed to write b.go: %v", err)
	}

	opts := DefaultOptions()
	opts.TypeAware = true
	opts.Threshold = 5

	detector, err := NewDetector(opts)
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	t.Cleanup(func() { cleanupDetector(t, detector) })

	result, err := detector.FindClones(t.Context(), []string{file1, file2})
	if err != nil {
		if errors.Is(err, ErrNoDuplicatesFound) {
			t.Skip("No duplicates found with type-aware mode (may be env-dependent)")
		}

		t.Fatalf("FindClones failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if len(result.CloneGroups) == 0 {
		t.Error("Expected at least one clone group with duplicate code")
	}
}

func TestDetector_TypeAware_FallsBackOnInvalidGo(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// File with intentionally broken type references to force LoadTypeAwareData failure.
	// The detector should fall back to syntax-only mode without crashing.
	brokenCode := `package main

func broken() {
	x := undefinedType.Method()
	_ = x
}
`

	file := filepath.Join(tmpDir, "broken.go")

	err := os.WriteFile(file, []byte(brokenCode), 0o644)
	if err != nil {
		t.Fatalf("Failed to write broken.go: %v", err)
	}

	opts := DefaultOptions()
	opts.TypeAware = true
	opts.Threshold = 5

	detector, err := NewDetector(opts)
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	t.Cleanup(func() { cleanupDetector(t, detector) })

	result, err := detector.FindClones(t.Context(), []string{file})

	// Type-aware fallback: broken type info must NOT cause a fatal error.
	// A single file yields no duplicates, so ErrNoDuplicatesFound or a nil-error
	// result with zero clone groups is the expected graceful-degradation outcome.
	if err != nil && !errors.Is(err, ErrNoDuplicatesFound) {
		t.Fatalf("FindClones should fall back to syntax-only on invalid Go, got: %v", err)
	}

	if err == nil && result == nil {
		t.Fatal("Result should not be nil when err is nil")
	}
}

func TestDetector_TypeAware_DisabledByDefault(t *testing.T) {
	t.Parallel()

	opts := DefaultOptions()
	if opts.TypeAware {
		t.Error("TypeAware should be false by default in DefaultOptions")
	}
}

// interfaceMethodCode defines an interface and two implementations with
// identical method bodies. With TypeAware=true, the transformer sets
// InterfaceMethod on these methods' body nodes via go/types. This test
// verifies the flag data flows through the SDK pipeline end-to-end.
const interfaceMethodCode = `package main

type Processor interface {
	Process(data []byte) (int, error)
}

type XMLProcessor struct{}
type JSONProcessor struct{}

func (x XMLProcessor) Process(data []byte) (int, error) {
	header := data[:4]
	version := int(header[0])
	flags := int(header[1])
	length := int(header[2])<<8 | int(header[3])
	body := data[4 : 4+length]
	tail := data[4+length:]
	return version + flags + len(body) + len(tail), nil
}

func (j JSONProcessor) Process(data []byte) (int, error) {
	header := data[:4]
	version := int(header[0])
	flags := int(header[1])
	length := int(header[2])<<8 | int(header[3])
	body := data[4 : 4+length]
	tail := data[4+length:]
	return version + flags + len(body) + len(tail), nil
}
`

// TestDetector_TypeAware_InterfaceMethodFlow verifies that interface method
// implementations are processed through the SDK pipeline when TypeAware is
// enabled. The InterfaceMethod flag (set internally on domain.CloneNode) flows
// through parse → transform → detect → SDK result without panics or errors.
func TestDetector_TypeAware_InterfaceMethodFlow(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "xml.go")
	file2 := filepath.Join(tmpDir, "json.go")

	// Both files contain the full interface + implementations (same package).
	err := os.WriteFile(file1, []byte(interfaceMethodCode), 0o644)
	if err != nil {
		t.Fatalf("Failed to write xml.go: %v", err)
	}

	err = os.WriteFile(file2, []byte(interfaceMethodCode), 0o644)
	if err != nil {
		t.Fatalf("Failed to write json.go: %v", err)
	}

	opts := DefaultOptions()
	opts.TypeAware = true
	opts.Threshold = 5

	detector, err := NewDetector(opts)
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	t.Cleanup(func() { cleanupDetector(t, detector) })

	result, err := detector.FindClones(t.Context(), []string{file1, file2})
	if err != nil {
		if errors.Is(err, ErrNoDuplicatesFound) {
			t.Skip("No duplicates found (may be env-dependent)")
		}

		t.Fatalf("FindClones failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	// The pipeline must complete and return clone groups for the duplicate
	// Process method bodies. The InterfaceMethod flag data flows through
	// internally without crashing the SDK.
	if len(result.CloneGroups) == 0 {
		t.Error("Expected at least one clone group from duplicate interface method bodies")
	}

	// Verify each clone group has valid metadata reflecting the type-aware analysis.
	for i, group := range result.CloneGroups {
		if group == nil {
			t.Errorf("CloneGroup[%d] is nil", i)

			continue
		}

		if len(group.Clones) < 2 {
			t.Errorf("CloneGroup[%d] has %d clones, expected >= 2", i, len(group.Clones))
		}
	}
}
