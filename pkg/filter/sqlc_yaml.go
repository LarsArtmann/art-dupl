package filter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LarsArtmann/art-dupl/pkg/logger"
	"gopkg.in/yaml.v3"
)

// SQLCConfig represents a sqlc.yaml configuration file structure.
type SQLCConfig struct {
	Version string       `yaml:"version"`
	SQL     []SQLCEngine `yaml:"sql"`
}

// SQLCEngine represents a single SQL engine configuration in sqlc.yaml.
type SQLCEngine struct {
	Schema string        `yaml:"schema"`
	Engine string        `yaml:"engine"`
	Gen    SQLCGenConfig `yaml:"gen"`
}

// SQLCGenConfig represents the generation configuration in sqlc.yaml.
type SQLCGenConfig struct {
	Go SQLCGoConfig `yaml:"go"`
}

// SQLCGoConfig represents the Go-specific generation configuration.
type SQLCGoConfig struct {
	Package string `yaml:"package"`
	Out     string `yaml:"out"`
}

// FindSQLCConfigs searches for sqlc.yaml or sqlc.yml files in the given paths.
func FindSQLCConfigs(paths []string) (map[string]string, error) {
	configs := make(map[string]string) // Map of config path to project root

	for _, path := range paths {
		err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Stop at depth to avoid walking too deep
			if info.IsDir() {
				// Skip hidden directories and common non-source directories
				name := info.Name()
				if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" {
					return filepath.SkipDir
				}
				return nil
			}

			// Check for sqlc.yaml or sqlc.yml
			filename := filepath.Base(filePath)
			if filename == "sqlc.yaml" || filename == "sqlc.yml" {
				configs[filePath] = filepath.Dir(filePath)
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error walking path %s: %w", path, err)
		}
	}

	return configs, nil
}

// ParseSQLCConfig reads and parses a sqlc.yaml file.
func ParseSQLCConfig(configPath string) (*SQLCConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading sqlc config %s: %w", configPath, err)
	}

	var config SQLCConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing sqlc config %s: %w", configPath, err)
	}

	return &config, nil
}

// GetSQLOutputDirs returns a list of output directories from sqlc configuration files.
func GetSQLOutputDirs(paths []string) ([]string, error) {
	configPaths, err := FindSQLCConfigs(paths)
	if err != nil {
		return nil, err
	}

	var outputDirs []string
	for configPath, projectRoot := range configPaths {
		config, err := ParseSQLCConfig(configPath)
		if err != nil {
			// Log but continue - one bad config shouldn't stop everything
			logger.Default.Warn("failed to parse sqlc config", "file", configPath, "err", err)
			continue
		}

		// Extract output directories from all SQL engines
		for _, sqlEngine := range config.SQL {
			if sqlEngine.Gen.Go.Out != "" {
				// Make output path absolute relative to project root
				outDir := filepath.Join(projectRoot, sqlEngine.Gen.Go.Out)
				// Normalize path
				outDir = filepath.Clean(outDir)
				outputDirs = append(outputDirs, outDir)
			}
		}
	}

	return outputDirs, nil
}
