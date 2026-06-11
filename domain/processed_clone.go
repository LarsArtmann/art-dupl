package domain

// CloneCategory represents the category of code that was duplicated.
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
	CategoryIdiom           CloneCategory = "idiom"
	CategoryUnknown         CloneCategory = "unknown"
)

// ClonePriority represents how important it is to address this clone.
type ClonePriority string

const (
	PriorityCritical ClonePriority = "critical"
	PriorityHigh     ClonePriority = "high"
	PriorityMedium   ClonePriority = "medium"
	PriorityLow      ClonePriority = "low"
)

// CloneActionability indicates whether a clone can realistically be deduplicated.
type CloneActionability string

const (
	// Actionable clones represent real logic duplication that can be extracted,
	// composed, or otherwise refactored to reduce duplication.
	Actionable CloneActionability = "actionable"
	// NonActionable clones are idiomatic patterns that cannot be deduplicated
	// without breaking Go semantics, interfaces, or standard conventions.
	NonActionable CloneActionability = "non-actionable"
)

// priorityData holds display data for priorities.
type priorityData struct {
	color string
	emoji string
}

var priorityDisplay = map[ClonePriority]priorityData{
	PriorityCritical: {color: "var(--error)", emoji: "\U0001f534"},
	PriorityHigh:     {color: "var(--warning)", emoji: "\U0001f7e0"},
	PriorityMedium:   {color: "var(--accent)", emoji: "\U0001f7e1"},
	PriorityLow:      {color: "var(--success)", emoji: "\U0001f7e2"},
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

var categoryEmojis = map[CloneCategory]string{
	CategoryFunction:        "\u26a1",
	CategoryMethod:          "\U0001f527",
	CategoryTest:            "\U0001f9ea",
	CategoryTestBoilerplate: "\U0001f9f9",
	CategoryTestFixture:     "\U0001f3af",
	CategoryStruct:          "\U0001f4e6",
	CategoryInterface:       "\U0001f50c",
	CategoryHandler:         "\U0001f3af",
	CategoryLoop:            "\U0001f504",
	CategoryConditional:     "\U0001f500",
	CategoryAssignment:      "\U0001f4dd",
	CategoryExpression:      "\U0001f4ca",
	CategoryIdiom:           "\U0001f4a0",
	CategoryUnknown:         "\U0001f4c4",
}

func (c CloneCategory) GetCategoryEmoji() string {
	if emoji, ok := categoryEmojis[c]; ok {
		return emoji
	}

	return "\U0001f4c4"
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
	Category      CloneCategory
	IsTest        bool
	Priority      ClonePriority
	Actionability CloneActionability
	Tokens        int
	Lines         int
	NodeType      string
	Suggestion    string
}

// ProcessedClone represents a single clone instance with extracted fragment data.
// Decouples printer output from syntax.Node internals.
type ProcessedClone struct {
	Filename       string
	LineStart      int
	LineEnd        int
	Fragment       []byte
	Size           int
	FileSize       int
	Classification CloneClassification
}

// ProcessedCloneGroup represents a group of duplicate code fragments.
type ProcessedCloneGroup struct {
	Hash   string
	Size   int
	Clones []ProcessedClone
}
