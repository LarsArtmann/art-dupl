package cmd

import (
	"fmt"
	"slices"
	"strings"
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
	warnTypeAwareIncremental(cmd)
	warnSemanticDeprecation(cmd)

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
		mergedConfig.DetectionMode = config.DetectionModeStructural
	} else if cmd.Flags().Changed("exact") {
		mergedConfig.DetectionMode = config.DetectionModeExact
	} else if cmd.Flags().Changed("semantic") {
		mergedConfig.DetectionMode = config.DetectionModeSemantic
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

	var err error

	err = applyIncludeGeneratedFlag(cmd, cfg)
	if err != nil {
		return nil, err
	}

	if len(args) > 0 {
		cfg.Paths = args
	}

	verboseCount, _ := cmd.Flags().GetCount("verbose")
	if verboseCount > 0 {
		cfg.Verbose = true
	}

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

	if cmd.Flags().Changed("disable-pattern") {
		cfg.DisabledPatterns, _ = cmd.Flags().GetStringSlice("disable-pattern")
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
		"include-generic":      &cfg.IncludeGeneric,
		"incremental":          &cfg.Incremental,
		"rich-text":            &cfg.RichText,
		"clear-cache":          &cfg.ClearCache,
		"suppress-test-low":    &cfg.SuppressTestLow,
		"ignore-tests":         &cfg.IgnoreTests,
		"quiet":                &cfg.Quiet,
		"type-aware":           &cfg.TypeAware,
		"no-actionability":     &cfg.NoActionability,
		"explain":              &cfg.Explain,
		"no-accept-directives": &cfg.NoAcceptDirectives,
		"include-ignored":      &cfg.IncludeIgnored,
	}

	for flagName, configPtr := range mappings {
		if cmd.Flags().Changed(flagName) {
			val, _ := cmd.Flags().GetBool(flagName)
			*configPtr = val
		}
	}

	// --include-tests overrides all test-specific filtering
	if cmd.Flags().Changed("include-tests") {
		val, _ := cmd.Flags().GetBool("include-tests")
		if val {
			cfg.IgnoreTests = false
			cfg.SuppressTestLow = false
			cfg.IncludeTests = true
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

	if cmd.Flags().Changed("search-workers") {
		val, _ := cmd.Flags().GetInt("search-workers")
		cfg.SearchWorkers = val
	}

	if cmd.Flags().Changed("test-threshold") {
		val, _ := cmd.Flags().GetInt("test-threshold")
		cfg.TestThreshold = val
	}

	if cmd.Flags().Changed("min-lines") {
		val, _ := cmd.Flags().GetInt("min-lines")
		cfg.MinLines = val
	}

	if cmd.Flags().Changed("max-cache-entries") {
		val, _ := cmd.Flags().GetInt("max-cache-entries")
		cfg.MaxCacheEntries = val
	}
}

// applyChangedStringFlags sets string config fields only when explicitly provided.
func applyChangedStringFlags(cmd *cobra.Command, cfg *config.Config) {
	if cmd.Flags().Changed("only") {
		val, _ := cmd.Flags().GetString("only")
		cfg.Only = config.FileType(val)
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

// validateMutualExclusion checks that the detection-mode flags are not
// combined incompatibly. --semantic, --exact, and --structural select one mode.
// --type-aware is only effective with --semantic (default) and silently wasted
// with --structural or --exact, so those combinations are rejected up front.
func validateMutualExclusion(cmd *cobra.Command) error {
	semanticSet := cmd.Flags().Changed("semantic")
	structuralSet := cmd.Flags().Changed("structural")
	exactSet := cmd.Flags().Changed("exact")
	typeAwareSet := cmd.Flags().Changed("type-aware")

	if semanticSet && structuralSet {
		return duplerrors.NewValidationError(
			"cannot use both --semantic and --structural flags; these are mutually exclusive",
			nil,
		)
	}

	if exactSet && structuralSet {
		return duplerrors.NewValidationError(
			"cannot use both --exact and --structural flags; these are mutually exclusive",
			nil,
		)
	}

	if typeAwareSet && structuralSet {
		return duplerrors.NewValidationError(
			"--type-aware is only effective with --semantic (default); --structural ignores all "+
				"identifiers so type information would be silently discarded",
			nil,
		)
	}

	if typeAwareSet && exactSet {
		return duplerrors.NewValidationError(
			"--type-aware is only effective with --semantic (default); --exact hashes identifiers "+
				"verbatim without canonicalization, so type information would be silently discarded",
			nil,
		)
	}

	return nil
}

// warnStructural prints a warning when structural flag is used.
func warnStructural(cmd *cobra.Command) {
	if cmd.Flags().Changed("structural") {
		fmt.Fprintf(
			cmd.ErrOrStderr(),
			"Note: --structural flag disables semantic detection. This may increase false positives from similar-looking but semantically different code.\n",
		)
	}
}

// warnTypeAwareIncremental is now a no-op — type-aware mode IS compatible with
// incremental mode since M07 (IncrementalParser threads typeInfos via SetTypeAwareData).
func warnTypeAwareIncremental(_ *cobra.Command) {}

// warnSemanticDeprecation prints a gentle notice when --semantic is explicitly
// used. Semantic is the default mode, so the flag is redundant.
func warnSemanticDeprecation(cmd *cobra.Command) {
	if cmd.Flags().Changed("semantic") {
		fmt.Fprintf(
			cmd.ErrOrStderr(),
			"Note: --semantic is the default detection mode; this flag is redundant "+
				"and may be removed in a future version.\n",
		)
	}
}

// flagStringReader reads a string flag value only when the user explicitly set it.
func flagStringReader(cmd *cobra.Command, name string) (string, bool) {
	if !cmd.Flags().Changed(name) {
		return "", false
	}

	val, _ := cmd.Flags().GetString(name)

	return val, true
}

// applyDetectionMethods sets detection methods from the -m flag.
func applyDetectionMethods(cmd *cobra.Command, cfg *config.Config) error {
	val, ok := flagStringReader(cmd, "detection-methods")
	if !ok {
		return nil
	}

	return setDetectionMethods(cfg, val)
}

// applyTimeoutFlag parses and applies the --timeout flag.
func applyTimeoutFlag(cmd *cobra.Command, cfg *config.Config) error {
	val, ok := flagStringReader(cmd, "timeout")
	if !ok || val == "" || val == "30m" {
		return nil
	}

	duration, err := time.ParseDuration(val)
	if err != nil {
		return duplerrors.WrapValidation(
			err,
			fmt.Sprintf("invalid timeout format %q (use '30m', '1h', etc.)", val),
		)
	}

	cfg.Timeout = duration

	return nil
}

// applyDiffModeFlag parses and applies the --diff flag.
func applyDiffModeFlag(cmd *cobra.Command, cfg *config.Config) error {
	val, ok := flagStringReader(cmd, "diff")
	if !ok || val == "" {
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

const (
	generatedCategorySQLC     = "sqlc"
	generatedCategoryTempl    = "templ"
	generatedCategoryProtobuf = "protobuf"
	generatedCategoryMockgen  = "mockgen"
	generatedCategoryStringer = "stringer"
	generatedCategoryGeneric  = "generic"
	generatedCategoryAll      = "all"
	flagIncludeGenerated      = "include-generated"
)

// generatedCategories returns the accepted values for --include-generated.
func generatedCategories() []string {
	return []string{
		generatedCategorySQLC,
		generatedCategoryTempl,
		generatedCategoryProtobuf,
		generatedCategoryMockgen,
		generatedCategoryStringer,
		generatedCategoryGeneric,
		generatedCategoryAll,
	}
}

// applyIncludeGeneratedFlag parses the --include-generated categories and sets
// the matching per-generator config fields. It accepts repeated flags and
// comma-separated values.
func applyIncludeGeneratedFlag(cmd *cobra.Command, cfg *config.Config) error {
	if !cmd.Flags().Changed(flagIncludeGenerated) {
		return nil
	}

	values, _ := cmd.Flags().GetStringArray(flagIncludeGenerated)

	categories, err := parseIncludeGeneratedCategories(values)
	if err != nil {
		return duplerrors.WrapValidation(err, "invalid --include-generated value")
	}

	for _, c := range categories {
		switch c {
		case generatedCategorySQLC:
			cfg.IncludeSQLC = true
		case generatedCategoryTempl:
			cfg.IncludeTempl = true
		case generatedCategoryProtobuf:
			cfg.IncludeProtobuf = true
		case generatedCategoryMockgen:
			cfg.IncludeMockgen = true
		case generatedCategoryStringer:
			cfg.IncludeStringer = true
		case generatedCategoryGeneric:
			cfg.IncludeGeneric = true
		case generatedCategoryAll:
			cfg.IncludeSQLC = true
			cfg.IncludeTempl = true
			cfg.IncludeProtobuf = true
			cfg.IncludeMockgen = true
			cfg.IncludeStringer = true
			cfg.IncludeGeneric = true
		}
	}

	return nil
}

// parseIncludeGeneratedCategories validates and normalizes raw --include-generated
// values. Each value may be a single category or a comma-separated list.
func parseIncludeGeneratedCategories(values []string) ([]string, error) {
	valid := generatedCategories()
	result := make([]string, 0, len(values)*2)

	for _, v := range values {
		for part := range strings.SplitSeq(v, ",") {
			c := strings.ToLower(strings.TrimSpace(part))
			if c == "" {
				continue
			}

			if !slices.Contains(valid, c) {
				return nil, duplerrors.NewValidationError(
					fmt.Sprintf("unknown --include-generated category %q (valid: %s)", c, strings.Join(valid, ", ")),
					nil,
				)
			}

			result = append(result, c)
		}
	}

	return result, nil
}
