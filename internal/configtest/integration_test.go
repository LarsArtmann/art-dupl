package configtest

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
)

func TestConfigurationIntegration(t *testing.T) {
	// Test configuration file loading
	fileConfig := &config.Config{
		Threshold:         30,
		IncludeVendor:     true,
		OutputFormat:      config.OutputFormatJSON,
		Verbose:           true,
		Paths:             []string{"./test"},
		IgnoreFiles:       []string{"*_test.go"},
		MaxChildrenSerial: 20000,
		OutputFile:        "output.json",
		DetectionMethods:  config.DetectionMethods{config.DetectionMethodHash},
	}

	cliConfig := &config.Config{
		Threshold:     50,                      // Should override file config
		OutputFormat:  config.OutputFormatHTML, // Should override file config
		IncludeVendor: true,                    // Should override file config
		DetectionMethods: config.DetectionMethods{
			config.DetectionMethodArtDupl,
		}, // Should override file config
	}

	merged := config.MergeConfigs(fileConfig, cliConfig)

	// Test CLI overrides
	if merged.Threshold != 50 {
		t.Errorf("Expected threshold 50 (CLI override), got %d", merged.Threshold)
	}

	if merged.OutputFormat != "html" {
		t.Errorf("Expected outputFormat html (CLI override), got %s", merged.OutputFormat)
	}

	if merged.IncludeVendor != true {
		t.Errorf("Expected IncludeVendor true (CLI override), got %v", merged.IncludeVendor)
	}

	// Test file config values preserved
	if !merged.Verbose {
		t.Error("Expected merged Verbose true")
	}

	if len(merged.Paths) != 1 || merged.Paths[0] != "./test" {
		t.Errorf("Expected merged paths [\"./test\"], got %v", merged.Paths)
	}

	if len(merged.IgnoreFiles) != 1 || merged.IgnoreFiles[0] != "*_test.go" {
		t.Errorf("Expected merged ignoreFiles [\"*_test.go\"], got %v", merged.IgnoreFiles)
	}

	if merged.MaxChildrenSerial != 20000 {
		t.Errorf("Expected MaxChildrenSerial 20000, got %d", merged.MaxChildrenSerial)
	}

	if merged.OutputFile != "output.json" {
		t.Errorf("Expected OutputFile output.json, got %s", merged.OutputFile)
	}

	if len(merged.DetectionMethods) != 1 ||
		merged.DetectionMethods[0] != config.DetectionMethodArtDupl {
		t.Errorf(
			"Expected DetectionMethods [art-dupl] (CLI override), got %v",
			merged.DetectionMethods,
		)
	}
}

// createConfigTestCase creates a test case for config validation.
func createConfigTestCase(name string, threshold int, isValid bool) struct {
	name    string
	config  *config.Config
	isValid bool
} {
	return struct {
		name    string
		config  *config.Config
		isValid bool
	}{
		name: name,
		config: &config.Config{
			Threshold:         threshold,
			OutputFormat:      config.OutputFormatText,
			MaxChildrenSerial: 10000,
			DetectionMethods:  config.DetectionMethods{config.DetectionMethodArtDupl},
		},
		isValid: isValid,
	}
}

func TestConfigurationValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		isValid bool
	}{
		createConfigTestCase("Valid config", 15, true),
		createConfigTestCase("Invalid threshold", 0, false),
		{
			name: "Invalid output format",
			config: &config.Config{
				Threshold:         15,
				OutputFormat:      config.OutputFormat("xml"),
				MaxChildrenSerial: 10000,
				DetectionMethods:  config.DetectionMethods{config.DetectionMethodArtDupl},
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.ValidateConfig(tt.config)
			if tt.isValid && err != nil {
				t.Errorf("Expected valid config, got error: %v", err)
			}

			if !tt.isValid && err == nil {
				t.Error("Expected invalid config, but got no error")
			}
		})
	}
}

func TestOutputFormatSelection(t *testing.T) {
	tests := []struct {
		outputFormat    string
		expectedPrinter string
	}{
		{"text", "text"},
		{"html", "html"},
		{"json", "json"},
		{"plumbing", "plumbing"},
		{"invalid", "text"}, // Should default to text
	}

	for _, tt := range tests {
		t.Run(tt.outputFormat, func(t *testing.T) {
			cfg := &config.Config{
				OutputFormat:      config.OutputFormat(tt.outputFormat),
				Threshold:         15,
				MaxChildrenSerial: 10000,
				DetectionMethods:  config.DetectionMethods{config.DetectionMethodArtDupl},
			}

			// This would be tested in the main CLI logic
			// For now, just validate the format is recognized
			err := config.ValidateConfig(cfg)
			if tt.outputFormat != "invalid" && err != nil {
				t.Errorf("Expected valid format %s, got error: %v", tt.outputFormat, err)
			}

			if tt.outputFormat == "invalid" && err == nil {
				t.Error("Expected invalid format to cause error")
			}
		})
	}
}
