package printer

import (
	"strings"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

type (
	CloneCategory       = domain.CloneCategory
	ClonePriority       = domain.ClonePriority
	CloneClassification = domain.CloneClassification
)

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

func nodeTypeToCategory(nodeType int32) CloneCategory {
	switch nodeType {
	case golang.FuncDecl:
		return domain.CategoryFunction
	case golang.FuncLit:
		return domain.CategoryMethod
	case golang.StructType:
		return domain.CategoryStruct
	case golang.InterfaceType:
		return domain.CategoryInterface
	case golang.ForStmt, golang.RangeStmt:
		return domain.CategoryLoop
	case golang.IfStmt, golang.SwitchStmt, golang.TypeSwitchStmt, golang.SelectStmt:
		return domain.CategoryConditional
	case golang.AssignStmt, golang.ValueSpec:
		return domain.CategoryAssignment
	case golang.GenDecl:
		return domain.CategoryExpression
	default:
		return domain.CategoryUnknown
	}
}

func nodeTypeToString(nodeType int32) string {
	if name, ok := nodeTypeNames[nodeType]; ok {
		return name
	}

	return "Unknown"
}

func isTestFile(filename string) bool {
	return strings.HasSuffix(filename, "_test.go")
}

func calculatePriority(category CloneCategory, isTest bool, tokens, lines int) ClonePriority {
	if isTest {
		return calculateTestPriority(tokens, lines)
	}

	return calculateProductionPriority(category, tokens, lines)
}

func calculateTestPriority(tokens, lines int) ClonePriority {
	if tokens > 100 || lines > 30 {
		return domain.PriorityMedium
	}

	return domain.PriorityLow
}

func calculateProductionPriority(category CloneCategory, tokens, lines int) ClonePriority {
	switch category {
	case domain.CategoryFunction, domain.CategoryMethod:
		return functionPriority(tokens, lines)
	case domain.CategoryStruct, domain.CategoryInterface:
		return typePriority(tokens)
	case domain.CategoryHandler:
		return domain.PriorityHigh
	case domain.CategoryLoop, domain.CategoryConditional:
		return controlFlowPriority(tokens)
	case domain.CategoryTest,
		domain.CategoryAssignment,
		domain.CategoryExpression,
		domain.CategoryUnknown:
		return otherPriority(tokens)
	default:
		return otherPriority(tokens)
	}
}

func functionPriority(tokens, lines int) ClonePriority {
	if tokens > 50 || lines > 20 {
		return domain.PriorityCritical
	}

	if tokens > 25 || lines > 10 {
		return domain.PriorityHigh
	}

	return domain.PriorityMedium
}

func typePriority(tokens int) ClonePriority {
	if tokens > 30 {
		return domain.PriorityHigh
	}

	return domain.PriorityMedium
}

func controlFlowPriority(tokens int) ClonePriority {
	if tokens > 30 {
		return domain.PriorityHigh
	}

	return domain.PriorityMedium
}

func otherPriority(tokens int) ClonePriority {
	if tokens > 50 {
		return domain.PriorityHigh
	}

	if tokens > 25 {
		return domain.PriorityMedium
	}

	return domain.PriorityLow
}

func getSuggestion(category CloneCategory, isTest bool, tokens int) string {
	if isTest {
		if tokens > 50 {
			return suggestTestHelper
		}

		return suggestSharedTestUtility
	}

	switch category {
	case domain.CategoryFunction, domain.CategoryMethod:
		return suggestExtractUtility
	case domain.CategoryStruct:
		return suggestComposition
	case domain.CategoryInterface:
		return suggestInterface
	case domain.CategoryHandler:
		return suggestHandler
	case domain.CategoryLoop:
		return suggestLoopHelper
	case domain.CategoryConditional:
		return suggestStrategy
	case domain.CategoryTest,
		domain.CategoryAssignment,
		domain.CategoryExpression,
		domain.CategoryUnknown:
		return suggestReviewExtract
	default:
		return suggestReviewExtract
	}
}
