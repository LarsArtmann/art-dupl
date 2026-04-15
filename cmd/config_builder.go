package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/LarsArtmann/art-dupl/cli"
	"github.com/LarsArtmann/art-dupl/config"
	duplerrors "github.com/LarsArtmann/art-dupl/errors"
	"github.com/spf13/cobra"
)

// FlagValues holds the extracted flag values needed for config building.
type FlagValues struct {
	ConfigFile        string
	Vendor            bool
	Verbose           bool
	Threshold         int
	Files             bool
	Profile           bool
	FilterGenerated   bool
	IncludeSQLC       bool
	IncludeTempl      bool
	IncludePatterns   []string
	ExcludePatterns   []string
	Semantic          bool
	Structural        bool
	Paths             []string
	DetectionMethods  string
	Timeout           string
	IncludeNodeModules bool
	Only              string
	Incremental       bool
	Since             string
	CacheDir          string
	ClearCache        bool
	Workers           int
	DiffMode          string
}

// BuildConfigFromFlags extracts flag values and builds a merged configuration.
// This function handles the common config-building logic shared between run and stats commands.
func BuildConfigFromFlags(cmd *cobra.Command, args []string) (*config.Config, error) {
	flags, err := extractFlagValues(cmd, args)
	if err != nil {
		return nil, err
	}

	if err := validateMutualExclusion(flags.Semantic, flags.Structural); err != nil {
		return nil, err
	}

	warnStructural(flags.Structural)

	appConfig := &config.Config{}

	if err := applyFlagValues(appConfig, flags); err != nil {
		return nil, err
	}

	fileConfig, err := config.LoadOptionalConfig(flags.ConfigFile)
	if err != nil {
		return nil, duplerrors.WrapConfig(err, fmt.Sprintf("loading config from file %q", flags.ConfigFile))
	}

	mergedConfig := config.MergeConfigs(fileConfig, appConfig)

	if flags.Structural {
		mergedConfig.Semantic = false
	}

	if err := config.ValidateConfig(mergedConfig); err != nil {
		return nil, duplerrors.WrapValidation(
			err,
			fmt.Sprintf("configuration validation failed (paths: %v)", mergedConfig.Paths),
		)
	}

	return mergedConfig, nil
}

// extractFlagValues extracts all flag values from the command.
func extractFlagValues(cmd *cobra.Command, args []string) (*FlagValues, error) {
	configFile, _ := cmd.Flags().GetString("config")
	vendor, _ := cmd.Flags().GetBool("vendor")
	verboseCount, _ := cmd.Flags().GetCount("verbose")
	verbose := verboseCount > 0
	threshold, _ := cmd.Flags().GetInt("threshold")
	files, _ := cmd.Flags().GetBool("files")
	profile, _ := cmd.Flags().GetBool("profile")
	timeoutStr, _ := cmd.Flags().GetString("timeout")
	filterGenerated, _ := cmd.Flags().GetBool("filter-generated")
	includeSQLC, _ := cmd.Flags().GetBool("include-sqlc")
	includeTempl, _ := cmd.Flags().GetBool("include-templ")
	includePatterns, _ := cmd.Flags().GetStringArray("include-pattern")
	excludePatterns, _ := cmd.Flags().GetStringArray("exclude-pattern")
	semantic, _ := cmd.Flags().GetBool("semantic")
	structural, _ := cmd.Flags().GetBool("structural")
	detectionMethods, _ := cmd.Flags().GetString("detection-methods")
	includeNodeModules, _ := cmd.Flags().GetBool("include-node-modules")
	only, _ := cmd.Flags().GetString("only")
	incremental, _ := cmd.Flags().GetBool("incremental")
	since, _ := cmd.Flags().GetString("since")
	cacheDir, _ := cmd.Flags().GetString("cache-dir")
	clearCache, _ := cmd.Flags().GetBool("clear-cache")
	workers, _ := cmd.Flags().GetInt("workers")
	diffMode, _ := cmd.Flags().GetString("diff")

	return &FlagValues{
		ConfigFile:        configFile,
		Vendor:            vendor,
		Verbose:           verbose,
		Threshold:         threshold,
		Files:             files,
		Profile:           profile,
		Timeout:           timeoutStr,
		FilterGenerated:   filterGenerated,
		IncludeSQLC:       includeSQLC,
		IncludeTempl:      includeTempl,
		IncludePatterns:   includePatterns,
		ExcludePatterns:   excludePatterns,
		Semantic:          semantic,
		Structural:        structural,
		Paths:             args,
		DetectionMethods:  detectionMethods,
		IncludeNodeModules: includeNodeModules,
		Only:              only,
		Incremental:       incremental,
		Since:             since,
		CacheDir:          cacheDir,
		ClearCache:        clearCache,
		Workers:           workers,
		DiffMode:          diffMode,
	}, nil
}

// validateMutualExclusion checks that semantic and structural are not both set.
func validateMutualExclusion(semantic, structural bool) error {
	if semantic && structural {
		return duplerrors.NewValidationError(
			"cannot use both --semantic and --structural flags; these are mutually exclusive",
			nil,
		)
	}
	return nil
}

// warnStructural prints a warning when structural flag is used.
func warnStructural(structural bool) {
	if structural {
		fmt.Fprintf(
			os.Stderr,
			"Note: --structural flag disables semantic detection. This may increase false positives from similar-looking but semantically different code.\n",
		)
	}
}

// applyFlagValues applies the extracted flag values to the config.
func applyFlagValues(cfg *config.Config, flags *FlagValues) error {
	if flags.Verbose {
		cfg.Verbose = true
	}

	if err := setDetectionMethods(cfg, flags.DetectionMethods); err != nil {
		return err
	}

	if flags.Threshold != cli.DefaultThreshold {
		cfg.Threshold = flags.Threshold
	}

	if flags.Vendor {
		cfg.IncludeVendor = flags.Vendor
	}

	if flags.IncludeNodeModules {
		cfg.IncludeNodeModules = flags.IncludeNodeModules
	}

	if flags.Files {
		cfg.FilesFromStdin = flags.Files
	}

	if flags.Profile {
		cfg.Profile = true
	}

	if flags.Timeout != "" && flags.Timeout != "30m" {
		duration, err := time.ParseDuration(flags.Timeout)
		if err != nil {
			return duplerrors.WrapValidation(
				err,
				fmt.Sprintf("invalid timeout format %q (use '30m', '1h', etc.)", flags.Timeout),
			)
		}
		cfg.Timeout = int(duration.Seconds())
	}

	if flags.FilterGenerated {
		cfg.FilterGenerated = true
	}

	if flags.IncludeSQLC {
		cfg.IncludeSQLC = true
	}

	if flags.IncludeTempl {
		cfg.IncludeTempl = true
	}

	if flags.Only != "" {
		cfg.Only = flags.Only
	}

	if len(flags.IncludePatterns) > 0 {
		cfg.IncludePatterns = flags.IncludePatterns
	}

	if len(flags.ExcludePatterns) > 0 {
		cfg.ExcludePatterns = flags.ExcludePatterns
	}

	if flags.Incremental {
		cfg.Incremental = true
	}

	if flags.Since != "" {
		cfg.Since = flags.Since
	}

	if flags.CacheDir != "" {
		cfg.CacheDir = flags.CacheDir
	}

	if flags.ClearCache {
		cfg.ClearCache = true
	}

	if flags.Semantic {
		cfg.Semantic = true
	}

	if flags.Workers != 0 {
		cfg.Workers = flags.Workers
	}

	if flags.DiffMode != "" {
		parsedDiffMode, err := config.ParseDiffMode(flags.DiffMode)
		if err != nil {
			return duplerrors.WrapValidation(
				err,
				fmt.Sprintf("invalid --diff value %q", flags.DiffMode),
			)
		}
		cfg.DiffMode = parsedDiffMode
	}

	if len(flags.Paths) > 0 {
		cfg.Paths = flags.Paths
	}

	return nil
}
