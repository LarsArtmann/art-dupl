package actionability

import (
	"strings"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// productionSuggestions maps clone categories to specific refactoring
// suggestions for production code. Categories not listed here fall back
// to suggestReviewExtract in getSuggestion.
var productionSuggestions = map[domain.CloneCategory]string{ //nolint:gochecknoglobals // static lookup table
	domain.CategoryFunction:    suggestExtractUtility,
	domain.CategoryMethod:      suggestExtractUtility,
	domain.CategoryStruct:      suggestComposition,
	domain.CategoryInterface:   suggestInterface,
	domain.CategoryHandler:     suggestHandler,
	domain.CategoryLoop:        suggestLoopHelper,
	domain.CategoryConditional: suggestStrategy,
	domain.CategoryAssignment:  suggestAssignment,
	domain.CategoryExpression:  suggestExpression,
	domain.CategoryBlock:       suggestBlock,
	domain.CategoryCall:        suggestCall,
	domain.CategoryReturn:      suggestReturn,
	domain.CategoryDefer:       suggestDefer,
}

const (
	suggestExtractUtility    = "Extract to shared utility function"
	suggestReviewExtract     = "Review and extract common logic"
	suggestComposition       = "Consider composition or shared base struct"
	suggestInterface         = "Extract common interface definition"
	suggestHandler           = "Extract handler logic to service layer"
	suggestLoopHelper        = "Extract loop body to helper function"
	suggestStrategy          = "Consider strategy pattern or early returns"
	suggestAssignment        = "Extract shared assignment or initialization pattern"
	suggestExpression        = "Extract repeated expression or builder chain"
	suggestBlock             = "Extract shared statement block to a helper function"
	suggestCall              = "Extract repeated call sequence or parameterize arguments"
	suggestReturn            = "Consolidate return patterns or extract wrapper"
	suggestDefer             = "Extract shared cleanup or resource-management pattern"
	suggestTestHelper        = "Extract test helper function or use table-driven tests"
	suggestSharedTestUtility = "Consider extracting to shared test utility"
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

func ClassifyClone(input domain.ClassificationInput) domain.CloneClassification {
	category := nodeTypeToCategory(input.NodeType)
	isTest := strings.HasSuffix(input.Filename, "_test.go")
	priority := calculatePriority(category, isTest, input.Tokens, input.Lines)
	suggestion := getSuggestion(category, isTest, input.Tokens)

	return domain.CloneClassification{
		Category:      category,
		IsTest:        isTest,
		Priority:      priority,
		Tokens:        input.Tokens,
		Lines:         input.Lines,
		Suggestion:    suggestion,
		Actionability: domain.Actionable,
	}
}

func nodeTypeToCategory(nodeType int32) domain.CloneCategory {
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
	case golang.BlockStmt:
		return domain.CategoryBlock
	case golang.CallExpr:
		return domain.CategoryCall
	case golang.ReturnStmt:
		return domain.CategoryReturn
	case golang.DeferStmt, golang.GoStmt:
		return domain.CategoryDefer
	case golang.DeclStmt, golang.BinaryExpr:
		return domain.CategoryExpression
	default:
		return domain.CategoryUnknown
	}
}

func calculatePriority(category domain.CloneCategory, isTest bool, tokens, lines int) domain.ClonePriority {
	if isTest {
		return calculateTestPriority(tokens, lines)
	}

	return calculateProductionPriority(category, tokens, lines)
}

func calculateTestPriority(tokens, lines int) domain.ClonePriority {
	if tokens > 30 || lines > 30 {
		return domain.PriorityMedium
	}

	return domain.PriorityLow
}

func calculateProductionPriority(category domain.CloneCategory, tokens, lines int) domain.ClonePriority {
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
		domain.CategoryBlock,
		domain.CategoryCall,
		domain.CategoryReturn,
		domain.CategoryDefer,
		domain.CategoryUnknown:
		return otherPriority(tokens)
	default:
		return otherPriority(tokens)
	}
}

func functionPriority(tokens, lines int) domain.ClonePriority {
	if tokens > 15 || lines > 20 {
		return domain.PriorityCritical
	}

	if tokens > 8 || lines > 10 {
		return domain.PriorityHigh
	}

	return domain.PriorityMedium
}

func typePriority(tokens int) domain.ClonePriority {
	return controlFlowPriority(tokens)
}

func controlFlowPriority(tokens int) domain.ClonePriority {
	if tokens > 10 {
		return domain.PriorityHigh
	}

	return domain.PriorityMedium
}

func otherPriority(tokens int) domain.ClonePriority {
	if tokens > 15 {
		return domain.PriorityHigh
	}

	if tokens > 8 {
		return domain.PriorityMedium
	}

	return domain.PriorityLow
}

func getSuggestion(category domain.CloneCategory, isTest bool, tokens int) string {
	if isTest {
		if tokens > 15 {
			return suggestTestHelper
		}

		return suggestSharedTestUtility
	}

	if suggestion, ok := productionSuggestions[category]; ok {
		return suggestion
	}

	return suggestReviewExtract
}

// patternLabelConfig defines how ApplyPatternLabel adjusts a clone's
// classification when a non-actionable pattern is detected.
type patternLabelConfig struct {
	category    domain.CloneCategory
	setCategory bool
	suggestion  string
	priority    domain.ClonePriority
}

var patternLabelConfigs = map[PatternLabel]patternLabelConfig{ //nolint:gochecknoglobals // static pattern label lookup
	PatternTestData: {
		category:    domain.CategoryTestFixture,
		setCategory: true,
		suggestion:  suggestTestDataPair,
		priority:    domain.PriorityLow,
	},
	PatternTableDrivenTest: {
		category:    domain.CategoryTestBoilerplate,
		setCategory: true,
		suggestion:  suggestTableDrivenTest,
		priority:    domain.PriorityLow,
	},
	PatternTestScaffolding: {
		category:    domain.CategoryTestBoilerplate,
		setCategory: true,
		suggestion:  suggestTestScaffolding,
		priority:    domain.PriorityLow,
	},
	PatternDataDominated:    {suggestion: suggestDataDominated, priority: domain.PriorityLow},
	PatternSignatureOnly:    {suggestion: suggestSignatureOnly, priority: domain.PriorityLow},
	PatternRAIIDefer:        {suggestion: suggestRAIIDefer, priority: domain.PriorityLow},
	PatternErrorPropagation: {suggestion: suggestErrorPropagation, priority: domain.PriorityLow},
	PatternErrorWrapping: {
		suggestion: "error-wrapping idiom (if err != nil { return fmt.Errorf(...) })",
		priority:   domain.PriorityLow,
	},
	PatternAssertionChain: {
		category:    domain.CategoryTestBoilerplate,
		setCategory: true,
		suggestion:  "test assertion chain (Expect/Assert/Require/Should/Must)",
		priority:    domain.PriorityLow,
	},
	PatternCobraBoilerplate: {suggestion: "cobra.Command/fang.Command boilerplate", priority: domain.PriorityLow},
	PatternInterfaceImpl:    {suggestion: suggestInterfaceImpl, priority: domain.PriorityLow},
	PatternDescribeTable: {
		category:    domain.CategoryTestBoilerplate,
		setCategory: true,
		suggestion:  suggestDescribeTable,
		priority:    domain.PriorityLow,
	},
	PatternBuilderCallback: {suggestion: suggestBuilderCallback, priority: domain.PriorityLow},
	PatternAssignErrorCheck: {
		suggestion: "assign + error-check boilerplate (err := f(); if err != nil { return })",
		priority:   domain.PriorityLow,
	},
	PatternSingleCallExpr: {
		suggestion: "single function call with different arguments",
		priority:   domain.PriorityLow,
	},
	PatternSingleSimpleStmt: {
		suggestion: "single terminal statement (return, assignment, var declaration)",
		priority:   domain.PriorityLow,
	},
	PatternBoolAccumulatorInitializer: {
		suggestion: "consecutive boolean flag initializers (name := true/false) — independent accumulators, not shared logic",
		priority:   domain.PriorityLow,
	},
	PatternInterfaceAssertion: {
		suggestion: "compile-time interface assertion (var _ I = (*T)(nil)) — intentional idiom, one per type",
		priority:   domain.PriorityLow,
	},
	PatternTypeAliasBlock: {
		suggestion: "type-alias re-export block (type X = pkg.Y) — intentional re-export shims",
		priority:   domain.PriorityLow,
	},
	PatternGuardClause: {
		suggestion: "guard clause (if cond { return }) — boilerplate control flow",
		priority:   domain.PriorityLow,
	},
}

// ApplyPatternLabel adjusts clone classification based on the AST-detected
// non-actionable pattern. This upgrades the category and suggestion to
// reflect the specific reason the clone is non-actionable, rather than
// relying solely on the first node's type.
func ApplyPatternLabel(cls domain.CloneClassification, label PatternLabel) domain.CloneClassification {
	cfg, ok := patternLabelConfigs[label]
	if !ok {
		return cls
	}

	if cfg.setCategory {
		cls.Category = cfg.category
	}

	cls.Suggestion = cfg.suggestion
	cls.Priority = cfg.priority

	return cls
}
