// Package printer provides functionality for formatting and outputting code clone reports.
package printer

import (
	"strings"

	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// CloneCategory represents the category of code that was duplicated.
type CloneCategory string

// Clone categories for classifying duplicate code.
const (
	CategoryFunction    CloneCategory = "function"    // Function declarations
	CategoryMethod      CloneCategory = "method"      // Method/function literals
	CategoryTest        CloneCategory = "test"        // Test code
	CategoryStruct      CloneCategory = "struct"      // Struct types
	CategoryInterface   CloneCategory = "interface"   // Interface types
	CategoryHandler     CloneCategory = "handler"     // HTTP handlers
	CategoryLoop        CloneCategory = "loop"        // Loop constructs
	CategoryConditional CloneCategory = "conditional" // Conditionals/switches
	CategoryAssignment  CloneCategory = "assignment"  // Variable assignments
	CategoryExpression  CloneCategory = "expression"  // Generic expressions
	CategoryUnknown     CloneCategory = "unknown"     // Unknown/other
)

// ClonePriority represents how important it is to address this clone.
type ClonePriority string

// Clone priority levels for actionability.
const (
	PriorityCritical ClonePriority = "critical" // Must fix - production code, large duplication
	PriorityHigh     ClonePriority = "high"     // Should fix - production code, medium duplication
	PriorityMedium   ClonePriority = "medium"   // Consider fixing - test code or small production
	PriorityLow      ClonePriority = "low"      // Optional - test helpers, tiny duplications
)

// CloneClassification provides metadata about a code clone for actionable reports.
type CloneClassification struct {
	Category   CloneCategory
	IsTest     bool
	Priority   ClonePriority
	Tokens     int
	Lines      int
	NodeType   string
	Suggestion string
}

// nodeTypeNames maps AST node types to human-readable names.
var nodeTypeNames = map[int32]string{
	golang.BadNode:        "BadNode",
	golang.File:           "File",
	golang.ArrayType:      "ArrayType",
	golang.AssignStmt:     "AssignStmt",
	golang.BasicLit:       "BasicLit",
	golang.BinaryExpr:     "BinaryExpr",
	golang.BlockStmt:      "BlockStmt",
	golang.BranchStmt:     "BranchStmt",
	golang.CallExpr:       "CallExpr",
	golang.CaseClause:     "CaseClause",
	golang.ChanType:       "ChanType",
	golang.CommClause:     "CommClause",
	golang.CompositeLit:   "CompositeLit",
	golang.DeclStmt:       "DeclStmt",
	golang.DeferStmt:      "DeferStmt",
	golang.Ellipsis:       "Ellipsis",
	golang.EmptyStmt:      "EmptyStmt",
	golang.ExprStmt:       "ExprStmt",
	golang.Field:          "Field",
	golang.FieldList:      "FieldList",
	golang.ForStmt:        "ForStmt",
	golang.FuncDecl:       "FuncDecl",
	golang.FuncLit:        "FuncLit",
	golang.FuncType:       "FuncType",
	golang.GenDecl:        "GenDecl",
	golang.GoStmt:         "GoStmt",
	golang.Ident:          "Ident",
	golang.IfStmt:         "IfStmt",
	golang.IncDecStmt:     "IncDecStmt",
	golang.InterfaceType:  "InterfaceType",
	golang.KeyValueExpr:   "KeyValueExpr",
	golang.LabeledStmt:    "LabeledStmt",
	golang.MapType:        "MapType",
	golang.ParenExpr:      "ParenExpr",
	golang.RangeStmt:      "RangeStmt",
	golang.ReturnStmt:     "ReturnStmt",
	golang.SelectStmt:     "SelectStmt",
	golang.SelectorExpr:   "SelectorExpr",
	golang.SendStmt:       "SendStmt",
	golang.SliceExpr:      "SliceExpr",
	golang.StarExpr:       "StarExpr",
	golang.StructType:     "StructType",
	golang.SwitchStmt:     "SwitchStmt",
	golang.TypeAssertExpr: "TypeAssertExpr",
	golang.TypeSpec:       "TypeSpec",
	golang.TypeSwitchStmt: "TypeSwitchStmt",
	golang.UnaryExpr:      "UnaryExpr",
	golang.ValueSpec:      "ValueSpec",
}

// ClassifyClone analyzes a clone and returns its classification.
func ClassifyClone(filename string, nodeType int32, tokens, lines int) CloneClassification {
	category := nodeTypeToCategory(nodeType)
	isTest := isTestFile(filename)
	priority := calculatePriority(category, isTest, tokens, lines)
	suggestion := getSuggestion(category, isTest, tokens)

	return CloneClassification{
		Category:   category,
		IsTest:     isTest,
		Priority:   priority,
		Tokens:     tokens,
		Lines:      lines,
		NodeType:   nodeTypeToString(nodeType),
		Suggestion: suggestion,
	}
}

// nodeTypeToCategory maps AST node types to human-readable categories.
func nodeTypeToCategory(nodeType int32) CloneCategory {
	switch nodeType {
	case golang.FuncDecl:
		return CategoryFunction
	case golang.FuncLit:
		return CategoryMethod
	case golang.StructType:
		return CategoryStruct
	case golang.InterfaceType:
		return CategoryInterface
	case golang.ForStmt, golang.RangeStmt:
		return CategoryLoop
	case golang.IfStmt, golang.SwitchStmt, golang.TypeSwitchStmt, golang.SelectStmt:
		return CategoryConditional
	case golang.AssignStmt, golang.ValueSpec:
		return CategoryAssignment
	case golang.GenDecl:
		return CategoryExpression
	default:
		return CategoryUnknown
	}
}

// nodeTypeToString returns a human-readable name for the node type.
func nodeTypeToString(nodeType int32) string {
	if name, ok := nodeTypeNames[nodeType]; ok {
		return name
	}

	return "Unknown"
}

// isTestFile detects if a file is a test file.
func isTestFile(filename string) bool {
	return strings.HasSuffix(filename, "_test.go")
}

// calculatePriority determines the priority level based on category, test status, and size.
func calculatePriority(category CloneCategory, isTest bool, tokens, lines int) ClonePriority {
	if isTest {
		return calculateTestPriority(tokens, lines)
	}

	return calculateProductionPriority(category, tokens, lines)
}

// calculateTestPriority returns priority for test files.
func calculateTestPriority(tokens, lines int) ClonePriority {
	if tokens > 100 || lines > 30 {
		return PriorityMedium
	}

	return PriorityLow
}

// calculateProductionPriority returns priority for production code.
func calculateProductionPriority(category CloneCategory, tokens, lines int) ClonePriority {
	switch category {
	case CategoryFunction, CategoryMethod:
		return functionPriority(tokens, lines)
	case CategoryStruct, CategoryInterface:
		return typePriority(tokens)
	case CategoryHandler:
		return PriorityHigh
	case CategoryLoop, CategoryConditional:
		return controlFlowPriority(tokens)
	case CategoryTest, CategoryAssignment, CategoryExpression, CategoryUnknown:
		return otherPriority(tokens)
	default:
		return otherPriority(tokens)
	}
}

// functionPriority calculates priority for function/method clones.
func functionPriority(tokens, lines int) ClonePriority {
	if tokens > 50 || lines > 20 {
		return PriorityCritical
	}

	if tokens > 25 || lines > 10 {
		return PriorityHigh
	}

	return PriorityMedium
}

// typePriority calculates priority for struct/interface clones.
func typePriority(tokens int) ClonePriority {
	if tokens > 30 {
		return PriorityHigh
	}

	return PriorityMedium
}

// controlFlowPriority calculates priority for loop/conditional clones.
func controlFlowPriority(tokens int) ClonePriority {
	if tokens > 30 {
		return PriorityHigh
	}

	return PriorityMedium
}

// otherPriority calculates priority for other category clones.
func otherPriority(tokens int) ClonePriority {
	if tokens > 50 {
		return PriorityHigh
	}

	if tokens > 25 {
		return PriorityMedium
	}

	return PriorityLow
}

// getSuggestion returns an actionable suggestion for the clone.
func getSuggestion(category CloneCategory, isTest bool, tokens int) string {
	if isTest {
		if tokens > 50 {
			return "Extract test helper function or use table-driven tests"
		}

		return "Consider extracting to shared test utility"
	}

	switch category {
	case CategoryFunction, CategoryMethod:
		return "Extract to shared utility function"
	case CategoryStruct:
		return "Consider composition or shared base struct"
	case CategoryInterface:
		return "Extract common interface definition"
	case CategoryHandler:
		return "Extract handler logic to service layer"
	case CategoryLoop:
		return "Extract loop body to helper function"
	case CategoryConditional:
		return "Consider strategy pattern or early returns"
	case CategoryTest, CategoryAssignment, CategoryExpression, CategoryUnknown:
		return "Review and extract common logic"
	default:
		return "Review and extract common logic"
	}
}

// GetPriorityColor returns a CSS color variable for the priority.
func (p ClonePriority) GetPriorityColor() string {
	switch p {
	case PriorityCritical:
		return "var(--error)"
	case PriorityHigh:
		return "var(--warning)"
	case PriorityMedium:
		return "var(--accent)"
	case PriorityLow:
		return "var(--success)"
	default:
		return "var(--text-secondary)"
	}
}

// GetPriorityEmoji returns an emoji indicator for the priority.
func (p ClonePriority) GetPriorityEmoji() string {
	switch p {
	case PriorityCritical:
		return "🔴"
	case PriorityHigh:
		return "🟠"
	case PriorityMedium:
		return "🟡"
	case PriorityLow:
		return "🟢"
	default:
		return "⚪"
	}
}

// GetCategoryEmoji returns an emoji indicator for the category.
func (c CloneCategory) GetCategoryEmoji() string {
	switch c {
	case CategoryFunction:
		return "⚡"
	case CategoryMethod:
		return "🔧"
	case CategoryTest:
		return "🧪"
	case CategoryStruct:
		return "📦"
	case CategoryInterface:
		return "🔌"
	case CategoryHandler:
		return "🎯"
	case CategoryLoop:
		return "🔄"
	case CategoryConditional:
		return "🔀"
	case CategoryAssignment:
		return "📝"
	case CategoryExpression:
		return "📊"
	case CategoryUnknown:
		return "📄"
	default:
		return "📄"
	}
}
