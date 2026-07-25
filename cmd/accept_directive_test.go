package cmd

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
)

func TestAcceptedSetIsAccepted(t *testing.T) {
	t.Parallel()

	fileWithDirective := "example.go"
	fileContent := `package example

import "fmt"

func foo() {
	//art-dupl:accept
	fmt.Println("hello")
	fmt.Println("world")
	fmt.Println("again")
}

func bar() {
	fmt.Println("hello")
	fmt.Println("world")
	fmt.Println("again")
}
`

	readFile := func(name string) ([]byte, error) {
		if name == fileWithDirective {
			return []byte(fileContent), nil
		}

		return nil, &osPathError{name: name}
	}

	tests := []struct {
		name  string
		group domain.ProcessedCloneGroup
		want  bool
	}{
		{
			name: "directive within clone range suppresses group",
			group: domain.ProcessedCloneGroup{
				Hash: "abc123",
				Clones: []domain.ProcessedClone{
					{CloneRef: domain.CloneRef{Filename: fileWithDirective, LineStart: 5, LineEnd: 10}},
				},
			},
			want: true,
		},
		{
			name: "directive 6 lines above LineStart does not suppress (beyond scan window)",
			group: domain.ProcessedCloneGroup{
				Hash: "abc123",
				Clones: []domain.ProcessedClone{
					{CloneRef: domain.CloneRef{Filename: fileWithDirective, LineStart: 12, LineEnd: 17}},
				},
			},
			want: false,
		},
		{
			name: "directive one line above LineStart suppresses (above-range scan)",
			group: domain.ProcessedCloneGroup{
				Hash: "abc123",
				Clones: []domain.ProcessedClone{
					{CloneRef: domain.CloneRef{Filename: fileWithDirective, LineStart: 7, LineEnd: 12}},
				},
			},
			want: true,
		},
		{
			name: "directive on boundary line (LineEnd) suppresses",
			group: domain.ProcessedCloneGroup{
				Hash: "abc123",
				Clones: []domain.ProcessedClone{
					{CloneRef: domain.CloneRef{Filename: fileWithDirective, LineStart: 4, LineEnd: 7}},
				},
			},
			want: true,
		},
		{
			name: "directive in any clone of group suppresses",
			group: domain.ProcessedCloneGroup{
				Hash: "abc123",
				Clones: []domain.ProcessedClone{
					{CloneRef: domain.CloneRef{Filename: "other.go", LineStart: 1, LineEnd: 5}},
					{CloneRef: domain.CloneRef{Filename: fileWithDirective, LineStart: 6, LineEnd: 8}},
				},
			},
			want: true,
		},
		{
			name: "file with no directives does not suppress",
			group: domain.ProcessedCloneGroup{
				Hash: "abc123",
				Clones: []domain.ProcessedClone{
					{CloneRef: domain.CloneRef{Filename: "other.go", LineStart: 1, LineEnd: 5}},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			set := NewAcceptedSet(readFile)
			got := set.IsAccepted(tt.group)

			if got != tt.want {
				t.Errorf("IsAccepted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAcceptedSetNilIsSafe(t *testing.T) {
	t.Parallel()

	var set *AcceptedSet

	group := domain.ProcessedCloneGroup{
		Hash: "abc123",
		Clones: []domain.ProcessedClone{
			{CloneRef: domain.CloneRef{Filename: "test.go", LineStart: 1, LineEnd: 5}},
		},
	}

	if set.IsAccepted(group) {
		t.Error("nil AcceptedSet should not accept any group")
	}
}

func TestAcceptedSetHashPrecision(t *testing.T) {
	t.Parallel()

	fileContent := `package example

func foo() {
	//art-dupl:accept abc123
	fmt.Println("hello")
}
`

	readFile := func(name string) ([]byte, error) {
		return []byte(fileContent), nil
	}

	set := NewAcceptedSet(readFile)

	matchingGroup := domain.ProcessedCloneGroup{
		Hash: "abc123",
		Clones: []domain.ProcessedClone{
			{CloneRef: domain.CloneRef{Filename: "test.go", LineStart: 1, LineEnd: 6}},
		},
	}

	nonMatchingGroup := domain.ProcessedCloneGroup{
		Hash: "xyz789",
		Clones: []domain.ProcessedClone{
			{CloneRef: domain.CloneRef{Filename: "test.go", LineStart: 1, LineEnd: 6}},
		},
	}

	if !set.IsAccepted(matchingGroup) {
		t.Error("directive with hash should accept group with matching hash")
	}

	if set.IsAccepted(nonMatchingGroup) {
		t.Error("directive with hash should NOT accept group with different hash")
	}
}

func TestAcceptedSetDescriptionText(t *testing.T) {
	t.Parallel()

	fileContent := `package example

func foo() {
	//art-dupl:accept idiomatic boilerplate that cannot be eliminated
	fmt.Println("hello")
	fmt.Println("world")
}
`

	readFile := func(name string) ([]byte, error) {
		return []byte(fileContent), nil
	}

	set := NewAcceptedSet(readFile)

	// A directive with multi-word descriptive text should be treated as a
	// bare accept (no hash), matching ANY group on those lines. The
	// description is for humans, not for precision matching.
	anyGroup := domain.ProcessedCloneGroup{
		Hash: "any-hash-value",
		Clones: []domain.ProcessedClone{
			{CloneRef: domain.CloneRef{Filename: "test.go", LineStart: 4, LineEnd: 7}},
		},
	}

	if !set.IsAccepted(anyGroup) {
		t.Error("directive with multi-word description should accept any group (description is not a hash)")
	}
}

func TestAcceptedSetInlineDirective(t *testing.T) {
	t.Parallel()

	// Regression test: a directive on the SAME line as code (trailing comment,
	// matching golangci-lint convention) must be recognized, not silently
	// dropped. Before the fix, scanFile used HasPrefix-after-TrimSpace which
	// only matched standalone comment lines.
	fileContent := `package example

func foo() {
	fmt.Println("hello") //art-dupl:accept inline trailing directive
	fmt.Println("world")
	fmt.Println("again")
}
`

	readFile := func(name string) ([]byte, error) {
		return []byte(fileContent), nil
	}

	set := NewAcceptedSet(readFile)

	group := domain.ProcessedCloneGroup{
		Hash: "abc123",
		Clones: []domain.ProcessedClone{
			{CloneRef: domain.CloneRef{Filename: "test.go", LineStart: 4, LineEnd: 7}},
		},
	}

	if !set.IsAccepted(group) {
		t.Error("inline trailing directive (code //art-dupl:accept) must suppress the group")
	}
}

func TestAcceptedSetInlineDirectiveWithHash(t *testing.T) {
	t.Parallel()

	// Inline directive WITH a hash token must still do precision matching.
	fileContent := "package example\n\n" +
		"func foo() {\n" +
		"\tfmt.Println(\"hello\") //art-dupl:accept deadbeef\n" +
		"}\n"

	readFile := func(name string) ([]byte, error) {
		return []byte(fileContent), nil
	}

	set := NewAcceptedSet(readFile)

	matching := domain.ProcessedCloneGroup{
		Hash: "deadbeef",
		Clones: []domain.ProcessedClone{
			{CloneRef: domain.CloneRef{Filename: "test.go", LineStart: 3, LineEnd: 5}},
		},
	}

	nonMatching := domain.ProcessedCloneGroup{
		Hash: "cafef00d",
		Clones: []domain.ProcessedClone{
			{CloneRef: domain.CloneRef{Filename: "test.go", LineStart: 3, LineEnd: 5}},
		},
	}

	if !set.IsAccepted(matching) {
		t.Error("inline directive with matching hash must accept the group")
	}

	if set.IsAccepted(nonMatching) {
		t.Error("inline directive with non-matching hash must NOT accept the group")
	}
}

func TestAcceptedSetCaching(t *testing.T) {
	readCount := 0

	fileContent := `package example

func foo() {
	//art-dupl:accept
	fmt.Println("hello")
}
`

	readFile := func(name string) ([]byte, error) {
		readCount++

		return []byte(fileContent), nil
	}

	set := NewAcceptedSet(readFile)

	group := domain.ProcessedCloneGroup{
		Hash: "abc123",
		Clones: []domain.ProcessedClone{
			{CloneRef: domain.CloneRef{Filename: "test.go", LineStart: 1, LineEnd: 6}},
		},
	}

	// Call multiple times
	_ = set.IsAccepted(group)
	_ = set.IsAccepted(group)
	_ = set.IsAccepted(group)

	if readCount != 1 {
		t.Errorf("expected file to be read once, got %d reads", readCount)
	}
}

func TestShouldSuppressGroupAcceptDirectives(t *testing.T) {
	t.Parallel()

	fileContent := `package example

func foo() {
	//art-dupl:accept
	fmt.Println("hello")
}
`

	readFile := func(name string) ([]byte, error) {
		return []byte(fileContent), nil
	}

	set := NewAcceptedSet(readFile)

	acceptedGroup := domain.ProcessedCloneGroup{
		Hash: "abc123",
		Clones: []domain.ProcessedClone{
			{CloneRef: domain.CloneRef{Filename: "test.go", LineStart: 1, LineEnd: 6}},
		},
	}

	nonAcceptedGroup := domain.ProcessedCloneGroup{
		Hash: "abc123",
		Clones: []domain.ProcessedClone{
			{CloneRef: domain.CloneRef{Filename: "test.go", LineStart: 10, LineEnd: 15}},
		},
	}

	suppression := SuppressionConfig{AcceptDirectives: set}

	if !shouldSuppressGroup(acceptedGroup, suppression) {
		t.Error("should suppress group with accept directive")
	}

	if shouldSuppressGroup(nonAcceptedGroup, suppression) {
		t.Error("should NOT suppress group without accept directive")
	}

	// Nil AcceptDirectives should not suppress
	nilSuppression := SuppressionConfig{}
	if shouldSuppressGroup(acceptedGroup, nilSuppression) {
		t.Error("nil AcceptDirectives should not suppress")
	}
}

// osPathError is a minimal error for test file readers.
type osPathError struct{ name string }

func (e *osPathError) Error() string { return "file not found: " + e.name }
