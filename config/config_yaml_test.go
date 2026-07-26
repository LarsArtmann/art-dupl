package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_YAML(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	content := `
threshold: 50
includeVendor: true
outputFormat: json
verbose: true
paths:
  - ./src
  - ./lib
ignoreFiles:
  - "*_test.go"
maxChildrenSerial: 20000
`

	err := os.WriteFile(configFile, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("Failed to write YAML config: %v", err)
	}

	cfg, err := LoadConfig(configFile)
	if err != nil {
		t.Fatalf("LoadConfig failed for YAML: %v", err)
	}

	if cfg.Threshold != 50 {
		t.Errorf("Threshold = %d, want 50", cfg.Threshold)
	}

	if !cfg.IncludeVendor {
		t.Errorf("IncludeVendor = false, want true")
	}

	if cfg.OutputFormat != "json" {
		t.Errorf("OutputFormat = %q, want %q", cfg.OutputFormat, "json")
	}

	if !cfg.Verbose {
		t.Errorf("Verbose = false, want true")
	}

	if cfg.MaxChildrenSerial != 20000 {
		t.Errorf("MaxChildrenSerial = %d, want 20000", cfg.MaxChildrenSerial)
	}

	if len(cfg.Paths) != 2 || cfg.Paths[0] != "./src" || cfg.Paths[1] != "./lib" {
		t.Errorf("Paths = %v, want [./src, ./lib]", cfg.Paths)
	}

	if len(cfg.IgnoreFiles) != 1 || cfg.IgnoreFiles[0] != "*_test.go" {
		t.Errorf("IgnoreFiles = %v, want [*_test.go]", cfg.IgnoreFiles)
	}
}

func TestLoadConfig_YAMLExtension(t *testing.T) {
	t.Parallel()

	for _, ext := range []string{".yaml", ".yml"} {
		t.Run(ext, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config"+ext)

			content := "threshold: 10\n"
			if err := os.WriteFile(configFile, []byte(content), 0o644); err != nil {
				t.Fatalf("Failed to write: %v", err)
			}

			cfg, err := LoadConfig(configFile)
			if err != nil {
				t.Fatalf("LoadConfig failed: %v", err)
			}

			if cfg.Threshold != 10 {
				t.Errorf("Threshold = %d, want 10", cfg.Threshold)
			}

			if !isYAMLFile(configFile) {
				t.Errorf("isYAMLFile(%q) = false, want true", configFile)
			}
		})
	}
}

func TestLoadConfig_MalformedYAML(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "bad.yaml")

	// Invalid YAML: unclosed quote and bad indentation
	content := `threshold: "unterminated
  bad: [}`

	err := os.WriteFile(configFile, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("Failed to write malformed YAML: %v", err)
	}

	_, err = LoadConfig(configFile)
	if err == nil {
		t.Error("LoadConfig should return error for malformed YAML, got nil")
	}
}

func TestLoadConfig_YAMLAndJSON_Equivalent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	yamlFile := filepath.Join(tmpDir, "c.yaml")
	jsonFile := filepath.Join(tmpDir, "c.json")

	yamlContent := "threshold: 25\noutputFormat: text\n"
	jsonContent := `{"threshold": 25, "outputFormat": "text"}`

	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(jsonFile, []byte(jsonContent), 0o644); err != nil {
		t.Fatal(err)
	}

	yamlCfg, err := LoadConfig(yamlFile)
	if err != nil {
		t.Fatalf("YAML load failed: %v", err)
	}

	jsonCfg, err := LoadConfig(jsonFile)
	if err != nil {
		t.Fatalf("JSON load failed: %v", err)
	}

	if yamlCfg.Threshold != jsonCfg.Threshold {
		t.Errorf("Threshold mismatch: YAML=%d, JSON=%d", yamlCfg.Threshold, jsonCfg.Threshold)
	}

	if yamlCfg.OutputFormat != jsonCfg.OutputFormat {
		t.Errorf("OutputFormat mismatch: YAML=%q, JSON=%q", yamlCfg.OutputFormat, jsonCfg.OutputFormat)
	}
}
