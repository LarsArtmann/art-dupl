package domain

import (
	"fmt"

	"github.com/LarsArtmann/art-dupl/pkg/enum"
)

// CloneCategory represents the category of code that was duplicated.
//
//nolint:recvcheck // standard Go JSON convention: MarshalJSON value receiver, UnmarshalJSON pointer receiver
type CloneCategory string

const (
	CategoryFunction        CloneCategory = "function"
	CategoryMethod          CloneCategory = "method"
	CategoryTest            CloneCategory = "test"
	CategoryStruct          CloneCategory = "struct"
	CategoryInterface       CloneCategory = "interface"
	CategoryHandler         CloneCategory = "handler"
	CategoryLoop            CloneCategory = "loop"
	CategoryConditional     CloneCategory = "conditional"
	CategoryTestBoilerplate CloneCategory = "test-boilerplate"
	CategoryTestFixture     CloneCategory = "test-fixture"
	CategoryAssignment      CloneCategory = "assignment"
	CategoryExpression      CloneCategory = "expression"
	CategoryBlock           CloneCategory = "block"
	CategoryCall            CloneCategory = "call"
	CategoryReturn          CloneCategory = "return"
	CategoryDefer           CloneCategory = "defer"
	CategoryUnknown         CloneCategory = "unknown"
)

// ClonePriority represents how important it is to address this clone.
//
//nolint:recvcheck // standard Go JSON convention: MarshalJSON value receiver, UnmarshalJSON pointer receiver
type ClonePriority string

const (
	PriorityCritical ClonePriority = "critical"
	PriorityHigh     ClonePriority = "high"
	PriorityMedium   ClonePriority = "medium"
	PriorityLow      ClonePriority = "low"
)

// CloneActionability indicates whether a clone can realistically be deduplicated.
//
//nolint:recvcheck // standard Go JSON convention: MarshalJSON value receiver, UnmarshalJSON pointer receiver
type CloneActionability string

const (
	// Actionable clones represent real logic duplication that can be extracted,
	// composed, or otherwise refactored to reduce duplication.
	Actionable CloneActionability = "actionable"
	// NonActionable clones are idiomatic patterns that cannot be deduplicated
	// without breaking Go semantics, interfaces, or standard conventions.
	NonActionable CloneActionability = "non-actionable"
)

// CloneType represents the standard code clone taxonomy (Bellon et al.).
//
//nolint:recvcheck // standard Go JSON convention: MarshalJSON value receiver, UnmarshalJSON pointer receiver
type CloneType string

const (
	// CloneType1 is an exact copy-paste: identical text (except whitespace/comments).
	// In semantic mode, all clones are Type 1 because identifiers and literals
	// are encoded into the token hash.
	CloneType1 CloneType = "type-1"
	// CloneType2 is a parameterized copy: same structure, different identifier
	// names or literal values. Detected when Name fields differ across fragments
	// but the AST structure is identical.
	CloneType2 CloneType = "type-2"
	// CloneType3 is a near-miss clone: small structural differences (added,
	// deleted, or modified statements). Future capability — not currently
	// detected by the suffix tree algorithm.
	CloneType3 CloneType = "type-3"
)

// priorityData holds display data for priorities.
type priorityData struct {
	color string
	emoji string
}

var priorityDisplay = map[ClonePriority]priorityData{ //nolint:gochecknoglobals // static display lookup table
	PriorityCritical: {color: "var(--error)", emoji: "\U0001f534"},
	PriorityHigh:     {color: "var(--warning)", emoji: "\U0001f7e0"},
	PriorityMedium:   {color: "var(--accent)", emoji: "\U0001f7e1"},
	PriorityLow:      {color: "var(--success)", emoji: "\U0001f7e2"},
}

// IsValid returns true if the priority is one of the defined constants.
func (p ClonePriority) IsValid() bool {
	switch p {
	case PriorityCritical, PriorityHigh, PriorityMedium, PriorityLow:
		return true
	default:
		return false
	}
}

// String returns the string representation of the priority.
func (p ClonePriority) String() string { return string(p) }

// Rank returns the ordinal rank of the priority (higher = more important).
// Used for sorting and comparison. Returns 0 for invalid values.
func (p ClonePriority) Rank() int {
	switch p {
	case PriorityCritical:
		return 4
	case PriorityHigh:
		return 3
	case PriorityMedium:
		return 2
	case PriorityLow:
		return 1
	default:
		return 0
	}
}

func (p ClonePriority) GetPriorityColor() string {
	if data, ok := priorityDisplay[p]; ok {
		return data.color
	}

	return "var(--text-secondary)"
}

func (p ClonePriority) GetPriorityEmoji() string {
	if data, ok := priorityDisplay[p]; ok {
		return data.emoji
	}

	return "\u26aa"
}

var categoryEmojis = map[CloneCategory]string{ //nolint:gochecknoglobals // static category lookup table
	CategoryFunction:        "\u26a1",
	CategoryMethod:          "\U0001f527",
	CategoryTest:            "\U0001f9ea",
	CategoryTestBoilerplate: "\U0001f9f9",
	CategoryTestFixture:     "\U0001f3af",
	CategoryStruct:          "\U0001f4e6",
	CategoryInterface:       "\U0001f50c",
	CategoryHandler:         "\U0001f3a8",
	CategoryLoop:            "\U0001f504",
	CategoryConditional:     "\U0001f500",
	CategoryAssignment:      "\U0001f4dd",
	CategoryExpression:      "\U0001f4ca",
	CategoryBlock:           "\U0001f4d1",
	CategoryCall:            "\U0001f4de",
	CategoryReturn:          "\U0001f519",
	CategoryDefer:           "\u23f1",
	CategoryUnknown:         "\U0001f4c4",
}

// IsValid returns true if the category is one of the defined constants.
func (c CloneCategory) IsValid() bool {
	switch c {
	case CategoryFunction, CategoryMethod, CategoryTest, CategoryStruct,
		CategoryInterface, CategoryHandler, CategoryLoop, CategoryConditional,
		CategoryTestBoilerplate, CategoryTestFixture, CategoryAssignment,
		CategoryExpression, CategoryBlock, CategoryCall, CategoryReturn,
		CategoryDefer, CategoryUnknown:
		return true
	default:
		return false
	}
}

// String returns the string representation of the category.
func (c CloneCategory) String() string { return string(c) }

func (c CloneCategory) GetCategoryEmoji() string {
	if emoji, ok := categoryEmojis[c]; ok {
		return emoji
	}

	return "\U0001f4c4"
}

// IsValid returns true if the actionability is one of the defined constants.
func (a CloneActionability) IsValid() bool {
	switch a {
	case Actionable, NonActionable:
		return true
	default:
		return false
	}
}

// String returns the string representation of the actionability.
func (a CloneActionability) String() string { return string(a) }

// IsValid returns true if the clone type is one of the defined constants.
func (t CloneType) IsValid() bool {
	switch t {
	case CloneType1, CloneType2, CloneType3:
		return true
	default:
		return false
	}
}

// String returns the string representation of the clone type.
func (t CloneType) String() string { return string(t) }

// MarshalJSON implements json.Marshaler for CloneType.
func (t CloneType) MarshalJSON() ([]byte, error) {
	return enum.MarshalJSON(t, CloneType.IsValid, ErrInvalidCloneType)
}

// UnmarshalJSON implements json.Unmarshaler for CloneType.
func (t *CloneType) UnmarshalJSON(data []byte) error {
	return enum.UnmarshalJSONInto(
		t,
		data,
		CloneType.IsValid,
		ErrInvalidCloneType,
	)
}

// MarshalJSON implements json.Marshaler for ClonePriority.
func (p ClonePriority) MarshalJSON() ([]byte, error) {
	return enum.MarshalJSON(p, ClonePriority.IsValid, ErrInvalidClonePriority)
}

// UnmarshalJSON implements json.Unmarshaler for ClonePriority.
func (p *ClonePriority) UnmarshalJSON(data []byte) error {
	return enum.UnmarshalJSONInto(
		p,
		data,
		ClonePriority.IsValid,
		ErrInvalidClonePriority,
	)
}

// MarshalJSON implements json.Marshaler for CloneCategory.
func (c CloneCategory) MarshalJSON() ([]byte, error) {
	return enum.MarshalJSON(c, CloneCategory.IsValid, ErrInvalidCloneCategory)
}

// UnmarshalJSON implements json.Unmarshaler for CloneCategory.
func (c *CloneCategory) UnmarshalJSON(data []byte) error {
	return enum.UnmarshalJSONInto(
		c,
		data,
		CloneCategory.IsValid,
		ErrInvalidCloneCategory,
	)
}

// MarshalJSON implements json.Marshaler for CloneActionability.
func (a CloneActionability) MarshalJSON() ([]byte, error) {
	return enum.MarshalJSON(a, CloneActionability.IsValid, ErrInvalidCloneActionability)
}

// UnmarshalJSON implements json.Unmarshaler for CloneActionability.
func (a *CloneActionability) UnmarshalJSON(data []byte) error {
	return enum.UnmarshalJSONInto(
		a,
		data,
		CloneActionability.IsValid,
		ErrInvalidCloneActionability,
	)
}

// ClassificationInput holds the data needed to classify a clone.
// Using a struct instead of primitives ensures the compiler catches missing fields
// when new classification inputs are added.
type ClassificationInput struct {
	Filename string
	NodeType int32
	Tokens   int
	Lines    int
}

// CloneClassification provides metadata about a code clone for actionable reports.
type CloneClassification struct {
	Category             CloneCategory
	IsTest               bool
	Priority             ClonePriority
	Actionability        CloneActionability
	NonActionablePattern string
	CloneType            CloneType
	Extractability       Extractability
	Tokens               int
	Lines                int
	Suggestion           string
}

// ProcessedClone represents a single clone instance with extracted fragment data.
// Decouples printer output from syntax.Node internals.
type ProcessedClone struct {
	CloneRef

	StartPos       int32
	EndPos         int32
	TokenCount     int
	FileSize       int
	Classification CloneClassification
}

// Validate checks that the clone's invariants hold.
// A valid clone must have a non-empty filename, LineEnd >= LineStart,
// and a non-negative TokenCount.
func (c ProcessedClone) Validate() error {
	if c.Filename == "" {
		return ErrEmptyFilename
	}

	if c.LineEnd < c.LineStart {
		return ErrLineEndBeforeStart
	}

	if c.TokenCount < 0 {
		return ErrNegativeTokenCount
	}

	return nil
}

// ProcessedCloneGroup represents a group of duplicate code fragments.
type ProcessedCloneGroup struct {
	Hash       string
	TokenCount int
	Clones     []ProcessedClone
}

// NewProcessedCloneGroup creates a group with TokenCount computed from clones.
// This is the canonical constructor — avoids inconsistent TokenCount states.
func NewProcessedCloneGroup(hash string, clones []ProcessedClone) ProcessedCloneGroup {
	g := ProcessedCloneGroup{
		Hash:       hash,
		Clones:     clones,
		TokenCount: 0,
	}
	g.TokenCount = g.TotalTokenCount()

	return g
}

// TotalTokenCount returns the sum of TokenCount across all clones in the group.
// This is the canonical way to compute total tokens — avoids manual iteration.
func (g ProcessedCloneGroup) TotalTokenCount() int {
	total := 0
	for _, c := range g.Clones {
		total += c.TokenCount
	}

	return total
}

// Validate checks that the group's invariants hold:
//   - Must have at least one clone.
//   - Every clone must pass its own Validate().
//   - Group TokenCount must equal the sum of clone TokenCounts.
func (g ProcessedCloneGroup) Validate() error {
	if len(g.Clones) == 0 {
		return ErrEmptyCloneGroup
	}

	for i, c := range g.Clones {
		err := c.Validate()
		if err != nil {
			return fmt.Errorf("clone %d: %w", i, err)
		}
	}

	expected := g.TotalTokenCount()
	if g.TokenCount != expected {
		return fmt.Errorf("%w: has %d, expected %d", ErrTokenCountMismatch, g.TokenCount, expected)
	}

	return nil
}
