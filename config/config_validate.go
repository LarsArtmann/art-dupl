package config

import (
	"fmt"
	"strings"

	"github.com/LarsArtmann/art-dupl/errors"
)

// Validate validates the configuration invariants in one place.
// It is the single entry point for all Config validation.
func (c *Config) Validate() error {
	return ValidateConfig(c)
}

// ValidateConfig validates the configuration.
func ValidateConfig(cfg *Config) error {
	validations := []func() error{
		func() error { return validateThreshold(cfg.Threshold) },
		func() error { return validateMaxChildrenSerial(cfg.MaxChildrenSerial) },
		func() error { return validateOutputFormat(cfg.OutputFormat) },
		func() error { return validateSortCriteria(cfg.SortBy) },
		func() error { return validateDiffMode(cfg.DiffMode) },
		func() error { return validateTestThreshold(cfg.TestThreshold) },
		func() error { return validateDetectionMethods(cfg.DetectionMethods) },
		func() error { return validateCacheFlags(cfg.CacheDir, cfg.ClearCache, cfg.Incremental) },
		func() error { return validateOnly(cfg.Only) },
		func() error { return validateNonNegative("workers", cfg.Workers) },
		func() error { return validateNonNegative("min-lines", cfg.MinLines) },
		func() error { return validateNonNegative("max-cache-entries", cfg.MaxCacheEntries) },
	}

	for _, validate := range validations {
		err := validate()
		if err != nil {
			return err
		}
	}

	return nil
}

func validateThreshold(threshold int) error {
	if threshold < 1 {
		return fmt.Errorf(
			"%w: %d (use --threshold/-t with a value of 1 or higher; default is 5)",
			ErrInvalidThreshold,
			threshold,
		)
	}

	if threshold > 1000 {
		return fmt.Errorf(
			"%w: %d (use a value between 1 and 1000; use --min-lines for line-based filtering instead)",
			ErrThresholdTooLarge,
			threshold,
		)
	}

	return nil
}

func validateMaxChildrenSerial(maxChildren int) error {
	if maxChildren < 1000 {
		return errors.NewValidationError("maxChildrenSerial should be at least 1000", nil)
	}

	if maxChildren > 100000 {
		return errors.NewValidationError("maxChildrenSerial seems too large (max 100000)", nil)
	}

	return nil
}

func validateOutputFormat(format OutputFormat) error {
	return validateEnumValue(format, "output format", "text, html, json, plumbing")
}

func validateSortCriteria(criteria SortCriteria) error {
	return validateEnumValue(criteria, "sort criteria", "size, occurrence, hash, total-tokens; use --sort/-s")
}

func validateDiffMode(mode DiffMode) error {
	return validateEnumValue(mode, "diff mode", "side-by-side, inline; use --diff")
}

func validateTestThreshold(testThreshold int) error {
	if testThreshold < 0 {
		return errors.NewValidationError(
			fmt.Sprintf("test threshold must be non-negative: %d", testThreshold),
			nil,
		)
	}

	return nil
}

func validateDetectionMethods(methods []DetectionMethod) error {
	if len(methods) == 0 {
		return errors.NewValidationError("at least one detection method must be specified", nil)
	}

	for _, method := range methods {
		if !method.IsValid() {
			return errors.NewValidationError(
				fmt.Sprintf(
					"invalid detection method: %s (valid: %s)",
					method,
					strings.Join(detectionMethodStrings(AllDetectionMethods()), ", "),
				),
				nil,
			)
		}
	}

	return nil
}

func validateCacheFlags(cacheDir string, clearCache, incremental bool) error {
	if (cacheDir != "" || clearCache) && !incremental {
		return errors.NewValidationError(
			"--cache-dir and --clear-cache require --incremental mode",
			nil,
		)
	}

	return nil
}

func validateOnly(only FileType) error {
	if !only.IsValid() {
		return errors.NewValidationError(
			fmt.Sprintf("invalid --only value: %q (valid: go, templ)", only),
			nil,
		)
	}

	return nil
}

// validateEnumValue checks an enum-typed value via IsValid and returns a
// uniform ValidationError on failure. Used for OutputFormat, SortCriteria,
// DiffMode, and any future single-value enum validators.
func validateEnumValue[T interface {
	IsValid() bool
	fmt.Stringer
}](value T, label, validValues string) error {
	if value.IsValid() {
		return nil
	}

	return errors.NewValidationError(
		fmt.Sprintf("invalid %s: %s (valid: %s)", label, value, validValues),
		nil,
	)
}

func validateNonNegative(name string, value int) error {
	if value < 0 {
		return errors.NewValidationError(
			fmt.Sprintf("%s must be non-negative: %d", name, value),
			nil,
		)
	}

	return nil
}
