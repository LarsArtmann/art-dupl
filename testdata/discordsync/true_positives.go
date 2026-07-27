package discordsync

// True positives: genuinely harmful duplication that should be reported.
// These are multi-statement function bodies that are exact copies.
// They should NOT be suppressed by any actionability pattern.

// calculateScore — identical multi-statement logic copy-pasted.
// Not a guard clause, not error handling, not a single statement.
func calculateScoreA(events []Event) int {
	score := 0

	for _, event := range events {
		if event.Type == "critical" {
			score += event.Value * criticalMultiplier
		} else {
			score += event.Value
		}
	}

	return score
}

func calculateScoreB(events []Event) int {
	score := 0

	for _, event := range events {
		if event.Type == "critical" {
			score += event.Value * criticalMultiplier
		} else {
			score += event.Value
		}
	}

	return score
}

// processBatch — identical batch processing logic copy-pasted.
func processBatchA(items []Item) []Result {
	results := make([]Result, 0, len(items))

	for _, item := range items {
		processed := transformItem(item)
		results = append(results, Result{Value: processed})
	}

	return results
}

func processBatchB(items []Item) []Result {
	results := make([]Result, 0, len(items))

	for _, item := range items {
		processed := transformItem(item)
		results = append(results, Result{Value: processed})
	}

	return results
}
