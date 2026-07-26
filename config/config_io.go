package config

import (
	"encoding/json/v2"
	"os"
	"path/filepath"

	"github.com/LarsArtmann/art-dupl/errors"
	yaml "github.com/go-faster/yaml"
)

// LoadConfig loads configuration from file. Auto-detects format by extension:
// .yml/.yaml → YAML, everything else → JSON.
func LoadConfig(filename string) (*Config, error) {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return nil, errors.NewConfigError("config file not found: "+filename, nil)
	}

	// #nosec G304 -- filename is controlled config path, not user input
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.NewIOError(filename, "failed to read config file", err)
	}

	config := DefaultConfig()
	if len(data) > 0 {
		if isYAMLFile(filename) {
			if err := loadYAMLIntoConfig(data, config, filename); err != nil {
				return nil, err
			}
		} else {
			err = errors.SafeUnmarshal(data, config, "config file: "+filename)
			if err != nil {
				return nil, errors.NewConfigError("failed to parse config file: "+filename, err)
			}
		}
	}

	return config, nil
}

func isYAMLFile(filename string) bool {
	ext := filepath.Ext(filename)

	return ext == ".yaml" || ext == ".yml"
}

// loadYAMLIntoConfig bridges YAML through JSON to reuse all existing json tags
// and custom MarshalJSON/UnmarshalJSON hooks on Config and its enum fields.
func loadYAMLIntoConfig(data []byte, config *Config, filename string) error {
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return errors.NewConfigError("failed to parse YAML config file: "+filename, err)
	}

	jsonData, err := json.Marshal(raw)
	if err != nil {
		return errors.NewConfigError("failed to convert YAML to JSON: "+filename, err)
	}

	if err := errors.SafeUnmarshal(jsonData, config, "YAML config file: "+filename); err != nil {
		return errors.NewConfigError("failed to parse YAML config file: "+filename, err)
	}

	return nil
}

// LoadOptionalConfig loads configuration from file if filename is not empty.
// Returns nil if filename is empty, allowing optional config file usage.
//
//nolint:nilnil // Intentional: nil config + nil error means "no config file, which is valid"
func LoadOptionalConfig(filename string) (*Config, error) {
	if filename == "" {
		return nil, nil
	}

	return LoadConfig(filename)
}

// SaveConfig saves configuration to file.
func SaveConfig(config *Config, filename string) error {
	dir := filepath.Dir(filename)

	err := os.MkdirAll(dir, 0o750)
	if err != nil {
		return errors.NewIOError(dir, "failed to create config directory", err)
	}

	data, err := errors.SafeMarshalIndent(config, "", "  ", "config")
	if err != nil {
		return err //nolint:wrapcheck // Error already wrapped by SafeMarshalIndent
	}

	err = os.WriteFile(filename, data, 0o600)
	if err != nil {
		return errors.NewIOError(filename, "failed to write config file", err)
	}

	return nil
}
