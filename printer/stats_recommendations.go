package printer

// printRecommendations prints actionable recommendations based on the health score.
func (p *stats) printRecommendations() {
	switch p.statsData.HealthScore {
	case "A":
		p.printSuccess("✓ Excellent code health! Duplication is minimal.")
		p.printLine("Keep up the good work. Maintain current practices.")
	case "B":
		p.printSuccess("✓ Good code health with minor duplication.")
		p.printLine("Consider extracting small duplicate patterns into shared functions.")
		p.printLine("Review the 'Top Files' section to identify problem areas.")
	case "C":
		p.printWarning("! Moderate code duplication detected.")
		p.printLine("Prioritize refactoring duplicate code blocks:")
		p.printLine("  1. Focus on large clones (50+ lines) first")
		p.printLine("  2. Create shared utility functions or base classes")
		p.printLine("  3. Consider domain-driven design patterns")
	case "D":
		p.printWarning("⚠ High code duplication - action needed.")
		p.printLine("Immediate actions recommended:")
		p.printLine("  1. Extract all medium/large duplicate blocks (>20 lines)")
		p.printLine("  2. Implement shared libraries or services")
		p.printLine("  3. Establish code review guidelines to prevent new duplication")
		p.printLine("  4. Consider architectural changes (e.g., introduce new abstractions)")
	case "F":
		p.printError("✗ Critical code duplication - immediate action required!")
		p.printLine("Urgent steps to take:")
		p.printLine("  1. Prioritize ALL duplicate code extraction immediately")
		p.printLine("  2. Halt new feature development until duplication is reduced")
		p.printLine("  3. Create comprehensive refactoring plan")
		p.printLine("  4. Invest in architectural review and design patterns")
		p.printLine("  5. Consider team training on DRY principles")
	default:
		p.printLine("No recommendations available.")
	}

	// Additional recommendations based on specific metrics
	p.printSection("Next Steps:")

	if p.statsData.TotalCloneGroups > 10 {
		p.printBullet(
			"You have %d clone groups - focus on the largest ones first",
			p.statsData.TotalCloneGroups,
		)
	}

	if p.statsData.AverageCloneSize > 50 {
		p.printBullet(
			"Average clone size is %d lines - prioritize extracting large blocks",
			p.statsData.AverageCloneSize,
		)
	}

	if p.statsData.ComplexityScore > 3.0 {
		p.printBullet(
			"Complexity score of %.2f suggests multiple clones per group - consider patterns",
			p.statsData.ComplexityScore,
		)
	}

	p.printBullet("Run with --threshold 50 to focus on large duplications only")
	p.printBullet("Use --format json for machine-readable output")
	p.printBullet("Integrate into CI/CD pipeline for continuous monitoring")
}
