package migration_test

import (
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/migration"
	"github.com/LarsArtmann/art-dupl/syntax"
)

var _ = Describe("Migration Path", func() {
	Context("When migrating from syntax to domain", func() {
		It("should create valid domain analysis", func() {
			nodes := [][]*syntax.Node{
				{
					{
						Filename: "test1.go",
						Pos:      100,
						End:      200,
					},
				},
			}

			migrationPath := migration.NewMigrationPath(nil, domain.DetectionOptions{
				Threshold: 10,
				Paths:     []string{"./src"},
			})

			analysis := migrationPath.FromSyntaxToNodes(nodes, 10)

			Expect(analysis.IsValid()).To(Succeed())
			Expect(analysis.Threshold).To(Equal(uint(10)))
			Expect(analysis.State).To(Equal(domain.DetectionStateCompleted))
			Expect(analysis.CloneGroups).To(HaveLen(1))
			Expect(analysis.CloneGroups[0].Clones).To(HaveLen(1))
		})
	})

	Context("When migrating configuration", func() {
		It("should migrate valid old config", func() {
			oldConfig := map[string]any{
				"threshold": 15,
				"paths":     []string{"./src", "./lib"},
			}

			result := migration.MigrateConfig(oldConfig)
			Expect(result.IsOk()).To(BeTrue())

			options, err := result.Unwrap()
			Expect(err).ToNot(HaveOccurred())
			Expect(options.Threshold).To(Equal(uint(15)))
			Expect(options.Paths).To(Equal([]string{"./src", "./lib"}))
			Expect(options.IsValid()).To(Succeed())
		})

		It("should reject invalid config", func() {
			oldConfig := map[string]any{
				// Missing threshold
				"paths": []string{"./src"},
			}

			result := migration.MigrateConfig(oldConfig)
			Expect(result.IsErr()).To(BeTrue())
			Expect(result.Error.Error()).To(ContainSubstring("threshold"))
		})
	})

	Context("When creating migration reports", func() {
		It("should generate comprehensive migration report", func() {
			// Create clones with proper StringID initialization
			clone1 := domain.Clone{}
			clone1.SetFilename("test.go")

			clone2 := domain.Clone{}
			clone2.SetFilename("test.go")

			clone3 := domain.Clone{}
			clone3.SetFilename("test2.go")

			before := domain.Analysis{
				Threshold: 10,
				CloneGroups: []domain.CloneGroup{
					{
						Size: 100,
						Clones: []domain.Clone{
							clone1,
						},
						Severity: domain.CloneSeverityMedium,
					},
				},
				Stats: domain.AnalysisStats{
					FilesAnalyzed:   5,
					TotalClones:     1,
					ComplexityScore: 0.5,
				},
				CreatedAt: "2023-01-01T00:00:00Z",
			}

			after := domain.Analysis{
				Threshold: 10,
				CloneGroups: []domain.CloneGroup{
					{
						Size: 100,
						Clones: []domain.Clone{
							clone2,
						},
						Severity: domain.CloneSeverityHigh, // Changed
					},
					{
						Size: 50,
						Clones: []domain.Clone{
							clone3,
						},
						Severity: domain.CloneSeverityLow, // New
					},
				},
				Stats: domain.AnalysisStats{
					FilesAnalyzed:   6,
					TotalClones:     2,
					ComplexityScore: 0.7, // Increased
				},
				CreatedAt: "2023-01-01T01:00:00Z",
			}

			migrationPath := migration.NewMigrationPath(nil, domain.DetectionOptions{})
			report := migrationPath.CreateMigrationReport(before, after)

			Expect(report.MigrationID).NotTo(BeEmpty())
			Expect(report.Differences.CloneGroupsAdded).To(Equal(1))
			Expect(report.Differences.ClonesAdded).To(Equal(1))
			Expect(report.Differences.ComplexityChanges).To(Equal(0.2))
			Expect(report.Validations).ToNot(BeEmpty())
			Expect(report.Recommendations).ToNot(BeEmpty())
		})
	})

	Context("When validating migration", func() {
		It("should accept valid migration", func() {
			analysis := domain.Analysis{
				Threshold: 10,
				State:     domain.DetectionStateCompleted,
				CreatedAt: time.Now().Format(time.RFC3339),
			}

			migrationPath := migration.NewMigrationPath(nil, domain.DetectionOptions{})
			result := migrationPath.ValidateMigration(analysis)

			Expect(result.IsOk()).To(BeTrue())
			_, err := result.Unwrap()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should reject invalid migration", func() {
			analysis := domain.Analysis{
				// Missing required fields
				Threshold: 0, // Invalid
			}

			migrationPath := migration.NewMigrationPath(nil, domain.DetectionOptions{})
			result := migrationPath.ValidateMigration(analysis)

			Expect(result.IsErr()).To(BeTrue())
			Expect(result.Error.Error()).To(ContainSubstring("invalid analysis"))
		})
	})

	Context("When JSON marshaling migration reports", func() {
		It("should serialize migration reports correctly", func() {
			report := migration.MigrationReport{
				MigrationID: "test-migration",
				CreatedAt:   "2023-01-01T00:00:00Z",
				BeforeState: domain.Analysis{
					Threshold: 10,
				},
				AfterState: domain.Analysis{
					Threshold: 15,
				},
				Differences: migration.AnalysisDifferences{
					CloneGroupsAdded: 1,
					ClonesAdded:      2,
				},
				Validations: []migration.ValidationResult{
					{
						Check:    "test-check",
						Status:   "passed",
						Message:  "Test passed",
						Severity: "info",
					},
				},
				Recommendations: []string{"Test recommendation"},
			}

			data, err := json.Marshal(report)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(ContainSubstring("test-migration"))
			Expect(string(data)).To(ContainSubstring("CloneGroupsAdded"))
			Expect(string(data)).To(ContainSubstring("test-check"))

			// Test unmarshaling
			var unmarshaled migration.MigrationReport
			err = json.Unmarshal(data, &unmarshaled)
			Expect(err).ToNot(HaveOccurred())
			Expect(unmarshaled.MigrationID).To(Equal("test-migration"))
			Expect(unmarshaled.Differences.CloneGroupsAdded).To(Equal(1))
			Expect(unmarshaled.Validations).To(HaveLen(1))
		})
	})
})
