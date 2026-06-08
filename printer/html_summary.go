package printer

import "github.com/LarsArtmann/art-dupl/domain"

// orderedCategories returns categories in a stable order for display.
func orderedCategories() []CloneCategory {
	return []CloneCategory{
		domain.CategoryFunction,
		domain.CategoryMethod,
		domain.CategoryHandler,
		domain.CategoryStruct,
		domain.CategoryInterface,
		domain.CategoryLoop,
		domain.CategoryConditional,
		domain.CategoryAssignment,
		domain.CategoryExpression,
		domain.CategoryIdiom,
		domain.CategoryTest,
		domain.CategoryUnknown,
	}
}

// orderedPriorities returns priorities in a stable order for display.
func orderedPriorities() []ClonePriority {
	return []ClonePriority{
		domain.PriorityCritical,
		domain.PriorityHigh,
		domain.PriorityMedium,
		domain.PriorityLow,
	}
}
