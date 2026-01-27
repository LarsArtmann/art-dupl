package domain_test

import (
	"encoding/json"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax"
)

var _ = Describe("Domain: Clone", func() {
	Context("When validating clones", func() {
		It("should accept valid clones", func() {
			clone := domain.Clone{
				Filename:   "test.go",
				StartLine:  10,
				EndLine:    20,
				StartPos:   100,
				EndPos:     200,
				Fragment:   "test code",
				Hash:       "abc123",
				Confidence: 0.95,
				Complexity: 5,
				Status:     domain.FileProcessingStateCompleted,
			}

			Expect(clone.IsValid()).To(Succeed())
		})

		It("should reject clones with invalid position", func() {
			clone := domain.Clone{
				Filename:  "test.go",
				StartLine: 10,
				EndLine:   20,
				StartPos:  200,
				EndPos:    100, // Invalid: end < start
			}

			Expect(clone.IsValid()).To(MatchError(ContainSubstring("end position must be > start position")))
		})

		It("should reject clones with invalid confidence", func() {
			clone := domain.Clone{
				Filename:   "test.go",
				Confidence: 1.5, // Invalid: > 1.0
			}

			Expect(clone.IsValid()).To(MatchError(ContainSubstring("confidence must be between 0 and 1")))
		})
	})

	Context("When converting syntax nodes", func() {
		It("should create valid clones from nodes", func() {
			node := &syntax.Node{
				Filename: "test.go",
				Pos:      50,
				End:      150,
			}

			fileContent := []byte("line 1\nline 2\nline 3\nline 4\nline 5\nfunc test() {}\nline 7\nline 8\nline 9\nline 10\nline 11\nline 12\nline 13\nline 14\nline 15\nline 16")

			clone := domain.NodeToClone(node, "test.go", fileContent)
			Expect(clone.Filename).To(BeEquivalentTo("test.go"))
			Expect(clone.StartPos).To(BeEquivalentTo(uint(50)))
			Expect(clone.EndPos).To(BeEquivalentTo(uint(150)))
			Expect(clone.Status).To(Equal(domain.FileProcessingStateCompleted))
			Expect(clone.IsValid()).To(Succeed())
		})
	})
})

var _ = Describe("Domain: CloneGroup", func() {
	Context("When validating clone groups", func() {
		It("should accept valid clone groups", func() {
			group := domain.CloneGroup{
				Hash: "abc123",
				Size: 100,
				Clones: []domain.Clone{
					{
						Filename:  "test1.go",
						StartLine: 10,
						EndLine:   20,
						Status:    domain.FileProcessingStateCompleted,
					},
					{
						Filename:  "test2.go",
						StartLine: 15,
						EndLine:   25,
						Status:    domain.FileProcessingStateCompleted,
					},
				},
				Severity: domain.CloneSeverityMedium,
				Status:   domain.FileProcessingStateCompleted,
			}

			Expect(group.IsValid()).To(Succeed())
		})

		It("should reject groups with empty clones", func() {
			group := domain.CloneGroup{
				Hash:     "abc123",
				Size:     100,
				Clones:   []domain.Clone{},
				Severity: domain.CloneSeverityMedium,
			}

			Expect(group.IsValid()).To(MatchError(ContainSubstring("must have at least one clone")))
		})

		It("should reject groups with invalid severity", func() {
			group := domain.CloneGroup{
				Hash: "abc123",
				Size: 100,
				Clones: []domain.Clone{
					{
						Filename:  "test.go",
						StartLine: 10,
						EndLine:   20,
						Status:    domain.FileProcessingStateCompleted,
					},
				},
				Severity: domain.CloneSeverity("invalid"),
			}

			Expect(group.IsValid()).To(MatchError(ContainSubstring("invalid clone severity")))
		})

		It("should propagate clone validation errors", func() {
			group := domain.CloneGroup{
				Hash: "abc123",
				Size: 100,
				Clones: []domain.Clone{
					{
						// Invalid: end line < start line
						Filename:  "test.go",
						StartLine: 20,
						EndLine:   10,
						Status:    domain.FileProcessingStateCompleted,
					},
				},
				Severity: domain.CloneSeverityMedium,
			}

			Expect(group.IsValid()).To(MatchError(And(
				ContainSubstring("clone 0"),
				ContainSubstring("end line must be >= start line"),
			)))
		})
	})
})

var _ = Describe("Domain: Analysis", func() {
	Context("When validating analysis", func() {
		It("should accept valid analysis", func() {
			analysis := domain.Analysis{
				State:     domain.DetectionStateCompleted,
				Mode:      domain.AnalysisModeFull,
				Threshold: 10,
				CreatedAt: time.Now().Format(time.RFC3339),
				CloneGroups: []domain.CloneGroup{
					{
						Hash: "abc123",
						Size: 100,
						Clones: []domain.Clone{
							{
								Filename:  "test.go",
								StartLine: 10,
								EndLine:   20,
								Status:    domain.FileProcessingStateCompleted,
							},
						},
						Severity: domain.CloneSeverityMedium,
						Status:   domain.FileProcessingStateCompleted,
					},
				},
				Stats: domain.AnalysisStats{
					FilesAnalyzed:    5,
					TotalClones:      1,
					TotalTokenSize:   100,
					ComplexityScore:  0.5,
					DuplicationRatio: 0.2,
					ProcessingTime:   1000,
				},
			}

			Expect(analysis.IsValid()).To(Succeed())
		})

		It("should reject analysis with zero threshold", func() {
			analysis := domain.Analysis{
				State:     domain.DetectionStateCompleted,
				Mode:      domain.AnalysisModeFull,
				Threshold: 0, // Invalid
				CreatedAt: time.Now().Format(time.RFC3339),
			}

			Expect(analysis.IsValid()).To(MatchError(ContainSubstring("threshold cannot be zero")))
		})
	})
})

var _ = Describe("Domain: CloneSeverity", func() {
	Context("When working with severity enums", func() {
		It("should validate all severity levels", func() {
			validSeverities := []domain.CloneSeverity{
				domain.CloneSeverityLow,
				domain.CloneSeverityMedium,
				domain.CloneSeverityHigh,
				domain.CloneSeverityCritical,
			}

			for _, severity := range validSeverities {
				Expect(severity.IsValid()).To(BeTrue(), fmt.Sprintf("Severity %s should be valid", severity))
			}
		})

		It("should reject invalid severity", func() {
			invalidSeverity := domain.CloneSeverity("invalid")
			Expect(invalidSeverity.IsValid()).To(BeFalse())
		})

		It("should marshal and unmarshal JSON correctly", func() {
			original := domain.CloneSeverityHigh
			data, err := json.Marshal(original)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal(`"high"`))

			var unmarshaled domain.CloneSeverity
			err = json.Unmarshal(data, &unmarshaled)
			Expect(err).ToNot(HaveOccurred())
			Expect(unmarshaled).To(Equal(original))
		})
	})
})

var _ = Describe("Domain: DetectionOptions", func() {
	Context("When validating detection options", func() {
		It("should accept valid options", func() {
			options := domain.DetectionOptions{
				Threshold:     10,
				Mode:          domain.AnalysisModeFull,
				IncludeVendor: false,
				Verbose:       false,
				Paths:         []string{"./src"},
				OutputFormat:  "json",
			}

			Expect(options.IsValid()).To(Succeed())
		})

		It("should reject options with zero threshold", func() {
			options := domain.DetectionOptions{
				Threshold:    0, // Invalid
				Mode:         domain.AnalysisModeFull,
				Paths:        []string{"./src"},
				OutputFormat: "json",
			}

			Expect(options.IsValid()).To(MatchError(ContainSubstring("threshold must be > 0")))
		})

		It("should reject options with empty paths", func() {
			options := domain.DetectionOptions{
				Threshold:    10,
				Mode:         domain.AnalysisModeFull,
				Paths:        []string{}, // Invalid
				OutputFormat: "json",
			}

			Expect(options.IsValid()).To(MatchError(ContainSubstring("at least one path must be specified")))
		})
	})
})

var _ = Describe("Domain: Business Logic", func() {
	Context("When calculating clone severity", func() {
		It("should calculate low severity for simple clones", func() {
			severity := domain.CalculateSeverity(20, 5)
			Expect(severity).To(Equal(domain.CloneSeverityLow))
		})

		It("should calculate medium severity for moderate clones", func() {
			severity := domain.CalculateSeverity(60, 15)
			Expect(severity).To(Equal(domain.CloneSeverityMedium))
		})

		It("should calculate high severity for complex clones", func() {
			severity := domain.CalculateSeverity(120, 25)
			Expect(severity).To(Equal(domain.CloneSeverityHigh))
		})

		It("should calculate critical severity for very complex clones", func() {
			severity := domain.CalculateSeverity(250, 60)
			Expect(severity).To(Equal(domain.CloneSeverityCritical))
		})
	})
})