package config

// MergeConfigs merges two configurations, with command line config taking precedence.
func MergeConfigs(fileConfig, cliConfig *Config) *Config {
	result := DefaultConfig()

	mergeFileConfig(result, fileConfig)
	mergeCLIConfig(result, cliConfig)

	return result
}

// mergeConfig merges source config into result config.
// If skipZeroValues is true, fields with zero/empty values are skipped.
// This provides a single source of truth for config merging.
//
//nolint:funlen,gocognit,gocyclo,cyclop // Config merging requires handling each field independently
func mergeConfig(result, cfg *Config, skipZeroValues bool) {
	if cfg == nil {
		return
	}

	// Threshold (int)
	if !skipZeroValues || cfg.Threshold != 0 {
		result.Threshold = cfg.Threshold
	}

	// IncludeVendor (bool)
	if !skipZeroValues || cfg.IncludeVendor {
		result.IncludeVendor = cfg.IncludeVendor
	}

	// FilesFromStdin (bool)
	if !skipZeroValues || cfg.FilesFromStdin {
		result.FilesFromStdin = cfg.FilesFromStdin
	}

	// OutputFormat (OutputFormat)
	if !skipZeroValues || cfg.OutputFormat != "" {
		result.OutputFormat = cfg.OutputFormat
	}

	// Verbose (bool)
	if !skipZeroValues || cfg.Verbose {
		result.Verbose = cfg.Verbose
	}

	// Paths ([]string)
	if !skipZeroValues || len(cfg.Paths) > 0 {
		result.Paths = cfg.Paths
	}

	// IgnoreFiles ([]string)
	if !skipZeroValues || len(cfg.IgnoreFiles) > 0 {
		result.IgnoreFiles = cfg.IgnoreFiles
	}

	// MaxChildrenSerial (int)
	if !skipZeroValues || cfg.MaxChildrenSerial != 0 {
		result.MaxChildrenSerial = cfg.MaxChildrenSerial
	}

	// OutputFile (string)
	if !skipZeroValues || cfg.OutputFile != "" {
		result.OutputFile = cfg.OutputFile
	}

	// SortBy (SortCriteria)
	if !skipZeroValues || cfg.SortBy != "" {
		result.SortBy = cfg.SortBy
	}

	// DetectionMethods (DetectionMethods)
	if !skipZeroValues || len(cfg.DetectionMethods) > 0 {
		result.DetectionMethods = cfg.DetectionMethods
	}

	// Profile (bool)
	if !skipZeroValues || cfg.Profile {
		result.Profile = cfg.Profile
	}

	// Timeout (int)
	if !skipZeroValues || cfg.Timeout != 0 {
		result.Timeout = cfg.Timeout
	}

	// FilterGenerated (bool)
	if !skipZeroValues || cfg.FilterGenerated {
		result.FilterGenerated = cfg.FilterGenerated
	}

	// IncludeSQLC (bool)
	if !skipZeroValues || cfg.IncludeSQLC {
		result.IncludeSQLC = cfg.IncludeSQLC
	}

	// IncludeTempl (bool)
	if !skipZeroValues || cfg.IncludeTempl {
		result.IncludeTempl = cfg.IncludeTempl
	}

	// IncludePatterns ([]string)
	if !skipZeroValues || len(cfg.IncludePatterns) > 0 {
		result.IncludePatterns = cfg.IncludePatterns
	}

	// ExcludePatterns ([]string)
	if !skipZeroValues || len(cfg.ExcludePatterns) > 0 {
		result.ExcludePatterns = cfg.ExcludePatterns
	}

	// Incremental (bool)
	if !skipZeroValues || cfg.Incremental {
		result.Incremental = cfg.Incremental
	}

	// Since (string)
	if !skipZeroValues || cfg.Since != "" {
		result.Since = cfg.Since
	}

	// CacheDir (string)
	if !skipZeroValues || cfg.CacheDir != "" {
		result.CacheDir = cfg.CacheDir
	}

	// ClearCache (bool)
	if !skipZeroValues || cfg.ClearCache {
		result.ClearCache = cfg.ClearCache
	}

	// Semantic (bool)
	if !skipZeroValues || cfg.Semantic {
		result.Semantic = cfg.Semantic
	}
}

func mergeFileConfig(result, cfg *Config) {
	mergeConfig(result, cfg, false)
}

func mergeCLIConfig(result, cfg *Config) {
	mergeConfig(result, cfg, true)
}
