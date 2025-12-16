package main

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
)

func TestConfigurationIntegration(t *testing.T) {
	// Test configuration file loading
	fileConfig := &config.Config{
		Threshold:         30,
		IncludeVendor:     true,
		OutputFormat:      "json",
		Verbose:           true,
		Paths:             []string{"./test"},
		IgnoreFiles:       []string{"*_test.go"},
		MaxChildrenSerial: 20000,
		OutputFile:        "output.json",
	}

	cliConfig := &config.Config{
		Threshold:     50,     // Should override file config
		OutputFormat:  "html", // Should override file config
		IncludeVendor: true,   // Should override file config
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
	if merged.Verbose != true {
		t.Errorf("Expected Verbose true (from file), got %v", merged.Verbose)
	}
	if len(merged.Paths) != 1 || merged.Paths[0] != "./test" {
		t.Errorf("Expected paths [\"./test\"], got %v", merged.Paths)
	}
	if len(merged.IgnoreFiles) != 1 || merged.IgnoreFiles[0] != "*_test.go" {
		t.Errorf("Expected ignoreFiles [\"*_test.go\"], got %v", merged.IgnoreFiles)
	}
	if merged.MaxChildrenSerial != 20000 {
		t.Errorf("Expected MaxChildrenSerial 20000, got %d", merged.MaxChildrenSerial)
	}
	if merged.OutputFile != "output.json" {
		t.Errorf("Expected OutputFile output.json, got %s", merged.OutputFile)
	}
}

func TestConfigurationValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		isValid bool
	}{
		{
			name: "Valid config",
			config: &config.Config{
				Threshold:         15,
				OutputFormat:      "text",
				MaxChildrenSerial: 10000,
				DetectionMethods:  config.DetectionMethods{config.DetectionMethodArtDupl},
			},
			isValid: true,
		},
		{
			name: "Invalid threshold",
			config: &config.Config{
				Threshold:         0,
				OutputFormat:      "text",
				MaxChildrenSerial: 10000,
				DetectionMethods:  config.DetectionMethods{config.DetectionMethodArtDupl},
			},
			isValid: false,
		},
		{
			name: "Invalid output format",
			config: &config.Config{
				Threshold:         15,
				OutputFormat:      "xml",
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
