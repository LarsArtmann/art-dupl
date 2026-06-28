package cmd

import (
	"slices"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/spf13/cobra"
)

func TestParseIncludeGeneratedCategories(t *testing.T) {
	tests := []struct {
		name    string
		values  []string
		want    []string
		wantErr bool
	}{
		{
			name:   "single value",
			values: []string{generatedCategorySQLC},
			want:   []string{generatedCategorySQLC},
		},
		{
			name:   "multiple values",
			values: []string{generatedCategorySQLC, generatedCategoryTempl},
			want:   []string{generatedCategorySQLC, generatedCategoryTempl},
		},
		{
			name:   "comma separated",
			values: []string{generatedCategorySQLC + "," + generatedCategoryTempl + "," + generatedCategoryProtobuf},
			want:   []string{generatedCategorySQLC, generatedCategoryTempl, generatedCategoryProtobuf},
		},
		{
			name:   "mixed repeated and comma separated",
			values: []string{generatedCategorySQLC + "," + generatedCategoryTempl, generatedCategoryMockgen},
			want:   []string{generatedCategorySQLC, generatedCategoryTempl, generatedCategoryMockgen},
		},
		{
			name:   "whitespace and case normalized",
			values: []string{"  SQLC , TEMPL "},
			want:   []string{generatedCategorySQLC, generatedCategoryTempl},
		},
		{
			name:   "all category",
			values: []string{generatedCategoryAll},
			want:   []string{generatedCategoryAll},
		},
		{
			name:    "invalid category",
			values:  []string{"unknown"},
			wantErr: true,
		},
		{
			name:    "mixed valid and invalid",
			values:  []string{generatedCategorySQLC, "bad"},
			wantErr: true,
		},
		{
			name:   "empty values skipped",
			values: []string{"", generatedCategorySQLC, ""},
			want:   []string{generatedCategorySQLC},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseIncludeGeneratedCategories(tt.values)

			if (err != nil) != tt.wantErr {
				t.Fatalf("parseIncludeGeneratedCategories(%v) error = %v, wantErr %v", tt.values, err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if !slices.Equal(got, tt.want) {
				t.Errorf("parseIncludeGeneratedCategories(%v) = %v, want %v", tt.values, got, tt.want)
			}
		})
	}
}

func TestApplyIncludeGeneratedFlag(t *testing.T) {
	flagWithDash := "--" + flagIncludeGenerated

	tests := []struct {
		name     string
		args     []string
		expected config.Config
		wantErr  bool
	}{
		{
			name: "single category",
			args: []string{flagWithDash, generatedCategorySQLC},
			expected: config.Config{
				IncludeSQLC: true,
			},
		},
		{
			name: "multiple categories",
			args: []string{
				flagWithDash, generatedCategorySQLC,
				flagWithDash, generatedCategoryTempl + "," + generatedCategoryProtobuf,
			},
			expected: config.Config{
				IncludeSQLC:     true,
				IncludeTempl:    true,
				IncludeProtobuf: true,
			},
		},
		{
			name: "all category",
			args: []string{flagWithDash, generatedCategoryAll},
			expected: config.Config{
				IncludeSQLC:     true,
				IncludeTempl:    true,
				IncludeProtobuf: true,
				IncludeMockgen:  true,
				IncludeStringer: true,
				IncludeGeneric:  true,
			},
		},
		{
			name:    "invalid category",
			args:    []string{flagWithDash, "bad"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			addSharedFlags(cmd)

			cmd.SetArgs(tt.args)

			if err := cmd.ParseFlags(tt.args); err != nil {
				t.Fatalf("ParseFlags(%v) error: %v", tt.args, err)
			}

			cfg := &config.Config{}
			err := applyIncludeGeneratedFlag(cmd, cfg)

			if (err != nil) != tt.wantErr {
				t.Fatalf("applyIncludeGeneratedFlag() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			assertIncludeGeneratedConfig(t, cfg, tt.expected)
		})
	}
}

func TestDeprecatedIncludeFlagsStillSetConfig(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	addSharedFlags(cmd)

	args := []string{
		"--include-sqlc",
		"--include-templ",
		"--include-protobuf",
		"--include-mockgen",
		"--include-stringer",
		"--include-generic",
	}
	cmd.SetArgs(args)

	if err := cmd.ParseFlags(args); err != nil {
		t.Fatalf("ParseFlags error: %v", err)
	}

	cfg, err := buildCLIConfig(cmd, nil)
	if err != nil {
		t.Fatalf("buildCLIConfig error: %v", err)
	}

	assertIncludeGeneratedConfig(t, cfg, config.Config{
		IncludeSQLC:     true,
		IncludeTempl:    true,
		IncludeProtobuf: true,
		IncludeMockgen:  true,
		IncludeStringer: true,
		IncludeGeneric:  true,
	})
}

func assertIncludeGeneratedConfig(t *testing.T, got *config.Config, want config.Config) {
	t.Helper()

	if got.IncludeSQLC != want.IncludeSQLC {
		t.Errorf("IncludeSQLC = %v, want %v", got.IncludeSQLC, want.IncludeSQLC)
	}

	if got.IncludeTempl != want.IncludeTempl {
		t.Errorf("IncludeTempl = %v, want %v", got.IncludeTempl, want.IncludeTempl)
	}

	if got.IncludeProtobuf != want.IncludeProtobuf {
		t.Errorf("IncludeProtobuf = %v, want %v", got.IncludeProtobuf, want.IncludeProtobuf)
	}

	if got.IncludeMockgen != want.IncludeMockgen {
		t.Errorf("IncludeMockgen = %v, want %v", got.IncludeMockgen, want.IncludeMockgen)
	}

	if got.IncludeStringer != want.IncludeStringer {
		t.Errorf("IncludeStringer = %v, want %v", got.IncludeStringer, want.IncludeStringer)
	}

	if got.IncludeGeneric != want.IncludeGeneric {
		t.Errorf("IncludeGeneric = %v, want %v", got.IncludeGeneric, want.IncludeGeneric)
	}
}
