package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/LarsArtmann/art-dupl/config"
	duplerrors "github.com/LarsArtmann/art-dupl/errors"
	"github.com/spf13/cobra"
)

// BuildConfigFromFlags extracts flag values and builds a merged configuration.
// This function handles the common config-building logic shared between run and stats commands.
func BuildConfigFromFlags(cmd *cobra.Command, args []string) (*config.Config, error) {
	err := validateMutualExclusion(cmd)
	if err != nil {
		return nil, err
	}

	warnStructural(cmd)

	appConfig, err := buildCLIConfig(cmd, args)
	if err != nil {
		return nil, err
	}

	fileConfig, err := config.LoadOptionalConfig(configFile(cmd))
	if err != nil {
		return nil, duplerrors.WrapConfig(
			err,
			fmt.Sprintf("loading config from file %q", configFile(cmd)),
		)
	}

	mergedConfig := config.MergeConfigs(fileConfig, appConfig)

	if cmd.Flags().Changed("structural") {
		mergedConfig.Semantic = false
	}

	err = config.ValidateConfig(mergedConfig)
	if err != nil {
		return nil, duplerrors.WrapValidation(
			err,
			fmt.Sprintf("configuration validation failed (paths: %v)", mergedConfig.Paths),
		)
	}

	return mergedConfig, nil
}

// buildCLIConfig constructs a Config from CLI flags.
// Only sets fields the user explicitly provided (via Changed() tracking),
// so zero values don't accidentally override file config.
func buildCLIConfig(cmd *cobra.Command, args []string) (*config.Config, error) {
	cfg := &config.Config{}

	applyChangedBoolFlags(cmd, cfg)
	applyChangedIntFlags(cmd, cfg)
	applyChangedStringFlags(cmd, cfg)
	applyChangedStringArrayFlags(cmd, cfg)

	if len(args) > 0 {
		cfg.Paths = args
	}

	verboseCount, _ := cmd.Flags().GetCount("verbose")
	if verboseCount > 0 {
		cfg.Verbose = true
	}

	var err error

	err = applyDetectionMethods(cmd, cfg)
	if err != nil {
		return nil, err
	}

	err = applyTimeoutFlag(cmd, cfg)
	if err != nil {
		return nil, err
	}

	err = applyDiffModeFlag(cmd, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// applyChangedBoolFlags sets bool config fields only when the user explicitly
// provided the flag (via Changed()). This fixes the "false can't override true" bug.
func applyChangedBoolFlags(cmd *cobra.Command, cfg *config.Config) {
	mappings := map[string]*bool{
		"vendor":               &cfg.IncludeVendor,
		"include-node-modules": &cfg.IncludeNodeModules,
		"files":                &cfg.FilesFromStdin,
		"profile":              &cfg.Profile,
		"include-sqlc":         &cfg.IncludeSQLC,
		"include-templ":        &cfg.IncludeTempl,
		"include-protobuf":     &cfg.IncludeProtobuf,
		"include-mockgen":      &cfg.IncludeMockgen,
		"include-stringer":     &cfg.IncludeStringer,
		"incremental":          &cfg.Incremental,
		"semantic":             &cfg.Semantic,
		"rich-text":            &cfg.RichText,
		"clear-cache":          &cfg.ClearCache,
	}

	for flagName, configPtr := range mappings {
		if cmd.Flags().Changed(flagName) {
			val, _ := cmd.Flags().GetBool(flagName)
			*configPtr = val
		}
	}
}

// applyChangedIntFlags sets int config fields only when explicitly provided.
func applyChangedIntFlags(cmd *cobra.Command, cfg *config.Config) {
	if cmd.Flags().Changed("threshold") {
		val, _ := cmd.Flags().GetInt("threshold")
		cfg.Threshold = val
	}

	if cmd.Flags().Changed("workers") {
		val, _ := cmd.Flags().GetInt("workers")
		cfg.Workers = val
	}
}

// applyChangedStringFlags sets string config fields only when explicitly provided.
func applyChangedStringFlags(cmd *cobra.Command, cfg *config.Config) {
	if cmd.Flags().Changed("only") {
		val, _ := cmd.Flags().GetString("only")
		cfg.Only = config.FileType(val)
	}

	if cmd.Flags().Changed("since") {
		val, _ := cmd.Flags().GetString("since")
		cfg.Since = val
	}

	if cmd.Flags().Changed("cache-dir") {
		val, _ := cmd.Flags().GetString("cache-dir")
		cfg.CacheDir = val
	}
}

// applyChangedStringArrayFlags sets string slice config fields only when explicitly provided.
func applyChangedStringArrayFlags(cmd *cobra.Command, cfg *config.Config) {
	if cmd.Flags().Changed("include-pattern") {
		val, _ := cmd.Flags().GetStringArray("include-pattern")
		cfg.IncludePatterns = val
	}

	if cmd.Flags().Changed("exclude-pattern") {
		val, _ := cmd.Flags().GetStringArray("exclude-pattern")
		cfg.ExcludePatterns = val
	}
}

// configFile returns the config file path from the --config flag.
func configFile(cmd *cobra.Command) string {
	val, _ := cmd.Flags().GetString("config")

	return val
}

// validateMutualExclusion checks that semantic and structural are not both explicitly set.
func validateMutualExclusion(cmd *cobra.Command) error {
	semanticSet := cmd.Flags().Changed("semantic")
	structuralSet := cmd.Flags().Changed("structural")

	if semanticSet && structuralSet {
		return duplerrors.NewValidationError(
			"cannot use both --semantic and --structural flags; these are mutually exclusive",
			nil,
		)
	}

	return nil
}

// warnStructural prints a warning when structural flag is used.
func warnStructural(cmd *cobra.Command) {
	if cmd.Flags().Changed("structural") {
		fmt.Fprintf(
			os.Stderr,
			"Note: --structural flag disables semantic detection. This may increase false positives from similar-looking but semantically different code.\n",
		)
	}
}

// applyDetectionMethods sets detection methods from the -m flag.
func applyDetectionMethods(cmd *cobra.Command, cfg *config.Config) error {
	if !cmd.Flags().Changed("detection-methods") {
		return nil
	}

	val, _ := cmd.Flags().GetString("detection-methods")

	return setDetectionMethods(cfg, val)
}

// applyTimeoutFlag parses and applies the --timeout flag.
func applyTimeoutFlag(cmd *cobra.Command, cfg *config.Config) error {
	if !cmd.Flags().Changed("timeout") {
		return nil
	}

	val, _ := cmd.Flags().GetString("timeout")
	if val == "" || val == "30m" {
		return nil
	}

	duration, err := time.ParseDuration(val)
	if err != nil {
		return duplerrors.WrapValidation(
			err,
			fmt.Sprintf("invalid timeout format %q (use '30m', '1h', etc.)", val),
		)
	}

	cfg.Timeout = int(duration.Seconds())

	return nil
}

// applyDiffModeFlag parses and applies the --diff flag.
func applyDiffModeFlag(cmd *cobra.Command, cfg *config.Config) error {
	if !cmd.Flags().Changed("diff") {
		return nil
	}

	val, _ := cmd.Flags().GetString("diff")
	if val == "" {
		return nil
	}

	parsed, err := config.ParseDiffMode(val)
	if err != nil {
		return duplerrors.WrapValidation(
			err,
			fmt.Sprintf("invalid --diff value %q", val),
		)
	}

	cfg.DiffMode = parsed

	return nil
}
