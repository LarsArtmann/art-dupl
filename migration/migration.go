package migration

import (
	"fmt"
	"strings"
	"time"

	"github.com/LarsArtmann/art-dupl/adapter"
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/samber/mo"
)

// MigrationPath handles conversion between old and new type systems.
type MigrationPath struct {
	legacyPrinter printer.Printer
	processor     *domain.DetectionOptions
}

// NewMigrationPath creates a new migration path.
func NewMigrationPath(legacyPrinter printer.Printer, options domain.DetectionOptions) *MigrationPath {
	return &MigrationPath{
		legacyPrinter: legacyPrinter,
		processor:     &options,
	}
}

// FromSyntaxToDomain converts syntax nodes to domain types.
func (mp *MigrationPath) FromSyntaxToNodes(dups [][]*syntax.Node, threshold uint) domain.Analysis {
	// Convert each duplicate group to domain clone group
	var cloneGroups []domain.CloneGroup

	for i, nodeGroup := range dups {
		groupID := fmt.Sprintf("group-%d", i)
		cloneGroup := adapter.CloneGroupFromNodes(groupID, [][]*syntax.Node{nodeGroup})
		cloneGroups = append(cloneGroups, cloneGroup)
	}

	// Create analysis from clone groups
	analysis := adapter.CreateAnalysisFromClones(cloneGroups, threshold)

	return analysis
}

// FromPrinterClonesToDomain converts printer clones to domain clones
// This function needs to be reimplemented as printer.Clone is not exported.
func (mp *MigrationPath) FromPrinterClonesToDomain(printerClones []printer.Clone) []domain.Clone {
	// This function is temporarily disabled due to missing printer.Clone type
	var domainClones []domain.Clone

	// Implement proper conversion when printer.Clone becomes available
	/*
		for _, pc := range printerClones {
			dc := adapter.CloneToDomain(pc)
			domainClones = append(domainClones, dc)
		}
	*/

	return domainClones
}

// ValidateMigration checks if migration is valid.
func (mp *MigrationPath) ValidateMigration(analysis domain.Analysis) mo.Result[domain.Analysis] {
	if err := analysis.IsValid(); err != nil {
		return mo.Errf[domain.Analysis]("invalid analysis for migration: %v", err)
	}
	return mo.Ok(analysis)
}

// CreateMigrationReport generates a migration report.
func (mp *MigrationPath) CreateMigrationReport(before, after domain.Analysis) MigrationReport {
	return MigrationReport{
		MigrationID:     generateMigrationID(),
		CreatedAt:       time.Now().Format(time.RFC3339),
		BeforeState:     before,
		AfterState:      after,
		Differences:     mp.calculateDifferences(before, after),
		Validations:     mp.validateMigration(before, after),
		Recommendations: mp.generateRecommendations(before, after),
	}
}

// MigrationReport tracks migration status and changes.
type MigrationReport struct {
	MigrationID     string              `json:"migrationId"`
	CreatedAt       string              `json:"createdAt"`
	BeforeState     domain.Analysis     `json:"beforeState"`
	AfterState      domain.Analysis     `json:"afterState"`
	Differences     AnalysisDifferences `json:"differences"`
	Validations     []ValidationResult  `json:"validations"`
	Recommendations []string            `json:"recommendations"`
}

// AnalysisDifferences tracks changes between analyses.
type AnalysisDifferences struct {
	CloneGroupsAdded   int            `json:"cloneGroupsAdded"`
	CloneGroupsRemoved int            `json:"cloneGroupsRemoved"`
	ClonesAdded        int            `json:"clonesAdded"`
	ClonesRemoved      int            `json:"clonesRemoved"`
	SeverityChanges    map[string]int `json:"severityChanges"`
	ComplexityChanges  float64        `json:"complexityChanges"`
}

// ValidationResult represents migration validation result.
type ValidationResult struct {
	Check    string `json:"check"`
	Status   string `json:"status"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

// Helper functions

func (mp *MigrationPath) calculateDifferences(before, after domain.Analysis) AnalysisDifferences {
	return AnalysisDifferences{
		CloneGroupsAdded:   len(after.CloneGroups) - len(before.CloneGroups),
		CloneGroupsRemoved: len(before.CloneGroups) - len(after.CloneGroups),
		ClonesAdded:        mp.calculateCloneDifference(after, before),
		ClonesRemoved:      mp.calculateCloneDifference(before, after),
		SeverityChanges:    mp.calculateSeverityChanges(before, after),
		ComplexityChanges:  after.Stats.ComplexityScore - before.Stats.ComplexityScore,
	}
}

func (mp *MigrationPath) calculateCloneDifference(a, b domain.Analysis) int {
	totalA := 0
	totalB := 0

	for _, group := range a.CloneGroups {
		totalA += len(group.Clones)
	}

	for _, group := range b.CloneGroups {
		totalB += len(group.Clones)
	}

	return totalA - totalB
}

func (mp *MigrationPath) calculateSeverityChanges(before, after domain.Analysis) map[string]int {
	changes := make(map[string]int)

	beforeCounts := mp.countSeverities(before.CloneGroups)
	afterCounts := mp.countSeverities(after.CloneGroups)

	for severity := range beforeCounts {
		diff := afterCounts[severity] - beforeCounts[severity]
		if diff != 0 {
			changes[string(severity)] = diff
		}
	}

	return changes
}

func (mp *MigrationPath) countSeverities(groups []domain.CloneGroup) map[domain.CloneSeverity]int {
	counts := make(map[domain.CloneSeverity]int)

	for _, group := range groups {
		counts[group.Severity]++
	}

	return counts
}

func (mp *MigrationPath) validateMigration(before, after domain.Analysis) []ValidationResult {
	var validations []ValidationResult

	validations = append(validations, mp.validateAnalysisState(before, "before"))
	validations = append(validations, mp.validateAnalysisState(after, "after"))

	// Validate migration logic
	if before.Threshold != after.Threshold {
		validations = append(validations, ValidationResult{
			Check:    "threshold-preserved",
			Status:   "warning",
			Message:  "Threshold changed during migration",
			Severity: "warning",
		})
	}

	return validations
}

func (mp *MigrationPath) validateAnalysisState(analysis domain.Analysis, state string) ValidationResult {
	if err := analysis.IsValid(); err != nil {
		return ValidationResult{
			Check:    state + "-state-valid",
			Status:   "failed",
			Message:  fmt.Sprintf("%s state invalid: %v", capitalize(state), err),
			Severity: "error",
		}
	}
	return ValidationResult{
		Check:    state + "-state-valid",
		Status:   "passed",
		Message:  capitalize(state) + " state is valid",
		Severity: "info",
	}
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(string(s[0])) + s[1:]
}

func (mp *MigrationPath) generateRecommendations(before, after domain.Analysis) []string {
	var recommendations []string

	if after.Stats.DuplicationRatio > before.Stats.DuplicationRatio {
		recommendations = append(recommendations, "Consider increasing threshold to reduce false positives")
	}

	if after.Stats.ComplexityScore > before.Stats.ComplexityScore {
		recommendations = append(recommendations, "Review newly detected clones for refactoring opportunities")
	}

	if len(after.CloneGroups) > int(float64(len(before.CloneGroups))*1.2) {
		recommendations = append(recommendations, "Migration detected significantly more clones - verify accuracy")
	}

	return recommendations
}

func generateMigrationID() string {
	return fmt.Sprintf("migration-%d", time.Now().Unix())
}

// MigrateConfig handles configuration migration.
func MigrateConfig(oldConfig map[string]any) mo.Result[domain.DetectionOptions] {
	options := domain.DetectionOptions{}

	// Extract threshold
	if threshold, ok := oldConfig["threshold"].(float64); ok {
		options.Threshold = domain.Threshold(uint(threshold))
	} else {
		return mo.Errf[domain.DetectionOptions]("missing or invalid threshold in config")
	}

	// Extract paths
	if paths, ok := oldConfig["paths"].([]any); ok {
		for _, path := range paths {
			if str, ok := path.(string); ok {
				options.Paths = append(options.Paths, domain.Filepath(str))
			}
		}
	}

	if len(options.Paths) == 0 {
		return mo.Errf[domain.DetectionOptions]("no paths found in config")
	}

	// Set defaults
	options.Mode = domain.AnalysisModeFull
	options.IncludeVendor = false
	options.Verbose = false
	options.OutputFormat = "json"

	// Validate final options
	if err := options.IsValid(); err != nil {
		return mo.Errf[domain.DetectionOptions]("invalid migrated config: %v", err)
	}

	return mo.Ok(options)
}
