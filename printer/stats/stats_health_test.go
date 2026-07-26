package stats

import (
	"bytes"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/internal/testutil"
)

func TestHealthScoreCalculation(t *testing.T) {
	// New algorithm: totalScore = duplication*0.7 + complexity*0.2 + impact*0.1
	// Grades: A: <5, B: <10, C: <15, D: <25, F: >=25
	tests := []struct {
		name             string
		duplicationRatio float64
		complexityScore  float64
		impactScore      int
		expectedGrade    domain.HealthScore
	}{
		{"Perfect health", 0.0, 0.0, 0, domain.HealthScoreA},
		{"Excellent health", 2.0, 1.0, 500, domain.HealthScoreA}, // totalScore ≈ 1.85
		{"Good health", 5.0, 2.0, 1000, domain.HealthScoreA},     // totalScore ≈ 4.4
		{
			"Moderate health",
			8.0,
			3.0,
			2000,
			domain.HealthScoreB,
		}, // totalScore ≈ 7.0
		{"Poor health", 12.0, 4.0, 3000, domain.HealthScoreC},     // totalScore ≈ 11.3
		{"Critical health", 20.0, 5.0, 5000, domain.HealthScoreD}, // totalScore ≈ 18.5
		{
			"Extreme duplication",
			40.0,
			10.0,
			10000,
			domain.HealthScoreF,
		}, // totalScore = 40*0.7 + 20*0.2 + 10*0.1 = 33
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			sp := NewStats(&buf, mockReadFile("package main\nfunc main(){}"), 15).(*stats)

			// Set up stats data
			sp.statsData.TotalDuplicateLines = int(tt.duplicationRatio * 10) // Simulating
			sp.statsData.TotalEstimatedLines = 1000
			sp.statsData.DuplicationRatio = tt.duplicationRatio // Set directly for health calculation
			sp.statsData.ComplexityScore = tt.complexityScore
			sp.statsData.ImpactScore = tt.impactScore

			grade := sp.calculateHealthScore()

			if grade != tt.expectedGrade {
				t.Errorf(
					"calculateHealthScore() = %q, want %q for inputs (ratio=%.1f%%, complexity=%.2f, impact=%d)",
					grade,
					tt.expectedGrade,
					tt.duplicationRatio,
					tt.complexityScore,
					tt.impactScore,
				)
			}
		})
	}
}

type gradeRecommendationTestCase struct {
	name             string
	healthScore      domain.HealthScore
	totalCloneGroups int
	averageCloneSize int
	complexityScore  float64
	shouldContain    []string
	shouldNotContain []string
}

func TestPrintRecommendations(t *testing.T) {
	gradeTestCase := func(name string, healthScore domain.HealthScore, totalCloneGroups, averageCloneSize int, complexityScore float64, shouldContain, shouldNotContain []string) gradeRecommendationTestCase {
		return gradeRecommendationTestCase{
			name:             name,
			healthScore:      healthScore,
			totalCloneGroups: totalCloneGroups,
			averageCloneSize: averageCloneSize,
			complexityScore:  complexityScore,
			shouldContain:    shouldContain,
			shouldNotContain: shouldNotContain,
		}
	}

	tests := []gradeRecommendationTestCase{
		{
			name:             "Grade A recommendations",
			healthScore:      "A",
			totalCloneGroups: 1,
			averageCloneSize: 5,
			complexityScore:  1.0,
			shouldContain:    []string{"Excellent", "Keep up the good work"},
			shouldNotContain: []string{"action needed", "Critical"},
		},
		gradeTestCase("Grade C recommendations with metrics", domain.HealthScoreC, 15, 60, 3.5,
			[]string{"Moderate", "50+ lines", "15 clone groups"},
			[]string{"Excellent", "Critical"}),
		gradeTestCase("Grade F critical recommendations", domain.HealthScoreF, 30, 80, 6.0,
			[]string{"Critical", "immediate action", "Halt new feature"},
			[]string{"Excellent", "minor"}),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			sp := NewStats(&buf, mockReadFile("package main\nfunc main(){}"), 15).(*stats)

			// Set up stats data
			sp.statsData.HealthScore = tt.healthScore
			sp.statsData.TotalCloneGroups = tt.totalCloneGroups
			sp.statsData.AverageCloneSize = tt.averageCloneSize
			sp.statsData.ComplexityScore = tt.complexityScore
			sp.statsData.TotalFilesScanned = 10
			sp.statsData.TotalDuplicateLines = 500
			sp.statsData.TotalEstimatedLines = 1000

			// Call printRecommendations
			sp.printRecommendations()

			output := buf.String()

			// Check that all expected strings are present
			for _, expected := range tt.shouldContain {
				testutil.AssertStringContains(
					t,
					output,
					expected,
					"Recommendations output missing expected text: "+expected,
				)
			}

			// Check that unexpected strings are NOT present
			for _, notExpected := range tt.shouldNotContain {
				if strings.Contains(output, notExpected) {
					t.Errorf(
						"Recommendations output should not contain: %q\nGot: %s",
						notExpected,
						output,
					)
				}
			}

			// Verify next steps section is present
			testutil.AssertStringContains(
				t,
				output,
				"Next Steps:",
				"Recommendations should include 'Next Steps:' section",
			)
		})
	}
}

func TestSetFilterStats(t *testing.T) {
	tests := []struct {
		name          string
		filesFiltered int
		breakdown     map[string]int
		wantFiltered  int
		wantBreakdown map[string]int
	}{
		{
			name:          "no filters applied",
			filesFiltered: 0,
			breakdown:     nil,
			wantFiltered:  0,
			wantBreakdown: nil,
		},
		{
			name:          "templ files filtered",
			filesFiltered: 12,
			breakdown:     map[string]int{testTempl: 12},
			wantFiltered:  12,
			wantBreakdown: map[string]int{testTempl: 12},
		},
		{
			name:          "multiple filter types",
			filesFiltered: 45,
			breakdown:     map[string]int{testTempl: 12, "sqlc": 8, "vendor": 25},
			wantFiltered:  45,
			wantBreakdown: map[string]int{testTempl: 12, "sqlc": 8, "vendor": 25},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			sp := NewStats(w, mockReadFile(string(mockReadFileContent())), 15).(*stats)

			sp.SetFilterStats(tt.filesFiltered, tt.breakdown)

			data := sp.GetStatsView()

			if data.FilesFiltered != tt.wantFiltered {
				t.Errorf("FilesFiltered = %d, want %d", data.FilesFiltered, tt.wantFiltered)
			}

			if tt.wantBreakdown == nil {
				if data.FilterBreakdown != nil {
					t.Errorf("Expected nil FilterBreakdown, got %v", data.FilterBreakdown)
				}
			} else {
				if len(data.FilterBreakdown) != len(tt.wantBreakdown) {
					t.Errorf(
						"FilterBreakdown length = %d, want %d",
						len(data.FilterBreakdown),
						len(tt.wantBreakdown),
					)
				}

				for key, want := range tt.wantBreakdown {
					if got := data.FilterBreakdown[key]; got != want {
						t.Errorf("FilterBreakdown[%q] = %d, want %d", key, got, want)
					}
				}
			}
		})
	}
}

func TestFilterStatsInJSONOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	sp := NewStats(buf, mockReadFile(string(mockReadFileContent())), 15).(*stats)

	// Set up stats with filter information
	sp.SetFilesCount(100)
	sp.SetFilterStats(25, map[string]int{testTempl: 10, "sqlc": 8, "vendor": 7})
	sp.statsData.TotalCloneGroups = 5
	sp.statsData.TotalClones = 10
	sp.statsData.DetectionMethods = "art-dupl"

	// Print footer to generate output
	err := sp.PrintFooter()
	if err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}

	// Verify the data was stored correctly
	data := sp.GetStatsView()
	testutil.AssertFieldValue(t, data.FilesFiltered, 25, "FilesFiltered")

	if len(data.FilterBreakdown) != 3 {
		t.Errorf("FilterBreakdown length = %d, want 3", len(data.FilterBreakdown))
	}

	if data.FilterBreakdown[testTempl] != 10 {
		t.Errorf("FilterBreakdown[templ] = %d, want 10", data.FilterBreakdown[testTempl])
	}
}
