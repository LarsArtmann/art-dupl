package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
)

// createTempDir creates a temporary directory for testing and returns a cleanup function.
func createTempDir(t *testing.T) (string, func()) {
	t.Helper()

	tmpDir := t.TempDir()

	return tmpDir, func() {
		err := os.RemoveAll(tmpDir)
		if err != nil {
			t.Logf("Failed to remove temp dir: %v", err)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()

	testutil.AssertConfigField(t, "Threshold", cfg.Threshold, 15)
	testutil.AssertConfigField(t, "IncludeVendor", cfg.IncludeVendor, false)
	testutil.AssertConfigField(t, "OutputFormat", cfg.OutputFormat, "text")

	if len(cfg.Paths) != 1 || cfg.Paths[0] != "." {
		t.Errorf("Expected default paths [\".\"], got %v", cfg.Paths)
	}
}

func TestLoadConfig(t *testing.T) {
	tmpDir, cleanup := createTempDir(t)
	defer cleanup()

	configFile := filepath.Join(tmpDir, "test-config.json")
	configContent := `{
		"threshold": 50,
		"includeVendor": true,
		"outputFormat": "json",
		"verbose": true,
		"paths": ["./src", "./lib"],
		"ignoreFiles": ["*_test.go", "mock_*.go"],
		"maxChildrenSerial": 20000
	}`

	err := os.WriteFile(configFile, []byte(configContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	config, err := LoadConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	testutil.AssertConfigField(t, "Threshold", config.Threshold, 50)
	testutil.AssertConfigField(t, "IncludeVendor", config.IncludeVendor, true)
	testutil.AssertConfigField(t, "OutputFormat", config.OutputFormat, "json")
	testutil.AssertConfigField(t, "Verbose", config.Verbose, true)
	testutil.AssertConfigField(t, "MaxChildrenSerial", config.MaxChildrenSerial, 20000)

	if len(config.Paths) != 2 || config.Paths[0] != "./src" || config.Paths[1] != "./lib" {
		t.Errorf("Expected paths [\"./src\", \"./lib\"], got %v", config.Paths)
	}

	if len(config.IgnoreFiles) != 2 || config.IgnoreFiles[0] != "*_test.go" ||
		config.IgnoreFiles[1] != "mock_*.go" {
		t.Errorf("Expected ignoreFiles [\"*_test.go\", \"mock_*.go\"], got %v", config.IgnoreFiles)
	}
}

func TestLoadConfigNotFound(t *testing.T) {
	t.Parallel()

	_, err := LoadConfig("nonexistent-config.json")
	if err == nil {
		t.Error("Expected error for nonexistent config file")
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir, cleanup := createTempDir(t)
	defer cleanup()

	configFile := filepath.Join(tmpDir, "saved-config.json")
	config := &Config{
		Threshold:     100,
		IncludeVendor: true,
		OutputFormat:  "html",
		Verbose:       true,
		Paths:         []string{"./test"},
	}

	err := SaveConfig(config, configFile)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Load and verify content
	loaded, err := LoadConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loaded.Threshold != config.Threshold {
		t.Errorf("Saved threshold %d, got %d", config.Threshold, loaded.Threshold)
	}

	if loaded.IncludeVendor != config.IncludeVendor {
		t.Errorf("Saved IncludeVendor %v, got %v", config.IncludeVendor, loaded.IncludeVendor)
	}

	if loaded.OutputFormat != config.OutputFormat {
		t.Errorf("Saved OutputFormat %s, got %s", config.OutputFormat, loaded.OutputFormat)
	}
}

// newInvalidConfigTestCase creates a test case for invalid configuration validation.
func newInvalidConfigTestCase(
	name string,
	threshold int,
	outputFormat OutputFormat,
	maxChildrenSerial int,
) struct {
	name    string
	config  *Config
	isValid bool
} {
	return struct {
		name    string
		config  *Config
		isValid bool
	}{
		name: name,
		config: &Config{
			Threshold:         threshold,
			OutputFormat:      outputFormat,
			MaxChildrenSerial: maxChildrenSerial,
		},
		isValid: false,
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		isValid bool
	}{
		{
			name: "Valid config",
			config: &Config{
				Threshold:         15,
				OutputFormat:      "text",
				MaxChildrenSerial: 10000,
				DetectionMethods:  DetectionMethods{DetectionMethodArtDupl},
			},
			isValid: true,
		},
		newInvalidConfigTestCase("Invalid threshold - too low", 0, OutputFormat("text"), 10000),
		newInvalidConfigTestCase("Invalid threshold - too high", 1001, OutputFormat("text"), 10000),
		newInvalidConfigTestCase("Invalid output format", 15, OutputFormat("xml"), 10000),
		newInvalidConfigTestCase(
			"Invalid maxChildrenSerial - too low",
			15,
			OutputFormat("text"),
			999,
		),
		newInvalidConfigTestCase(
			"Invalid maxChildrenSerial - too high",
			15,
			OutputFormat("text"),
			100001,
		),
		{
			name: "Cache flags require incremental - cache-dir without incremental",
			config: &Config{
				Threshold:         15,
				OutputFormat:      "text",
				MaxChildrenSerial: 10000,
				DetectionMethods:  DetectionMethods{DetectionMethodArtDupl},
				CacheDir:          ".cache/art-dupl",
			},
			isValid: false,
		},
		{
			name: "Cache flags require incremental - clear-cache without incremental",
			config: &Config{
				Threshold:         15,
				OutputFormat:      "text",
				MaxChildrenSerial: 10000,
				DetectionMethods:  DetectionMethods{DetectionMethodArtDupl},
				ClearCache:        true,
			},
			isValid: false,
		},
		{
			name: "Cache flags valid with incremental",
			config: &Config{
				Threshold:         15,
				OutputFormat:      "text",
				MaxChildrenSerial: 10000,
				DetectionMethods:  DetectionMethods{DetectionMethodArtDupl},
				Incremental:       true,
				CacheDir:          ".cache/art-dupl",
				ClearCache:        true,
			},
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AssertConfigValidation(t, tt.config, tt.isValid)
		})
	}
}

// AssertConfigValidation validates a config and asserts the expected result.
func AssertConfigValidation(t *testing.T, cfg *Config, expectValid bool) {
	t.Helper()

	err := ValidateConfig(cfg)
	if expectValid && err != nil {
		t.Errorf("Expected valid config, got error: %v", err)
	}

	if !expectValid && err == nil {
		t.Error("Expected invalid config, but got no error")
	}
}

func TestMergeConfigs(t *testing.T) {
	fileConfig := &Config{
		Threshold:    50,
		OutputFormat: "json",
		Verbose:      true,
		Paths:        []string{"./src"},
		IgnoreFiles:  []string{"*_test.go"},
	}

	cliConfig := &Config{
		Threshold:     25,     // Override
		IncludeVendor: true,   // New value
		OutputFormat:  "html", // Override
		// Verbose: not set, should keep file config value
		Paths: []string{"./cmd"}, // Override
		// IgnoreFiles: not set, should keep file config value
	}

	combined := MergeConfigs(fileConfig, cliConfig)

	testutil.AssertConfigField(t, "Threshold", combined.Threshold, 25)
	testutil.AssertConfigField(t, "IncludeVendor", combined.IncludeVendor, true)
	testutil.AssertConfigField(t, "OutputFormat", combined.OutputFormat, "html")

	if !combined.Verbose {
		t.Error("Expected combined Verbose true")
	}

	if len(combined.Paths) != 1 || combined.Paths[0] != "./cmd" {
		t.Errorf("Expected combined paths [\"./cmd\"], got %v", combined.Paths)
	}

	if len(combined.IgnoreFiles) != 1 || combined.IgnoreFiles[0] != "*_test.go" {
		t.Errorf("Expected combined ignoreFiles [\"*_test.go\"], got %v", combined.IgnoreFiles)
	}
}

func TestMergeConfigsWithNil(t *testing.T) {
	testCases := []struct {
		name            string
		config          *Config
		isNilFileConfig bool
		expectedFields  map[string]any
	}{
		{
			name: "NilFileConfig",
			config: &Config{
				Threshold:     30,
				OutputFormat:  "json",
				IncludeVendor: true,
			},
			isNilFileConfig: true,
			expectedFields: map[string]any{
				"threshold":     30,
				"outputFormat":  "json",
				"includeVendor": true,
				"verbose":       false,
			},
		},
		{
			name: "NilCLIConfig",
			config: &Config{
				Threshold:    40,
				OutputFormat: "html",
				Verbose:      true,
			},
			isNilFileConfig: false,
			expectedFields: map[string]any{
				"threshold":     40,
				"outputFormat":  "html",
				"verbose":       true,
				"includeVendor": false,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var merged *Config

			if tc.isNilFileConfig {
				merged = MergeConfigs(nil, tc.config)
			} else {
				merged = MergeConfigs(tc.config, nil)
			}

			if threshold, ok := tc.expectedFields["threshold"].(int); ok {
				if merged.Threshold != threshold {
					t.Errorf("Expected threshold %d, got %d", threshold, merged.Threshold)
				}
			}

			if outputFormat, ok := tc.expectedFields["outputFormat"].(string); ok {
				if merged.OutputFormat.String() != outputFormat {
					t.Errorf(
						"Expected OutputFormat %s, got %s",
						outputFormat,
						merged.OutputFormat.String(),
					)
				}
			}

			if verbose, ok := tc.expectedFields["verbose"].(bool); ok {
				if merged.Verbose != verbose {
					t.Errorf("Expected Verbose %v, got %v", verbose, merged.Verbose)
				}
			}

			if includeVendor, ok := tc.expectedFields["includeVendor"].(bool); ok {
				if merged.IncludeVendor != includeVendor {
					t.Errorf(
						"Expected IncludeVendor %v, got %v",
						includeVendor,
						merged.IncludeVendor,
					)
				}
			}
		})
	}
}

func TestDetectionMethods(t *testing.T) {
	t.Parallel()
	// Test String method
	dm := DetectionMethodHash
	if dm.String() != "hash" {
		t.Errorf("Expected 'hash', got %s", dm.String())
	}

	// Test IsValid
	if !DetectionMethodHash.IsValid() {
		t.Error("Expected hash to be valid")
	}

	if DetectionMethod("invalid").IsValid() {
		t.Error("Expected invalid method to be invalid")
	}

	// Test AllDetectionMethods
	methods := AllDetectionMethods()
	if len(methods) != 4 {
		t.Errorf("Expected 4 methods, got %d", len(methods))
	}

	// Test ParseDetectionMethods
	parsed, err := ParseDetectionMethods("hash,art-dupl")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(parsed) != 2 {
		t.Errorf("Expected 2 methods, got %d", len(parsed))
	}

	// Test empty string defaults
	parsed, err = ParseDetectionMethods("")
	if err != nil {
		t.Errorf("Expected no error for empty string, got %v", err)
	}

	if len(parsed) != 1 || parsed[0] != DetectionMethodArtDupl {
		t.Error("Expected default art-dupl method for empty string")
	}

	// Test Contains
	dms := DetectionMethods{DetectionMethodHash, DetectionMethodArtDupl}
	if !dms.Contains(DetectionMethodHash) {
		t.Error("Expected Contains to find hash method")
	}

	if dms.Contains(DetectionMethod("invalid")) {
		t.Error("Expected Contains to not find invalid method")
	}

	// Test IsDefault
	defaultDms := DetectionMethods{DetectionMethodArtDupl}
	if !defaultDms.IsDefault() {
		t.Error("Expected IsDefault to return true for art-dupl only")
	}

	mixedDms := DetectionMethods{DetectionMethodHash, DetectionMethodArtDupl}
	if mixedDms.IsDefault() {
		t.Error("Expected IsDefault to return false for mixed methods")
	}
}

func TestOutputFormats(t *testing.T) {
	t.Parallel()
	// Test AllOutputFormats
	formats := AllOutputFormats()
	testutil.AssertLen(t, formats, 7, "output formats")
	// Test AllSortCriteria
	criteria := AllSortCriteria()
	testutil.AssertLen(t, criteria, 4, "sort criteria")
}

func TestJSONMarshalUnmarshal(t *testing.T) {
	t.Parallel()
	// Test DetectionMethod JSON marshal/unmarshal
	dm := DetectionMethodHash

	data, err := dm.MarshalJSON()
	if err != nil {
		t.Errorf("Expected no error marshaling, got %v", err)
	}

	var parsedDM DetectionMethod

	err = json.Unmarshal(data, &parsedDM)
	if err != nil {
		t.Errorf("Expected no error unmarshaling, got %v", err)
	}

	if parsedDM != dm {
		t.Error("Expected parsed detection method to match original")
	}

	// Test OutputFormat JSON marshal/unmarshal
	of := OutputFormatJSON

	data, err = of.MarshalJSON()
	if err != nil {
		t.Errorf("Expected no error marshaling output format, got %v", err)
	}

	var parsedOF OutputFormat

	err = json.Unmarshal(data, &parsedOF)
	if err != nil {
		t.Errorf("Expected no error unmarshaling output format, got %v", err)
	}

	if parsedOF != of {
		t.Error("Expected parsed output format to match original")
	}
}

func TestSemanticField(t *testing.T) {
	t.Parallel()

	// Test default value is false
	cfg := DefaultConfig()
	testutil.AssertFieldValue(t, cfg.Semantic, false, "Semantic")

	// Test Semantic can be loaded from config file
	t.Run("LoadFromConfigFile", func(t *testing.T) {
		t.Parallel()

		tmpDir, cleanup := createTempDir(t)
		defer cleanup()

		configFile := filepath.Join(tmpDir, "semantic-config.json")
		configContent := `{
			"threshold": 15,
			"semantic": true
		}`

		err := os.WriteFile(configFile, []byte(configContent), 0o644)
		if err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		loaded, err := LoadConfig(configFile)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		testutil.AssertFieldValue(t, loaded.Semantic, true, "Semantic")
	})

	// Test Semantic is preserved in save/load round trip
	t.Run("SaveLoadRoundTrip", func(t *testing.T) {
		t.Parallel()

		tmpDir, cleanup := createTempDir(t)
		defer cleanup()

		configFile := filepath.Join(tmpDir, "roundtrip-config.json")
		cfg := &Config{
			Threshold:         20,
			OutputFormat:      "json",
			Semantic:          true,
			DetectionMethods:  DetectionMethods{DetectionMethodArtDupl},
			MaxChildrenSerial: 10000,
		}

		err := SaveConfig(cfg, configFile)
		if err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		loaded, err := LoadConfig(configFile)
		if err != nil {
			t.Fatalf("Failed to load saved config: %v", err)
		}

		if loaded.Semantic != cfg.Semantic {
			t.Errorf("Expected Semantic %v, got %v", cfg.Semantic, loaded.Semantic)
		}
	})

	// Test Semantic is merged correctly (CLI zero values don't override file values for booleans)
	t.Run("MergeConfigs", func(t *testing.T) {
		fileConfig := &Config{
			Threshold:         15,
			Semantic:          true,
			MaxChildrenSerial: 10000,
			DetectionMethods:  DetectionMethods{DetectionMethodArtDupl},
		}

		cliConfig := &Config{
			Threshold: 20,
			// Semantic not set (zero value = false)
			// CLI skips zero values for booleans, so file value should be preserved
		}

		merged := MergeConfigs(fileConfig, cliConfig)

		// File value should be preserved since CLI has zero value
		if merged.Semantic != true {
			t.Errorf(
				"Expected merged Semantic true (file value preserved), got %v",
				merged.Semantic,
			)
		}
	})
}
