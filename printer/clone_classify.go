package printer

import (
	"strings"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

const (
	suggestExtractUtility    = "Extract to shared utility function"
	suggestReviewExtract     = "Review and extract common logic"
	suggestComposition       = "Consider composition or shared base struct"
	suggestInterface         = "Extract common interface definition"
	suggestHandler           = "Extract handler logic to service layer"
	suggestLoopHelper        = "Extract loop body to helper function"
	suggestStrategy          = "Consider strategy pattern or early returns"
	suggestTestHelper        = "Extract test helper function or use table-driven tests"
	suggestSharedTestUtility = "Consider extracting to shared test utility"
	suggestIdiom             = "Structural idiom — typically not actionable"
	suggestTestDataPair      = "Test fixture files — expected structural similarity"
	suggestTableDrivenTest   = "Table-driven test body — framework pattern, not logic"
	suggestTestScaffolding   = "Test setup/assertion pattern — only data differs"
	suggestDataDominated     = "Data-dominated clone — struct literals, not logic"
	suggestSignatureOnly     = "Signature-only match — no extractable body"
	suggestRAIIDefer         = "RAII cleanup defer — idiomatic resource management"
	suggestErrorPropagation  = "Error propagation — standard Go error handling pattern"
	suggestInterfaceImpl     = "Interface method implementation — shared signature, different behavior"
	suggestDescribeTable     = "Ginkgo DescribeTable entry — framework-generated structural repetition"
	suggestBuilderCallback   = "Builder/callback chain — intentional fluent API design"
)

// idiomTokenThreshold is the maximum number of tokens for a clone to be
// classified as an "idiom" rather than a specific AST category. At this
// token count, clones are almost always structural artifacts (function
// signatures, import patterns, standard library calls) rather than
// meaningful duplication. Based on real-world feedback showing 100% false
// positive rate for 2-4 token clones at threshold 15.
const idiomTokenThreshold = 5

type (
	CloneCategory       = domain.CloneCategory
	ClonePriority       = domain.ClonePriority
	CloneClassification = domain.CloneClassification
)

func ClassifyClone(input domain.ClassificationInput) CloneClassification {
	if input.Tokens < idiomTokenThreshold {
		return idiomClassification(input)
	}

	category := nodeTypeToCategory(input.NodeType)
	isTest := isTestFile(input.Filename)
	priority := calculatePriority(category, isTest, input.Tokens, input.Lines)
	suggestion := getSuggestion(category, isTest, input.Tokens)

	return CloneClassification{
		Category:      category,
		IsTest:        isTest,
		Priority:      priority,
		Tokens:        input.Tokens,
		Lines:         input.Lines,
		NodeTypeName:  nodeTypeToString(input.NodeType),
		Suggestion:    suggestion,
		Actionability: domain.Actionable,
	}
}

func idiomClassification(input domain.ClassificationInput) CloneClassification {
	isTest := isTestFile(input.Filename)

	return CloneClassification{
		Category:      domain.CategoryIdiom,
		IsTest:        isTest,
		Priority:      domain.PriorityLow,
		Tokens:        input.Tokens,
		Lines:         input.Lines,
		NodeTypeName:  nodeTypeToString(input.NodeType),
		Suggestion:    suggestIdiom,
		Actionability: domain.NonActionable,
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
	return golang.TypeName(nodeType)
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
		domain.CategoryTestBoilerplate,
		domain.CategoryTestFixture,
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
	if category == domain.CategoryIdiom {
		return suggestIdiom
	}

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

// applyPatternLabel adjusts clone classification based on the AST-detected
// non-actionable pattern. This upgrades the category and suggestion to
// reflect the specific reason the clone is non-actionable, rather than
// relying solely on the first node's type.
func applyPatternLabel(cls domain.CloneClassification, label PatternLabel) domain.CloneClassification {
	switch label {
	case PatternTestData:
		cls.Category = domain.CategoryTestFixture
		cls.Suggestion = suggestTestDataPair
		cls.Priority = domain.PriorityLow
	case PatternTableDrivenTest:
		cls.Category = domain.CategoryTestBoilerplate
		cls.Suggestion = suggestTableDrivenTest
		cls.Priority = domain.PriorityLow
	case PatternTestScaffolding:
		cls.Category = domain.CategoryTestBoilerplate
		cls.Suggestion = suggestTestScaffolding
		cls.Priority = domain.PriorityLow
	case PatternDataDominated:
		cls.Suggestion = suggestDataDominated
		cls.Priority = domain.PriorityLow
	case PatternSignatureOnly:
		cls.Suggestion = suggestSignatureOnly
		cls.Priority = domain.PriorityLow
	case PatternRAIIDefer:
		cls.Suggestion = suggestRAIIDefer
		cls.Priority = domain.PriorityLow
	case PatternErrorPropagation:
		cls.Suggestion = suggestErrorPropagation
		cls.Priority = domain.PriorityLow
	case PatternInterfaceImpl:
		cls.Suggestion = suggestInterfaceImpl
		cls.Priority = domain.PriorityLow
	case PatternDescribeTable:
		cls.Category = domain.CategoryTestBoilerplate
		cls.Suggestion = suggestDescribeTable
		cls.Priority = domain.PriorityLow
	case PatternBuilderCallback:
		cls.Suggestion = suggestBuilderCallback
		cls.Priority = domain.PriorityLow
	case PatternNone:
		// No pattern detected — keep original classification
	}

	return cls
}
