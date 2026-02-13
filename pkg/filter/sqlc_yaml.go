package filter

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/internal/utils"
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
// Searches both the provided paths and their parent directories (up to 3 levels up).
// Returns a map of config file path to project root directory.
func FindSQLCConfigs(paths []string) (map[string]string, error) {
	configs := make(map[string]string)

	for _, path := range paths {
		if err := findSQLCConfigsInPath(path, configs); err != nil {
			return nil, err
		}
	}

	return configs, nil
}

// findSQLCConfigsInPath searches for sqlc configs in a single path.
func findSQLCConfigsInPath(path string, configs map[string]string) error {
	if err := walkPathForSQLCConfigs(path, configs); err != nil {
		return errors.WrapFile(err, path, "walking path") //nolint:wrapcheck // Error already wrapped by WrapFile
	}

	findSQLCConfigsInParent(path, configs)
	return nil
}

// walkPathForSQLCConfigs walks a path to find sqlc config files.
func walkPathForSQLCConfigs(path string, configs map[string]string) error {
	return filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return handleDirectoryWalk(info.Name())
		}

		recordSQLCConfig(filePath, configs)
		return nil
	})
}

// handleDirectoryWalk determines whether to skip a directory during walk.
func handleDirectoryWalk(name string) error {
	if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" {
		return filepath.SkipDir
	}
	return nil
}

// recordSQLCConfig records a sqlc config file if it matches.
func recordSQLCConfig(filePath string, configs map[string]string) {
	filename := filepath.Base(filePath)
	if filename == "sqlc.yaml" || filename == "sqlc.yml" {
		configs[filePath] = filepath.Dir(filePath)
	}
}

// findSQLCConfigsInParent searches parent directories for sqlc config.
func findSQLCConfigsInParent(path string, configs map[string]string) {
	parentPath, err := utils.FindProjectRoot(path, []string{"sqlc.yaml", "sqlc.yml"})
	if err != nil || parentPath == "" {
		return
	}

	tryAddSQLCConfig(parentPath, "sqlc.yaml", configs)
	tryAddSQLCConfig(parentPath, "sqlc.yml", configs)
}

// tryAddSQLCConfig adds a config to the map if the file exists.
func tryAddSQLCConfig(parentPath, filename string, configs map[string]string) {
	configPath := filepath.Join(parentPath, filename)
	if _, err := os.Stat(configPath); err == nil {
		configs[configPath] = parentPath
	}
}

// ParseSQLCConfig reads and parses a sqlc.yaml file.
func ParseSQLCConfig(configPath string) (*SQLCConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, errors.WrapFile(err, configPath, "reading sqlc config") //nolint:wrapcheck // Error already wrapped by WrapFile
	}

	var config SQLCConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, errors.WrapConfig(err, "parsing sqlc config") //nolint:wrapcheck // Error already wrapped by WrapConfig
	}

	return &config, nil
}

// GetSQLOutputDirs returns a list of output directories from sqlc configuration files.
func GetSQLOutputDirs(paths []string) ([]string, error) {
	configPaths, err := FindSQLCConfigs(paths)
	if err != nil {
		return nil, err
	}

	// Warn if multiple config files found
	if len(configPaths) > 1 {
		logger.Default.Warn("multiple sqlc config files found", "count", len(configPaths))
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
