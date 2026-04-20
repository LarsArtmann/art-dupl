package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
)

// =============================================================================
// DetectionMethods helpers
// =============================================================================

func TestDetectionMethods_IsEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		methods  DetectionMethods
		expected bool
	}{
		{"empty slice", DetectionMethods{}, true},
		{"single method", DetectionMethods{DetectionMethodArtDupl}, false},
		{"multiple methods", DetectionMethods{DetectionMethodHash, DetectionMethodArtDupl}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := tc.methods.IsEmpty()
			if got != tc.expected {
				t.Errorf("IsEmpty() = %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestDetectionMethods_IsHashOnly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		methods  DetectionMethods
		expected bool
	}{
		{"empty", DetectionMethods{}, false},
		{"hash only", DetectionMethods{DetectionMethodHash}, true},
		{"art-dupl only", DetectionMethods{DetectionMethodArtDupl}, false},
		{"both methods", DetectionMethods{DetectionMethodHash, DetectionMethodArtDupl}, false},
		{"hash + other", DetectionMethods{DetectionMethodHash, DetectionMethodTodos}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := tc.methods.IsHashOnly()
			if got != tc.expected {
				t.Errorf("IsHashOnly() = %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestDetectionMethods_IsDefault(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		methods  DetectionMethods
		expected bool
	}{
		{"empty", DetectionMethods{}, false},
		{"art-dupl only", DetectionMethods{DetectionMethodArtDupl}, true},
		{"hash only", DetectionMethods{DetectionMethodHash}, false},
		{"both", DetectionMethods{DetectionMethodArtDupl, DetectionMethodHash}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := tc.methods.IsDefault()
			if got != tc.expected {
				t.Errorf("IsDefault() = %v, want %v", got, tc.expected)
			}
		})
	}
}

// =============================================================================
// GetThresholdAsDomain / SetThresholdFromDomain
// =============================================================================

func TestConfig_GetThresholdAsDomain(t *testing.T) {
	t.Parallel()

	cfg := &Config{Threshold: 15}
	dom, err := cfg.GetThresholdAsDomain()
	if err != nil {
		t.Fatalf("GetThresholdAsDomain() error: %v", err)
	}

	if dom.Uint() != 15 {
		t.Errorf("domain threshold = %d, want 15", dom.Uint())
	}
}

func TestConfig_GetThresholdAsDomain_Invalid(t *testing.T) {
	t.Parallel()

	cfg := &Config{Threshold: 0}
	_, err := cfg.GetThresholdAsDomain()
	if err == nil {
		t.Error("expected error for threshold 0")
	}
}

func TestConfig_SetThresholdFromDomain(t *testing.T) {
	t.Parallel()

	cfg := &Config{}
	dom, err := domain.NewThreshold(42)
	if err != nil {
		t.Fatalf("NewThreshold() error: %v", err)
	}

	err = cfg.SetThresholdFromDomain(dom)
	if err != nil {
		t.Fatalf("SetThresholdFromDomain() error: %v", err)
	}

	if cfg.Threshold != 42 {
		t.Errorf("Threshold = %d, want 42", cfg.Threshold)
	}
}

// =============================================================================
// LoadOptionalConfig
// =============================================================================

func TestLoadOptionalConfig_Empty(t *testing.T) {
	t.Parallel()

	cfg, err := LoadOptionalConfig("")
	if err != nil {
		t.Fatalf("LoadOptionalConfig('') error: %v", err)
	}

	if cfg != nil {
		t.Errorf("LoadOptionalConfig('') returned %v, want nil", cfg)
	}
}

func TestLoadOptionalConfig_FileExists(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "opt.json")

	cfg := &Config{Threshold: 99}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(cfgFile, data, 0o644); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadOptionalConfig(cfgFile)
	if err != nil {
		t.Fatalf("LoadOptionalConfig(file) error: %v", err)
	}

	if loaded == nil {
		t.Fatal("expected non-nil config")
	}

	if loaded.Threshold != 99 {
		t.Errorf("Threshold = %d, want 99", loaded.Threshold)
	}
}

// =============================================================================
// LoadConfig error paths
// =============================================================================

func TestLoadConfig_StatError(t *testing.T) {
	t.Parallel()

	_, err := LoadConfig("/proc/fake/file.json")
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestLoadConfig_EmptyFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "empty.json")

	if err := os.WriteFile(cfgFile, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(cfgFile)
	if err != nil {
		t.Fatalf("LoadConfig(empty) error: %v", err)
	}

	if cfg.Threshold != 15 {
		t.Errorf("empty file should use defaults, Threshold = %d, want 15", cfg.Threshold)
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "bad.json")

	if err := os.WriteFile(cfgFile, []byte("{invalid json}"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadConfig(cfgFile)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// =============================================================================
// SaveConfig error path
// =============================================================================

func TestSaveConfig_DirectoryCreationFails(t *testing.T) {
	t.Parallel()

	cfg := &Config{Threshold: 50}
	badPath := "/proc/fake/subdir/config.json"

	err := SaveConfig(cfg, badPath)
	if err == nil {
		t.Error("expected error when directory creation fails")
	}
}

func TestSaveConfig_FileExists(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "existing.json")

	if err := os.WriteFile(cfgFile, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &Config{Threshold: 77}
	if err := SaveConfig(cfg, cfgFile); err != nil {
		t.Fatalf("SaveConfig error: %v", err)
	}

	loaded, err := LoadConfig(cfgFile)
	if err != nil {
		t.Fatalf("LoadConfig after save error: %v", err)
	}

	if loaded.Threshold != 77 {
		t.Errorf("Threshold = %d, want 77", loaded.Threshold)
	}
}

// =============================================================================
// Validation functions
// =============================================================================

func TestValidateDetectionMethods_Empty(t *testing.T) {
	t.Parallel()

	// Empty/nil slice: validation requires at least one method
	err := validateDetectionMethods(nil)
	if err == nil {
		t.Error("expected error for nil (no methods specified)")
	}

	err = validateDetectionMethods([]DetectionMethod{})
	if err == nil {
		t.Error("expected error for empty (no methods specified)")
	}
}

func TestValidateDetectionMethods_InvalidMethod(t *testing.T) {
	t.Parallel()

	err := validateDetectionMethods([]DetectionMethod{"invalid_method"})
	if err == nil {
		t.Error("expected error for invalid method")
	}
}

func TestValidateDetectionMethods_Valid(t *testing.T) {
	t.Parallel()

	err := validateDetectionMethods([]DetectionMethod{DetectionMethodHash, DetectionMethodArtDupl})
	if err != nil {
		t.Errorf("unexpected error for valid methods: %v", err)
	}
}

func TestValidateOnly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		only    string
		wantErr bool
	}{
		{"empty", "", false},
		{"go", "go", false},
		{"templ", "templ", false},
		{"invalid", "rust", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validateOnly(tc.only)
			if tc.wantErr && err == nil {
				t.Error("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateCacheFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		cacheDir string
		clear    bool
		incr     bool
		wantErr  bool
	}{
		{"no flags", "", false, false, false},
		{"incremental only", "", false, true, false},
		{"cache-dir with incremental", ".cache", false, true, false},
		{"clear-cache with incremental", "", true, true, false},
		{"cache-dir without incremental", ".cache", false, false, true},
		{"clear-cache without incremental", "", true, false, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validateCacheFlags(tc.cacheDir, tc.clear, tc.incr)
			if tc.wantErr && err == nil {
				t.Error("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// =============================================================================
// DetectionMethod enum
// =============================================================================

func TestDetectionMethod_String(t *testing.T) {
	t.Parallel()

	if DetectionMethodHash.String() != "hash" {
		t.Errorf("DetectionMethodHash.String() = %q, want %q", DetectionMethodHash.String(), "hash")
	}

	if DetectionMethodArtDupl.String() != "art-dupl" {
		t.Errorf("DetectionMethodArtDupl.String() = %q, want %q", DetectionMethodArtDupl.String(), "art-dupl")
	}
}

func TestDetectionMethod_IsValid(t *testing.T) {
	t.Parallel()

	if !DetectionMethodHash.IsValid() {
		t.Error("hash should be valid")
	}

	if !DetectionMethodArtDupl.IsValid() {
		t.Error("art-dupl should be valid")
	}

	if DetectionMethod("unknown").IsValid() {
		t.Error("unknown should be invalid")
	}
}

func TestDetectionMethod_MarshalUnmarshal(t *testing.T) {
	t.Parallel()

	// Only test methods in the valid map (hash, art-dupl)
	for _, method := range []DetectionMethod{DetectionMethodHash, DetectionMethodArtDupl} {
		data, err := method.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON(%s) error: %v", method, err)
		}

		var parsed DetectionMethod
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("UnmarshalJSON(%s) error: %v", method, err)
		}

		if parsed != method {
			t.Errorf("round-trip %s: got %s", method, parsed)
		}
	}
}

func TestDetectionMethod_UnmarshalJSON_Invalid(t *testing.T) {
	t.Parallel()

	var dm DetectionMethod
	err := dm.UnmarshalJSON([]byte(`"not_a_method"`))
	if err == nil {
		t.Error("expected error for invalid detection method")
	}

	// Value is NOT set when error is returned (pointer unchanged)
	if dm != "" {
		t.Errorf("dm = %q on error, want empty string", dm)
	}
}

func TestDetectionMethod_UnmarshalJSON_Default(t *testing.T) {
	t.Parallel()

	var dm DetectionMethod
	err := dm.UnmarshalJSON([]byte(`"invalid"`))
	if err == nil {
		t.Error("expected error for invalid detection method")
	}

	// Value remains empty on error
	if dm != "" {
		t.Errorf("dm = %q on error, want empty string", dm)
	}
}

func TestParseDetectionMethods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    []DetectionMethod
		wantErr bool
	}{
		{"empty", "", []DetectionMethod{DetectionMethodArtDupl}, false},
		{"hash", "hash", []DetectionMethod{DetectionMethodHash}, false},
		{"art-dupl", "art-dupl", []DetectionMethod{DetectionMethodArtDupl}, false},
		{"both comma-separated", "hash, art-dupl ", []DetectionMethod{DetectionMethodHash, DetectionMethodArtDupl}, false},
		{"with spaces", "hash , art-dupl", []DetectionMethod{DetectionMethodHash, DetectionMethodArtDupl}, false},
		{"with empty parts", "hash,,art-dupl", []DetectionMethod{DetectionMethodHash, DetectionMethodArtDupl}, false},
		{"invalid", "hash,invalid", nil, true},
		// todos and legacy are defined but NOT in valid map, so they fail
		{"invalid todos", "hash,todos", nil, true},
		{"invalid legacy", "hash,legacy", nil, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseDetectionMethods(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got) != len(tc.want) {
				t.Errorf("ParseDetectionMethods(%q) = %v, want %v", tc.input, got, tc.want)
				return
			}

			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("ParseDetectionMethods(%q)[%d] = %s, want %s", tc.input, i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestValidateDetectionMethods_Function(t *testing.T) {
	t.Parallel()

	// nil/empty is INVALID (validation requires at least one method)
	err := ValidateDetectionMethods(nil)
	if err == nil {
		t.Error("expected error for nil")
	}

	err = ValidateDetectionMethods([]DetectionMethod{})
	if err == nil {
		t.Error("expected error for empty")
	}

	err = ValidateDetectionMethods([]DetectionMethod{"bad"})
	if err == nil {
		t.Error("expected error for invalid method")
	}

	err = ValidateDetectionMethods([]DetectionMethod{DetectionMethodHash})
	if err != nil {
		t.Errorf("unexpected error for valid: %v", err)
	}
}

func TestDefaultDetectionMethod(t *testing.T) {
	t.Parallel()

	if DefaultDetectionMethod() != DetectionMethodArtDupl {
		t.Errorf("DefaultDetectionMethod() = %s, want %s", DefaultDetectionMethod(), DetectionMethodArtDupl)
	}
}

// =============================================================================
// DiffMode enum
// =============================================================================

func TestDiffMode_String(t *testing.T) {
	t.Parallel()

	if DiffModeSideBySide.String() != "side-by-side" {
		t.Errorf("DiffModeSideBySide.String() = %q, want %q", DiffModeSideBySide.String(), "side-by-side")
	}

	if DiffModeInline.String() != "inline" {
		t.Errorf("DiffModeInline.String() = %q, want %q", DiffModeInline.String(), "inline")
	}

	if DiffModeDisabled.String() != "disabled" {
		t.Errorf("DiffModeDisabled.String() = %q, want %q", DiffModeDisabled.String(), "disabled")
	}
}

func TestDiffMode_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		mode  DiffMode
		valid bool
	}{
		{DiffModeDisabled, true},
		{DiffModeSideBySide, true},
		{DiffModeInline, true},
		{DiffMode("unknown"), false},
		{DiffMode(""), false},
	}

	for _, tc := range tests {
		t.Run(string(tc.mode), func(t *testing.T) {
			t.Parallel()
			got := tc.mode.IsValid()
			if got != tc.valid {
				t.Errorf("IsValid(%s) = %v, want %v", tc.mode, got, tc.valid)
			}
		})
	}
}

func TestDiffMode_IsEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		mode    DiffMode
		enabled bool
	}{
		{DiffModeDisabled, false},
		{DiffModeSideBySide, true},
		{DiffModeInline, true},
		{DiffMode("invalid"), false},
	}

	for _, tc := range tests {
		t.Run(string(tc.mode), func(t *testing.T) {
			t.Parallel()
			got := tc.mode.IsEnabled()
			if got != tc.enabled {
				t.Errorf("IsEnabled(%s) = %v, want %v", tc.mode, got, tc.enabled)
			}
		})
	}
}

func TestDiffMode_MarshalUnmarshal(t *testing.T) {
	t.Parallel()

	for _, mode := range []DiffMode{DiffModeDisabled, DiffModeSideBySide, DiffModeInline} {
		data, err := mode.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON(%s) error: %v", mode, err)
		}

		var parsed DiffMode
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("UnmarshalJSON(%s) error: %v", mode, err)
		}

		if parsed != mode {
			t.Errorf("round-trip %s: got %s", mode, parsed)
		}
	}
}

func TestDiffMode_UnmarshalJSON_Invalid(t *testing.T) {
	t.Parallel()

	var dm DiffMode
	err := dm.UnmarshalJSON([]byte(`"bad_mode"`))
	if err == nil {
		t.Error("expected error for invalid diff mode")
	}

	// Value remains empty on error
	if dm != "" {
		t.Errorf("dm = %q on error, want empty string", dm)
	}
}

func TestDiffMode_UnmarshalJSON_Default(t *testing.T) {
	t.Parallel()

	var dm DiffMode
	err := dm.UnmarshalJSON([]byte(`"invalid"`))
	if err == nil {
		t.Error("expected error for invalid diff mode")
	}

	// Value remains empty on error
	if dm != "" {
		t.Errorf("dm = %q on error, want empty string", dm)
	}
}

func TestParseDiffMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input     string
		wantMode  DiffMode
		wantErr   bool
	}{
		{"", DiffModeDisabled, false},
		{"false", DiffModeDisabled, false},
		{"disabled", DiffModeDisabled, false},
		{"true", DiffModeSideBySide, false},
		{"side-by-side", DiffModeSideBySide, false},
		{"inline", DiffModeInline, false},
		{"invalid_value", DiffModeDisabled, true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			got, err := ParseDiffMode(tc.input)
			if tc.wantErr && err == nil {
				t.Error("expected error")
				return
			}

			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if got != tc.wantMode {
				t.Errorf("ParseDiffMode(%q) = %s, want %s", tc.input, got, tc.wantMode)
			}
		})
	}
}

func TestAllDiffModes(t *testing.T) {
	t.Parallel()

	modes := AllDiffModes()
	if len(modes) != 3 {
		t.Errorf("AllDiffModes() returned %d modes, want 3", len(modes))
	}

	for _, mode := range modes {
		if !mode.IsValid() {
			t.Errorf("AllDiffModes() contains invalid mode: %s", mode)
		}
	}
}

func TestDefaultDiffMode(t *testing.T) {
	t.Parallel()

	if DefaultDiffMode() != DiffModeDisabled {
		t.Errorf("DefaultDiffMode() = %s, want %s", DefaultDiffMode(), DiffModeDisabled)
	}
}

// =============================================================================
// SortCriteria enum
// =============================================================================

func TestSortCriteria_String(t *testing.T) {
	t.Parallel()

	if SortBySize.String() != "size" {
		t.Errorf("SortBySize.String() = %q, want %q", SortBySize.String(), "size")
	}
}

func TestSortCriteria_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		crit  SortCriteria
		valid bool
	}{
		{SortBySize, true},
		{SortByOccurrence, true},
		{SortByHash, true},
		{SortByTotalTokens, true},
		{SortCriteria("invalid"), false},
		{SortCriteria(""), false},
	}

	for _, tc := range tests {
		t.Run(string(tc.crit), func(t *testing.T) {
			t.Parallel()
			got := tc.crit.IsValid()
			if got != tc.valid {
				t.Errorf("IsValid(%s) = %v, want %v", tc.crit, got, tc.valid)
			}
		})
	}
}

func TestSortCriteria_MarshalUnmarshal(t *testing.T) {
	t.Parallel()

	for _, sc := range []SortCriteria{SortBySize, SortByOccurrence, SortByHash, SortByTotalTokens} {
		data, err := sc.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON(%s) error: %v", sc, err)
		}

		var parsed SortCriteria
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("UnmarshalJSON(%s) error: %v", sc, err)
		}

		if parsed != sc {
			t.Errorf("round-trip %s: got %s", sc, parsed)
		}
	}
}

func TestSortCriteria_UnmarshalJSON_Invalid(t *testing.T) {
	t.Parallel()

	var sc SortCriteria
	err := sc.UnmarshalJSON([]byte(`"bad_criteria"`))
	if err == nil {
		t.Error("expected error for invalid sort criteria")
	}

	if sc != "" {
		t.Errorf("sc = %q on error, want empty string", sc)
	}
}

func TestSortCriteria_UnmarshalJSON_Default(t *testing.T) {
	t.Parallel()

	var sc SortCriteria
	err := sc.UnmarshalJSON([]byte(`"invalid"`))
	if err == nil {
		t.Error("expected error for invalid sort criteria")
	}

	if sc != "" {
		t.Errorf("sc = %q on error, want empty string", sc)
	}
}

func TestAllSortCriteria(t *testing.T) {
	t.Parallel()

	criteria := AllSortCriteria()
	if len(criteria) != 4 {
		t.Errorf("AllSortCriteria() returned %d criteria, want 4", len(criteria))
	}

	for _, sc := range criteria {
		if !sc.IsValid() {
			t.Errorf("AllSortCriteria() contains invalid: %s", sc)
		}
	}
}

func TestDefaultSortCriteria(t *testing.T) {
	t.Parallel()

	if DefaultSortCriteria() != SortBySize {
		t.Errorf("DefaultSortCriteria() = %s, want %s", DefaultSortCriteria(), SortBySize)
	}
}

// =============================================================================
// OutputFormat enum
// =============================================================================

func TestOutputFormat_String(t *testing.T) {
	t.Parallel()

	if OutputFormatText.String() != "text" {
		t.Errorf("OutputFormatText.String() = %q, want %q", OutputFormatText.String(), "text")
	}
}

func TestOutputFormat_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		fmt   OutputFormat
		valid bool
	}{
		{OutputFormatText, true},
		{OutputFormatHTML, true},
		{OutputFormatJSON, true},
		{OutputFormatCSV, true},
		{OutputFormatPlumbing, true},
		{OutputFormatSimpleJSON, true},
		{OutputFormatSARIF, true},
		{OutputFormat("unknown"), false},
	}

	for _, tc := range tests {
		t.Run(string(tc.fmt), func(t *testing.T) {
			t.Parallel()
			got := tc.fmt.IsValid()
			if got != tc.valid {
				t.Errorf("IsValid(%s) = %v, want %v", tc.fmt, got, tc.valid)
			}
		})
	}
}

func TestOutputFormat_MarshalUnmarshal(t *testing.T) {
	t.Parallel()

	for _, fmt := range []OutputFormat{OutputFormatText, OutputFormatHTML, OutputFormatJSON, OutputFormatCSV, OutputFormatPlumbing, OutputFormatSimpleJSON, OutputFormatSARIF} {
		data, err := fmt.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON(%s) error: %v", fmt, err)
		}

		var parsed OutputFormat
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("UnmarshalJSON(%s) error: %v", fmt, err)
		}

		if parsed != fmt {
			t.Errorf("round-trip %s: got %s", fmt, parsed)
		}
	}
}

func TestOutputFormat_UnmarshalJSON_Invalid(t *testing.T) {
	t.Parallel()

	var of OutputFormat
	err := of.UnmarshalJSON([]byte(`"bad_format"`))
	if err == nil {
		t.Error("expected error for invalid output format")
	}

	if of != "" {
		t.Errorf("of = %q on error, want empty string", of)
	}
}

func TestOutputFormat_UnmarshalJSON_Default(t *testing.T) {
	t.Parallel()

	var of OutputFormat
	err := of.UnmarshalJSON([]byte(`"invalid"`))
	if err == nil {
		t.Error("expected error for invalid output format")
	}

	if of != "" {
		t.Errorf("of = %q on error, want empty string", of)
	}
}

func TestAllOutputFormats(t *testing.T) {
	t.Parallel()

	formats := AllOutputFormats()
	if len(formats) != 7 {
		t.Errorf("AllOutputFormats() returned %d formats, want 7", len(formats))
	}

	for _, fmt := range formats {
		if !fmt.IsValid() {
			t.Errorf("AllOutputFormats() contains invalid: %s", fmt)
		}
	}
}

func TestDefaultOutputFormat(t *testing.T) {
	t.Parallel()

	if DefaultOutputFormat() != OutputFormatText {
		t.Errorf("DefaultOutputFormat() = %s, want %s", DefaultOutputFormat(), OutputFormatText)
	}
}

// =============================================================================
// FileType enum
// =============================================================================

func TestFileType_String(t *testing.T) {
	t.Parallel()

	if FileTypeGo.String() != "go" {
		t.Errorf("FileTypeGo.String() = %q, want %q", FileTypeGo.String(), "go")
	}

	if FileTypeTempl.String() != "templ" {
		t.Errorf("FileTypeTempl.String() = %q, want %q", FileTypeTempl.String(), "templ")
	}
}

func TestFileType_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		ft    FileType
		valid bool
	}{
		{FileTypeGo, true},
		{FileTypeTempl, true},
		{FileTypeAll, true},
		{FileType("unknown"), false},
		{FileType("rust"), false},
	}

	for _, tc := range tests {
		t.Run(string(tc.ft), func(t *testing.T) {
			t.Parallel()
			got := tc.ft.IsValid()
			if got != tc.valid {
				t.Errorf("IsValid(%s) = %v, want %v", tc.ft, got, tc.valid)
			}
		})
	}
}

func TestFileType_MarshalJSON(t *testing.T) {
	t.Parallel()

	data, err := FileTypeGo.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON(FileTypeGo) error: %v", err)
	}

	if string(data) != `"go"` {
		t.Errorf("MarshalJSON(FileTypeGo) = %s, want %q", data, `"go"`)
	}

	data, err = FileTypeAll.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON(FileTypeAll) error: %v", err)
	}

	if string(data) != "null" {
		t.Errorf("MarshalJSON(FileTypeAll) = %s, want %q", data, "null")
	}
}

func TestFileType_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	for _, ft := range []FileType{FileTypeGo, FileTypeTempl} {
		data, _ := json.Marshal(string(ft))
		var parsed FileType
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("UnmarshalJSON(%s) error: %v", ft, err)
		}

		if parsed != ft {
			t.Errorf("round-trip %s: got %s", ft, parsed)
		}
	}
}

func TestFileType_UnmarshalJSON_Invalid(t *testing.T) {
	t.Parallel()

	var ft FileType
	err := ft.UnmarshalJSON([]byte(`"rust"`))
	if err == nil {
		t.Error("expected error for invalid file type")
	}

	if ft != "" {
		t.Errorf("ft = %q on error, want empty string", ft)
	}
}

func TestFileType_UnmarshalJSON_Default(t *testing.T) {
	t.Parallel()

	var ft FileType
	err := ft.UnmarshalJSON([]byte(`"invalid"`))
	if err == nil {
		t.Error("expected error for invalid file type")
	}

	if ft != "" {
		t.Errorf("ft = %q on error, want empty string", ft)
	}
}

func TestFileType_Matches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		ft      FileType
		path    string
		matches bool
	}{
		{FileTypeGo, "foo.go", true},
		{FileTypeGo, "bar.go", true},
		{FileTypeGo, "foo.templ", false},
		{FileTypeGo, "foo.go.txt", false},
		{FileTypeTempl, "comp.templ", true},
		{FileTypeTempl, "comp.html", false},
		{FileTypeAll, "anything.go", true},
		{FileTypeAll, "anything.templ", true},
		{FileTypeAll, "anything.anything", true},
		{FileType("unknown"), "foo.go", true},
	}

	for _, tc := range tests {
		t.Run(string(tc.ft)+"_"+tc.path, func(t *testing.T) {
			t.Parallel()
			got := tc.ft.Matches(tc.path)
			if got != tc.matches {
				t.Errorf("Matches(%s, %q) = %v, want %v", tc.ft, tc.path, got, tc.matches)
			}
		})
	}
}

func TestHasSuffix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		s      string
		suffix string
		want   bool
	}{
		{"foo.go", ".go", true},
		{"bar.templ", ".templ", true},
		{"foo.go", ".templ", false},
		{"foo", ".go", false},
		{"", ".go", false},
		{"foo.go", "", true},
		{"go", "go", true},
		{"foo.go.txt", ".go", false},
	}

	for _, tc := range tests {
		t.Run(tc.s+"_"+tc.suffix, func(t *testing.T) {
			t.Parallel()
			got := hasSuffix(tc.s, tc.suffix)
			if got != tc.want {
				t.Errorf("hasSuffix(%q, %q) = %v, want %v", tc.s, tc.suffix, got, tc.want)
			}
		})
	}
}

func TestParseFileType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input     string
		wantType  FileType
		wantErr   bool
	}{
		{"", FileTypeAll, false},
		{"go", FileTypeGo, false},
		{"templ", FileTypeTempl, false},
		{"rust", FileTypeAll, true},
		{"c++", FileTypeAll, true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			got, err := ParseFileType(tc.input)
			if tc.wantErr && err == nil {
				t.Error("expected error")
				return
			}

			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if got != tc.wantType {
				t.Errorf("ParseFileType(%q) = %s, want %s", tc.input, got, tc.wantType)
			}
		})
	}
}

// =============================================================================
// Enum helpers
// =============================================================================

func TestIsValidStringType(t *testing.T) {
	t.Parallel()

	validMap := map[string]bool{"valid": true, "also_valid": true}
	if !isValidStringType("valid", validMap) {
		t.Error("expected 'valid' to be valid")
	}

	if isValidStringType("invalid", validMap) {
		t.Error("expected 'invalid' to be invalid")
	}
}

func TestUnmarshalStringType_InvalidJSON(t *testing.T) {
	t.Parallel()

	isValidFn := func(v string) bool { return v == "valid" }
	_, err := unmarshalStringType([]byte("not json"), isValidFn, "default", "string type")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestUnmarshalStringType_InvalidValue(t *testing.T) {
	t.Parallel()

	isValidFn := func(v string) bool { return v == "valid" }
	got, err := unmarshalStringType([]byte(`"invalid"`), isValidFn, "default", "string type")
	if err == nil {
		t.Error("expected error for invalid value")
	}

	if got != "default" {
		t.Errorf("returned default = %q, want %q", got, "default")
	}
}

func TestUnmarshalStringTypeToPointer(t *testing.T) {
	t.Parallel()

	isValidFn := func(v string) bool { return v == "valid" }
	var target string
	err := unmarshalStringTypeToPointer([]byte(`"valid"`), isValidFn, "default", "string type", &target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if target != "valid" {
		t.Errorf("target = %q, want %q", target, "valid")
	}
}

func TestUnmarshalStringTypeToPointer_InvalidValue(t *testing.T) {
	t.Parallel()

	isValidFn := func(v string) bool { return v == "valid" }
	var target string
	err := unmarshalStringTypeToPointer([]byte(`"bad"`), isValidFn, "default", "string type", &target)
	if err == nil {
		t.Error("expected error for invalid value")
	}

	// Pointer is NOT set on error
	if target != "" {
		t.Errorf("target = %q on error, want empty string", target)
	}
}

func TestMarshalStringType_Invalid(t *testing.T) {
	t.Parallel()

	isValidFn := func(v string) bool { return v == "valid" }
	data, err := marshalStringType("invalid", isValidFn, "string type")
	if err != nil {
		t.Fatalf("marshalStringType error: %v", err)
	}

	if string(data) != "null" {
		t.Errorf("marshalStringType(invalid) = %s, want %q", data, "null")
	}
}

func TestMarshalStringType_Valid(t *testing.T) {
	t.Parallel()

	isValidFn := func(v string) bool { return v == "valid" }
	data, err := marshalStringType("valid", isValidFn, "string type")
	if err != nil {
		t.Fatalf("marshalStringType error: %v", err)
	}

	if string(data) != `"valid"` {
		t.Errorf("marshalStringType(valid) = %s, want %q", data, `"valid"`)
	}
}

// =============================================================================
// Errors
// =============================================================================

func TestErrInvalidFileType(t *testing.T) {
	t.Parallel()

	if !errors.Is(ErrInvalidFileType, ErrInvalidFileType) {
		t.Error("ErrInvalidFileType should be ErrInvalidFileType")
	}
}

func TestErrInvalidDetectionMethod(t *testing.T) {
	t.Parallel()

	if !errors.Is(ErrInvalidDetectionMethod, ErrInvalidDetectionMethod) {
		t.Error("ErrInvalidDetectionMethod should be ErrInvalidDetectionMethod")
	}
}
